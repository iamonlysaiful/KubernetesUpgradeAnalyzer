package recommendation

import (
	"testing"
)

func TestHealthFindingEnhancement(t *testing.T) {
	tests := []struct {
		ruleID        string
		wantImpact    ImpactLevel
		wantHasAction bool
		wantIfIgnored bool
	}{
		{"health.node.notReady", ImpactHigh, true, true},
		{"health.node.pressure", ImpactMedium, true, true},
		{"health.node.kubeletSkew", ImpactLow, true, true},
		{"health.workload.unavailable", ImpactMedium, true, true},
		{"health.pvc.unbound", ImpactHigh, true, true},
		{"health.event.warning", ImpactLow, true, true},
		{"unknown.rule", ImpactMedium, true, true}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			impact, action, ifIgnored := healthFindingEnhancement(tt.ruleID)

			if impact == nil {
				t.Error("impact should not be nil")
				return
			}
			if impact.Level != tt.wantImpact {
				t.Errorf("impact.Level = %q, want %q", impact.Level, tt.wantImpact)
			}
			if tt.wantHasAction && action == nil {
				t.Error("action should not be nil")
			}
			if action != nil && action.Description == "" {
				t.Error("action.Description should not be empty")
			}
			if tt.wantIfIgnored && ifIgnored == "" {
				t.Error("ifIgnored should not be empty")
			}
		})
	}
}

func TestAPIFindingEnhancement(t *testing.T) {
	tests := []struct {
		name        string
		isRemoved   bool
		replacement string
		removedIn   string
		wantImpact  ImpactLevel
	}{
		{
			name:        "removed API",
			isRemoved:   true,
			replacement: "apps/v1",
			removedIn:   "1.25",
			wantImpact:  ImpactHigh,
		},
		{
			name:        "deprecated API",
			isRemoved:   false,
			replacement: "networking.k8s.io/v1",
			removedIn:   "1.26",
			wantImpact:  ImpactMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact, action, ifIgnored := apiFindingEnhancement(tt.isRemoved, tt.replacement, tt.removedIn)

			if impact == nil {
				t.Error("impact should not be nil")
				return
			}
			if impact.Level != tt.wantImpact {
				t.Errorf("impact.Level = %q, want %q", impact.Level, tt.wantImpact)
			}
			if action == nil {
				t.Error("action should not be nil")
			}
			if action != nil && action.Description == "" {
				t.Error("action.Description should not be empty")
			}
			if ifIgnored == "" {
				t.Error("ifIgnored should not be empty")
			}
		})
	}
}

func TestComponentFindingEnhancement(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		isUnknown     bool
		isPartial     bool
		wantImpact    ImpactLevel
	}{
		{"unknown version", "EMQX", true, false, ImpactMedium},
		{"partial version", "CoreDNS", false, true, ImpactLow},
		{"known version", "NGINX", false, false, ImpactLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact, action, ifIgnored := componentFindingEnhancement(tt.componentName, tt.isUnknown, tt.isPartial)

			if impact == nil {
				t.Error("impact should not be nil")
				return
			}
			if impact.Level != tt.wantImpact {
				t.Errorf("impact.Level = %q, want %q", impact.Level, tt.wantImpact)
			}
			if action == nil {
				t.Error("action should not be nil")
			}
			if ifIgnored == "" {
				t.Error("ifIgnored should not be empty")
			}
		})
	}
}

func TestProviderFindingEnhancement(t *testing.T) {
	tests := []struct {
		code       string
		wantImpact ImpactLevel
	}{
		{"PROVIDER_UNAVAILABLE", ImpactMedium},
		{"FILE_EVIDENCE_SOURCE", ImpactLow},
		{"UNKNOWN_CODE", ImpactMedium},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			impact, action, ifIgnored := providerFindingEnhancement(tt.code)

			if impact == nil {
				t.Error("impact should not be nil")
				return
			}
			if impact.Level != tt.wantImpact {
				t.Errorf("impact.Level = %q, want %q", impact.Level, tt.wantImpact)
			}
			if action == nil {
				t.Error("action should not be nil")
			}
			if ifIgnored == "" {
				t.Error("ifIgnored should not be empty")
			}
		})
	}
}

func TestEnhanceHealthFinding(t *testing.T) {
	f := &Finding{
		ID:       "health.node.notReady",
		Category: CategoryHealth,
		Severity: SeverityBlocker,
		Summary:  "Node not ready",
	}

	EnhanceHealthFinding(f)

	if f.Impact == nil {
		t.Error("Impact should be set")
	}
	if f.Action == nil {
		t.Error("Action should be set")
	}
	if f.IfIgnored == "" {
		t.Error("IfIgnored should be set")
	}
	if f.Impact.Level != ImpactHigh {
		t.Errorf("Impact.Level = %q, want HIGH", f.Impact.Level)
	}
}

func TestEnhanceAPIFinding(t *testing.T) {
	f := &Finding{
		ID:       "API_batch/v1beta1_CronJob",
		Category: CategoryAPI,
		Severity: SeverityBlocker,
		Summary:  "API removed",
	}

	EnhanceAPIFinding(f, true, "batch/v1", "1.25")

	if f.Impact == nil {
		t.Error("Impact should be set")
	}
	if f.Action == nil {
		t.Error("Action should be set")
	}
	if f.IfIgnored == "" {
		t.Error("IfIgnored should be set")
	}
	if f.Impact.Level != ImpactHigh {
		t.Errorf("Impact.Level = %q, want HIGH", f.Impact.Level)
	}
}

func TestEnhanceComponentFinding(t *testing.T) {
	f := &Finding{
		ID:       "COMPONENT_emqx_UNKNOWN",
		Category: CategoryComponent,
		Severity: SeverityUnknown,
		Summary:  "EMQX version unknown",
	}

	EnhanceComponentFinding(f, "EMQX", true, false)

	if f.Impact == nil {
		t.Error("Impact should be set")
	}
	if f.Action == nil {
		t.Error("Action should be set")
	}
	if f.IfIgnored == "" {
		t.Error("IfIgnored should be set")
	}
	if f.Impact.Level != ImpactMedium {
		t.Errorf("Impact.Level = %q, want MEDIUM", f.Impact.Level)
	}
}
