package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/catalog"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/external/kubent"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/health"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider/aks"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/report"
)

const (
	ExitReady        = 0
	ExitNotReady     = 2
	ExitInconclusive = 3
	ExitUsage        = 4
	ExitExecution    = 5
)

const (
	schemaVersion = "kua.assessment.v1"
	binaryName    = "kua"
)

type BuildInfo struct {
	Version        string
	Commit         string
	BuildDate      string
	CatalogVersion string
}

type Dependencies struct {
	PreflightRunner    PreflightRunner
	InventoryCollector InventoryCollector
	ProviderFactory    ProviderFactory
	APIAnalyzer        APIAnalyzer
	Clock              func() time.Time
	Stdin              io.Reader // set to os.Stdin for interactive terminal sessions
}

type PreflightRunner interface {
	Run(preflight.KubeconfigOptions) (preflight.Result, error)
}

type InventoryCollector interface {
	CollectCore(preflight.KubeconfigOptions, preflight.Result) (inventory.Snapshot, error)
	CollectAssessment(preflight.KubeconfigOptions, preflight.Result) (inventory.Snapshot, error)
}

type ProviderFactory interface {
	NewProvider(inventory.Snapshot, Config) provider.Provider
}

type APIAnalyzer interface {
	Analyze(context.Context, Config, string) ([]kubent.Finding, recommendation.Limitation)
}

type defaultProviderFactory struct{}

func (defaultProviderFactory) NewProvider(snapshot inventory.Snapshot, cfg Config) provider.Provider {
	return aks.NewAKSProvider(aks.IdentitySignals{
		ExplicitSubscription:  cfg.Subscription,
		ExplicitResourceGroup: cfg.ResourceGroup,
		ExplicitClusterName:   cfg.ClusterName,
		ContextName:           snapshot.Cluster.Context.Name,
	})
}

func Run(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	return RunWithDependencies(args, stdout, stderr, build, Dependencies{
		PreflightRunner:    preflight.LiveRunner{},
		InventoryCollector: inventory.LiveCollector{},
		ProviderFactory:    defaultProviderFactory{},
		APIAnalyzer:        kubentAnalyzer{},
		Stdin:              os.Stdin,
	})
}

func RunWithDependencies(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo, deps Dependencies) int {
	cfg, positional, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err.Message)
		printUsage(stderr)
		return err.Code
	}
	if cfg.Help {
		printUsage(stdout)
		return ExitReady
	}
	_ = newLogger(stderr, cfg.LogLevel)

	if len(positional) == 0 {
		printUsage(stderr)
		return ExitUsage
	}

	switch positional[0] {
	case "version":
		printVersion(stdout, build)
		return ExitReady
	case "inventory":
		return runInventory(cfg, stdout, stderr, deps.PreflightRunner, deps.InventoryCollector)
	case "analyze":
		return runAnalyze(cfg, stdout, stderr, deps)
	case "health":
		return runAnalyzeSubset(cfg, stdout, stderr, deps, "HEALTH")
	case "compatibility":
		return runAnalyzeSubset(cfg, stdout, stderr, deps, "COMPATIBILITY")
	case "report":
		if len(positional) > 2 {
			fmt.Fprintln(stderr, "report accepts at most one positional input path")
			printUsage(stderr)
			return ExitUsage
		}
		if len(positional) == 2 {
			if cfg.InputPath != "" {
				fmt.Fprintln(stderr, "report input is ambiguous: use either positional path or --input, not both")
				printUsage(stderr)
				return ExitUsage
			}
			cfg.InputPath = positional[1]
		}
		return runReport(cfg, stdout, stderr)
	case "component-overrides":
		return runComponentOverrides(cfg, stdout, stderr)
	case "help", "-h", "--help":
		printUsage(stdout)
		return ExitReady
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", positional[0])
		printUsage(stderr)
		return ExitUsage
	}
}

