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
		detectedVersions, hasAmbiguousEvidence := detector.versions(workload)
		if len(detectedVersions) == 0 {
			detections = append(detections, Detection{
				ComponentID: detector.componentID,
				Name:        detector.name,
				Version:     UnknownVersion,
				Confidence:  ConfidenceUnknown,
				Status:      StatusUnknown,
				Evidence:    []ResourceRef{ResourceFromInventory(workload.Ref)},
				Limitations: []Limitation{{
					Code:    "VERSION_UNKNOWN",
					Summary: "component version evidence is missing or ambiguous",
				}},
			})
			continue
		}
		for _, version := range detectedVersions {
			detection := Detection{
				ComponentID: detector.componentID,
				Name:        detector.name,
				Version:     version,
				Confidence:  ConfidenceHigh,
				Status:      StatusFound,
				Evidence:    []ResourceRef{ResourceFromInventory(workload.Ref)},
			}
			if hasAmbiguousEvidence {
				detection.Limitations = append(detection.Limitations, Limitation{
					Code:    "VERSION_PARTIAL",
					Summary: "component has known version evidence plus ambiguous matching evidence",
				})
			}
			detections = append(detections, detection)
		}
	}
	return detections
}

func (detector WorkloadDetector) matchesWorkload(workload inventory.Workload) bool {
	// Match on the workload's own name only. Matching on namespace would
	// misclassify unrelated workloads that merely share a namespace with
	// the component (e.g. a config-reloader deployed in an "emqx" namespace).
	resourceName := strings.ToLower(workload.Ref.Name)
	for _, hint := range detector.nameHints {
		if strings.Contains(resourceName, hint) {
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

func (detector WorkloadDetector) versions(workload inventory.Workload) ([]string, bool) {
	// Prefer the explicit app.kubernetes.io/version label over image-tag
	// extraction: it reflects the application version, not the image build tag.
	if workload.VersionLabel != "" {
		version, confidence, status := NormalizeLabelVersion(workload.VersionLabel)
		if status == StatusFound && confidence != ConfidenceUnknown {
			return []string{version}, false
		}
	}

	seen := map[string]bool{}
	var versions []string
	hasAmbiguousEvidence := false
	for _, container := range workload.Containers {
		if !detector.matchesImage(container.Image) {
			continue
		}
		version, confidence, status := NormalizeVersion(container.Image)
		if status == StatusUnknown || confidence != ConfidenceHigh {
			hasAmbiguousEvidence = true
			continue
		}
		if !seen[version] {
			seen[version] = true
			versions = append(versions, version)
		}
	}
	return versions, hasAmbiguousEvidence
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
