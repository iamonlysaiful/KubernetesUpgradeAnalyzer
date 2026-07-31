package recommendation

// Decision represents the traffic light upgrade decision.
type Decision string

const (
	// DecisionGo indicates confidence ≥90% with no blockers.
	DecisionGo Decision = "GO"
	// DecisionCaution indicates confidence 70-89% with no blockers.
	DecisionCaution Decision = "GO_WITH_CAUTION"
	// DecisionStop indicates blockers present or confidence <70%.
	DecisionStop Decision = "DO_NOT_PROCEED"
)

// ImpactLevel indicates the severity of a finding's impact.
type ImpactLevel string

const (
	// ImpactLow indicates minimal effect on upgrade success.
	ImpactLow ImpactLevel = "Low"
	// ImpactMedium indicates moderate effect on upgrade success.
	ImpactMedium ImpactLevel = "Medium"
	// ImpactHigh indicates significant effect on upgrade success.
	ImpactHigh ImpactLevel = "High"
)

// RiskProfile defines confidence thresholds for decisions.
type RiskProfile string

const (
	// RiskProfileConservative uses higher thresholds (GO ≥95%, CAUTION ≥85%).
	RiskProfileConservative RiskProfile = "conservative"
	// RiskProfileBalanced uses default thresholds (GO ≥90%, CAUTION ≥70%).
	RiskProfileBalanced RiskProfile = "balanced"
	// RiskProfileAggressive uses lower thresholds (GO ≥80%, CAUTION ≥60%).
	RiskProfileAggressive RiskProfile = "aggressive"
)

// ConfidenceFactor identifies a factor contributing to confidence.
type ConfidenceFactor string

const (
	// FactorAPICompatibility represents API deprecation/removal analysis.
	FactorAPICompatibility ConfidenceFactor = "API_COMPATIBILITY"
	// FactorComponentCompatibility represents component version analysis.
	FactorComponentCompatibility ConfidenceFactor = "COMPONENT_COMPATIBILITY"
	// FactorClusterHealth represents node and workload health.
	FactorClusterHealth ConfidenceFactor = "CLUSTER_HEALTH"
	// FactorProviderEvidence represents upgrade availability evidence.
	FactorProviderEvidence ConfidenceFactor = "PROVIDER_EVIDENCE"
	// FactorStorageHealth represents PVC and storage status.
	FactorStorageHealth ConfidenceFactor = "STORAGE_HEALTH"
	// FactorAnalysisCoverage represents analyzer rule coverage.
	FactorAnalysisCoverage ConfidenceFactor = "ANALYSIS_COVERAGE"
)

// FactorWeight defines the weight for each confidence factor.
// Weights sum to 1.0.
var FactorWeight = map[ConfidenceFactor]float64{
	FactorAPICompatibility:       0.25,
	FactorComponentCompatibility: 0.20,
	FactorClusterHealth:          0.20,
	FactorProviderEvidence:       0.15,
	FactorStorageHealth:          0.10,
	FactorAnalysisCoverage:       0.10,
}

// ContributionFactor represents one factor's contribution to confidence.
type ContributionFactor struct {
	// Factor identifies which confidence factor this is.
	Factor ConfidenceFactor `json:"factor"`
	// Weight is the factor's weight in the overall score (0.0-1.0).
	Weight float64 `json:"weight"`
	// Confidence is this factor's confidence (0.0-1.0).
	Confidence float64 `json:"confidence"`
	// Contribution is weight × confidence.
	Contribution float64 `json:"contribution"`
	// Evidence explains why this confidence was assigned.
	Evidence string `json:"evidence"`
}

// ConfidenceModel contains the complete confidence calculation.
type ConfidenceModel struct {
	// Score is the overall confidence (0.0-1.0).
	Score float64 `json:"score"`
	// Percentage is Score as a percentage (0-100).
	Percentage int `json:"percentage"`
	// Factors contains each factor's contribution.
	Factors []ContributionFactor `json:"factors"`
	// HasBlocker indicates if any blocker was found.
	HasBlocker bool `json:"hasBlocker"`
	// Decision is the traffic light decision.
	Decision Decision `json:"decision"`
	// Profile is the risk profile used for thresholds.
	Profile RiskProfile `json:"profile"`
}

// FindingImpact describes the impact of a finding.
type FindingImpact struct {
	// Level is the impact severity.
	Level ImpactLevel `json:"level"`
	// Explanation describes what this affects.
	Explanation string `json:"explanation"`
}

// ActionItem describes a concrete action to address a finding.
type ActionItem struct {
	// Description is what to do.
	Description string `json:"description"`
	// Command is an optional command to run.
	Command string `json:"command,omitempty"`
	// Effort is estimated time to complete.
	Effort string `json:"effort,omitempty"`
}

// RiskProfileThresholds defines the GO/CAUTION thresholds for a profile.
type RiskProfileThresholds struct {
	GoThreshold      float64
	CautionThreshold float64
}