func runAnalyze(cfg Config, stdout io.Writer, stderr io.Writer, deps Dependencies) int {
	if cfg.Preflight && cfg.DayOf {
		fmt.Fprintln(stderr, "kua analyze: --preflight and --day-of are mutually exclusive")
		return ExitUsage
	}

	interactive := isInteractiveTTY(deps.Stdin)

	// Resolve display label for confirmation and --yes banner.
	contextName, serverURL := preflight.ResolveContextDisplay(
		preflight.KubeconfigOptions{Path: cfg.Kubeconfig, Context: cfg.Context},
	)
	clusterLabel := contextLabel(contextName, serverURL)

	if interactive {
		if !confirmAnalysis(bufio.NewReader(deps.Stdin), stdout, clusterLabel) {
			fmt.Fprintln(stdout, "Analysis cancelled.")
			return ExitReady
		}
	} else if !cfg.Yes {
		fmt.Fprintln(stderr, "kua analyze: requires confirmation; pass --yes to run non-interactively")
		return ExitUsage
	} else if cfg.Format == "console" {
		fmt.Fprintf(stdout, "Analyzing cluster: %s\n", clusterLabel)
	}

	if cfg.Format == "console" {
		if cfg.Preflight {
			fmt.Fprintln(stdout, "Mode: PRE-FLIGHT (API + component + provider checks; day-of health checks skipped)")
		} else if cfg.DayOf {
			fmt.Fprintln(stdout, "Mode: DAY-OF (health checks + cached pre-flight data)")
		}
	}

	doc, appErr := buildAssessmentDocument(cfg, deps)
	if appErr != nil {
		fmt.Fprintln(stderr, errorMessageForOutput(appErr.Message, cfg.Redacted))
		return appErr.Code
	}

	// After first pass: if no destination and no explicit target, prompt interactively.
	if interactive && cfg.Format == "console" && cfg.TargetVersion == "" && doc.Destination == "" {
		reader := bufio.NewReader(deps.Stdin)
		if v := promptTargetVersion(reader, stdout); v != "" {
			cfg.TargetVersion = v
			doc, appErr = buildAssessmentDocument(cfg, deps)
			if appErr != nil {
				fmt.Fprintln(stderr, errorMessageForOutput(appErr.Message, cfg.Redacted))
				return appErr.Code
			}
		}
	}

	// One interactive loop: prompt for unknown component versions then re-analyze.
	if interactive && cfg.Format == "console" &&
		doc.ComponentVersionOverrides != nil &&
		len(doc.ComponentVersionOverrides.Components) > 0 {

		reader := bufio.NewReader(deps.Stdin)
		answers := promptComponentVersions(reader, stdout, doc.ComponentVersionOverrides.Components)
		if len(answers) > 0 {
			cfg.inlineOverrides = answers
			fmt.Fprintln(stdout, "\nRe-analyzing with provided component versions...")
			doc, appErr = buildAssessmentDocument(cfg, deps)
			if appErr != nil {
				fmt.Fprintln(stderr, errorMessageForOutput(appErr.Message, cfg.Redacted))
				return appErr.Code
			}
		}
	}

	exitCode := renderAssessment(cfg, doc, stdout, stderr)
	if cfg.Preflight && cfg.Format == "console" {
		cachePath := cfg.resolvedPreflightCachePath()
		fmt.Fprintf(stdout, "\nPre-flight cache saved: %s\n", cachePath)
		fmt.Fprintf(stdout, "Run day-of:  kua analyze --day-of --yes\n")
	}
	return exitCode
}

// isInteractiveTTY reports whether r is a character device (interactive terminal).
func isInteractiveTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// contextLabel builds a display string for the confirmation prompt.
func contextLabel(name, server string) string {
	if name == "" {
		name = "(current context)"
	}
	if server != "" {
		return name + " (" + server + ")"
	}
	return name
}

// confirmAnalysis prompts the operator and returns true only on explicit yes.
func confirmAnalysis(reader *bufio.Reader, stdout io.Writer, clusterLabel string) bool {
	fmt.Fprintf(stdout, "\nCluster: %s\n", clusterLabel)
	fmt.Fprintf(stdout, "Analyze this cluster? [y/N]: ")
	line, _ := reader.ReadString('\n')
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}

// promptComponentVersions asks for missing component versions and returns the answers.
func promptComponentVersions(reader *bufio.Reader, stdout io.Writer, requests []report.ComponentVersionOverrideRequest) map[string]string {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Some component versions could not be detected automatically:")
	overrides := make(map[string]string)
	for _, req := range requests {
		fmt.Fprintf(stdout, "\n  %s", req.Name)
		if len(req.ObservedVersions) > 0 {
			fmt.Fprintf(stdout, " (observed: %s)", strings.Join(req.ObservedVersions, ", "))
		}
		fmt.Fprintf(stdout, "\n  Reason: %s\n", req.Reason)
		fmt.Fprintf(stdout, "  Version [skip]: ")
		line, _ := reader.ReadString('\n')
		answer := strings.TrimSpace(line)
		if answer != "" && answer != "skip" {
			overrides[req.ID] = answer
		}
	}
	return overrides
}

// promptTargetVersion asks the operator for a destination version when the provider cannot supply one.
func promptTargetVersion(reader *bufio.Reader, stdout io.Writer) string {
	fmt.Fprintln(stdout, "\nUpgrade targets could not be detected from the provider.")
	fmt.Fprintln(stdout, "To fetch available upgrades from AKS:")
	fmt.Fprintln(stdout, "  az aks get-upgrades --resource-group <rg> --name <cluster> -o json > upgrades.json")
	fmt.Fprintln(stdout, "  kua analyze --provider-evidence upgrades.json")
	fmt.Fprintf(stdout, "\nEnter target version (e.g. 1.35.0) [skip]: ")
	line, _ := reader.ReadString('\n')
	answer := strings.TrimSpace(line)
	if answer == "" || answer == "skip" {
		return ""
	}
	return answer
}

