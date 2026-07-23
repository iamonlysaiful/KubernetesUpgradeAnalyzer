package components

import "github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"

type Confidence string

const (
	ConfidenceHigh    Confidence = "HIGH"
	ConfidenceMedium  Confidence = "MEDIUM"
	ConfidenceLow     Confidence = "LOW"
	ConfidenceUnknown Confidence = "UNKNOWN"
)

type Status string

const (
	StatusFound    Status = "FOUND"
	StatusNotFound Status = "NOT_FOUND"
	StatusUnknown  Status = "UNKNOWN"
)

type ResourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}

type Detection struct {
	ComponentID string
	Name        string
	Version     string
	Confidence  Confidence
	Status      Status
	Evidence    []ResourceRef
	Limitations []Limitation
}

type Limitation struct {
	Code    string
	Summary string
}

func ResourceFromInventory(ref inventory.ResourceRef) ResourceRef {
	return ResourceRef{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Namespace:  ref.Namespace,
		Name:       ref.Name,
	}
}
