package health

import (
	"reflect"
	"testing"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

func TestStoragePVCUnknownRuleFindsMissingStorageEvidenceWhenWorkloadsExist(t *testing.T) {
	rule := StoragePVCUnknownRule{}
	snapshot := inventory.Snapshot{
		Cluster: inventory.Cluster{Identity: inventory.ResourceRef{Kind: "Cluster", Name: "cluster-a"}},
		Inventory: inventory.Inventory{
			Workloads: []inventory.Workload{
				workload("Deployment", "default", "api", 1, 1),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{})

	got := findingKeys(findings)
	want := []string{
		"health.storage.pvcUnknown|INFO|UNKNOWN|Cluster|cluster-a|storage inventory evidence is unavailable",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestStoragePVCUnknownRulePassesWhenStorageEvidenceExistsOrNoWorkloadsExist(t *testing.T) {
	rule := StoragePVCUnknownRule{}
	withStorage := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Workloads: []inventory.Workload{workload("Deployment", "default", "api", 1, 1)},
			Storage: []inventory.ResourceRef{
				{APIVersion: "v1", Kind: "PersistentVolumeClaim", Namespace: "default", Name: "data"},
			},
		},
	}
	withoutWorkloads := inventory.Snapshot{}

	if findings := rule.Evaluate(withStorage, Options{}); len(findings) != 0 {
		t.Fatalf("withStorage findings = %#v, want none", findings)
	}
	if findings := rule.Evaluate(withoutWorkloads, Options{}); len(findings) != 0 {
		t.Fatalf("withoutWorkloads findings = %#v, want none", findings)
	}
}

func TestEventWarningRuleFindsWarningEventsInsideLookback(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	rule := EventWarningRule{}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Events: []inventory.Event{
				event("WARNING", "FailedScheduling", "Pod", "default", "api", now.Add(-5*time.Minute)),
				event("WARNING", "OldWarning", "Pod", "default", "old", now.Add(-31*time.Minute)),
				event("NORMAL", "Pulled", "Pod", "default", "ok", now.Add(-5*time.Minute)),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{Now: func() time.Time { return now }})

	got := findingKeys(findings)
	want := []string{
		"health.event.warning|WARNING|WARN|Pod|api|warning event observed for Pod api",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestEventUnknownTypeRuleFindsUnknownEventsInsideLookback(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	rule := EventUnknownTypeRule{}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Events: []inventory.Event{
				event("UNKNOWN", "StrangeType", "Deployment", "default", "api", now.Add(-10*time.Minute)),
				event("UNKNOWN", "OldUnknown", "Deployment", "default", "old", now.Add(-40*time.Minute)),
				event("WARNING", "Failed", "Deployment", "default", "warn", now.Add(-10*time.Minute)),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{Now: func() time.Time { return now }})

	got := findingKeys(findings)
	want := []string{
		"health.event.unknownType|INFO|UNKNOWN|Deployment|api|event type is unknown for Deployment api",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestDefaultRulesIncludeAllPhase3Rules(t *testing.T) {
	rules := DefaultRules()

	got := make([]string, 0, len(rules))
	for _, rule := range rules {
		got = append(got, rule.ID())
	}
	want := []string{
		"health.node.notReady",
		"health.node.pressure",
		"health.node.kubeletSkew",
		"health.workload.unavailable",
		"health.storage.pvcUnknown",
		"health.event.warning",
		"health.event.unknownType",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rule IDs = %#v, want %#v", got, want)
	}
}

func event(eventType string, reason string, kind string, namespace string, name string, lastSeenAt time.Time) inventory.Event {
	return inventory.Event{
		Ref: inventory.ResourceRef{
			APIVersion: "v1",
			Kind:       kind,
			Namespace:  namespace,
			Name:       name,
		},
		Type:       eventType,
		Reason:     reason,
		LastSeenAt: lastSeenAt.Format(time.RFC3339),
	}
}
