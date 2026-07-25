package recommendation

import (
	"fmt"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/external/kubent"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/health"
)

// Aggregator combines findings from multiple sources into recommendation findings.
type Aggregator struct{}

// NewAggregator creates a new finding aggregator.
func NewAggregator() *Aggregator {
	return &Aggregator{}
}

// AggregateHealth converts health findings to recommendation findings.
func (a *Aggregator) AggregateHealth(healthFindings []health.Finding) []Finding {
	findings := make([]Finding, 0, len(healthFindings))

	for _, hf := range healthFindings {
		// Skip passing health checks
		if hf.Status == health.StatusPass {
			continue
		}

		f := Finding{
			ID:       hf.RuleID,
			Category: CategoryHealth,
			Severity: a.healthSeverityToRecommendation(hf.Severity, hf.Status),
			Summary:  hf.Summary,
			Resource: &ResourceRef{
				Kind:      hf.Resource.Kind,
				Namespace: hf.Resource.Namespace,
				Name:      hf.Resource.Name,
			},
		}

		// Add remediation based on rule
		f.Remediation = a.healthRemediation(hf.RuleID)

		findings = append(findings, f)
	}

	return findings
}

// healthSeverityToRecommendation maps health severity to recommendation severity.
func (a *Aggregator) healthSeverityToRecommendation(sev health.Severity, status health.Status) FindingSeverity {
	if status == health.StatusUnknown {
		return SeverityUnknown
	}

	switch sev {
	case health.SeverityBlocker:
		return SeverityBlocker
	case health.SeverityWarning:
		return SeverityWarning
	case health.SeverityInfo:
		return SeverityInfo
	default:
		return SeverityUnknown
	}
}

// healthRemediation returns remediation guidance for health rules.
func (a *Aggregator) healthRemediation(ruleID string) string {
	remediations := map[string]string{
		"NODE_NOT_READY":        "Investigate and resolve node conditions before upgrade",
		"NODE_MEMORY_PRESSURE":  "Address memory pressure on nodes before upgrade",
		"NODE_DISK_PRESSURE":    "Address disk pressure on nodes before upgrade",
		"NODE_PID_PRESSURE":     "Address PID pressure on nodes before upgrade",
		"WORKLOAD_UNAVAILABLE":  "Ensure workload replicas are available before upgrade",
		"PVC_UNBOUND":           "Resolve unbound PVCs before upgrade",
		"EVENTS_WARNING":        "Review warning events and address underlying issues",
	}

	if r, ok := remediations[ruleID]; ok {
		return r
	}
	return "Review and address the issue before upgrade"
}

// AggregateAPI converts kubent API findings to recommendation findings.
func (a *Aggregator) AggregateAPI(apiFindings []kubent.Finding, targetVersion string) []Finding {
	findings := make([]Finding, 0, len(apiFindings))

	for _, af := range apiFindings {
		// Handle findings with limitations
		if len(af.Limitations) > 0 {
			for _, lim := range af.Limitations {
				findings = append(findings, Finding{
					ID:       "API_" + lim.Code,
					Category: CategoryAPI,
					Severity: SeverityUnknown,
					Summary:  lim.Summary,
					Stage:    targetVersion,
				})
			}
			continue
		}

		// Skip passing API checks
		if af.Status == kubent.FindingPass {
			continue
		}

		severity := SeverityWarning
		if af.Status == kubent.FindingFail {
			severity = SeverityBlocker
		} else if af.Status == kubent.FindingUnknown {
			severity = SeverityUnknown
		}

		summary := fmt.Sprintf("API %s %s", af.APIVersion, af.Kind)
		if af.RemovedIn != "" {
			summary = fmt.Sprintf("API %s %s is removed in %s", af.APIVersion, af.Kind, af.RemovedIn)
		}

		detail := ""
		if af.Replacement != "" {
			detail = fmt.Sprintf("Replace with %s", af.Replacement)
		}

		f := Finding{
			ID:          fmt.Sprintf("API_%s_%s", af.APIVersion, af.Kind),
			Category:    CategoryAPI,
			Severity:    severity,
			Summary:     summary,
			Detail:      detail,
			Stage:       targetVersion,
			Remediation: a.apiRemediation(af),
		}

		if af.Resource.Kind != "" {
			f.Resource = &ResourceRef{
				Kind:      af.Resource.Kind,
				Namespace: af.Resource.Namespace,
				Name:      af.Resource.Name,
			}
		}

		findings = append(findings, f)
	}

	return findings
}

// apiRemediation returns remediation guidance for API findings.
func (a *Aggregator) apiRemediation(af kubent.Finding) string {
	if af.Replacement != "" {
		return fmt.Sprintf("Migrate to %s before upgrading to %s", af.Replacement, af.RemovedIn)
	}
	if af.RemovedIn != "" {
		return fmt.Sprintf("Address deprecated API usage before upgrading to %s", af.RemovedIn)
	}
	return "Review API compatibility before upgrade"
}

// AggregateComponents converts component detections to recommendation findings.
func (a *Aggregator) AggregateComponents(detections []components.Detection, targetMinor int) []Finding {
	findings := make([]Finding, 0)

	for _, d := range detections {
		// Skip not found components
		if d.Status == components.StatusNotFound {
			continue
		}

		// Handle unknown version or status
		if d.Status == components.StatusUnknown || d.Version == "UNKNOWN" {
			findings = append(findings, Finding{
				ID:          fmt.Sprintf("COMPONENT_%s_UNKNOWN", d.ComponentID),
				Category:    CategoryComponent,
				Severity:    SeverityUnknown,
				Summary:     fmt.Sprintf("Component %s version is unknown", d.Name),
				Remediation: fmt.Sprintf("Verify %s version and compatibility", d.Name),
			})
			continue
		}

		// Handle detection limitations
		for _, lim := range d.Limitations {
			findings = append(findings, Finding{
				ID:          fmt.Sprintf("COMPONENT_%s_%s", d.ComponentID, lim.Code),
				Category:    CategoryComponent,
				Severity:    SeverityWarning,
				Summary:     fmt.Sprintf("%s: %s", d.Name, lim.Summary),
				Remediation: "Review component compatibility documentation",
			})
		}
	}

	return findings
}

// AggregateLimitations extracts limitations from various sources.
func (a *Aggregator) AggregateLimitations(
	apiFindings []kubent.Finding,
	detections []components.Detection,
) []Limitation {
	limitations := make([]Limitation, 0)

	// API limitations
	for _, af := range apiFindings {
		for _, lim := range af.Limitations {
			limitations = append(limitations, Limitation{
				Code:    lim.Code,
				Summary: lim.Summary,
				Impact:  "API compatibility analysis may be incomplete",
			})
		}
	}

	// Component limitations
	for _, d := range detections {
		for _, lim := range d.Limitations {
			limitations = append(limitations, Limitation{
				Code:    lim.Code,
				Summary: fmt.Sprintf("%s: %s", d.Name, lim.Summary),
				Impact:  "Component compatibility analysis may be incomplete",
			})
		}
	}

	return limitations
}