func runAnalyzeSubset(cfg Config, stdout io.Writer, stderr io.Writer, deps Dependencies, subset string) int {
	doc, appErr := buildAssessmentDocument(cfg, deps)
	if appErr != nil {
		fmt.Fprintln(stderr, errorMessageForOutput(appErr.Message, cfg.Redacted))
		return appErr.Code
	}
	switch subset {
	case "HEALTH":
		doc.Findings = filterFindings(doc.Findings, recommendation.CategoryHealth)
	case "COMPATIBILITY":
		doc.Findings = filterFindings(doc.Findings, recommendation.CategoryAPI, recommendation.CategoryComponent)
	}
	return renderAssessment(cfg, doc, stdout, stderr)
}

func errorMessageForOutput(message string, redacted bool) string {
	if !redacted {
		return message
	}
	return redactCLIError(message)
}

var (
	urlHostPattern     = regexp.MustCompile(`https?://[^/\s:]+(?::[0-9]+)?`)
	aksHostPattern     = regexp.MustCompile(`[A-Za-z0-9.-]+\.azmk8s\.io(?::[0-9]+)?`)
	dnsLookupPattern   = regexp.MustCompile(`lookup [^:\s]+`)
	hostBracketPattern = regexp.MustCompile(`Host: "[^"]+"`)
)

func redactCLIError(message string) string {
	redacted := urlHostPattern.ReplaceAllString(message, "https://redacted-host")
	redacted = aksHostPattern.ReplaceAllString(redacted, "redacted-host")
	redacted = dnsLookupPattern.ReplaceAllString(redacted, "lookup redacted-host")
	redacted = hostBracketPattern.ReplaceAllString(redacted, `Host: "redacted-host"`)
	return redacted
}

func buildAssessmentDocument(cfg Config, deps Dependencies) (report.Document, *AppError) {
	if cfg.Preflight {
		return buildPreflightDocument(cfg, deps)
	}
	if cfg.DayOf {
		return buildDayOfDocument(cfg, deps)
	}
	if deps.PreflightRunner == nil {
		return report.Document{}, ExecutionError("analyze preflight runner is not configured", nil)
	}
	if deps.InventoryCollector == nil {
		return report.Document{}, ExecutionError("analyze inventory collector is not configured", nil)
	}
	if deps.ProviderFactory == nil {
		deps.ProviderFactory = defaultProviderFactory{}
	}
	if deps.APIAnalyzer == nil {
		deps.APIAnalyzer = kubentAnalyzer{}
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}

	options := preflight.KubeconfigOptions{Path: cfg.Kubeconfig, Context: cfg.Context}
	preflightResult, err := deps.PreflightRunner.Run(options)
	if err != nil {
		return report.Document{}, ExecutionError("analyze preflight failed: "+err.Error(), err)
	}
	if preflightResult.HasRequiredFailure() {
		rec := inconclusiveRecommendation(preflightResult.ServerVersion, clock(), "PREFLIGHT_REQUIRED_FAILURE", "Required Kubernetes preflight check failed")
		return documentFromRecommendation(rec, clock()), nil
	}

	snapshot, err := deps.InventoryCollector.CollectAssessment(options, preflightResult)
	if err != nil {
		return report.Document{}, ExecutionError("analyze inventory collection failed: "+err.Error(), err)
	}
	if err := inventory.ValidateCoreSnapshot(snapshot); err != nil {
		return report.Document{}, ExecutionError("analyze inventory snapshot validation failed: "+err.Error(), err)
	}

	healthFindings := health.NewRunner(health.DefaultRules()...).Evaluate(snapshot, health.Options{Now: clock})
	detections := components.NewRunner(components.InitialDetectorCohort()...).Detect(snapshot)
	overrides, err := loadComponentOverrides(cfg.ComponentOverrides)
	if err != nil {
		return report.Document{}, ExecutionError("read component overrides failed: "+err.Error(), err)
	}
	if len(cfg.inlineOverrides) > 0 {
		overrides = mergeInlineOverrides(overrides, cfg.inlineOverrides)
	}
	detections = applyComponentOverrides(detections, overrides)

	providerEvidence, providerLimit := collectProviderEvidence(context.Background(), cfg, deps.ProviderFactory.NewProvider(snapshot, cfg))
	targetVersion := cfg.TargetVersion
	if targetVersion == "" && providerEvidence != nil && len(providerEvidence.AvailableUpgrades) > 0 {
		targetVersion = highestProviderVersion(providerEvidence.AvailableUpgrades)
	}
	apiFindings, apiLimit := deps.APIAnalyzer.Analyze(context.Background(), cfg, targetVersion)

	resGroupResolved, clusterResolved := resolveClusterIdentity(cfg, providerEvidence)
	engine := recommendation.NewEngine().WithClock(clock)
	rec, err := engine.Generate(recommendation.Input{
		CurrentVersion:      trimVersionPrefix(snapshot.Kubernetes.ServerVersion),
		HealthFindings:      healthFindings,
		APIFindings:         apiFindings,
		ComponentDetections: detections,
		ProviderEvidence:    providerEvidence,
		InventorySnapshot:   &snapshot.Inventory,
		ClusterName:         clusterResolved,
		ResourceGroup:       resGroupResolved,
	}, recommendation.RecommendationOptions{
		TargetVersion: cfg.TargetVersion,
		MaxMinorSkip:  4,
	})
	if err != nil {
		return report.Document{}, ExecutionError("generate recommendation failed: "+err.Error(), err)
	}
	if apiLimit.Code != "" {
		rec.Limitations = append(rec.Limitations, apiLimit)
	}
	if providerLimit.Code != "" {
		rec.Limitations = append(rec.Limitations, providerLimit)
	}

	doc := documentFromRecommendation(rec, clock())
	if cfg.ComponentOverrides == "" {
		doc.ComponentVersionOverrides = buildComponentOverrideTemplate(detections, cfg)
	}
	return doc, nil
}

