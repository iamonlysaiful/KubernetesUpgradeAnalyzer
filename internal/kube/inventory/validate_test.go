package inventory

import (
	"strings"
	"testing"
)

func TestValidateCoreSnapshotAcceptsGoldenSnapshot(t *testing.T) {
	if err := ValidateCoreSnapshot(collectGoldenSnapshot(t)); err != nil {
		t.Fatalf("ValidateCoreSnapshot(golden) returned error: %v", err)
	}
}

func TestValidateCoreSnapshotRejectsInvalidSnapshot(t *testing.T) {
	snapshot := collectGoldenSnapshot(t)
	snapshot.SchemaVersion = "kua.cluster-snapshot.v2"
	snapshot.CapturedAt = "not-a-time"
	snapshot.Kubernetes.ServerVersion = "1.29.0"
	snapshot.Inventory.Nodes[0].Conditions = nil
	snapshot.Limitations[0].Severity = "warning"

	err := ValidateCoreSnapshot(snapshot)
	if err == nil {
		t.Fatalf("ValidateCoreSnapshot(invalid) returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"schemaVersion",
		"capturedAt",
		"kubernetes.serverVersion",
		"conditions",
		"severity",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("ValidateCoreSnapshot error missing %q in: %s", want, message)
		}
	}
}

func TestValidateCoreSnapshotRejectsInvalidWorkload(t *testing.T) {
	snapshot := collectWorkloadGoldenSnapshot(t)
	snapshot.Inventory.Workloads[0].Ref.APIVersion = ""
	snapshot.Inventory.Workloads[0].Critical = "PASS"
	snapshot.Inventory.Workloads[0].DesiredReplicas = -1
	snapshot.Inventory.Workloads[0].Containers[0].Image = ""

	err := ValidateCoreSnapshot(snapshot)
	if err == nil {
		t.Fatalf("ValidateCoreSnapshot(invalid workload) returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"apiVersion",
		"critical",
		"desiredReplicas",
		"image",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("ValidateCoreSnapshot workload error missing %q in: %s", want, message)
		}
	}
}

func TestValidateCoreSnapshotRejectsInvalidCRD(t *testing.T) {
	snapshot := collectCRDGoldenSnapshot(t)
	snapshot.Inventory.CRDs[0].APIVersion = "v1"
	snapshot.Inventory.CRDs[0].Kind = "ConfigMap"
	snapshot.Inventory.CRDs[0].Namespace = "team-a"

	err := ValidateCoreSnapshot(snapshot)
	if err == nil {
		t.Fatalf("ValidateCoreSnapshot(invalid CRD) returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"apiVersion",
		"kind",
		"namespace",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("ValidateCoreSnapshot CRD error missing %q in: %s", want, message)
		}
	}
}
