package plan

import (
	"testing"
	"time"
)

func TestGenerator_Generate_AKS(t *testing.T) {
	gen := NewGenerator()

	input := PlanInput{
		Provider:           "AKS",
		ClusterName:        "my-cluster",
		ResourceGroup:      "my-rg",
		CurrentVersion:     "1.30.0",
		TargetVersion:      "1.31.0",
		NodePoolCount:      3,
		TotalNodeCount:     10,
		HasStatefulSets:    true,
		HasIngress:         true,
		DetectedComponents: []string{"EMQX", "CoreDNS"},
		RiskLevel:          "MEDIUM",
	}

	plan := gen.Generate(input)

	// Verify steps generated
	if len(plan.Steps) < 5 {
		t.Errorf("Steps count = %d, want >= 5", len(plan.Steps))
	}

	// Verify first step is backup
	if plan.Steps[0].Description != "Take AKS backup or etcd snapshot" {
		t.Errorf("First step = %q, want backup step", plan.Steps[0].Description)
	}

	// Verify control plane upgrade step
	found := false
	for _, step := range plan.Steps {
		if step.Command != "" && containsAll(step.Command, "az aks upgrade", "--control-plane-only") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Missing control plane upgrade step")
	}

	// Verify estimated time is reasonable
	if plan.EstimatedTime < 30*time.Minute {
		t.Errorf("EstimatedTime = %v, want >= 30m", plan.EstimatedTime)
	}

	// Verify validation steps
	if len(plan.ValidationSteps) < 5 {
		t.Errorf("ValidationSteps count = %d, want >= 5", len(plan.ValidationSteps))
	}

	// Verify rollback guidance
	if plan.RollbackGuidance == "" {
		t.Error("RollbackGuidance should not be empty")
	}
}

func TestGenerator_Generate_Generic(t *testing.T) {
	gen := NewGenerator()

	input := PlanInput{
		Provider:       "GKE",
		CurrentVersion: "1.30.0",
		TargetVersion:  "1.31.0",
		NodePoolCount:  2,
		TotalNodeCount: 5,
	}

	plan := gen.Generate(input)

	// Verify generic plan has steps
	if len(plan.Steps) < 3 {
		t.Errorf("Steps count = %d, want >= 3", len(plan.Steps))
	}

	// Verify first step is backup
	if plan.Steps[0].Description != "Take cluster backup" {
		t.Errorf("First step = %q, want backup step", plan.Steps[0].Description)
	}
}

func TestGenerator_EstimateTime(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		input   PlanInput
		minTime time.Duration
		maxTime time.Duration
	}{
		{
			name: "small cluster",
			input: PlanInput{
				NodePoolCount:  1,
				TotalNodeCount: 3,
			},
			minTime: 45 * time.Minute,
			maxTime: 60 * time.Minute,
		},
		{
			name: "medium cluster",
			input: PlanInput{
				NodePoolCount:  3,
				TotalNodeCount: 30,
			},
			minTime: 60 * time.Minute,
			maxTime: 90 * time.Minute,
		},
		{
			name: "large cluster",
			input: PlanInput{
				NodePoolCount:  5,
				TotalNodeCount: 100,
			},
			minTime: 80 * time.Minute,
			maxTime: 120 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimate := gen.estimateTime(tt.input)

			if estimate < tt.minTime {
				t.Errorf("EstimateTime = %v, want >= %v", estimate, tt.minTime)
			}
			if estimate > tt.maxTime {
				t.Errorf("EstimateTime = %v, want <= %v", estimate, tt.maxTime)
			}
		})
	}
}

func TestGenerator_GenerateValidationSteps(t *testing.T) {
	gen := NewGenerator()

	input := PlanInput{
		HasStatefulSets:    true,
		HasIngress:         true,
		DetectedComponents: []string{"EMQX"},
	}

	steps := gen.generateValidationSteps(input)

	// Base validations
	if len(steps) < 5 {
		t.Errorf("ValidationSteps count = %d, want >= 5", len(steps))
	}

	// Check for ingress validation
	hasIngress := false
	for _, step := range steps {
		if containsAll(step.Description, "Ingress") {
			hasIngress = true
			break
		}
	}
	if !hasIngress {
		t.Error("Missing Ingress validation step")
	}

	// Check for EMQX validation
	hasEMQX := false
	for _, step := range steps {
		if containsAll(step.Description, "EMQX") {
			hasEMQX = true
			break
		}
	}
	if !hasEMQX {
		t.Error("Missing EMQX validation step")
	}
}

func TestGenerator_ValidationSteps_DeduplicatesComponents(t *testing.T) {
	gen := NewGenerator()

	// Simulate 2 EMQX detections + 4 CoreDNS detections (real cluster scenario)
	input := PlanInput{
		DetectedComponents: []string{"EMQX", "EMQX", "CoreDNS", "CoreDNS", "CoreDNS", "CoreDNS"},
	}

	steps := gen.generateValidationSteps(input)

	emqxCount, corednsCount := 0, 0
	for _, s := range steps {
		if containsAll(s.Description, "EMQX") {
			emqxCount++
		}
		if containsAll(s.Description, "CoreDNS") {
			corednsCount++
		}
	}
	if emqxCount != 1 {
		t.Errorf("EMQX validation steps = %d, want 1", emqxCount)
	}
	if corednsCount != 1 {
		t.Errorf("CoreDNS validation steps = %d, want 1", corednsCount)
	}
}

func TestGenerator_GenerateRollbackGuidance(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		riskLevel  string
		shouldHave string
	}{
		{"HIGH", "ROLLBACK PROCEDURE (HIGH RISK)"},
		{"MEDIUM", "ROLLBACK GUIDANCE (MEDIUM RISK)"},
		{"LOW", "ROLLBACK GUIDANCE"},
	}

	for _, tt := range tests {
		t.Run(tt.riskLevel, func(t *testing.T) {
			input := PlanInput{
				RiskLevel:     tt.riskLevel,
				ResourceGroup: "rg",
				ClusterName:   "cluster",
			}
			guidance := gen.generateRollbackGuidance(input)

			if !containsAll(guidance, tt.shouldHave) {
				t.Errorf("Guidance should contain %q", tt.shouldHave)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Minute, "30m"},
		{45 * time.Minute, "45m"},
		{60 * time.Minute, "1h 0m"},
		{90 * time.Minute, "1h 30m"},
		{120 * time.Minute, "2h 0m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// containsAll returns true if s contains all substrings.
func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
