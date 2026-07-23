package components

import (
	"reflect"
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

func TestRunnerSkipsNilDetectorsAndSortsDetections(t *testing.T) {
	runner := NewRunner(
		DetectorFunc{
			DetectorID: "test.b",
			Run: func(snapshot inventory.Snapshot) []Detection {
				return []Detection{
					detection("metrics-server", "kube-system", "Deployment", "metrics-server", "0.7.2"),
					detection("coredns", "kube-system", "Deployment", "coredns", "1.11.3"),
				}
			},
		},
		nil,
		DetectorFunc{
			DetectorID: "test.a",
			Run: func(snapshot inventory.Snapshot) []Detection {
				return []Detection{
					detection("coredns", "alpha", "Deployment", "coredns", "1.11.1"),
					detection("coredns", "alpha", "Deployment", "coredns", "1.11.0"),
				}
			},
		},
	)

	detections := runner.Detect(inventory.Snapshot{})

	got := detectionKeys(detections)
	want := []string{
		"coredns|alpha|Deployment|coredns|1.11.0",
		"coredns|alpha|Deployment|coredns|1.11.1",
		"coredns|kube-system|Deployment|coredns|1.11.3",
		"metrics-server|kube-system|Deployment|metrics-server|0.7.2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detections = %#v, want %#v", got, want)
	}
}

func TestDetectorFuncWithNilRunReturnsNoDetections(t *testing.T) {
	detector := DetectorFunc{DetectorID: "test.nil"}
	if got := detector.Detect(inventory.Snapshot{}); got != nil {
		t.Fatalf("Detect = %#v, want nil", got)
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

func detection(componentID string, namespace string, kind string, name string, version string) Detection {
	return Detection{
		ComponentID: componentID,
		Name:        componentID,
		Version:     version,
		Confidence:  ConfidenceHigh,
		Status:      StatusFound,
		Evidence: []ResourceRef{
			{Namespace: namespace, Kind: kind, Name: name},
		},
	}
}

func detectionKeys(detections []Detection) []string {
	result := make([]string, 0, len(detections))
	for _, detection := range detections {
		ref := firstEvidence(detection)
		result = append(result, detection.ComponentID+"|"+ref.Namespace+"|"+ref.Kind+"|"+ref.Name+"|"+detection.Version)
	}
	return result
}
