package recommendation

import (
	"fmt"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/kube/inventory"
)

// EvidenceSummary summarizes what was analyzed, providing transparency
// into the recommendation's evidence base.
type EvidenceSummary struct {
	// Deployments summarizes Deployment workloads analyzed.
	Deployments CountSummary `json:"deployments"`
	// StatefulSets summarizes StatefulSet workloads analyzed.
	StatefulSets CountSummary `json:"statefulSets"`
	// DaemonSets summarizes DaemonSet workloads analyzed.
	DaemonSets CountSummary `json:"daemonSets"`
	// Nodes summarizes cluster nodes analyzed.
	Nodes CountSummary `json:"nodes"`
	// PVCs summarizes PersistentVolumeClaims analyzed.
	PVCs CountSummary `json:"pvcs"`
	// CRDs summarizes CustomResourceDefinitions analyzed.
	CRDs CountSummary `json:"crds"`
	// Components lists detected components with their status.
	Components []ComponentStatus `json:"components"`
	// DeprecatedAPIs is the count of deprecated APIs found.
	DeprecatedAPIs int `json:"deprecatedAPIs"`
	// AnalysisTime is when the analysis was performed.
	AnalysisTime time.Time `json:"analysisTime"`
}

// CountSummary provides total and healthy counts for a resource type.
type CountSummary struct {
	// Total is the total count of resources.
	Total int `json:"total"`
	// Healthy is the count of healthy/ready resources.
	Healthy int `json:"healthy"`
}

// ComponentStatus represents a detected component's compatibility status.
type ComponentStatus struct {
	// Name is the component display name.
	Name string `json:"name"`
	// Version is the detected version (may be "unknown").
	Version string `json:"version"`
	// Status indicates compatibility: "compatible", "conditional", "incompatible", "unknown".
	Status string `json:"status"`
}

// EvidenceBuilder constructs an EvidenceSummary from raw inputs.
type EvidenceBuilder struct {
	clock func() time.Time
}

// NewEvidenceBuilder creates a new evidence builder.
func NewEvidenceBuilder() *EvidenceBuilder {
	return &EvidenceBuilder{
		clock: time.Now,
	}
}

// WithClock sets a custom clock for testing.
func (b *EvidenceBuilder) WithClock(clock func() time.Time) *EvidenceBuilder {
	b.clock = clock
	return b
}

// Build constructs an EvidenceSummary from the inventory snapshot and component detections.
func (b *EvidenceBuilder) Build(
	inv *inventory.Inventory,
	detections []components.Detection,
	deprecatedAPICount int,
) *EvidenceSummary {
	summary := &EvidenceSummary{
		AnalysisTime:   b.clock().UTC(),
		DeprecatedAPIs: deprecatedAPICount,
	}

	if inv == nil {
		return summary
	}

	// Count nodes
	summary.Nodes = b.countNodes(inv.Nodes)

	// Count workloads by kind
	summary.Deployments, summary.StatefulSets, summary.DaemonSets = b.countWorkloads(inv.Workloads)

	// Count PVCs from storage references
	summary.PVCs = b.countPVCs(inv.Storage)

	// Count CRDs
	summary.CRDs = CountSummary{
		Total:   len(inv.CRDs),
		Healthy: len(inv.CRDs), // CRDs don't have a "healthy" concept
	}

	// Build component status list
	summary.Components = b.buildComponentStatuses(detections)

	return summary
}

// countNodes counts total and healthy nodes.
func (b *EvidenceBuilder) countNodes(nodes []inventory.Node) CountSummary {
	total := len(nodes)
	healthy := 0

	for _, node := range nodes {
		if isNodeReady(node) {
			healthy++
		}
	}

	return CountSummary{Total: total, Healthy: healthy}
}

// isNodeReady checks if a node has Ready=True condition.
func isNodeReady(node inventory.Node) bool {
	for _, cond := range node.Conditions {
		if cond.Type == "Ready" && cond.Status == "TRUE" {
			return true
		}
	}
	return false
}

// countWorkloads counts workloads by kind.
func (b *EvidenceBuilder) countWorkloads(workloads []inventory.Workload) (deployments, statefulSets, daemonSets CountSummary) {
	for _, w := range workloads {
		isHealthy := w.ReadyReplicas >= w.DesiredReplicas

		switch w.Ref.Kind {
		case "Deployment":
			deployments.Total++
			if isHealthy {
				deployments.Healthy++
			}
		case "StatefulSet":
			statefulSets.Total++
			if isHealthy {
				statefulSets.Healthy++
			}
		case "DaemonSet":
			daemonSets.Total++
			if isHealthy {
				daemonSets.Healthy++
			}
		}
	}
	return
}

// countPVCs counts PVCs from storage references.
// We can't determine bound/unbound from ResourceRef alone, so we use findings.
func (b *EvidenceBuilder) countPVCs(storage []inventory.ResourceRef) CountSummary {
	total := 0
	for _, ref := range storage {
		if ref.Kind == "PersistentVolumeClaim" {
			total++
		}
	}
	// Without phase info, assume all are healthy (findings will flag unbound ones)
	return CountSummary{Total: total, Healthy: total}
}

// buildComponentStatuses converts component detections to status list.
func (b *EvidenceBuilder) buildComponentStatuses(detections []components.Detection) []ComponentStatus {
	statuses := make([]ComponentStatus, 0, len(detections))

	for _, d := range detections {
		// Skip not-found components
		if d.Status == components.StatusNotFound {
			continue
		}

		status := "unknown"
		switch d.Status {
		case components.StatusFound:
			status = "detected"
		case components.StatusUnknown:
			status = "unknown"
		}

		version := d.Version
		if version == "" || version == "UNKNOWN" {
			version = "unknown"
		}

		statuses = append(statuses, ComponentStatus{
			Name:    d.Name,
			Version: version,
			Status:  status,
		})
	}

	return statuses
}

// AllHealthy returns true if Total == Healthy for a CountSummary.
func (c CountSummary) AllHealthy() bool {
	return c.Total == c.Healthy
}

// String returns a human-readable summary like "3/3" or "2/3".
func (c CountSummary) String() string {
	if c.Total == 0 {
		return "0"
	}
	return fmt.Sprintf("%d/%d", c.Healthy, c.Total)
}
