package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"runtime"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
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
}

type PreflightRunner interface {
	Run(preflight.KubeconfigOptions) (preflight.Result, error)
}

type InventoryCollector interface {
	CollectCore(preflight.KubeconfigOptions, preflight.Result) (inventory.Snapshot, error)
}

func Run(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo) int {
	return RunWithDependencies(args, stdout, stderr, build, Dependencies{
		PreflightRunner:    preflight.LiveRunner{},
		InventoryCollector: inventory.LiveCollector{},
	})
}

func RunWithDependencies(args []string, stdout io.Writer, stderr io.Writer, build BuildInfo, deps Dependencies) int {
	cfg, positional, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err.Message)
		printUsage(stderr)
		return err.Code
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
	case "analyze", "health", "compatibility", "report":
		appErr := UnimplementedError(positional[0])
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	case "help", "-h", "--help":
		printUsage(stdout)
		return ExitReady
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", positional[0])
		printUsage(stderr)
		return ExitUsage
	}
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
  report         Render a saved assessment
  version        Print build and contract versions
`, binaryName, binaryName)
}

func printVersion(w io.Writer, build BuildInfo) {
	fmt.Fprintf(w, "%s version: %s\n", binaryName, defaultString(build.Version, "0.0.0-dev"))
	fmt.Fprintf(w, "commit: %s\n", defaultString(build.Commit, "unknown"))
	fmt.Fprintf(w, "buildDate: %s\n", defaultString(build.BuildDate, "unknown"))
	fmt.Fprintf(w, "go: %s\n", runtime.Version())
	fmt.Fprintf(w, "assessmentSchema: %s\n", schemaVersion)
	fmt.Fprintf(w, "catalogVersion: %s\n", defaultString(build.CatalogVersion, "unavailable"))
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
