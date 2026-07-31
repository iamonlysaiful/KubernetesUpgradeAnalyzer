package recommendation

import (
	"fmt"
	"sort"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/external/kubent"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/health"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/plan"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

const schemaVersion = "kua.recommendation.v1"

// Input contains all data needed for recommendation.
type Input struct {
	// CurrentVersion is the cluster's current Kubernetes version.
	CurrentVersion string
	// HealthFindings are from the health analyzer.
	HealthFindings []health.Finding
	// APIFindings are from the API compatibility analyzer.
	APIFindings []kubent.Finding
	// ComponentDetections are from the component detector.
	ComponentDetections []components.Detection
	// ProviderEvidence is from the provider adapter.
	ProviderEvidence *provider.ProviderEvidence
	// InventorySnapshot provides raw cluster inventory for evidence summary.
	InventorySnapshot *inventory.Inventory
	// ClusterName is used for plan generation (optional).
	ClusterName string
	// ResourceGroup is used for AKS plan generation (optional).
	ResourceGroup string
}

// Engine produces upgrade recommendations.
type Engine struct {
	aggregator      *Aggregator
	policy          *Policy
	evidenceBuilder *EvidenceBuilder
	planGenerator   *plan.Generator
	clock           func() time.Time
}

// NewEngine creates a new recommendation engine.
func NewEngine() *Engine {
	return &Engine{
		aggregator:      NewAggregator(),
		policy:          NewPolicy(),
		evidenceBuilder: NewEvidenceBuilder(),
		planGenerator:   plan.NewGenerator(),
		clock:           time.Now,
	}
}

// WithClock sets a custom clock for testing.
func (e *Engine) WithClock(clock func() time.Time) *Engine {
	e.clock = clock
	e.evidenceBuilder = e.evidenceBuilder.WithClock(clock)
	return e
}

// Generate produces a recommendation from the input.
func (e *Engine) Generate(input Input, opts RecommendationOptions) (*Recommendation, error) {
	if opts.MaxMinorSkip == 0 {
		opts = DefaultOptions()
	}

	// Use risk profile from options, defaulting to balanced
	policy := e.policy
	if opts.RiskProfile != "" && opts.RiskProfile != e.policy.RiskProfile() {
		policy = NewPolicyWithProfile(opts.RiskProfile)
	}

	rec := &Recommendation{
		SchemaVersion:  schemaVersion,
		CurrentVersion: input.CurrentVersion,
		Findings:       make([]Finding, 0),
		Limitations:    make([]Limitation, 0),
		Path:           make([]UpgradeStage, 0),
		GeneratedAt:    e.clock().UTC(),
	}

	// Aggregate findings from all sources
	rec.Findings = append(rec.Findings, e.aggregator.AggregateHealth(input.HealthFindings)...)

	// Aggregate API findings for current target
	if len(input.APIFindings) > 0 {
		targetVersion := ""
		if input.ProviderEvidence != nil && len(input.ProviderEvidence.AvailableUpgrades) > 0 {
			// Use highest available version as target for API analysis
			targetVersion = e.findHighestVersion(input.ProviderEvidence.AvailableUpgrades)
		}
		rec.Findings = append(rec.Findings, e.aggregator.AggregateAPI(input.APIFindings, targetVersion)...)
	}

	// Aggregate component findings
	targetMinor := 0
	if input.ProviderEvidence != nil {
		// Parse target minor from provider evidence
		if cv, err := provider.ParseVersion(input.CurrentVersion); err == nil {
			targetMinor = cv.Minor + opts.MaxMinorSkip
		}
	}
	rec.Findings = append(rec.Findings, e.aggregator.AggregateComponents(input.ComponentDetections, targetMinor)...)

	// Aggregate limitations
	rec.Limitations = e.aggregator.AggregateLimitations(input.APIFindings, input.ComponentDetections)

	// Check provider evidence availability
	if input.ProviderEvidence == nil {
		rec.Limitations = append(rec.Limitations, Limitation{
			Code:    "PROVIDER_UNAVAILABLE",
			Summary: "Provider evidence not available",
			Impact:  "Cannot determine available upgrade versions",
		})
	} else {
		// Add provider limitations
		for _, lim := range input.ProviderEvidence.Limitations {
			rec.Limitations = append(rec.Limitations, Limitation{
				Code:    lim.Code,
				Summary: lim.Summary,
			})
		}
	}

	// Build upgrade path only when provider evidence includes candidate versions.
	if input.ProviderEvidence != nil && input.ProviderEvidence.CurrentVersion != "" && len(input.ProviderEvidence.AvailableUpgrades) > 0 {
		path, pathLimitations, err := e.buildPath(input, opts)
		if err != nil {
			rec.Limitations = append(rec.Limitations, Limitation{
				Code:    "PATH_BUILD_FAILED",
				Summary: err.Error(),
				Impact:  "Cannot determine upgrade path",
			})
		} else {
			rec.Path = path
			rec.Limitations = append(rec.Limitations, pathLimitations...)
			for _, stage := range path {
				rec.Findings = append(rec.Findings, stage.Findings...)
			}

			// Set destination from path
			if len(path) > 0 {
				rec.Destination = path[len(path)-1].To
			}
		}
	}

	// Fall back to the operator-provided target version when provider evidence
	// could not confirm a destination or upgrade path (e.g. no CLI access or
	// evidence file). The destination is then unvalidated against the provider.
	if rec.Destination == "" && opts.TargetVersion != "" {
		if tv, err := provider.ParseVersion(opts.TargetVersion); err == nil {
			rec.Destination = tv.String()
			rec.Limitations = append(rec.Limitations, Limitation{
				Code:    "DESTINATION_UNVALIDATED",
				Summary: fmt.Sprintf("Destination %s set from operator-provided target version; not validated against provider evidence", tv.String()),
				Impact:  "Upgrade path stages and provider validation are unavailable; verify supported upgrade steps manually",
			})
		}
	}

	// Evaluate readiness and risk (legacy)
	rec.Readiness = policy.EvaluateReadiness(rec.Findings, rec.Limitations)
	rec.Risk = policy.EvaluateRisk(rec.Findings, rec.Limitations)

	// Evaluate confidence and decision (Phase 10)
	confidence := policy.EvaluateConfidence(rec.Findings, rec.Limitations)
	rec.Decision = confidence.Decision
	rec.Confidence = &confidence

	// Build evidence summary (Phase 10)
	deprecatedAPICount := e.countDeprecatedAPIs(rec.Findings)
	rec.Evidence = e.evidenceBuilder.Build(input.InventorySnapshot, input.ComponentDetections, deprecatedAPICount)

	// Generate upgrade plan (Phase 10)
	if rec.Destination != "" {
		rec.UpgradePlan = e.generateUpgradePlan(input, rec)
	}

	// Sort findings by severity
	e.sortFindings(rec.Findings)

	return rec, nil
}

// buildPath constructs the sequential upgrade path.
func (e *Engine) buildPath(input Input, opts RecommendationOptions) ([]UpgradeStage, []Limitation, error) {
	if input.ProviderEvidence == nil {
		return nil, nil, fmt.Errorf("provider evidence required for path construction")
	}

	// Build candidate set
	candidates, err := provider.BuildCandidateSet(input.ProviderEvidence, opts.AllowPreview)
	if err != nil {
		return nil, nil, fmt.Errorf("build candidate set: %w", err)
	}

	// Determine destination
	destination, err := e.selectDestination(candidates, opts)
	if err != nil {
		return nil, nil, err
	}

	// Build sequential path
	path, err := provider.BuildSequentialPath(candidates, destination)
	if err != nil {
		directPath, ok := e.providerDirectPath(candidates, destination)
		if !ok {
			return nil, nil, fmt.Errorf("build sequential path: %w", err)
		}
		return directPath, providerDirectLimitations(candidates.Current, destination), nil
	}
	if !path.IsValid {
		directPath, ok := e.providerDirectPath(candidates, destination)
		if ok {
			return directPath, providerDirectLimitations(candidates.Current, destination), nil
		}
	}

	// Convert to UpgradeStages
	stages := make([]UpgradeStage, 0, len(path.Steps))
	for _, step := range path.Steps {
		stage := UpgradeStage{
			From:            step.From.String(),
			To:              step.To.String(),
			IsProviderValid: step.IsProviderValid,
			Findings:        make([]Finding, 0),
		}

		// Add stage-specific finding if not provider valid
		if !step.IsProviderValid {
			stage.Findings = append(stage.Findings, Finding{
				ID:          "PROVIDER_TRANSITION_INVALID",
				Category:    CategoryProvider,
				Severity:    SeverityBlocker,
				Summary:     fmt.Sprintf("Upgrade from %s to %s is not provider-validated", step.From.String(), step.To.String()),
				Stage:       step.To.String(),
				Remediation: "Verify upgrade availability with provider or use a different target version",
			})
		}

		stages = append(stages, stage)
	}

	// Convert path limitations
	limitations := make([]Limitation, 0, len(path.Limitations))
	for _, lim := range path.Limitations {
		limitations = append(limitations, Limitation{
			Code:    lim.Code,
			Summary: lim.Summary,
		})
	}

	return stages, limitations, nil
}

func providerDirectLimitations(current provider.SemanticVersion, destination provider.SemanticVersion) []Limitation {
	return []Limitation{{
		Code:    "PROVIDER_DIRECT_MULTI_MINOR",
		Summary: fmt.Sprintf("Provider advertises direct upgrade from %s to %s without lower intermediate minors", current.String(), destination.String()),
		Impact:  "Review provider release notes and maintenance window before upgrade",
	}}
}

func (e *Engine) providerDirectPath(candidates *provider.CandidateSet, destination provider.SemanticVersion) ([]UpgradeStage, bool) {
	for _, opt := range candidates.AllVersions {
		version, err := provider.ParseVersion(opt.Version)
		if err != nil {
			continue
		}
		if version.Compare(destination) != 0 {
			continue
		}
		return []UpgradeStage{{
			From:            candidates.Current.String(),
			To:              destination.String(),
			IsProviderValid: true,
			Findings: []Finding{{
				ID:          "PROVIDER_DIRECT_MULTI_MINOR",
				Category:    CategoryProvider,
				Severity:    SeverityWarning,
				Summary:     fmt.Sprintf("Provider advertises direct upgrade from %s to %s without lower intermediate minors", candidates.Current.String(), destination.String()),
				Stage:       destination.String(),
				Remediation: "Review AKS release notes, backup, and maintenance window before using provider-direct multi-minor upgrade",
			}},
		}}, true
	}
	return nil, false
}

// selectDestination chooses the highest suitable candidate.
func (e *Engine) selectDestination(candidates *provider.CandidateSet, opts RecommendationOptions) (provider.SemanticVersion, error) {
	if len(candidates.AllVersions) == 0 {
		return provider.SemanticVersion{}, fmt.Errorf("no upgrade candidates available")
	}

	// If explicit target specified, use it
	if opts.TargetVersion != "" {
		return provider.ParseVersion(opts.TargetVersion)
	}

	// Find highest version within max minor skip
	maxMinor := candidates.Current.Minor + opts.MaxMinorSkip

	var highest provider.SemanticVersion
	for _, opt := range candidates.AllVersions {
		v, err := provider.ParseVersion(opt.Version)
		if err != nil {
			continue
		}
		if v.Minor > maxMinor {
			continue
		}
		if highest.Raw == "" || v.Compare(highest) > 0 {
			highest = v
		}
	}

	if highest.Raw == "" {
		return provider.SemanticVersion{}, fmt.Errorf("no suitable destination within %d minor versions", opts.MaxMinorSkip)
	}

	return highest, nil
}

// findHighestVersion returns the highest version string from upgrade options.
func (e *Engine) findHighestVersion(upgrades []provider.UpgradeOption) string {
	if len(upgrades) == 0 {
		return ""
	}

	var highest provider.SemanticVersion
	for _, opt := range upgrades {
		v, err := provider.ParseVersion(opt.Version)
		if err != nil {
			continue
		}
		if highest.Raw == "" || v.Compare(highest) > 0 {
			highest = v
		}
	}

	return highest.String()
}

// sortFindings sorts findings by severity (blockers first), then by category and ID.
func (e *Engine) sortFindings(findings []Finding) {
	severityOrder := map[FindingSeverity]int{
		SeverityBlocker: 0,
		SeverityWarning: 1,
		SeverityUnknown: 2,
		SeverityInfo:    3,
	}

	sort.Slice(findings, func(i, j int) bool {
		// Sort by severity first
		if severityOrder[findings[i].Severity] != severityOrder[findings[j].Severity] {
			return severityOrder[findings[i].Severity] < severityOrder[findings[j].Severity]
		}
		// Then by category
		if findings[i].Category != findings[j].Category {
			return findings[i].Category < findings[j].Category
		}
		// Then by ID
		return findings[i].ID < findings[j].ID
	})
}

// countDeprecatedAPIs counts findings related to deprecated APIs.
func (e *Engine) countDeprecatedAPIs(findings []Finding) int {
	count := 0
	for _, f := range findings {
		if f.Category == CategoryAPI && (f.ID == "API_DEPRECATED" || f.ID == "API_REMOVED") {
			count++
		}
	}
	return count
}

// generateUpgradePlan creates an upgrade plan based on inputs and recommendation.
func (e *Engine) generateUpgradePlan(input Input, rec *Recommendation) *plan.UpgradePlan {
	planInput := plan.PlanInput{
		CurrentVersion: input.CurrentVersion,
		TargetVersion:  rec.Destination,
		ClusterName:    input.ClusterName,
		ResourceGroup:  input.ResourceGroup,
		RiskLevel:      string(rec.Risk),
	}

	// Determine provider
	if input.ProviderEvidence != nil {
		planInput.Provider = string(input.ProviderEvidence.Cluster.Provider)
		// Count node pools
		planInput.NodePoolCount = len(input.ProviderEvidence.NodePools)
		if planInput.NodePoolCount == 0 {
			planInput.NodePoolCount = 1 // Assume at least 1
		}
	}

	// Count nodes and workloads from inventory
	if input.InventorySnapshot != nil {
		planInput.TotalNodeCount = len(input.InventorySnapshot.Nodes)

		for _, w := range input.InventorySnapshot.Workloads {
			if w.Ref.Kind == "StatefulSet" {
				planInput.HasStatefulSets = true
			}
		}

		for _, n := range input.InventorySnapshot.Networking {
			if n.Kind == "Ingress" {
				planInput.HasIngress = true
				break
			}
		}
	}

	// Extract detected component names
	for _, d := range input.ComponentDetections {
		if d.Status == components.StatusFound || d.Status == components.StatusUnknown {
			planInput.DetectedComponents = append(planInput.DetectedComponents, d.Name)
		}
	}

	return e.planGenerator.Generate(planInput)
}
