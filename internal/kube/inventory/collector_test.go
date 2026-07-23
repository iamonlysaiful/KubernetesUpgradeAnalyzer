package inventory

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
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
	if err := ValidateCoreSnapshot(snapshot); err != nil {
		t.Fatalf("ValidateCoreSnapshot returned error: %v", err)
	}

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

func TestCollectSnapshotWithWorkloadsMatchesGoldenFixture(t *testing.T) {
	snapshot := collectWorkloadGoldenSnapshot(t)
	if err := ValidateCoreSnapshot(snapshot); err != nil {
		t.Fatalf("ValidateCoreSnapshot returned error: %v", err)
	}

	got, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent returned error: %v", err)
	}
	got = append(got, '\n')

	want, err := os.ReadFile("../../../schemas/fixtures/cluster-snapshot/valid/p2-03-workload-inventory.json")
	if err != nil {
		t.Fatalf("ReadFile golden fixture returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Generated workload snapshot does not match golden fixture.\nGot:\n%s\nWant:\n%s", string(got), string(want))
	}
}

func TestCollectSnapshotWithWorkloadsAndCRDsMatchesGoldenFixture(t *testing.T) {
	snapshot := collectCRDGoldenSnapshot(t)
	if err := ValidateCoreSnapshot(snapshot); err != nil {
		t.Fatalf("ValidateCoreSnapshot returned error: %v", err)
	}

	got, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent returned error: %v", err)
	}
	got = append(got, '\n')

	want, err := os.ReadFile("../../../schemas/fixtures/cluster-snapshot/valid/p2-03-crd-inventory.json")
	if err != nil {
		t.Fatalf("ReadFile golden fixture returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Generated CRD snapshot does not match golden fixture.\nGot:\n%s\nWant:\n%s", string(got), string(want))
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

func collectWorkloadGoldenSnapshot(t *testing.T) Snapshot {
	t.Helper()

	replicas := int32(3)
	parallelism := int32(2)
	suspended := true

	client := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-a"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-b"}},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-a"},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.30.7"},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-b"},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Template: podTemplate(
					container("worker", "registry-001/app-worker:2.1.0"),
					container("api", "registry-001/app-api:1.2.3"),
				),
			},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 2},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "agent", Namespace: "team-a"},
			Spec: appsv1.DaemonSetSpec{
				Template: podTemplate(container("agent", "registry-001/agent@sha256:abc123")),
			},
			Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 4, NumberReady: 3},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "team-a"},
			Spec: appsv1.StatefulSetSpec{
				Template: podTemplate(container("db", "registry-001/db:12")),
			},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 1},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Name: "legacy", Namespace: "team-a"},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
				Template: podTemplate(container("legacy", "registry-001/legacy")),
			},
			Status: appsv1.ReplicaSetStatus{ReadyReplicas: 3},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "migrate", Namespace: "team-b"},
			Spec: batchv1.JobSpec{
				Parallelism: &parallelism,
				Template:    podTemplate(container("migrate", "registry-001/migrate:2026.07")),
			},
			Status: batchv1.JobStatus{Succeeded: 1},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{Name: "nightly", Namespace: "team-b"},
			Spec: batchv1.CronJobSpec{
				Suspend:     &suspended,
				Schedule:    "0 0 * * *",
				JobTemplate: batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: podTemplate(container("nightly", "registry-001/nightly:1.0.0"))}},
			},
		},
	)

	collector := NewCollector(client)
	collector.Clock = func() time.Time {
		return time.Date(2026, 7, 23, 4, 5, 6, 0, time.UTC)
	}

	snapshot, err := collector.CollectSnapshotWithWorkloads(context.Background(), preflight.Result{
		Context: preflight.ContextSelection{
			Name:             "ctx-workload-golden",
			KubeconfigSource: preflight.KubeconfigSourceDefault,
		},
		ServerVersion: "v1.30.7",
	})
	if err != nil {
		t.Fatalf("CollectSnapshotWithWorkloads returned error: %v", err)
	}
	return snapshot
}

func collectCRDGoldenSnapshot(t *testing.T) Snapshot {
	t.Helper()

	collector := NewCollectorWithAPIExtensions(
		workloadGoldenClient(),
		apiextensionsfake.NewSimpleClientset(
			&apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "alerts.example.test"},
			},
			&apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "widgets.example.test"},
			},
		),
	)
	collector.Clock = func() time.Time {
		return time.Date(2026, 7, 23, 5, 6, 7, 0, time.UTC)
	}

	snapshot, err := collector.CollectSnapshotWithWorkloadsAndCRDs(context.Background(), preflight.Result{
		Context: preflight.ContextSelection{
			Name:             "ctx-crd-golden",
			KubeconfigSource: preflight.KubeconfigSourceDefault,
		},
		ServerVersion: "v1.30.7",
	})
	if err != nil {
		t.Fatalf("CollectSnapshotWithWorkloadsAndCRDs returned error: %v", err)
	}
	return snapshot
}

func workloadGoldenClient() *fake.Clientset {
	replicas := int32(3)
	parallelism := int32(2)
	suspended := true

	return fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-a"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "team-b"}},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-a"},
			Status: corev1.NodeStatus{
				NodeInfo: corev1.NodeSystemInfo{KubeletVersion: "v1.30.7"},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "team-b"},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Template: podTemplate(
					container("worker", "registry-001/app-worker:2.1.0"),
					container("api", "registry-001/app-api:1.2.3"),
				),
			},
			Status: appsv1.DeploymentStatus{ReadyReplicas: 2},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "agent", Namespace: "team-a"},
			Spec: appsv1.DaemonSetSpec{
				Template: podTemplate(container("agent", "registry-001/agent@sha256:abc123")),
			},
			Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 4, NumberReady: 3},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "team-a"},
			Spec: appsv1.StatefulSetSpec{
				Template: podTemplate(container("db", "registry-001/db:12")),
			},
			Status: appsv1.StatefulSetStatus{ReadyReplicas: 1},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Name: "legacy", Namespace: "team-a"},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
				Template: podTemplate(container("legacy", "registry-001/legacy")),
			},
			Status: appsv1.ReplicaSetStatus{ReadyReplicas: 3},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "migrate", Namespace: "team-b"},
			Spec: batchv1.JobSpec{
				Parallelism: &parallelism,
				Template:    podTemplate(container("migrate", "registry-001/migrate:2026.07")),
			},
			Status: batchv1.JobStatus{Succeeded: 1},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{Name: "nightly", Namespace: "team-b"},
			Spec: batchv1.CronJobSpec{
				Suspend:     &suspended,
				Schedule:    "0 0 * * *",
				JobTemplate: batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: podTemplate(container("nightly", "registry-001/nightly:1.0.0"))}},
			},
		},
	)
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
