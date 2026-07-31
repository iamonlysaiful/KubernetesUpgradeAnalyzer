package recommendation

import (
	"testing"
)

func TestConfidenceCalculator_AllPass(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	// No findings, no limitations
	// Component Compatibility: 0.7 (no components detected)
	// All others: 1.0
	// Expected: 0.25 + (0.20*0.7) + 0.20 + 0.15 + 0.10 + 0.10 = 0.94 = 94%
	model := calc.Calculate(nil, nil)

	if model.Percentage != 94 {
		t.Errorf("expected 94%% confidence (no components detected), got %d%%", model.Percentage)
	}
	if model.Decision != DecisionGo {
		t.Errorf("expected GO decision, got %s", model.Decision)
	}
	if model.HasBlocker {
		t.Error("expected no blocker")
	}
	if len(model.Factors) != 6 {
		t.Errorf("expected 6 factors, got %d", len(model.Factors))
	}
}

func TestConfidenceCalculator_WithBlocker(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	findings := []Finding{
		{
			ID:       "API_REMOVED",
			Category: CategoryAPI,
			Severity: SeverityBlocker,
			Summary:  "extensions/v1beta1 Ingress removed in 1.22",
		},
	}

	model := calc.Calculate(findings, nil)

	if model.Decision != DecisionStop {
		t.Errorf("expected DO_NOT_PROCEED with blocker, got %s", model.Decision)
	}
	if !model.HasBlocker {
		t.Error("expected HasBlocker to be true")
	}
}

func TestConfidenceCalculator_KubentCoverageUnverified(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	limitations := []Limitation{
		{
			Code:    "TARGET_COVERAGE_UNVERIFIED",
			Summary: "kubent target-rule coverage is not verified",
		},
	}

	model := calc.Calculate(nil, limitations)

	// API Compatibility: 0.25 * 0.5 = 0.125
	// Component Compatibility: 0.20 * 0.7 = 0.14 (no components detected)
	// Cluster Health: 0.20 * 1.0 = 0.20
	// Provider Evidence: 0.15 * 1.0 = 0.15
	// Storage Health: 0.10 * 1.0 = 0.10
	// Analysis Coverage: 0.10 * 0.5 = 0.05
	// Expected: 0.125 + 0.14 + 0.20 + 0.15 + 0.10 + 0.05 = 0.765 = 76%

	if model.Percentage < 74 || model.Percentage > 78 {
		t.Errorf("expected ~76%% confidence with unverified coverage, got %d%%", model.Percentage)
	}
	if model.Decision != DecisionCaution {
		t.Errorf("expected GO_WITH_CAUTION (76%% with balanced), got %s", model.Decision)
	}
}

func TestConfidenceCalculator_FileEvidence(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	limitations := []Limitation{
		{
			Code:    "FILE_EVIDENCE_SOURCE",
			Summary: "Evidence loaded from file",
		},
	}

	model := calc.Calculate(nil, limitations)

	// API Compatibility: 0.25 * 1.0 = 0.25
	// Component Compatibility: 0.20 * 0.7 = 0.14 (no components detected)
	// Cluster Health: 0.20 * 1.0 = 0.20
	// Provider Evidence: 0.15 * 0.8 = 0.12
	// Storage Health: 0.10 * 1.0 = 0.10
	// Analysis Coverage: 0.10 * 1.0 = 0.10
	// Expected: 0.25 + 0.14 + 0.20 + 0.12 + 0.10 + 0.10 = 0.91 = 91%

	if model.Percentage < 89 || model.Percentage > 93 {
		t.Errorf("expected ~91%% confidence with file evidence, got %d%%", model.Percentage)
	}
	if model.Decision != DecisionGo {
		t.Errorf("expected GO (91%% with balanced), got %s", model.Decision)
	}
}

func TestConfidenceCalculator_ProviderUnavailable(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	limitations := []Limitation{
		{
			Code:    "PROVIDER_UNAVAILABLE",
			Summary: "Provider upgrade availability unknown",
		},
	}

	model := calc.Calculate(nil, limitations)

	// API Compatibility: 0.25 * 1.0 = 0.25
	// Component Compatibility: 0.20 * 0.7 = 0.14 (no components detected)
	// Cluster Health: 0.20 * 1.0 = 0.20
	// Provider Evidence: 0.15 * 0.5 = 0.075
	// Storage Health: 0.10 * 1.0 = 0.10
	// Analysis Coverage: 0.10 * 1.0 = 0.10
	// Expected: 0.25 + 0.14 + 0.20 + 0.075 + 0.10 + 0.10 = 0.865 = 86%

	if model.Percentage < 84 || model.Percentage > 88 {
		t.Errorf("expected ~86%% confidence with unavailable provider, got %d%%", model.Percentage)
	}
}

func TestConfidenceCalculator_UnknownComponents(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	findings := []Finding{
		{
			ID:       "COMPONENT_EMQX",
			Category: CategoryComponent,
			Severity: SeverityUnknown,
			Summary:  "EMQX version unknown",
		},
		{
			ID:       "COMPONENT_COREDNS",
			Category: CategoryComponent,
			Severity: SeverityInfo,
			Summary:  "CoreDNS 1.9.4 compatible",
		},
	}

	model := calc.Calculate(findings, nil)

	// Component Compatibility: 0.20 * ~0.85 = 0.17 (1 unknown out of 2)
	// Expected: 0.25 + 0.17 + 0.20 + 0.15 + 0.10 + 0.10 = 0.97 = 97%

	if model.Percentage < 95 || model.Percentage > 99 {
		t.Errorf("expected ~97%% confidence with one unknown component, got %d%%", model.Percentage)
	}
}

