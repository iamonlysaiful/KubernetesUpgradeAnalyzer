package health

import (
	"reflect"
	"testing"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

func TestRunnerAppliesDefaultsAndSortsFindings(t *testing.T) {
	fixedNow := time.Date(2026, 7, 23, 1, 2, 3, 0, time.UTC)
	runner := NewRunner(
		RuleFunc{
			RuleID: "test.defaultOptions",
			Run: func(snapshot inventory.Snapshot, options Options) []Finding {
				if got := options.Now(); !got.After(time.Time{}) {
					t.Fatalf("options.Now returned zero time")
				}
				if options.EventLookback != DefaultEventLookback {
					t.Fatalf("EventLookback = %s, want %s", options.EventLookback, DefaultEventLookback)
				}
				return []Finding{
					finding("health.z", SeverityInfo, StatusUnknown, "default", "ConfigMap", "zeta", "last"),
					finding("health.a", SeverityBlocker, StatusFail, "default", "Deployment", "api", "first"),
				}
			},
		},
		RuleFunc{
			RuleID: "test.customOptions",
			Run: func(snapshot inventory.Snapshot, options Options) []Finding {
				return []Finding{
					{
						RuleID:   "health.clock",
						Severity: SeverityWarning,
						Status:   StatusWarn,
						Resource: ResourceRef{Namespace: "default", Kind: "Event", Name: "clock"},
						Summary:  options.Now().Format(time.RFC3339),
					},
				}
			},
		},
		nil,
	)

	findings := runner.Evaluate(inventory.Snapshot{}, Options{Now: func() time.Time { return fixedNow }})

	got := summaries(findings)
	want := []string{"first", "2026-07-23T01:02:03Z", "last"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summaries = %#v, want %#v", got, want)
	}
}

func TestRunnerUsesInjectedClockAndEventLookback(t *testing.T) {
	fixedNow := time.Date(2026, 7, 23, 4, 5, 6, 0, time.UTC)
	customLookback := 45 * time.Minute
	runner := NewRunner(RuleFunc{
		RuleID: "test.options",
		Run: func(snapshot inventory.Snapshot, options Options) []Finding {
			if got := options.Now(); !got.Equal(fixedNow) {
				t.Fatalf("Now = %s, want %s", got, fixedNow)
			}
			if options.EventLookback != customLookback {
				t.Fatalf("EventLookback = %s, want %s", options.EventLookback, customLookback)
			}
			return nil
		},
	})

	runner.Evaluate(inventory.Snapshot{}, Options{
		Now:           func() time.Time { return fixedNow },
		EventLookback: customLookback,
	})
}

func TestSortFindingsUsesContractOrder(t *testing.T) {
	findings := []Finding{
		finding("health.workload.unavailable", SeverityWarning, StatusWarn, "zeta", "Deployment", "api", "namespace later"),
		finding("health.workload.unavailable", SeverityWarning, StatusWarn, "alpha", "StatefulSet", "db", "kind later"),
		finding("health.event.unknownType", SeverityInfo, StatusUnknown, "alpha", "Event", "unknown", "info later"),
		finding("health.node.notReady", SeverityBlocker, StatusFail, "", "Node", "node-b", "blocker first"),
		finding("health.workload.unavailable", SeverityWarning, StatusWarn, "alpha", "Deployment", "web", "name later"),
		finding("health.workload.unavailable", SeverityWarning, StatusWarn, "alpha", "Deployment", "api", "summary b"),
		finding("health.workload.unavailable", SeverityWarning, StatusWarn, "alpha", "Deployment", "api", "summary a"),
	}

	SortFindings(findings)

	got := summaries(findings)
	want := []string{
		"blocker first",
		"summary a",
		"summary b",
		"name later",
		"kind later",
		"namespace later",
		"info later",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("summaries = %#v, want %#v", got, want)
	}
}

func TestResourceFromInventoryDropsUIDAlias(t *testing.T) {
	got := ResourceFromInventory(inventory.ResourceRef{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "default",
		Name:       "api",
		UIDAlias:   "uid-redacted",
	})
	want := ResourceRef{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespace:  "default",
		Name:       "api",
	}
	if got != want {
		t.Fatalf("ResourceFromInventory = %#v, want %#v", got, want)
	}
}

func finding(ruleID string, severity Severity, status Status, namespace string, kind string, name string, summary string) Finding {
	return Finding{
		RuleID:   ruleID,
		Severity: severity,
		Status:   status,
		Resource: ResourceRef{Namespace: namespace, Kind: kind, Name: name},
		Summary:  summary,
	}
}

func summaries(findings []Finding) []string {
	result := make([]string, 0, len(findings))
	for _, finding := range findings {
		result = append(result, finding.Summary)
	}
	return result
}
