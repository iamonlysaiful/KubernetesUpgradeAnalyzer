package health

import (
	"reflect"
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

func TestNodeReadinessRuleFindsNotReadyAndUnknownEvidence(t *testing.T) {
	rule := NodeReadinessRule{}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Nodes: []inventory.Node{
				node("node-a", "1.30.7", condition("Ready", "TRUE", "KubeletReady")),
				node("node-b", "1.30.7", condition("Ready", "FALSE", "KubeletNotReady")),
				node("node-c", "1.30.7"),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{})

	got := findingKeys(findings)
	want := []string{
		"health.node.notReady|BLOCKER|FAIL|Node|node-b|node node-b is not Ready",
		"health.node.notReady|INFO|UNKNOWN|Node|node-c|node node-c has no Ready condition evidence",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestNodePressureRuleFindsTruePressureConditions(t *testing.T) {
	rule := NodePressureRule{}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Nodes: []inventory.Node{
				node("node-a", "1.30.7",
					condition("Ready", "TRUE", ""),
					condition("MemoryPressure", "TRUE", "LowMemory"),
					condition("DiskPressure", "FALSE", "SufficientDisk"),
					condition("PIDPressure", "TRUE", "TooManyProcesses"),
				),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{})
	SortFindings(findings)

	got := findingKeys(findings)
	want := []string{
		"health.node.pressure|WARNING|WARN|Node|node-a|node node-a reports MemoryPressure",
		"health.node.pressure|WARNING|WARN|Node|node-a|node node-a reports PIDPressure",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestNodeKubeletSkewRuleFindsMinorSkewAndUnknownEvidence(t *testing.T) {
	rule := NodeKubeletSkewRule{}
	snapshot := inventory.Snapshot{
		Kubernetes: inventory.Kubernetes{ServerVersion: "1.30.7"},
		Inventory: inventory.Inventory{
			Nodes: []inventory.Node{
				node("node-a", "v1.30.9"),
				node("node-b", "v1.31.1"),
				node("node-c", ""),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{})
	SortFindings(findings)

	got := findingKeys(findings)
	want := []string{
		"health.node.kubeletSkew|WARNING|WARN|Node|node-b|node node-b kubelet minor differs from server minor",
		"health.node.kubeletSkew|INFO|UNKNOWN|Node|node-c|node node-c kubelet version evidence is unavailable",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestNodeKubeletSkewRuleUnknownWhenServerVersionMissing(t *testing.T) {
	rule := NodeKubeletSkewRule{}
	snapshot := inventory.Snapshot{
		Cluster: inventory.Cluster{Identity: inventory.ResourceRef{Kind: "Cluster", Name: "cluster-a"}},
		Inventory: inventory.Inventory{
			Nodes: []inventory.Node{node("node-a", "v1.30.9")},
		},
	}

	findings := rule.Evaluate(snapshot, Options{})

	got := findingKeys(findings)
	want := []string{
		"health.node.kubeletSkew|INFO|UNKNOWN|Cluster|cluster-a|cluster server version evidence is unavailable for kubelet skew analysis",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestWorkloadUnavailableRuleFindsUnavailableWorkloads(t *testing.T) {
	rule := WorkloadUnavailableRule{}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Workloads: []inventory.Workload{
				workload("Deployment", "default", "api", 3, 2),
				workload("StatefulSet", "default", "db", 2, 2),
				workload("DaemonSet", "kube-system", "agent", 5, 4),
			},
		},
	}

	findings := rule.Evaluate(snapshot, Options{})
	SortFindings(findings)

	got := findingKeys(findings)
	want := []string{
		"health.workload.unavailable|WARNING|WARN|Deployment|api|workload default/api has fewer ready replicas than desired",
		"health.workload.unavailable|WARNING|WARN|DaemonSet|agent|workload kube-system/agent has fewer ready replicas than desired",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestNodeAndWorkloadRulesRunTogether(t *testing.T) {
	runner := NewRunner(NodeAndWorkloadRules()...)
	snapshot := inventory.Snapshot{
		Kubernetes: inventory.Kubernetes{ServerVersion: "1.30.7"},
		Inventory: inventory.Inventory{
			Nodes: []inventory.Node{
				node("node-b", "v1.31.1", condition("Ready", "FALSE", "KubeletNotReady")),
			},
			Workloads: []inventory.Workload{
				workload("Deployment", "default", "api", 3, 2),
			},
		},
	}

	findings := runner.Evaluate(snapshot, Options{})

	got := findingKeys(findings)
	want := []string{
		"health.node.notReady|BLOCKER|FAIL|Node|node-b|node node-b is not Ready",
		"health.node.kubeletSkew|WARNING|WARN|Node|node-b|node node-b kubelet minor differs from server minor",
		"health.workload.unavailable|WARNING|WARN|Deployment|api|workload default/api has fewer ready replicas than desired",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func node(name string, kubeletVersion string, conditions ...inventory.Condition) inventory.Node {
	return inventory.Node{
		Ref:            inventory.ResourceRef{Kind: "Node", Name: name},
		KubeletVersion: kubeletVersion,
		Conditions:     conditions,
	}
}

func condition(conditionType string, status string, reason string) inventory.Condition {
	return inventory.Condition{Type: conditionType, Status: status, Reason: reason}
}

func workload(kind string, namespace string, name string, desired int, ready int) inventory.Workload {
	return inventory.Workload{
		Ref: inventory.ResourceRef{
			APIVersion: "apps/v1",
			Kind:       kind,
			Namespace:  namespace,
			Name:       name,
		},
		DesiredReplicas: desired,
		ReadyReplicas:   ready,
		Containers: []inventory.Container{
			{Name: "app", Image: "example/app:1.0"},
		},
	}
}

func findingKeys(findings []Finding) []string {
	result := make([]string, 0, len(findings))
	for _, finding := range findings {
		result = append(result, finding.RuleID+"|"+string(finding.Severity)+"|"+string(finding.Status)+"|"+finding.Resource.Kind+"|"+finding.Resource.Name+"|"+finding.Summary)
	}
	return result
}