// buildPreflightDocument runs pre-flight analyzers only (API, component, provider)
// and writes the results to a local cache file for later --day-of reuse.
func buildPreflightDocument(cfg Config, deps Dependencies) (report.Document, *AppError) {
	if deps.PreflightRunner == nil {
		return report.Document{}, ExecutionError("analyze preflight runner is not configured", nil)
	}
	if deps.InventoryCollector == nil {
		return report.Document{}, ExecutionError("analyze inventory collector is not configured", nil)
	}
	if deps.ProviderFactory == nil {
		deps.ProviderFactory = defaultProviderFactory{}
	}
	if deps.APIAnalyzer == nil {
		deps.APIAnalyzer = kubentAnalyzer{}
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}

	options := preflight.KubeconfigOptions{Path: cfg.Kubeconfig, Context: cfg.Context}
	preflightResult, err := deps.PreflightRunner.Run(options)
	if err != nil {
		return report.Document{}, ExecutionError("analyze preflight failed: "+err.Error(), err)
	}
	if preflightResult.HasRequiredFailure() {
		rec := inconclusiveRecommendation(preflightResult.ServerVersion, clock(), "PREFLIGHT_REQUIRED_FAILURE", "Required Kubernetes preflight check failed")
		return documentFromRecommendation(rec, clock()), nil
	}

	snapshot, err := deps.InventoryCollector.CollectAssessment(options, preflightResult)
	if err != nil {
		return report.Document{}, ExecutionError("analyze inventory collection failed: "+err.Error(), err)
	}
	if err := inventory.ValidateCoreSnapshot(snapshot); err != nil {
		return report.Document{}, ExecutionError("analyze inventory snapshot validation failed: "+err.Error(), err)
	}

	detections := components.NewRunner(components.InitialDetectorCohort()...).Detect(snapshot)
	overrides, err := loadComponentOverrides(cfg.ComponentOverrides)
	if err != nil {
		return report.Document{}, ExecutionError("read component overrides failed: "+err.Error(), err)
	}
	if len(cfg.inlineOverrides) > 0 {
		overrides = mergeInlineOverrides(overrides, cfg.inlineOverrides)
	}
	detections = applyComponentOverrides(detections, overrides)

	providerEvidence, providerLimit := collectProviderEvidence(context.Background(), cfg, deps.ProviderFactory.NewProvider(snapshot, cfg))
	targetVersion := cfg.TargetVersion
	if targetVersion == "" && providerEvidence != nil && len(providerEvidence.AvailableUpgrades) > 0 {
		targetVersion = highestProviderVersion(providerEvidence.AvailableUpgrades)
	}
	apiFindings, apiLimit := deps.APIAnalyzer.Analyze(context.Background(), cfg, targetVersion)

	cacheEntry := PreflightCacheEntry{
		SchemaVersion:       preflightCacheSchema,
		CachedAt:            clock().UTC(),
		ContextName:         preflightResult.Context.Name,
		TargetVersion:       targetVersion,
		APIFindings:         apiFindings,
		APILimitation:       apiLimit,
		ComponentDetections: detections,
		ProviderEvidence:    providerEvidence,
		ProviderLimitation:  providerLimit,
	}
	if saveErr := savePreflightCache(cfg.resolvedPreflightCachePath(), cacheEntry); saveErr != nil {
		return report.Document{}, ExecutionError("save preflight cache failed: "+saveErr.Error(), saveErr)
	}

	resGroupResolved, clusterResolved := resolveClusterIdentity(cfg, providerEvidence)
	engine := recommendation.NewEngine().WithClock(clock)
	rec, err := engine.Generate(recommendation.Input{
		CurrentVersion:      trimVersionPrefix(snapshot.Kubernetes.ServerVersion),
		APIFindings:         apiFindings,
		ComponentDetections: detections,
		ProviderEvidence:    providerEvidence,
		InventorySnapshot:   &snapshot.Inventory,
		ClusterName:         clusterResolved,
		ResourceGroup:       resGroupResolved,
	}, recommendation.RecommendationOptions{
		TargetVersion: cfg.TargetVersion,
		MaxMinorSkip:  4,
	})
	if err != nil {
		return report.Document{}, ExecutionError("generate recommendation failed: "+err.Error(), err)
	}
	if apiLimit.Code != "" {
		rec.Limitations = append(rec.Limitations, apiLimit)
	}
	if providerLimit.Code != "" {
		rec.Limitations = append(rec.Limitations, providerLimit)
	}
	rec.Limitations = append(rec.Limitations, recommendation.Limitation{
		Code:    "PREFLIGHT_MODE",
		Summary: "Pre-flight mode: day-of health checks (node readiness, pod status, PVC binding) were not run",
		Impact:  "Health findings not included; run kua analyze --day-of to add them",
	})

	doc := documentFromRecommendation(rec, clock())
	if cfg.ComponentOverrides == "" {
		doc.ComponentVersionOverrides = buildComponentOverrideTemplate(detections, cfg)
	}
	return doc, nil
}

