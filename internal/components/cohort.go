package components

import (
	"strings"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

type WorkloadDetector struct {
	componentID string
	name        string
	imageHints  []string
	nameHints   []string
}

func InitialDetectorCohort() []Detector {
	return []Detector{
		WorkloadDetector{componentID: "azure-disk-csi", name: "Azure Disk CSI", imageHints: []string{"azuredisk-csi"}, nameHints: []string{"azuredisk-csi"}},
		WorkloadDetector{componentID: "azure-file-csi", name: "Azure File CSI", imageHints: []string{"azurefile-csi"}, nameHints: []string{"azurefile-csi"}},
		WorkloadDetector{componentID: "coredns", name: "CoreDNS", imageHints: []string{"coredns"}, nameHints: []string{"coredns"}},
		WorkloadDetector{componentID: "emqx", name: "EMQX", imageHints: []string{"emqx"}, nameHints: []string{"emqx"}},
		WorkloadDetector{componentID: "fluent-bit", name: "Fluent Bit", imageHints: []string{"fluent-bit", "fluentbit"}, nameHints: []string{"fluent-bit", "fluentbit"}},
		WorkloadDetector{componentID: "metrics-server", name: "Metrics Server", imageHints: []string{"metrics-server"}, nameHints: []string{"metrics-server"}},
		WorkloadDetector{componentID: "nginx-ingress", name: "NGINX Ingress", imageHints: []string{"ingress-nginx", "nginx-ingress"}, nameHints: []string{"ingress-nginx", "nginx-ingress"}},
	}
}

func (detector WorkloadDetector) ID() string {
	return detector.componentID
}

func (detector WorkloadDetector) Detect(snapshot inventory.Snapshot) []Detection {
	var detections []Detection
	for _, workload := range snapshot.Inventory.Workloads {
		if !detector.matchesWorkload(workload) {
			continue
		}
		version, confidence, status := detector.version(workload)
		detection := Detection{
			ComponentID: detector.componentID,
			Name:        detector.name,
			Version:     version,
			Confidence:  confidence,
			Status:      status,
			Evidence:    []ResourceRef{ResourceFromInventory(workload.Ref)},
		}
		if status == StatusUnknown {
			detection.Limitations = append(detection.Limitations, Limitation{
				Code:    "VERSION_UNKNOWN",
				Summary: "component version evidence is missing or ambiguous",
			})
		}
		detections = append(detections, detection)
	}
	return detections
}

func (detector WorkloadDetector) matchesWorkload(workload inventory.Workload) bool {
	resourceText := strings.ToLower(workload.Ref.Name + " " + workload.Ref.Namespace)
	for _, hint := range detector.nameHints {
		if strings.Contains(resourceText, hint) {
			return true
		}
	}
	for _, container := range workload.Containers {
		image := strings.ToLower(container.Image)
		for _, hint := range detector.imageHints {
			if strings.Contains(image, hint) {
				return true
			}
		}
	}
	return false
}

func (detector WorkloadDetector) version(workload inventory.Workload) (string, Confidence, Status) {
	var foundVersion string
	for _, container := range workload.Containers {
		if !detector.matchesImage(container.Image) {
			continue
		}
		version, confidence, status := NormalizeVersion(container.Image)
		if status == StatusUnknown {
			return UnknownVersion, ConfidenceUnknown, StatusUnknown
		}
		if foundVersion != "" && foundVersion != version {
			return UnknownVersion, ConfidenceUnknown, StatusUnknown
		}
		foundVersion = version
		if confidence != ConfidenceHigh {
			return UnknownVersion, ConfidenceUnknown, StatusUnknown
		}
	}
	if foundVersion == "" {
		return UnknownVersion, ConfidenceUnknown, StatusUnknown
	}
	return foundVersion, ConfidenceHigh, StatusFound
}

func (detector WorkloadDetector) matchesImage(image string) bool {
	lower := strings.ToLower(image)
	for _, hint := range detector.imageHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}
