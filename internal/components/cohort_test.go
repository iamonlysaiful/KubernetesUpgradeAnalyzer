package components

import (
	"reflect"
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

func TestInitialDetectorCohortFindsKnownComponents(t *testing.T) {
	runner := NewRunner(InitialDetectorCohort()...)
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Workloads: []inventory.Workload{
				workload("Deployment", "kube-system", "coredns", "coredns", "registry.k8s.io/coredns/coredns:v1.11.3"),
				workload("Deployment", "kube-system", "metrics-server", "metrics-server", "registry.k8s.io/metrics-server/metrics-server:v0.7.2"),
				workload("DaemonSet", "kube-system", "azuredisk-csi-node", "azuredisk", "mcr.microsoft.com/oss/kubernetes-csi/azuredisk-csi:v1.30.4"),
				workload("DaemonSet", "kube-system", "azurefile-csi-node", "azurefile", "mcr.microsoft.com/oss/kubernetes-csi/azurefile-csi:v1.30.5"),
				workload("DaemonSet", "logging", "fluent-bit", "fluent-bit", "cr.fluentbit.io/fluent/fluent-bit:3.1.9"),
				workload("StatefulSet", "mqtt", "emqx", "emqx", "emqx/emqx:5.8.1"),
				workload("Deployment", "ingress-nginx", "ingress-nginx-controller", "controller", "registry.k8s.io/ingress-nginx/controller:v1.12.1"),
				workload("Deployment", "default", "business-api", "api", "example/api:v1.0.0"),
			},
		},
	}

	detections := runner.Detect(snapshot)

	got := detectionKeys(detections)
	want := []string{
		"azure-disk-csi|kube-system|DaemonSet|azuredisk-csi-node|1.30.4",
		"azure-file-csi|kube-system|DaemonSet|azurefile-csi-node|1.30.5",
		"coredns|kube-system|Deployment|coredns|1.11.3",
		"emqx|mqtt|StatefulSet|emqx|5.8.1",
		"fluent-bit|logging|DaemonSet|fluent-bit|3.1.9",
		"metrics-server|kube-system|Deployment|metrics-server|0.7.2",
		"nginx-ingress|ingress-nginx|Deployment|ingress-nginx-controller|1.12.1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detections = %#v, want %#v", got, want)
	}
}

func TestDetectorReportsUnknownVersionForLatestTag(t *testing.T) {
	detector := WorkloadDetector{componentID: "coredns", name: "CoreDNS", imageHints: []string{"coredns"}, nameHints: []string{"coredns"}}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Workloads: []inventory.Workload{
				workload("Deployment", "kube-system", "coredns", "coredns", "registry.k8s.io/coredns/coredns:latest"),
			},
		},
	}

	detections := detector.Detect(snapshot)
	if len(detections) != 1 {
		t.Fatalf("detections = %d, want 1", len(detections))
	}
	if detections[0].Status != StatusUnknown || detections[0].Version != UnknownVersion {
		t.Fatalf("detection = %#v, want UNKNOWN version/status", detections[0])
	}
	if len(detections[0].Limitations) != 1 || detections[0].Limitations[0].Code != "VERSION_UNKNOWN" {
		t.Fatalf("limitations = %#v, want VERSION_UNKNOWN", detections[0].Limitations)
	}
}

func TestDetectorReportsUnknownVersionForConflictingImages(t *testing.T) {
	detector := WorkloadDetector{componentID: "emqx", name: "EMQX", imageHints: []string{"emqx"}, nameHints: []string{"emqx"}}
	snapshot := inventory.Snapshot{
		Inventory: inventory.Inventory{
			Workloads: []inventory.Workload{
				{
					Ref: inventory.ResourceRef{APIVersion: "apps/v1", Kind: "StatefulSet", Namespace: "mqtt", Name: "emqx"},
					Containers: []inventory.Container{
						{Name: "emqx-a", Image: "emqx/emqx:5.8.1"},
						{Name: "emqx-b", Image: "emqx/emqx:5.8.2"},
					},
				},
			},
		},
	}

	detections := detector.Detect(snapshot)
	if len(detections) != 1 {
		t.Fatalf("detections = %d, want 1", len(detections))
	}
	if detections[0].Status != StatusUnknown || detections[0].Confidence != ConfidenceUnknown {
		t.Fatalf("detection = %#v, want UNKNOWN", detections[0])
	}
}

func TestInitialDetectorCohortIDsAreStable(t *testing.T) {
	detectors := InitialDetectorCohort()

	got := make([]string, 0, len(detectors))
	for _, detector := range detectors {
		got = append(got, detector.ID())
	}
	want := []string{
		"azure-disk-csi",
		"azure-file-csi",
		"coredns",
		"emqx",
		"fluent-bit",
		"metrics-server",
		"nginx-ingress",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detector IDs = %#v, want %#v", got, want)
	}
}

func workload(kind string, namespace string, name string, containerName string, image string) inventory.Workload {
	return inventory.Workload{
		Ref: inventory.ResourceRef{
			APIVersion: "apps/v1",
			Kind:       kind,
			Namespace:  namespace,
			Name:       name,
		},
		DesiredReplicas: 1,
		ReadyReplicas:   1,
		Containers: []inventory.Container{
			{Name: containerName, Image: image},
		},
	}
}