// buildDayOfDocument loads cached pre-flight results and runs day-of health checks.
func buildDayOfDocument(cfg Config, deps Dependencies) (report.Document, *AppError) {
	if deps.PreflightRunner == nil {
		return report.Document{}, ExecutionError("analyze preflight runner is not configured", nil)
	}
	if deps.InventoryCollector == nil {
		return report.Document{}, ExecutionError("analyze inventory collector is not configured", nil)
	}
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}

	cachePath := cfg.resolvedPreflightCachePath()
	cacheEntry, loadErr := loadPreflightCache(cachePath)
	if loadErr != nil {
		return report.Document{}, ExecutionError("load preflight cache failed: "+loadErr.Error()+" — run kua analyze --preflight first", loadErr)
	}

	options := preflight.KubeconfigOptions{Path: cfg.Kubeconfig, Context: cfg.Context}
	preflightResult, err := deps.PreflightRunner.Run(options)
	if err != nil {
		return report.Document{}, ExecutionError("analyze preflight failed: "+err.Error(), err)
	}
	if preflightResult.HasRequiredFailure() {
		rec := inconclusiveRecommendation(preflightResult.ServerVersion, clock(), "PREFLIGHT_REQUIRED_FAILURE", "Required Kubernetes preflight check failed")
		return documentFromRecommendation(rec, clock()), nil
	}

	snapshot, err := deps.InventoryCollector.CollectAssessment(options, preflightResult)
	if err != nil {
		return report.Document{}, ExecutionError("analyze inventory collection failed: "+err.Error(), err)
	}
	if err := inventory.ValidateCoreSnapshot(snapshot); err != nil {
		return report.Document{}, ExecutionError("analyze inventory snapshot validation failed: "+err.Error(), err)
	}

	healthFindings := health.NewRunner(health.DefaultRules()...).Evaluate(snapshot, health.Options{Now: clock})

	engine := recommendation.NewEngine().WithClock(clock)
	resGroupResolved, clusterResolved := resolveClusterIdentity(cfg, cacheEntry.ProviderEvidence)
	rec, err := engine.Generate(recommendation.Input{
		CurrentVersion:      trimVersionPrefix(snapshot.Kubernetes.ServerVersion),
		HealthFindings:      healthFindings,
		APIFindings:         cacheEntry.APIFindings,
		ComponentDetections: cacheEntry.ComponentDetections,
		ProviderEvidence:    cacheEntry.ProviderEvidence,
		InventorySnapshot:   &snapshot.Inventory,
		ClusterName:         clusterResolved,
		ResourceGroup:       resGroupResolved,
	}, recommendation.RecommendationOptions{
		TargetVersion: cacheEntry.TargetVersion,
		MaxMinorSkip:  4,
	})
	if err != nil {
		return report.Document{}, ExecutionError("generate recommendation failed: "+err.Error(), err)
	}
	if cacheEntry.APILimitation.Code != "" {
		rec.Limitations = append(rec.Limitations, cacheEntry.APILimitation)
	}
	if cacheEntry.ProviderLimitation.Code != "" {
		rec.Limitations = append(rec.Limitations, cacheEntry.ProviderLimitation)
	}
	age := clock().UTC().Sub(cacheEntry.CachedAt.UTC()).Round(time.Minute)
	rec.Limitations = append(rec.Limitations, recommendation.Limitation{
		Code:    "DAY_OF_PREFLIGHT_CACHE",
		Summary: fmt.Sprintf("Pre-flight data loaded from cache (cached %s ago)", formatDuration(age)),
		Impact:  "API/component/provider findings reflect cache state, not live cluster",
	})

	return documentFromRecommendation(rec, clock()), nil
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

