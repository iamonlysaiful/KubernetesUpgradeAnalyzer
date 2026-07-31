// Package recommendation provides the upgrade recommendation engine.
package recommendation

import (
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/plan"
)

// ReadinessState indicates overall upgrade readiness.
type ReadinessState string

const (
	// ReadinessReady indicates no blockers and sufficient evidence.
	ReadinessReady ReadinessState = "READY"
	// ReadinessReadyWithWarnings indicates no blockers but warnings need review.
	ReadinessReadyWithWarnings ReadinessState = "READY_WITH_WARNINGS"
	// ReadinessNotReady indicates one or more blockers prevent upgrade.
	ReadinessNotReady ReadinessState = "NOT_READY"
	// ReadinessInconclusive indicates required evidence is absent or ambiguous.
	ReadinessInconclusive ReadinessState = "INCONCLUSIVE"
)

// RiskLevel indicates the risk level of the upgrade.
type RiskLevel string

const (
	// RiskLow indicates all analyzers pass with adequate evidence.
	RiskLow RiskLevel = "LOW"
	// RiskMedium indicates no blockers but material warnings exist.
	RiskMedium RiskLevel = "MEDIUM"
	// RiskHigh indicates any blocker or explicitly unsupported condition.
	RiskHigh RiskLevel = "HIGH"
	// RiskUnknown indicates required evidence is insufficient.
	RiskUnknown RiskLevel = "UNKNOWN"
)

// FindingSeverity indicates the severity of a finding.
type FindingSeverity string

const (
	// SeverityBlocker prevents upgrade.
	SeverityBlocker FindingSeverity = "BLOCKER"
	// SeverityWarning allows upgrade with review.
	SeverityWarning FindingSeverity = "WARNING"
	// SeverityInfo is informational only.
	SeverityInfo FindingSeverity = "INFO"
	// SeverityUnknown indicates insufficient evidence.
	SeverityUnknown FindingSeverity = "UNKNOWN"
)

// FindingCategory identifies the source of a finding.
type FindingCategory string

const (
	// CategoryHealth indicates a health-related finding.
	CategoryHealth FindingCategory = "HEALTH"
	// CategoryAPI indicates an API compatibility finding.
	CategoryAPI FindingCategory = "API"
	// CategoryComponent indicates a component compatibility finding.
	CategoryComponent FindingCategory = "COMPONENT"
	// CategoryProvider indicates a provider evidence finding.
	CategoryProvider FindingCategory = "PROVIDER"
)

// Finding represents a single recommendation finding.
type Finding struct {
	// ID is a unique identifier for the finding type.
	ID string `json:"id"`
	// Category identifies the finding source.
	Category FindingCategory `json:"category"`
	// Severity indicates the impact.
	Severity FindingSeverity `json:"severity"`
	// Summary is a human-readable description.
	Summary string `json:"summary"`
	// Detail provides additional context.
	Detail string `json:"detail,omitempty"`
	// Resource identifies the affected resource if applicable.
	Resource *ResourceRef `json:"resource,omitempty"`
	// Stage indicates which upgrade stage is affected.
	Stage string `json:"stage,omitempty"`
	// Remediation suggests how to address the finding.
	Remediation string `json:"remediation,omitempty"`
	// Impact describes the finding's effect (Phase 10).
	Impact *FindingImpact `json:"impact,omitempty"`
	// Action is the recommended action to address this (Phase 10).
	Action *ActionItem `json:"action,omitempty"`
	// IfIgnored describes consequences of not addressing this (Phase 10).
	IfIgnored string `json:"ifIgnored,omitempty"`
}

// ResourceRef identifies a Kubernetes resource.
type ResourceRef struct {
	// Kind is the resource kind.
	Kind string `json:"kind"`
	// Namespace is the resource namespace.
	Namespace string `json:"namespace,omitempty"`
	// Name is the resource name.
	Name string `json:"name"`
}

// Limitation describes an evidence gap.
type Limitation struct {
	// Code is a unique identifier for the limitation type.
	Code string `json:"code"`
	// Summary is a human-readable description.
	Summary string `json:"summary"`
	// Impact describes how this affects the recommendation.
	Impact string `json:"impact,omitempty"`
}

// UpgradeStage represents one step in the upgrade path.
type UpgradeStage struct {
	// From is the source version.
	From string `json:"from"`
	// To is the target version.
	To string `json:"to"`
	// IsProviderValid indicates if provider confirms this transition.
	IsProviderValid bool `json:"isProviderValid"`
	// Findings contains stage-specific findings.
	Findings []Finding `json:"findings,omitempty"`
}

// Recommendation contains the complete upgrade recommendation.
type Recommendation struct {
	// SchemaVersion identifies the recommendation schema.
	SchemaVersion string `json:"schemaVersion"`
	// CurrentVersion is the cluster's current version.
	CurrentVersion string `json:"currentVersion"`
	// Destination is the recommended target version.
	Destination string `json:"destination,omitempty"`
	// Path is the sequential upgrade stages.
	Path []UpgradeStage `json:"path,omitempty"`
	// Readiness indicates overall upgrade readiness (legacy).
	Readiness ReadinessState `json:"readiness"`
	// Risk indicates the risk level (legacy).
	Risk RiskLevel `json:"risk"`
	// Decision is the traffic light decision (Phase 10).
	Decision Decision `json:"decision,omitempty"`
	// Confidence is the confidence model (Phase 10).
	Confidence *ConfidenceModel `json:"confidence,omitempty"`
	// Evidence summarizes what was analyzed (Phase 10).
	Evidence *EvidenceSummary `json:"evidence,omitempty"`
	// UpgradePlan provides step-by-step upgrade instructions (Phase 10).
	UpgradePlan *plan.UpgradePlan `json:"upgradePlan,omitempty"`
	// Findings contains all blockers, warnings, and info.
	Findings []Finding `json:"findings"`
	// Limitations lists evidence gaps.
	Limitations []Limitation `json:"limitations"`
	// GeneratedAt is when the recommendation was produced.
	GeneratedAt time.Time `json:"generatedAt"`
}

// RecommendationOptions configures the recommendation engine.
type RecommendationOptions struct {
	// TargetVersion is an explicit destination (optional).
	TargetVersion string
	// AllowPreview includes preview versions.
	AllowPreview bool
	// MaxMinorSkip limits the destination (default: 4 minors).
	MaxMinorSkip int
	// RiskProfile controls confidence thresholds (default: balanced).
	RiskProfile RiskProfile
}

// DefaultOptions returns the default recommendation options.
func DefaultOptions() RecommendationOptions {
	return RecommendationOptions{
		MaxMinorSkip: 4,
		AllowPreview: false,
		RiskProfile:  RiskProfileBalanced,
	}
}
