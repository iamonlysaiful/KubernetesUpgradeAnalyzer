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
