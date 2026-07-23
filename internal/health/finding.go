package health

import "github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"

type Severity string

const (
	SeverityBlocker Severity = "BLOCKER"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

type Status string

const (
	StatusFail    Status = "FAIL"
	StatusWarn    Status = "WARN"
	StatusPass    Status = "PASS"
	StatusUnknown Status = "UNKNOWN"
)

type ResourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

type Finding struct {
	RuleID   string
	Severity Severity
	Status   Status
	Resource ResourceRef
	Summary  string
	Evidence map[string]string
}

func ResourceFromInventory(ref inventory.ResourceRef) ResourceRef {
	return ResourceRef{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Namespace:  ref.Namespace,
		Name:       ref.Name,
	}
}