// ProfileThresholds maps risk profiles to their thresholds.
var ProfileThresholds = map[RiskProfile]RiskProfileThresholds{
	RiskProfileConservative: {GoThreshold: 0.95, CautionThreshold: 0.85},
	RiskProfileBalanced:     {GoThreshold: 0.90, CautionThreshold: 0.70},
	RiskProfileAggressive:   {GoThreshold: 0.80, CautionThreshold: 0.60},
}

// ConfidenceCalculator calculates confidence from findings and limitations.
type ConfidenceCalculator struct {
	profile RiskProfile
}

// NewConfidenceCalculator creates a new calculator with the given profile.
func NewConfidenceCalculator(profile RiskProfile) *ConfidenceCalculator {
	if profile == "" {
		profile = RiskProfileBalanced
	}
	return &ConfidenceCalculator{profile: profile}
}

// Calculate computes confidence from findings and limitations.
func (c *ConfidenceCalculator) Calculate(findings []Finding, limitations []Limitation) ConfidenceModel {
	model := ConfidenceModel{
		Profile: c.profile,
		Factors: make([]ContributionFactor, 0, len(FactorWeight)),
	}

	// Check for blockers first
	for _, f := range findings {
		if f.Severity == SeverityBlocker {
			model.HasBlocker = true
			break
		}
	}

	// Calculate each factor's confidence
	factors := []struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{
		c.calculateAPICompatibility(findings, limitations),
		c.calculateComponentCompatibility(findings, limitations),
		c.calculateClusterHealth(findings, limitations),
		c.calculateProviderEvidence(findings, limitations),
		c.calculateStorageHealth(findings, limitations),
		c.calculateAnalysisCoverage(findings, limitations),
	}

	// Build contribution factors and sum score
	for _, f := range factors {
		weight := FactorWeight[f.factor]
		contribution := weight * f.confidence
		model.Score += contribution

		model.Factors = append(model.Factors, ContributionFactor{
			Factor:       f.factor,
			Weight:       weight,
			Confidence:   f.confidence,
			Contribution: contribution,
			Evidence:     f.evidence,
		})
	}

	// Convert to percentage
	model.Percentage = int(model.Score * 100)

	// Determine decision based on thresholds
	model.Decision = c.determineDecision(model.Score, model.HasBlocker)

	return model
}

// determineDecision returns the traffic light decision.
func (c *ConfidenceCalculator) determineDecision(score float64, hasBlocker bool) Decision {
	if hasBlocker {
		return DecisionStop
	}

	thresholds := ProfileThresholds[c.profile]
	if score >= thresholds.GoThreshold {
		return DecisionGo
	}
	if score >= thresholds.CautionThreshold {
		return DecisionCaution
	}
	return DecisionStop
}

// Factor calculation helpers

func (c *ConfidenceCalculator) calculateAPICompatibility(findings []Finding, limitations []Limitation) struct {
	factor     ConfidenceFactor
	confidence float64
	evidence   string
} {
	result := struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{factor: FactorAPICompatibility, confidence: 1.0, evidence: "No deprecated or removed APIs found"}

	// Check for API-related findings
	for _, f := range findings {
		if f.Category == CategoryAPI {
			if f.Severity == SeverityBlocker {
				result.confidence = 0.0
				result.evidence = "Removed API in use: " + f.Summary
				return result
			}
			if f.Severity == SeverityWarning {
				result.confidence = 0.8
				result.evidence = "Deprecated API detected: " + f.Summary
			}
		}
	}

	// Check for kubent coverage limitations
	for _, lim := range limitations {
		if lim.Code == "TARGET_COVERAGE_UNVERIFIED" || lim.Code == "TARGET_COVERAGE_MISSING" {
			result.confidence = 0.5
			result.evidence = "API analysis coverage not verified for target version"
			return result
		}
		if lim.Code == "KUBENT_UNAVAILABLE" || lim.Code == "KUBENT_EXECUTION_FAILED" {
			result.confidence = 0.3
			result.evidence = "kubent analyzer unavailable"
			return result
		}
		if lim.Code == "API_TARGET_UNAVAILABLE" {
			result.confidence = 0.5
			result.evidence = "API compatibility cannot be verified without a target version"
			return result
		}
	}

	return result
}