func TestConfidenceCalculator_HealthWarning(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)

	findings := []Finding{
		{
			ID:       "POD_RESTARTING",
			Category: CategoryHealth,
			Severity: SeverityWarning,
			Summary:  "Pod has restarted 3 times in last hour",
		},
	}

	model := calc.Calculate(findings, nil)

	// API Compatibility: 0.25 * 1.0 = 0.25
	// Component Compatibility: 0.20 * 0.7 = 0.14 (no components detected)
	// Cluster Health: 0.20 * 0.8 = 0.16
	// Provider Evidence: 0.15 * 1.0 = 0.15
	// Storage Health: 0.10 * 1.0 = 0.10
	// Analysis Coverage: 0.10 * 1.0 = 0.10
	// Expected: 0.25 + 0.14 + 0.16 + 0.15 + 0.10 + 0.10 = 0.90 = 90%

	if model.Percentage < 88 || model.Percentage > 92 {
		t.Errorf("expected ~90%% confidence with health warning, got %d%%", model.Percentage)
	}
}

func TestConfidenceCalculator_RiskProfiles(t *testing.T) {
	// Test that different profiles produce different decisions at boundary

	findings := []Finding{}
	limitations := []Limitation{
		{Code: "TARGET_COVERAGE_UNVERIFIED", Summary: "test"},
		{Code: "PROVIDER_UNAVAILABLE", Summary: "test"},
	}

	// This should give ~75-80% confidence

	t.Run("Conservative", func(t *testing.T) {
		calc := NewConfidenceCalculator(RiskProfileConservative)
		model := calc.Calculate(findings, limitations)

		// ~75% is below conservative CAUTION threshold (85%)
		if model.Percentage > 80 && model.Decision != DecisionStop {
			t.Errorf("conservative profile should be STOP at %d%%", model.Percentage)
		}
	})

	t.Run("Balanced", func(t *testing.T) {
		calc := NewConfidenceCalculator(RiskProfileBalanced)
		model := calc.Calculate(findings, limitations)

		// ~75% is above balanced CAUTION threshold (70%)
		if model.Percentage >= 70 && model.Percentage < 90 && model.Decision != DecisionCaution {
			t.Errorf("balanced profile should be CAUTION at %d%%, got %s", model.Percentage, model.Decision)
		}
	})

	t.Run("Aggressive", func(t *testing.T) {
		calc := NewConfidenceCalculator(RiskProfileAggressive)
		model := calc.Calculate(findings, limitations)

		// ~75% is above aggressive CAUTION threshold (60%)
		if model.Percentage >= 60 && model.Percentage < 80 && model.Decision != DecisionCaution {
			t.Errorf("aggressive profile should be CAUTION at %d%%, got %s", model.Percentage, model.Decision)
		}
	})
}

func TestConfidenceCalculator_FactorWeightsSum(t *testing.T) {
	var sum float64
	for _, w := range FactorWeight {
		sum += w
	}
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("factor weights should sum to 1.0, got %f", sum)
	}
}

func TestConfidenceCalculator_AllFactorsPresent(t *testing.T) {
	calc := NewConfidenceCalculator(RiskProfileBalanced)
	model := calc.Calculate(nil, nil)

	factorsSeen := make(map[ConfidenceFactor]bool)
	for _, f := range model.Factors {
		factorsSeen[f.Factor] = true
	}

	expectedFactors := []ConfidenceFactor{
		FactorAPICompatibility,
		FactorComponentCompatibility,
		FactorClusterHealth,
		FactorProviderEvidence,
		FactorStorageHealth,
		FactorAnalysisCoverage,
	}

	for _, ef := range expectedFactors {
		if !factorsSeen[ef] {
			t.Errorf("missing factor: %s", ef)
		}
	}
}

func TestConfidenceCalculator_Thresholds(t *testing.T) {
	tests := []struct {
		name       string
		profile    RiskProfile
		confidence float64
		hasBlocker bool
		expected   Decision
	}{
		{"balanced_go", RiskProfileBalanced, 0.95, false, DecisionGo},
		{"balanced_caution", RiskProfileBalanced, 0.85, false, DecisionCaution},
		{"balanced_stop_low", RiskProfileBalanced, 0.65, false, DecisionStop},
		{"balanced_stop_blocker", RiskProfileBalanced, 0.95, true, DecisionStop},
		{"conservative_go", RiskProfileConservative, 0.96, false, DecisionGo},
		{"conservative_caution", RiskProfileConservative, 0.90, false, DecisionCaution},
		{"conservative_stop", RiskProfileConservative, 0.80, false, DecisionStop},
		{"aggressive_go", RiskProfileAggressive, 0.85, false, DecisionGo},
		{"aggressive_caution", RiskProfileAggressive, 0.65, false, DecisionCaution},
		{"aggressive_stop", RiskProfileAggressive, 0.55, false, DecisionStop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewConfidenceCalculator(tt.profile)
			decision := calc.determineDecision(tt.confidence, tt.hasBlocker)
			if decision != tt.expected {
				t.Errorf("expected %s, got %s (confidence=%.2f, blocker=%v)",
					tt.expected, decision, tt.confidence, tt.hasBlocker)
			}
		})
	}
}
