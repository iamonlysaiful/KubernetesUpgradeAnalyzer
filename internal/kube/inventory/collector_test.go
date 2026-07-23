package inventory

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/preflight"
)

func TestCollectCoreBuildsPartialSnapshot(t *testing.T) {
	snapshot := collectGoldenSnapshot(t)

	if snapshot.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %q, want %q", snapshot.SchemaVersion, SchemaVersion)
	}
	if snapshot.SnapshotID != "ctx-golden-20260723T010203Z" {
		t.Fatalf("SnapshotID = %q, want ctx-golden-20260723T010203Z", snapshot.SnapshotID)
	}
	if snapshot.CapturedAt != "2026-07-23T01:02:03Z" {
		t.Fatalf("CapturedAt = %q, want 2026-07-23T01:02:03Z", snapshot.CapturedAt)
	}
	if snapshot.Cluster.Context.Name != "ctx-golden" {
		t.Fatalf("Context name = %q, want ctx-golden", snapshot.Cluster.Context.Name)
	}
	if snapshot.Cluster.Context.KubeconfigSource != "DEFAULT" {
		t.Fatalf("KubeconfigSource = %q, want DEFAULT", snapshot.Cluster.Context.KubeconfigSource)
	}
	if snapshot.Kubernetes.ServerVersion != "1.30.7" {
		t.Fatalf("ServerVersion = %q, want 1.30.7", snapshot.Kubernetes.ServerVersion)
	}

	if got := names(snapshot.Inventory.Namespaces); got != "alpha,zeta" {
		t.Fatalf("Namespaces = %q, want alpha,zeta", got)
	}
	if got := nodeNames(snapshot.Inventory.Nodes); got != "node-a,node-b" {
		t.Fatalf("Nodes = %q, want node-a,node-b", got)
	}
	if snapshot.Inventory.Nodes[1].ProviderIDPresent != true {
		t.Fatalf("ProviderIDPresent for node-b = false, want true")
	}
	if snapshot.Inventory.Nodes[1].NodePool != "pool-b" {
		t.Fatalf("NodePool for node-b = %q, want pool-b", snapshot.Inventory.Nodes[1].NodePool)
	}
	if len(snapshot.Inventory.Nodes[1].Conditions) != 2 {
		t.Fatalf("Node-b conditions = %d, want 2", len(snapshot.Inventory.Nodes[1].Conditions))
	}
	if snapshot.Inventory.Nodes[1].Conditions[1].Status != "TRUE" {
		t.Fatalf("Ready status = %q, want TRUE", snapshot.Inventory.Nodes[1].Conditions[1].Status)
	}
	if len(snapshot.Inventory.Workloads) != 0 ||
		len(snapshot.Inventory.Storage) != 0 ||
		len(snapshot.Inventory.Networking) != 0 ||
		len(snapshot.Inventory.CRDs) != 0 ||
		len(snapshot.Inventory.Events) != 0 {
		t.Fatalf("Future inventory groups should be empty in P2-02: %#v", snapshot.Inventory)
	}
	if len(snapshot.Limitations) != 1 || snapshot.Limitations[0].Code != "PARTIAL_INVENTORY_P2_02" {
		t.Fatalf("Limitations = %#v, want PARTIAL_INVENTORY_P2_02", snapshot.Limitations)
	}
}

func TestCollectCoreMatchesGoldenFixture(t *testing.T) {
	snapshot := collectGoldenSnapshot(t)
	got, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent returned error: %v", err)
	}
	got = append(got, '\n')

	want, err := os.ReadFile("../../../schemas/fixtures/cluster-snapshot/valid/p2-02-core-inventory.json")
	if err != nil {
		t.Fatalf("ReadFile golden fixture returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Generated snapshot does not match golden fixture.\nGot:\n%s\nWant:\n%s", string(got), string(want))
	}
}

func collectGoldenSnapshot(t *testing.T) Snapshot {
	t.Helper()

	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "zeta"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "alpha"}},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-b",
				Labels: map[string]string{
					"kubernetes.azure.com/agentpool": "pool-b",
				},
			},
			Spec: corev1.NodeSpec{ProviderID: "azure:///redacted-in-test"},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.30.7"},
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue, Reason: "KubeletReady"},
					{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionFalse, Reason: "SufficientMemory"},
				},
			},
		},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-a"},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.30.6"},
			},
		},
	)

	collector := NewCollector(client)
	collector.Clock = func() time.Time {
		return time.Date(2026, 7, 23, 1, 2, 3, 0, time.UTC)
	}

	snapshot, err := collector.CollectCore(context.Background(), preflight.Result{
		Context: preflight.ContextSelection{
			Name:             "ctx-golden",
			KubeconfigSource: preflight.KubeconfigSourceDefault,
		},
		ServerVersion: "v1.30.7",
	})
	if err != nil {
		t.Fatalf("CollectCore returned error: %v", err)
	}
	return snapshot
}

func TestCollectCoreRequiresClient(t *testing.T) {
	_, err := NewCollector(nil).CollectCore(context.Background(), preflight.Result{})
	if err == nil {
		t.Fatalf("CollectCore with nil client returned nil error")
	}
}

func names(refs []ResourceRef) string {
	var result string
	for i, ref := range refs {
		if i > 0 {
			result += ","
		}
		result += ref.Name
	}
	return result
}

func nodeNames(nodes []Node) string {
	var result string
	for i, node := range nodes {
		if i > 0 {
			result += ","
		}
		result += node.Ref.Name
	}
	return result
}
