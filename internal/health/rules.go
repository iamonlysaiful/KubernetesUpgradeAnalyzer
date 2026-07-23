package health

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

const (
	RuleNodeNotReady        = "health.node.notReady"
	RuleNodePressure        = "health.node.pressure"
	RuleNodeKubeletSkew     = "health.node.kubeletSkew"
	RuleWorkloadUnavailable = "health.workload.unavailable"
	RuleStoragePVCUnknown   = "health.storage.pvcUnknown"
	RuleEventWarning        = "health.event.warning"
	RuleEventUnknownType    = "health.event.unknownType"
)

var minorVersionPattern = regexp.MustCompile(`^v?1\.([0-9]+)(?:\.([0-9]+))?`)

func NodeAndWorkloadRules() []Rule {
	return []Rule{
		NodeReadinessRule{},
		NodePressureRule{},
		NodeKubeletSkewRule{},
		WorkloadUnavailableRule{},
	}
}

func StorageAndEventRules() []Rule {
	return []Rule{
		StoragePVCUnknownRule{},
		EventWarningRule{},
		EventUnknownTypeRule{},
	}
}

func DefaultRules() []Rule {
	rules := NodeAndWorkloadRules()
	rules = append(rules, StorageAndEventRules()...)
	return rules
}

type NodeReadinessRule struct{}

func (NodeReadinessRule) ID() string {
	return RuleNodeNotReady
}

func (NodeReadinessRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	var findings []Finding
	for _, node := range snapshot.Inventory.Nodes {
		condition, ok := findCondition(node.Conditions, "Ready")
		if !ok {
			findings = append(findings, Finding{
				RuleID:   RuleNodeNotReady,
				Severity: SeverityInfo,
				Status:   StatusUnknown,
				Resource: ResourceFromInventory(node.Ref),
				Summary:  fmt.Sprintf("node %s has no Ready condition evidence", node.Ref.Name),
				Evidence: map[string]string{"condition": "Ready"},
			})
			continue
		}
		if condition.Status != "TRUE" {
			findings = append(findings, Finding{
				RuleID:   RuleNodeNotReady,
				Severity: SeverityBlocker,
				Status:   StatusFail,
				Resource: ResourceFromInventory(node.Ref),
				Summary:  fmt.Sprintf("node %s is not Ready", node.Ref.Name),
				Evidence: conditionEvidence(condition),
			})
		}
	}
	return findings
}

type NodePressureRule struct{}

func (NodePressureRule) ID() string {
	return RuleNodePressure
}

func (NodePressureRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	var findings []Finding
	for _, node := range snapshot.Inventory.Nodes {
		for _, conditionName := range []string{"MemoryPressure", "DiskPressure", "PIDPressure"} {
			condition, ok := findCondition(node.Conditions, conditionName)
			if !ok || condition.Status != "TRUE" {
				continue
			}
			findings = append(findings, Finding{
				RuleID:   RuleNodePressure,
				Severity: SeverityWarning,
				Status:   StatusWarn,
				Resource: ResourceFromInventory(node.Ref),
				Summary:  fmt.Sprintf("node %s reports %s", node.Ref.Name, conditionName),
				Evidence: conditionEvidence(condition),
			})
		}
	}
	return findings
}

type NodeKubeletSkewRule struct{}

func (NodeKubeletSkewRule) ID() string {
	return RuleNodeKubeletSkew
}

func (NodeKubeletSkewRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	serverMinor, ok := kubernetesMinor(snapshot.Kubernetes.ServerVersion)
	if !ok {
		return []Finding{{
			RuleID:   RuleNodeKubeletSkew,
			Severity: SeverityInfo,
			Status:   StatusUnknown,
			Resource: ResourceFromInventory(snapshot.Cluster.Identity),
			Summary:  "cluster server version evidence is unavailable for kubelet skew analysis",
			Evidence: map[string]string{"serverVersion": snapshot.Kubernetes.ServerVersion},
		}}
	}

	var findings []Finding
	for _, node := range snapshot.Inventory.Nodes {
		nodeMinor, ok := kubernetesMinor(node.KubeletVersion)
		if !ok {
			findings = append(findings, Finding{
				RuleID:   RuleNodeKubeletSkew,
				Severity: SeverityInfo,
				Status:   StatusUnknown,
				Resource: ResourceFromInventory(node.Ref),
				Summary:  fmt.Sprintf("node %s kubelet version evidence is unavailable", node.Ref.Name),
				Evidence: map[string]string{"kubeletVersion": node.KubeletVersion},
			})
			continue
		}
		if nodeMinor != serverMinor {
			findings = append(findings, Finding{
				RuleID:   RuleNodeKubeletSkew,
				Severity: SeverityWarning,
				Status:   StatusWarn,
				Resource: ResourceFromInventory(node.Ref),
				Summary:  fmt.Sprintf("node %s kubelet minor differs from server minor", node.Ref.Name),
				Evidence: map[string]string{
					"serverMinor":    serverMinor,
					"kubeletMinor":   nodeMinor,
					"kubeletVersion": node.KubeletVersion,
				},
			})
		}
	}
	return findings
}

type WorkloadUnavailableRule struct{}