type kubentAnalyzer struct{}

func (kubentAnalyzer) Analyze(ctx context.Context, cfg Config, targetVersion string) ([]kubent.Finding, recommendation.Limitation) {
	if targetVersion == "" {
		return nil, recommendation.Limitation{
			Code:    "API_TARGET_UNAVAILABLE",
			Summary: "API compatibility target version is unavailable",
			Impact:  "API compatibility remains inconclusive",
		}
	}
	adapter := kubent.Adapter{Runner: processRunner{}}
	version, err := adapter.ValidateVersion(ctx)
	if err != nil {
		return nil, recommendation.Limitation{
			Code:    "KUBENT_UNAVAILABLE",
			Summary: "kubent API compatibility evidence is unavailable",
			Impact:  "API compatibility remains inconclusive",
		}
	}
	report, err := adapter.RunJSON(ctx, targetVersion, cfg.Kubeconfig, cfg.Context)
	if err != nil {
		return nil, recommendation.Limitation{
			Code:    "KUBENT_EXECUTION_FAILED",
			Summary: "kubent API compatibility execution failed",
			Impact:  "API compatibility remains inconclusive",
		}
	}
	coverage := kubent.VerifyCoverage(targetVersion, kubent.DefaultCoveragePolicy())
	findings := kubent.NormalizeFindings(report, version, targetVersion, coverage)
	return findings, recommendation.Limitation{}
}

type processRunner struct{}