func (c *ConfidenceCalculator) calculateComponentCompatibility(findings []Finding, limitations []Limitation) struct {
	factor     ConfidenceFactor
	confidence float64
	evidence   string
} {
	result := struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{factor: FactorComponentCompatibility, confidence: 1.0, evidence: "All detected components verified compatible"}

	componentCount := 0
	unknownCount := 0
	partialCount := 0
	incompatibleCount := 0

	for _, f := range findings {
		if f.Category == CategoryComponent {
			componentCount++
			switch f.Severity {
			case SeverityBlocker:
				incompatibleCount++
			case SeverityUnknown:
				unknownCount++
			case SeverityWarning:
				partialCount++
			}
		}
	}

	if incompatibleCount > 0 {
		result.confidence = 0.0
		result.evidence = "Incompatible component detected"
		return result
	}

	if componentCount == 0 {
		result.confidence = 0.7
		result.evidence = "No components detected for verification"
		return result
	}

	if unknownCount > 0 || partialCount > 0 {
		// Fully unknown components count fully against the known ratio;
		// partial/ambiguous evidence counts at half weight.
		penalized := float64(unknownCount) + 0.5*float64(partialCount)
		knownRatio := (float64(componentCount) - penalized) / float64(componentCount)
		if knownRatio < 0 {
			knownRatio = 0
		}
		result.confidence = 0.7 + (0.3 * knownRatio) // Range: 0.7-1.0
		result.evidence = "Some component versions unverified or ambiguous"
		return result
	}

	return result
}

func (c *ConfidenceCalculator) calculateClusterHealth(findings []Finding, limitations []Limitation) struct {
	factor     ConfidenceFactor
	confidence float64
	evidence   string
} {
	result := struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{factor: FactorClusterHealth, confidence: 1.0, evidence: "All nodes and workloads healthy"}

	hasHealthWarning := false

	for _, f := range findings {
		if f.Category == CategoryHealth {
			if f.Severity == SeverityBlocker {
				result.confidence = 0.0
				result.evidence = "Critical health issue: " + f.Summary
				return result
			}
			if f.Severity == SeverityWarning {
				hasHealthWarning = true
			}
		}
	}

	if hasHealthWarning {
		result.confidence = 0.8
		result.evidence = "Minor health warnings present"
	}

	return result
}

func (c *ConfidenceCalculator) calculateProviderEvidence(findings []Finding, limitations []Limitation) struct {
	factor     ConfidenceFactor
	confidence float64
	evidence   string
} {
	result := struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{factor: FactorProviderEvidence, confidence: 1.0, evidence: "Fresh provider upgrade evidence available"}

	for _, lim := range limitations {
		if lim.Code == "PROVIDER_UNAVAILABLE" {
			result.confidence = 0.5
			result.evidence = "Provider upgrade availability unknown"
			return result
		}
		if lim.Code == "FILE_EVIDENCE_SOURCE" {
			result.confidence = 0.8
			result.evidence = "Provider evidence from file (may not reflect current state)"
			return result
		}
		if lim.Code == "PROVIDER_EVIDENCE_ERROR" {
			result.confidence = 0.3
			result.evidence = "Provider evidence fetch failed"
			return result
		}
		if lim.Code == "PROVIDER_EVIDENCE_UNAVAILABLE" {
			result.confidence = 0.5
			result.evidence = "Provider upgrade evidence unavailable (no CLI access or evidence file)"
			return result
		}
	}

	// Check for provider-related findings
	for _, f := range findings {
		if f.Category == CategoryProvider {
			if f.Severity == SeverityBlocker {
				result.confidence = 0.0
				result.evidence = "Provider blocks upgrade: " + f.Summary
				return result
			}
			if f.Severity == SeverityWarning {
				result.confidence = 0.7
				result.evidence = "Provider warning: " + f.Summary
			}
		}
	}

	return result
}

func (c *ConfidenceCalculator) calculateStorageHealth(findings []Finding, limitations []Limitation) struct {
	factor     ConfidenceFactor
	confidence float64
	evidence   string
} {
	result := struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{factor: FactorStorageHealth, confidence: 1.0, evidence: "All PVCs bound, no storage pressure"}

	for _, f := range findings {
		if f.Category == CategoryHealth && (f.ID == "PVC_UNBOUND" || f.ID == "STORAGE_PRESSURE") {
			if f.Severity == SeverityBlocker {
				result.confidence = 0.0
				result.evidence = "Storage blocker: " + f.Summary
				return result
			}
			if f.Severity == SeverityWarning {
				result.confidence = 0.8
				result.evidence = "Storage warning: " + f.Summary
			}
		}
	}

	return result
}

func (c *ConfidenceCalculator) calculateAnalysisCoverage(findings []Finding, limitations []Limitation) struct {
	factor     ConfidenceFactor
	confidence float64
	evidence   string
} {
	result := struct {
		factor     ConfidenceFactor
		confidence float64
		evidence   string
	}{factor: FactorAnalysisCoverage, confidence: 1.0, evidence: "All analyzers completed with verified coverage"}

	for _, lim := range limitations {
		switch lim.Code {
		case "TARGET_COVERAGE_UNVERIFIED":
			result.confidence = 0.5
			result.evidence = "Analyzer rule coverage not verified for target version"
			return result
		case "ANALYZER_FAILED":
			result.confidence = 0.3
			result.evidence = "One or more analyzers failed"
			return result
		case "PARTIAL_INVENTORY":
			result.confidence = 0.7
			result.evidence = "Partial cluster inventory collected"
		}
	}

	return result
}