func (WorkloadUnavailableRule) ID() string {
	return RuleWorkloadUnavailable
}

type StoragePVCUnknownRule struct{}

func (StoragePVCUnknownRule) ID() string {
	return RuleStoragePVCUnknown
}

func (StoragePVCUnknownRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	if len(snapshot.Inventory.Storage) > 0 || len(snapshot.Inventory.Workloads) == 0 {
		return nil
	}
	return []Finding{{
		RuleID:   RuleStoragePVCUnknown,
		Severity: SeverityInfo,
		Status:   StatusUnknown,
		Resource: ResourceFromInventory(snapshot.Cluster.Identity),
		Summary:  "storage inventory evidence is unavailable",
		Evidence: map[string]string{
			"storageItems": fmt.Sprintf("%d", len(snapshot.Inventory.Storage)),
			"workloads":    fmt.Sprintf("%d", len(snapshot.Inventory.Workloads)),
		},
	}}
}

type EventWarningRule struct{}

func (EventWarningRule) ID() string {
	return RuleEventWarning
}

func (EventWarningRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	options = options.withDefaults()
	var findings []Finding
	for _, event := range snapshot.Inventory.Events {
		if event.Type != "WARNING" || !eventInLookback(event, options) {
			continue
		}
		findings = append(findings, Finding{
			RuleID:   RuleEventWarning,
			Severity: SeverityWarning,
			Status:   StatusWarn,
			Resource: ResourceFromInventory(event.Ref),
			Summary:  fmt.Sprintf("warning event observed for %s %s", event.Ref.Kind, event.Ref.Name),
			Evidence: eventEvidence(event),
		})
	}
	return findings
}

type EventUnknownTypeRule struct{}

func (EventUnknownTypeRule) ID() string {
	return RuleEventUnknownType
}

func (EventUnknownTypeRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	options = options.withDefaults()
	var findings []Finding
	for _, event := range snapshot.Inventory.Events {
		if event.Type != "UNKNOWN" || !eventInLookback(event, options) {
			continue
		}
		findings = append(findings, Finding{
			RuleID:   RuleEventUnknownType,
			Severity: SeverityInfo,
			Status:   StatusUnknown,
			Resource: ResourceFromInventory(event.Ref),
			Summary:  fmt.Sprintf("event type is unknown for %s %s", event.Ref.Kind, event.Ref.Name),
			Evidence: eventEvidence(event),
		})
	}
	return findings
}

func (WorkloadUnavailableRule) Evaluate(snapshot inventory.Snapshot, options Options) []Finding {
	var findings []Finding
	for _, workload := range snapshot.Inventory.Workloads {
		if workload.DesiredReplicas < 0 || workload.ReadyReplicas < 0 {
			findings = append(findings, Finding{
				RuleID:   RuleWorkloadUnavailable,
				Severity: SeverityInfo,
				Status:   StatusUnknown,
				Resource: ResourceFromInventory(workload.Ref),
				Summary:  fmt.Sprintf("workload %s/%s has incomplete replica evidence", workload.Ref.Namespace, workload.Ref.Name),
				Evidence: replicaEvidence(workload),
			})
			continue
		}
		if workload.ReadyReplicas < workload.DesiredReplicas {
			findings = append(findings, Finding{
				RuleID:   RuleWorkloadUnavailable,
				Severity: SeverityWarning,
				Status:   StatusWarn,
				Resource: ResourceFromInventory(workload.Ref),
				Summary:  fmt.Sprintf("workload %s/%s has fewer ready replicas than desired", workload.Ref.Namespace, workload.Ref.Name),
				Evidence: replicaEvidence(workload),
			})
		}
	}
	return findings
}

func findCondition(conditions []inventory.Condition, conditionType string) (inventory.Condition, bool) {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition, true
		}
	}
	return inventory.Condition{}, false
}

func conditionEvidence(condition inventory.Condition) map[string]string {
	evidence := map[string]string{
		"condition": condition.Type,
		"status":    condition.Status,
	}
	if condition.Reason != "" {
		evidence["reason"] = condition.Reason
	}
	return evidence
}

func kubernetesMinor(version string) (string, bool) {
	match := minorVersionPattern.FindStringSubmatch(strings.TrimSpace(version))
	if len(match) < 2 {
		return "", false
	}
	return match[1], true
}

func replicaEvidence(workload inventory.Workload) map[string]string {
	return map[string]string{
		"desiredReplicas": fmt.Sprintf("%d", workload.DesiredReplicas),
		"readyReplicas":   fmt.Sprintf("%d", workload.ReadyReplicas),
	}
}

func eventInLookback(event inventory.Event, options Options) bool {
	lastSeenAt, err := time.Parse(time.RFC3339, event.LastSeenAt)
	if err != nil {
		return false
	}
	now := options.Now()
	if lastSeenAt.After(now) {
		return true
	}
	return !lastSeenAt.Before(now.Add(-options.EventLookback))
}

func eventEvidence(event inventory.Event) map[string]string {
	evidence := map[string]string{
		"type":       event.Type,
		"lastSeenAt": event.LastSeenAt,
	}
	if event.Reason != "" {
		evidence["reason"] = event.Reason
	}
	return evidence
}
