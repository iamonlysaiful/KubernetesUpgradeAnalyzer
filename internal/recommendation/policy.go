package recommendation

// Policy evaluates findings to determine readiness and risk.
type Policy struct{}

// NewPolicy creates a new policy evaluator.
func NewPolicy() *Policy {
	return &Policy{}
}

// EvaluateReadiness determines the overall readiness state from findings.
func (p *Policy) EvaluateReadiness(findings []Finding, limitations []Limitation) ReadinessState {
	hasBlocker := false
	hasWarning := false
	hasUnknown := false

	for _, f := range findings {
		switch f.Severity {
		case SeverityBlocker:
			hasBlocker = true
		case SeverityWarning:
			hasWarning = true
		case SeverityUnknown:
			hasUnknown = true
		}
	}

	// Check limitations for critical evidence gaps
	for _, lim := range limitations {
		if p.isCriticalLimitation(lim) {
			hasUnknown = true
		}
	}

	// Determine readiness state
	if hasBlocker {
		return ReadinessNotReady
	}
	if hasUnknown && !hasWarning {
		return ReadinessInconclusive
	}
	if hasWarning {
		return ReadinessReadyWithWarnings
	}
	if hasUnknown {
		return ReadinessInconclusive
	}
	return ReadinessReady
}

// EvaluateRisk determines the risk level from findings and limitations.
func (p *Policy) EvaluateRisk(findings []Finding, limitations []Limitation) RiskLevel {
	hasBlocker := false
	hasWarning := false
	hasUnknown := false

	for _, f := range findings {
		switch f.Severity {
		case SeverityBlocker:
			hasBlocker = true
		case SeverityWarning:
			hasWarning = true
		case SeverityUnknown:
			hasUnknown = true
		}
	}

	// Check limitations for critical evidence gaps
	for _, lim := range limitations {
		if p.isCriticalLimitation(lim) {
			hasUnknown = true
		}
	}

	// Determine risk level (rule-based, not scored)
	if hasBlocker {
		return RiskHigh
	}
	if hasUnknown {
		return RiskUnknown
	}
	if hasWarning {
		return RiskMedium
	}
	return RiskLow
}

// isCriticalLimitation returns true if the limitation affects recommendation validity.
func (p *Policy) isCriticalLimitation(lim Limitation) bool {
	criticalCodes := map[string]bool{
		"TARGET_COVERAGE_UNVERIFIED": true,
		"PROVIDER_UNAVAILABLE":       true,
		"CATALOG_CORRUPT":            true,
		"ANALYZER_FAILED":            true,
	}
	return criticalCodes[lim.Code]
}

// IsBlocker returns true if a finding is a blocker.
func (p *Policy) IsBlocker(f Finding) bool {
	return f.Severity == SeverityBlocker
}

// BlockerCodes defines the finding IDs that are always blockers.
var BlockerCodes = map[string]bool{
	// API blockers
	"API_REMOVED": true,

	// Health blockers
	"NODE_NOT_READY":       true,
	"WORKLOAD_UNAVAILABLE": true,
	"PVC_UNBOUND":          true,

	// Provider blockers
	"PROVIDER_TRANSITION_INVALID": true,

	// Component blockers
	"COMPONENT_INCOMPATIBLE": true,
}

// WarningCodes defines the finding IDs that are warnings.
var WarningCodes = map[string]bool{
	// API warnings
	"API_DEPRECATED": true,

	// Health warnings
	"NODE_MEMORY_PRESSURE": true,
	"NODE_DISK_PRESSURE":   true,
	"NODE_PID_PRESSURE":    true,
	"EVENTS_WARNING":       true,

	// Component warnings
	"COMPONENT_CONDITIONAL": true,
}

// ClassifyFinding determines the severity of a finding based on policy.
func (p *Policy) ClassifyFinding(f Finding) FindingSeverity {
	if f.Severity != "" {
		return f.Severity
	}

	if BlockerCodes[f.ID] {
		return SeverityBlocker
	}
	if WarningCodes[f.ID] {
		return SeverityWarning
	}
	return SeverityInfo
}
