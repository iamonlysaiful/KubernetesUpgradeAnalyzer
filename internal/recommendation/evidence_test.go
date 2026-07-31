package recommendation

import (
	"testing"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

func evidenceFixedClock() time.Time {
	return time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
}

func TestEvidenceBuilder_Build_EmptyInventory(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	summary := builder.Build(nil, nil, 0)

	if summary.AnalysisTime != evidenceFixedClock() {
		t.Errorf("AnalysisTime = %v, want %v", summary.AnalysisTime, evidenceFixedClock())
	}
	if summary.Nodes.Total != 0 {
		t.Errorf("Nodes.Total = %d, want 0", summary.Nodes.Total)
	}
}

func TestEvidenceBuilder_Build_CountsNodes(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	inv := &inventory.Inventory{
		Nodes: []inventory.Node{
			{
				Ref:        inventory.ResourceRef{Kind: "Node", Name: "node-a"},
				Conditions: []inventory.Condition{{Type: "Ready", Status: "TRUE"}},
			},
			{
				Ref:        inventory.ResourceRef{Kind: "Node", Name: "node-b"},
				Conditions: []inventory.Condition{{Type: "Ready", Status: "FALSE"}},
			},
			{
				Ref:        inventory.ResourceRef{Kind: "Node", Name: "node-c"},
				Conditions: []inventory.Condition{{Type: "Ready", Status: "TRUE"}},
			},
		},
	}

	summary := builder.Build(inv, nil, 0)

	if summary.Nodes.Total != 3 {
		t.Errorf("Nodes.Total = %d, want 3", summary.Nodes.Total)
	}
	if summary.Nodes.Healthy != 2 {
		t.Errorf("Nodes.Healthy = %d, want 2", summary.Nodes.Healthy)
	}
}

func TestEvidenceBuilder_Build_CountsWorkloads(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	inv := &inventory.Inventory{
		Workloads: []inventory.Workload{
			{Ref: inventory.ResourceRef{Kind: "Deployment"}, DesiredReplicas: 3, ReadyReplicas: 3},
			{Ref: inventory.ResourceRef{Kind: "Deployment"}, DesiredReplicas: 2, ReadyReplicas: 1},
			{Ref: inventory.ResourceRef{Kind: "StatefulSet"}, DesiredReplicas: 2, ReadyReplicas: 2},
			{Ref: inventory.ResourceRef{Kind: "DaemonSet"}, DesiredReplicas: 4, ReadyReplicas: 4},
		},
	}

	summary := builder.Build(inv, nil, 0)

	if summary.Deployments.Total != 2 {
		t.Errorf("Deployments.Total = %d, want 2", summary.Deployments.Total)
	}
	if summary.Deployments.Healthy != 1 {
		t.Errorf("Deployments.Healthy = %d, want 1", summary.Deployments.Healthy)
	}
	if summary.StatefulSets.Total != 1 {
		t.Errorf("StatefulSets.Total = %d, want 1", summary.StatefulSets.Total)
	}
	if summary.StatefulSets.Healthy != 1 {
		t.Errorf("StatefulSets.Healthy = %d, want 1", summary.StatefulSets.Healthy)
	}
	if summary.DaemonSets.Total != 1 {
		t.Errorf("DaemonSets.Total = %d, want 1", summary.DaemonSets.Total)
	}
	if summary.DaemonSets.Healthy != 1 {
		t.Errorf("DaemonSets.Healthy = %d, want 1", summary.DaemonSets.Healthy)
	}
}

func TestEvidenceBuilder_Build_CountsPVCs(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	inv := &inventory.Inventory{
		Storage: []inventory.ResourceRef{
			{Kind: "PersistentVolumeClaim", Name: "pvc-1"},
			{Kind: "PersistentVolumeClaim", Name: "pvc-2"},
			{Kind: "PersistentVolume", Name: "pv-1"},
			{Kind: "StorageClass", Name: "default"},
		},
	}

	summary := builder.Build(inv, nil, 0)

	if summary.PVCs.Total != 2 {
		t.Errorf("PVCs.Total = %d, want 2", summary.PVCs.Total)
	}
}

func TestEvidenceBuilder_Build_CountsCRDs(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	inv := &inventory.Inventory{
		CRDs: []inventory.ResourceRef{
			{Kind: "CustomResourceDefinition", Name: "emqxbrokers.emqx.io"},
			{Kind: "CustomResourceDefinition", Name: "ingressroutes.traefik.io"},
		},
	}

	summary := builder.Build(inv, nil, 0)

	if summary.CRDs.Total != 2 {
		t.Errorf("CRDs.Total = %d, want 2", summary.CRDs.Total)
	}
}

func TestEvidenceBuilder_Build_BuildsComponentStatuses(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	detections := []components.Detection{
		{ComponentID: "nginx-ingress", Name: "NGINX Ingress", Version: "1.12.1", Status: components.StatusFound},
		{ComponentID: "coredns", Name: "CoreDNS", Version: "1.11.3", Status: components.StatusFound},
		{ComponentID: "emqx", Name: "EMQX", Version: "UNKNOWN", Status: components.StatusUnknown},
		{ComponentID: "notfound", Name: "Not Found", Status: components.StatusNotFound},
	}

	summary := builder.Build(&inventory.Inventory{}, detections, 0)

	if len(summary.Components) != 3 {
		t.Fatalf("Components count = %d, want 3 (not-found excluded)", len(summary.Components))
	}

	// Check first component
	if summary.Components[0].Name != "NGINX Ingress" {
		t.Errorf("Components[0].Name = %q, want NGINX Ingress", summary.Components[0].Name)
	}
	if summary.Components[0].Version != "1.12.1" {
		t.Errorf("Components[0].Version = %q, want 1.12.1", summary.Components[0].Version)
	}
	if summary.Components[0].Status != "detected" {
		t.Errorf("Components[0].Status = %q, want detected", summary.Components[0].Status)
	}

	// Check unknown component
	if summary.Components[2].Status != "unknown" {
		t.Errorf("Components[2].Status = %q, want unknown", summary.Components[2].Status)
	}
	if summary.Components[2].Version != "unknown" {
		t.Errorf("Components[2].Version = %q, want unknown", summary.Components[2].Version)
	}
}

func TestEvidenceBuilder_Build_DeprecatedAPIs(t *testing.T) {
	builder := NewEvidenceBuilder().WithClock(evidenceFixedClock)

	summary := builder.Build(&inventory.Inventory{}, nil, 5)

	if summary.DeprecatedAPIs != 5 {
		t.Errorf("DeprecatedAPIs = %d, want 5", summary.DeprecatedAPIs)
	}
}

func TestCountSummary_AllHealthy(t *testing.T) {
	tests := []struct {
		name   string
		cs     CountSummary
		expect bool
	}{
		{"all healthy", CountSummary{Total: 5, Healthy: 5}, true},
		{"some unhealthy", CountSummary{Total: 5, Healthy: 3}, false},
		{"zero total", CountSummary{Total: 0, Healthy: 0}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.AllHealthy(); got != tt.expect {
				t.Errorf("AllHealthy() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestCountSummary_String(t *testing.T) {
	tests := []struct {
		name   string
		cs     CountSummary
		expect string
	}{
		{"all healthy", CountSummary{Total: 5, Healthy: 5}, "5/5"},
		{"some unhealthy", CountSummary{Total: 5, Healthy: 3}, "3/5"},
		{"zero total", CountSummary{Total: 0, Healthy: 0}, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.String(); got != tt.expect {
				t.Errorf("String() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestEngine_Generate_PopulatesEvidence(t *testing.T) {
	engine := NewEngine().WithClock(fixedClock)

	inv := &inventory.Inventory{
		Nodes: []inventory.Node{
			{Ref: inventory.ResourceRef{Kind: "Node", Name: "node-a"}, Conditions: []inventory.Condition{{Type: "Ready", Status: "TRUE"}}},
			{Ref: inventory.ResourceRef{Kind: "Node", Name: "node-b"}, Conditions: []inventory.Condition{{Type: "Ready", Status: "TRUE"}}},
		},
		Workloads: []inventory.Workload{
			{Ref: inventory.ResourceRef{Kind: "Deployment", Name: "api"}, DesiredReplicas: 3, ReadyReplicas: 3},
		},
	}

	input := Input{
		CurrentVersion:    "1.30.0",
		InventorySnapshot: inv,
		ComponentDetections: []components.Detection{
			{ComponentID: "nginx-ingress", Name: "NGINX Ingress", Version: "1.12.1", Status: components.StatusFound},
		},
		ProviderEvidence: &provider.ProviderEvidence{
			CurrentVersion:    "1.30.0",
			AvailableUpgrades: []provider.UpgradeOption{{Version: "1.31.0"}},
		},
	}

	rec, err := engine.Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if rec.Evidence == nil {
		t.Fatal("Evidence should be populated")
	}
	if rec.Evidence.Nodes.Total != 2 {
		t.Errorf("Evidence.Nodes.Total = %d, want 2", rec.Evidence.Nodes.Total)
	}
	if rec.Evidence.Nodes.Healthy != 2 {
		t.Errorf("Evidence.Nodes.Healthy = %d, want 2", rec.Evidence.Nodes.Healthy)
	}
	if rec.Evidence.Deployments.Total != 1 {
		t.Errorf("Evidence.Deployments.Total = %d, want 1", rec.Evidence.Deployments.Total)
	}
	if len(rec.Evidence.Components) != 1 {
		t.Errorf("Evidence.Components count = %d, want 1", len(rec.Evidence.Components))
	}
}
