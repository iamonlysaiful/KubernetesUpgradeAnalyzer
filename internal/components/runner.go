package components

import (
	"sort"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

type Detector interface {
	ID() string
	Detect(snapshot inventory.Snapshot) []Detection
}

type DetectorFunc struct {
	DetectorID string
	Run        func(snapshot inventory.Snapshot) []Detection
}

func (detector DetectorFunc) ID() string {
	return detector.DetectorID
}

func (detector DetectorFunc) Detect(snapshot inventory.Snapshot) []Detection {
	if detector.Run == nil {
		return nil
	}
	return detector.Run(snapshot)
}

type Runner struct {
	detectors []Detector
}

func NewRunner(detectors ...Detector) Runner {
	copied := append([]Detector(nil), detectors...)
	return Runner{detectors: copied}
}

func (runner Runner) Detect(snapshot inventory.Snapshot) []Detection {
	var detections []Detection
	for _, detector := range runner.detectors {
		if detector == nil {
			continue
		}
		detections = append(detections, detector.Detect(snapshot)...)
	}
	SortDetections(detections)
	return detections
}

func SortDetections(detections []Detection) {
	sort.SliceStable(detections, func(i, j int) bool {
		left := detections[i]
		right := detections[j]
		leftRef := firstEvidence(left)
		rightRef := firstEvidence(right)

		if left.ComponentID != right.ComponentID {
			return left.ComponentID < right.ComponentID
		}
		if leftRef.Namespace != rightRef.Namespace {
			return leftRef.Namespace < rightRef.Namespace
		}
		if leftRef.Kind != rightRef.Kind {
			return leftRef.Kind < rightRef.Kind
		}
		if leftRef.Name != rightRef.Name {
			return leftRef.Name < rightRef.Name
		}
		return left.Version < right.Version
	})
}

func firstEvidence(detection Detection) ResourceRef {
	if len(detection.Evidence) == 0 {
		return ResourceRef{}
	}
	return detection.Evidence[0]
}
