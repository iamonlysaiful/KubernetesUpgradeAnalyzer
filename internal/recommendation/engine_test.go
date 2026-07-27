package recommendation

import (
	"testing"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/external/kubent"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/health"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

func fixedClock() time.Time {
	return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
}

func TestEngine_Generate_Ready(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		HealthFindings: []health.Finding{
			{RuleID: "NODE_READY", Status: health.StatusPass},
		},
		APIFindings: []kubent.Finding{
			{Status: kubent.FindingPass, TargetVersion: "1.33"},
		},
		ComponentDetections: []components.Detection{
			{ComponentID: "nginx-ingress", Name: "NGINX Ingress", Version: "1.12.1", Status: components.StatusFound},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.31.2", IsPreview: false},
				{Version: "1.32.1", IsPreview: false},
				{Version: "1.33.0", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Readiness != ReadinessReady {
		t.Errorf("Readiness = %v, want READY", rec.Readiness)
	}
	if rec.Risk != RiskLow {
		t.Errorf("Risk = %v, want LOW", rec.Risk)
	}
	if rec.Destination != "1.33.0" {
		t.Errorf("Destination = %v, want 1.33.0", rec.Destination)
	}
	if len(rec.Path) != 3 {
		t.Errorf("Path length = %d, want 3", len(rec.Path))
	}
}

func TestEngine_Generate_ReadyWithWarnings(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		HealthFindings: []health.Finding{
			{
				RuleID:   "NODE_MEMORY_PRESSURE",
				Severity: health.SeverityWarning,
				Status:   health.StatusWarn,
				Summary:  "Node has memory pressure",
			},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.31.0", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Readiness != ReadinessReadyWithWarnings {
		t.Errorf("Readiness = %v, want READY_WITH_WARNINGS", rec.Readiness)
	}
	if rec.Risk != RiskMedium {
		t.Errorf("Risk = %v, want MEDIUM", rec.Risk)
	}
}

func TestEngine_Generate_NotReady_Blocker(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		HealthFindings: []health.Finding{
			{
				RuleID:   "NODE_NOT_READY",
				Severity: health.SeverityBlocker,
				Status:   health.StatusFail,
				Summary:  "Node is not ready",
				Resource: health.ResourceRef{Kind: "Node", Name: "node-1"},
			},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.31.0", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Readiness != ReadinessNotReady {
		t.Errorf("Readiness = %v, want NOT_READY", rec.Readiness)
	}
	if rec.Risk != RiskHigh {
		t.Errorf("Risk = %v, want HIGH", rec.Risk)
	}

	// Should have blocker finding
	hasBlocker := false
	for _, f := range rec.Findings {
		if f.Severity == SeverityBlocker {
			hasBlocker = true
			break
		}
	}
	if !hasBlocker {
		t.Error("expected blocker finding")
	}
}

func TestEngine_Generate_Inconclusive_NoProvider(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		HealthFindings: []health.Finding{
			{RuleID: "NODE_READY", Status: health.StatusPass},
		},
		ProviderEvidence: nil, // No provider evidence
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Readiness != ReadinessInconclusive {
		t.Errorf("Readiness = %v, want INCONCLUSIVE", rec.Readiness)
	}
	if rec.Risk != RiskUnknown {
		t.Errorf("Risk = %v, want UNKNOWN", rec.Risk)
	}

	// Should have provider limitation
	hasProviderLimitation := false
	for _, lim := range rec.Limitations {
		if lim.Code == "PROVIDER_UNAVAILABLE" {
			hasProviderLimitation = true
			break
		}
	}
	if !hasProviderLimitation {
		t.Error("expected PROVIDER_UNAVAILABLE limitation")
	}
}

func TestEngine_Generate_APIBlocker(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		APIFindings: []kubent.Finding{
			{
				Status:     kubent.FindingFail,
				APIVersion: "policy/v1beta1",
				Kind:       "PodSecurityPolicy",
				RemovedIn:  "1.25",
				Resource: kubent.ResourceRef{
					Kind:      "PodSecurityPolicy",
					Namespace: "",
					Name:      "restricted",
				},
			},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.31.0", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Readiness != ReadinessNotReady {
		t.Errorf("Readiness = %v, want NOT_READY", rec.Readiness)
	}

	// Should have API blocker finding
	hasAPIBlocker := false
	for _, f := range rec.Findings {
		if f.Category == CategoryAPI && f.Severity == SeverityBlocker {
			hasAPIBlocker = true
			break
		}
	}
	if !hasAPIBlocker {
		t.Error("expected API blocker finding")
	}
}

func TestEngine_Generate_SequentialPath(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.4",
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.4",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.30.6", IsPreview: false},
				{Version: "1.31.2", IsPreview: false},
				{Version: "1.32.1", IsPreview: false},
				{Version: "1.33.0", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Verify sequential path: 1.30.4 -> 1.31.2 -> 1.32.1 -> 1.33.0
	if len(rec.Path) != 3 {
		t.Fatalf("Path length = %d, want 3", len(rec.Path))
	}

	expectedPath := []struct {
		from string
		to   string
	}{
		{"1.30.4", "1.31.2"},
		{"1.31.2", "1.32.1"},
		{"1.32.1", "1.33.0"},
	}

	for i, exp := range expectedPath {
		if rec.Path[i].From != exp.from {
			t.Errorf("Path[%d].From = %v, want %v", i, rec.Path[i].From, exp.from)
		}
		if rec.Path[i].To != exp.to {
			t.Errorf("Path[%d].To = %v, want %v", i, rec.Path[i].To, exp.to)
		}
		if !rec.Path[i].IsProviderValid {
			t.Errorf("Path[%d] should be provider valid", i)
		}
	}
}

func TestEngine_Generate_ProviderDirectMultiMinorPath(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		APIFindings: []kubent.Finding{
			{Status: kubent.FindingPass, TargetVersion: "1.34.9"},
		},
		ComponentDetections: []components.Detection{
			{ComponentID: "nginx-ingress", Name: "NGINX Ingress", Version: "1.12.1", Status: components.StatusFound},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.34.9", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if rec.Destination != "1.34.9" {
		t.Fatalf("Destination = %s, want 1.34.9", rec.Destination)
	}
	if len(rec.Path) != 1 || rec.Path[0].From != "1.30.0" || rec.Path[0].To != "1.34.9" {
		t.Fatalf("Path = %#v, want direct 1.30.0 to 1.34.9", rec.Path)
	}
	if rec.Readiness != ReadinessReadyWithWarnings {
		t.Fatalf("Readiness = %s, want READY_WITH_WARNINGS", rec.Readiness)
	}
	if rec.Risk != RiskMedium {
		t.Fatalf("Risk = %s, want MEDIUM", rec.Risk)
	}
	if len(rec.Findings) == 0 || rec.Findings[0].ID != "PROVIDER_DIRECT_MULTI_MINOR" {
		t.Fatalf("Findings = %#v, want provider direct warning", rec.Findings)
	}
}

func TestEngine_Generate_ExplicitTarget(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.30.0",
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.31.0", IsPreview: false},
				{Version: "1.32.0", IsPreview: false},
				{Version: "1.33.0", IsPreview: false},
			},
		},
	}

	opts := DefaultOptions()
	opts.TargetVersion = "1.32.0"

	rec, err := engine.Generate(input, opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Destination != "1.32.0" {
		t.Errorf("Destination = %v, want 1.32.0", rec.Destination)
	}
	if len(rec.Path) != 2 {
		t.Errorf("Path length = %d, want 2", len(rec.Path))
	}
}

func TestEngine_Generate_NoUpgradesAvailable(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	input := Input{
		CurrentVersion: "1.33.0",
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion:    "1.33.0",
			AvailableUpgrades: []provider.UpgradeOption{}, // No upgrades
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Empty provider candidates should not produce an internal path-build error.
	for _, lim := range rec.Limitations {
		if lim.Code == "PATH_BUILD_FAILED" {
			t.Fatalf("unexpected PATH_BUILD_FAILED limitation: %#v", rec.Limitations)
		}
	}
	if len(rec.Path) != 0 {
		t.Fatalf("Path length = %d, want 0", len(rec.Path))
	}
}

func TestPolicy_EvaluateReadiness(t *testing.T) {
	policy := NewPolicy()

	tests := []struct {
		name        string
		findings    []Finding
		limitations []Limitation
		want        ReadinessState
	}{
		{
			name:     "no findings",
			findings: []Finding{},
			want:     ReadinessReady,
		},
		{
			name: "info only",
			findings: []Finding{
				{Severity: SeverityInfo},
			},
			want: ReadinessReady,
		},
		{
			name: "warning",
			findings: []Finding{
				{Severity: SeverityWarning},
			},
			want: ReadinessReadyWithWarnings,
		},
		{
			name: "blocker",
			findings: []Finding{
				{Severity: SeverityBlocker},
			},
			want: ReadinessNotReady,
		},
		{
			name: "unknown only",
			findings: []Finding{
				{Severity: SeverityUnknown},
			},
			want: ReadinessInconclusive,
		},
		{
			name:     "critical limitation",
			findings: []Finding{},
			limitations: []Limitation{
				{Code: "TARGET_COVERAGE_UNVERIFIED"},
			},
			want: ReadinessInconclusive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.EvaluateReadiness(tt.findings, tt.limitations)
			if got != tt.want {
				t.Errorf("EvaluateReadiness() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolicy_EvaluateRisk(t *testing.T) {
	policy := NewPolicy()

	tests := []struct {
		name        string
		findings    []Finding
		limitations []Limitation
		want        RiskLevel
	}{
		{
			name:     "no findings",
			findings: []Finding{},
			want:     RiskLow,
		},
		{
			name: "warning",
			findings: []Finding{
				{Severity: SeverityWarning},
			},
			want: RiskMedium,
		},
		{
			name: "blocker",
			findings: []Finding{
				{Severity: SeverityBlocker},
			},
			want: RiskHigh,
		},
		{
			name: "unknown",
			findings: []Finding{
				{Severity: SeverityUnknown},
			},
			want: RiskUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.EvaluateRisk(tt.findings, tt.limitations)
			if got != tt.want {
				t.Errorf("EvaluateRisk() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAKS130ValidationCase validates the AKS 1.30 -> 1.33 case from docs/recommendation-model.md
func TestAKS130ValidationCase(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	// Simulate the validation case from docs
	input := Input{
		CurrentVersion: "1.30.0",
		HealthFindings: []health.Finding{
			{RuleID: "NODE_READY", Status: health.StatusPass},
			{RuleID: "WORKLOAD_AVAILABLE", Status: health.StatusPass},
		},
		APIFindings: []kubent.Finding{
			{Status: kubent.FindingPass, TargetVersion: "1.33"},
		},
		ComponentDetections: []components.Detection{
			{ComponentID: "nginx-ingress", Name: "NGINX Ingress", Version: "1.12.1", Status: components.StatusFound, Confidence: components.ConfidenceHigh},
			{ComponentID: "emqx", Name: "EMQX", Version: "5.8.8", Status: components.StatusFound, Confidence: components.ConfidenceHigh},
			{ComponentID: "fluent-bit", Name: "Fluent Bit", Version: "4.0.3", Status: components.StatusFound, Confidence: components.ConfidenceHigh},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion: "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{
				{Version: "1.30.6", IsPreview: false},
				{Version: "1.31.4", IsPreview: false},
				{Version: "1.32.2", IsPreview: false},
				{Version: "1.33.12", IsPreview: false},
			},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Expected per docs/recommendation-model.md:
	// - destination: 1.33.12
	// - path: 1.30.x → 1.31.x → 1.32.x → 1.33.12
	// - readiness: READY
	// - risk: LOW

	if rec.Destination != "1.33.12" {
		t.Errorf("Destination = %v, want 1.33.12", rec.Destination)
	}

	if rec.Readiness != ReadinessReady {
		t.Errorf("Readiness = %v, want READY", rec.Readiness)
	}

	if rec.Risk != RiskLow {
		t.Errorf("Risk = %v, want LOW", rec.Risk)
	}

	// Verify path has 3 stages (1.30->1.31->1.32->1.33)
	if len(rec.Path) != 3 {
		t.Fatalf("Path length = %d, want 3", len(rec.Path))
	}

	// All stages should be provider valid
	for i, stage := range rec.Path {
		if !stage.IsProviderValid {
			t.Errorf("Path[%d] should be provider valid", i)
		}
	}

	t.Logf("AKS 1.30 validation case passed: %s -> %s, Readiness=%s, Risk=%s",
		rec.CurrentVersion, rec.Destination, rec.Readiness, rec.Risk)
}