func (processRunner) Run(ctx context.Context, command kubent.Command) (kubent.Result, error) {
	timeout := command.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, command.Path, command.Args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := kubent.Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		return result, nil
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func collectProviderEvidence(ctx context.Context, cfg Config, p provider.Provider) (*provider.ProviderEvidence, recommendation.Limitation) {
	if p == nil {
		return nil, recommendation.Limitation{Code: "PROVIDER_UNAVAILABLE", Summary: "provider adapter is not configured", Impact: "provider availability remains unknown"}
	}
	evidence, err := p.Evidence(ctx, provider.EvidenceOptions{
		Mode:          provider.SourceMode(cfg.ProviderSource),
		FilePath:      cfg.ProviderEvidence,
		Subscription:  cfg.Subscription,
		ResourceGroup: cfg.ResourceGroup,
		ClusterName:   cfg.ClusterName,
		Timeout:       aks.DefaultTimeout,
	})
	if err != nil {
		return nil, recommendation.Limitation{Code: "PROVIDER_EVIDENCE_ERROR", Summary: sanitizeProviderError(err.Error()), Impact: "provider availability remains unknown"}
	}
	return evidence, recommendation.Limitation{}
}

// resolveClusterIdentity returns resource group and cluster name, preferring
// explicit cfg flags over values parsed from provider evidence.
func resolveClusterIdentity(cfg Config, evidence *provider.ProviderEvidence) (resourceGroup, clusterName string) {
	rg, cn := cfg.ResourceGroup, cfg.ClusterName
	if evidence != nil {
		if rg == "" {
			rg = evidence.Cluster.ResourceGroup
		}
		if cn == "" {
			cn = evidence.Cluster.ClusterName
		}
	}
	return rg, cn
}

func sanitizeProviderError(message string) string {
	if message == "" {
		return "provider evidence unavailable"
	}
	return "provider evidence unavailable; verify Azure CLI authentication or provide --provider-evidence"
}

func highestProviderVersion(upgrades []provider.UpgradeOption) string {
	var highest provider.SemanticVersion
	for _, upgrade := range upgrades {
		version, err := provider.ParseVersion(upgrade.Version)
		if err != nil {
			continue
		}
		if highest.Raw == "" || version.Compare(highest) > 0 {
			highest = version
		}
	}
	return highest.String()
}

func inconclusiveRecommendation(current string, now time.Time, code string, summary string) *recommendation.Recommendation {
	return &recommendation.Recommendation{
		SchemaVersion:  "kua.recommendation.v1",
		CurrentVersion: trimVersionPrefix(current),
		Readiness:      recommendation.ReadinessInconclusive,
		Risk:           recommendation.RiskUnknown,
		Findings:       []recommendation.Finding{},
		Limitations:    []recommendation.Limitation{{Code: code, Summary: summary, Impact: "assessment cannot continue"}},
		GeneratedAt:    now.UTC(),
	}
}

func documentFromRecommendation(rec *recommendation.Recommendation, now time.Time) report.Document {
	return report.Document{
		SchemaVersion: "kua.assessment.v1",
		AssessmentID:  fmt.Sprintf("assessment-%d", now.UTC().Unix()),
		GeneratedAt:   rec.GeneratedAt,
		Redacted:      false,
		Current:       rec.CurrentVersion,
		Destination:   rec.Destination,
		Readiness:     rec.Readiness,
		Risk:          rec.Risk,
		Path:          rec.Path,
		Decision:      rec.Decision,
		Confidence:    rec.Confidence,
		Evidence:      rec.Evidence,
		UpgradePlan:   rec.UpgradePlan,
		Findings:      rec.Findings,
		Limitations:   rec.Limitations,
	}
}

func renderAssessment(cfg Config, doc report.Document, stdout io.Writer, stderr io.Writer) int {
	format := report.RenderFormat(cfg.Format)
	content, err := report.Render(doc, report.RenderOptions{Format: format, Redacted: cfg.Redacted})
	if err != nil {
		appErr := ExecutionError("render assessment failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	if cfg.OutputPath != "" {
		if err := report.WriteAtomic(cfg.OutputPath, content); err != nil {
			appErr := ExecutionError("write assessment output failed: "+err.Error(), err)
			fmt.Fprintln(stderr, appErr.Message)
			return appErr.Code
		}
		return exitCodeForReadiness(doc.Readiness)
	}
	if _, err := stdout.Write(content); err != nil {
		appErr := ExecutionError("write assessment output failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	return exitCodeForReadiness(doc.Readiness)
}

func runReport(cfg Config, stdout io.Writer, stderr io.Writer) int {
	inputPath, appErr := resolveReportInputPath(cfg.InputPath)
	if appErr != nil {
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	data, err := os.ReadFile(inputPath)
	if err != nil {
		appErr := ExecutionError("read report input failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	var doc report.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		appErr := ExecutionError("parse report input failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	return renderAssessment(cfg, doc, stdout, stderr)
}

func resolveReportInputPath(explicitPath string) (string, *AppError) {
	if explicitPath != "" {
		return explicitPath, nil
	}
	candidates := []string{
		"assessment.json",
		"local-output/analyze.final.redacted.json",
		"local-output/analyze.redacted.json",
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", UsageError("missing report input: use kua report <assessment.json> or --input <assessment.json>; checked assessment.json, local-output/analyze.final.redacted.json, local-output/analyze.redacted.json")
}

func filterFindings(findings []recommendation.Finding, categories ...recommendation.FindingCategory) []recommendation.Finding {
	allowed := map[recommendation.FindingCategory]bool{}
	for _, category := range categories {
		allowed[category] = true
	}
	var out []recommendation.Finding
	for _, finding := range findings {
		if allowed[finding.Category] {
			out = append(out, finding)
		}
	}
	return out
}

func exitCodeForReadiness(readiness recommendation.ReadinessState) int {
	switch readiness {
	case recommendation.ReadinessNotReady:
		return ExitNotReady
	case recommendation.ReadinessInconclusive:
		return ExitInconclusive
	default:
		return ExitReady
	}
}

func trimVersionPrefix(version string) string {
	if len(version) > 0 && version[0] == 'v' {
		return version[1:]
	}
	return version
}

func runInventory(cfg Config, stdout io.Writer, stderr io.Writer, runner PreflightRunner, collector InventoryCollector) int {
	if runner == nil {
		appErr := ExecutionError("inventory preflight runner is not configured", nil)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}

	options := preflight.KubeconfigOptions{
		Path:    cfg.Kubeconfig,
		Context: cfg.Context,
	}
	result, err := runner.Run(options)
	if err != nil {
		appErr := ExecutionError("inventory preflight failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}

	if cfg.Format == "json" {
		if result.HasRequiredFailure() {
			if err := printInventoryPreflightJSON(stdout, result); err != nil {
				appErr := ExecutionError("inventory preflight JSON render failed: "+err.Error(), err)
				fmt.Fprintln(stderr, appErr.Message)
				return appErr.Code
			}
			return ExitInconclusive
		}
		return runInventorySnapshotJSON(options, result, stdout, stderr, collector)
	}

	fmt.Fprintln(stdout, "inventory preflight only")
	fmt.Fprintf(stdout, "context: %s\n", result.Context.Name)
	fmt.Fprintf(stdout, "kubeconfigSource: %s\n", result.Context.KubeconfigSource)
	fmt.Fprintf(stdout, "serverVersion: %s\n", result.ServerVersion)
	fmt.Fprintf(stdout, "discovery: %s\n", result.DiscoveryStatus)
	fmt.Fprintf(stdout, "requiredFailure: %t\n", result.HasRequiredFailure())
	fmt.Fprintf(stdout, "permissionChecks: %d\n", len(result.PermissionChecks))
	fmt.Fprintf(stdout, "limitations: %d\n", len(result.Limitations))

	if result.HasRequiredFailure() {
		return ExitInconclusive
	}
	return ExitReady
}

func runInventorySnapshotJSON(options preflight.KubeconfigOptions, result preflight.Result, stdout io.Writer, stderr io.Writer, collector InventoryCollector) int {
	if collector == nil {
		appErr := ExecutionError("inventory collector is not configured", nil)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}

	snapshot, err := collector.CollectCore(options, result)
	if err != nil {
		appErr := ExecutionError("inventory collection failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	if err := inventory.ValidateCoreSnapshot(snapshot); err != nil {
		appErr := ExecutionError("inventory snapshot validation failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(snapshot); err != nil {
		appErr := ExecutionError("inventory snapshot JSON render failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	return ExitReady
}

type inventoryPreflightDocument struct {
	SchemaVersion    string                      `json:"schemaVersion"`
	Kind             string                      `json:"kind"`
	PreflightOnly    bool                        `json:"preflightOnly"`
	Context          string                      `json:"context"`
	KubeconfigSource string                      `json:"kubeconfigSource"`
	ServerVersion    string                      `json:"serverVersion"`
	Discovery        preflight.Status            `json:"discovery"`
	RequiredFailure  bool                        `json:"requiredFailure"`
	PermissionChecks []preflight.PermissionCheck `json:"permissionChecks"`
	Limitations      []preflight.Limitation      `json:"limitations"`
}

func printInventoryPreflightJSON(stdout io.Writer, result preflight.Result) error {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(inventoryPreflightDocument{
		SchemaVersion:    schemaVersion,
		Kind:             "InventoryPreflight",
		PreflightOnly:    true,
		Context:          result.Context.Name,
		KubeconfigSource: string(result.Context.KubeconfigSource),
		ServerVersion:    result.ServerVersion,
		Discovery:        result.DiscoveryStatus,
		RequiredFailure:  result.HasRequiredFailure(),
		PermissionChecks: result.PermissionChecks,
		Limitations:      result.Limitations,
	})
}

func newLogger(w io.Writer, level string) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slogLevel}))
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `%s analyzes Kubernetes upgrade readiness.

Usage:
  %s <command>

Commands:
  analyze        Run full upgrade readiness assessment
  inventory      Collect and summarize inventory
  health         Run health checks
  compatibility  Run API and component compatibility checks
report         Render a saved assessment (%s report <assessment.json> or --input <assessment.json>)
  version        Print build and contract versions

Flags (analyze):
  --yes            Skip confirmation prompt (required when not running interactively)
  --preflight      Run pre-flight checks only; save results to cache for --day-of
  --day-of         Run day-of health checks using cached pre-flight results
  --preflight-cache <path>  Override pre-flight cache file path (default: kua-preflight.json)

Flags (report):
  --input <path>   Input assessment JSON file
  --format <fmt>   Output format: console|json|markdown|html
  --output <path>  Write rendered output to file
  --redacted       Redact sensitive host details in output

Global flags:
  --help           Print this usage text and exit
`, binaryName, binaryName, binaryName)
}

func printVersion(w io.Writer, build BuildInfo) {
	fmt.Fprintf(w, "%s version: %s\n", binaryName, defaultString(build.Version, "0.0.0-dev"))
	fmt.Fprintf(w, "commit: %s\n", defaultString(build.Commit, "unknown"))
	fmt.Fprintf(w, "buildDate: %s\n", defaultString(build.BuildDate, "unknown"))
	fmt.Fprintf(w, "go: %s\n", runtime.Version())
	fmt.Fprintf(w, "assessmentSchema: %s\n", schemaVersion)
	fmt.Fprintf(w, "catalogVersion: %s\n", defaultString(build.CatalogVersion, embeddedCatalogVersion()))
}

// embeddedCatalogVersion reads the catalog version baked into the binary via
// go:embed, used when no build-time catalogVersion ldflag was set.
func embeddedCatalogVersion() string {
	bundle, err := catalog.LoadEmbedded()
	if err != nil || bundle.CatalogVersion == "" {
		return "unavailable"
	}
	return bundle.CatalogVersion
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
