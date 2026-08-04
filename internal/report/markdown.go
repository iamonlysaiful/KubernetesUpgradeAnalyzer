package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/plan"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
)

// RenderMarkdown renders a comprehensive Markdown report matching console output.
func RenderMarkdown(doc Document) ([]byte, error) {
	var b strings.Builder

	// ── Header ──
	fmt.Fprintln(&b, "# Kubernetes Upgrade Advisor")
	fmt.Fprintln(&b)

	// ── Decision and Confidence ──
	if doc.Decision != "" {
		fmt.Fprintf(&b, "**Recommendation:** %s\n\n", mdDecisionDisplay(doc.Decision))
		if doc.Confidence != nil {
			fmt.Fprintf(&b, "**Confidence:** %d%%\n\n", doc.Confidence.Percentage)
		}
	} else {
		fmt.Fprintf(&b, "**Readiness:** %s\n\n", doc.Readiness)
	}

	// ── Version Info ──
	fmt.Fprintf(&b, "| | |\n")
	fmt.Fprintf(&b, "|---|---|\n")
	fmt.Fprintf(&b, "| Current | %s |\n", escapeMarkdown(doc.Current))
	fmt.Fprintf(&b, "| Destination | %s |\n", escapeMarkdown(defaultString(doc.Destination, "n/a")))
	if len(doc.Path) > 0 {
		steps := make([]string, 0, len(doc.Path)+1)
		steps = append(steps, doc.Path[0].From)
		for _, stage := range doc.Path {
			steps = append(steps, stage.To)
		}
		fmt.Fprintf(&b, "| Upgrade Path | %s |\n", strings.Join(steps, " → "))
	} else if doc.Destination != "" {
		fmt.Fprintf(&b, "| Upgrade Path | %s → %s |\n", escapeMarkdown(doc.Current), escapeMarkdown(doc.Destination))
	}
	if doc.UpgradePlan != nil && doc.UpgradePlan.EstimatedTime > 0 {
		fmt.Fprintf(&b, "| Estimated Time | %s |\n", mdFormatDuration(doc.UpgradePlan.EstimatedTime))
	}
	fmt.Fprintln(&b)

	// ── Blockers ──
	blockers := mdFindingsBySeverity(doc.Findings, recommendation.SeverityBlocker)
	fmt.Fprintf(&b, "## Blockers (%d)\n\n", len(blockers))
	if len(blockers) == 0 {
		fmt.Fprintln(&b, "✅ None — no issues blocking upgrade.")
		fmt.Fprintln(&b)
	} else {
		for i, f := range blockers {
			mdWriteFinding(&b, i+1, f)
		}
		fmt.Fprintln(&b)
	}

	// ── Warnings ──
	warnings := mdFindingsBySeverity(doc.Findings, recommendation.SeverityWarning)
	if len(warnings) > 0 {
		fmt.Fprintf(&b, "## Warnings (%d)\n\n", len(warnings))
		for i, f := range warnings {
			mdWriteFinding(&b, i+1, f)
		}
		fmt.Fprintln(&b)
	}

	// ── Evidence Gaps ──
	if len(doc.Limitations) > 0 {
		fmt.Fprintf(&b, "## Evidence Gaps (%d)\n\n", len(doc.Limitations))
		for i, l := range doc.Limitations {
			fmt.Fprintf(&b, "%d. **%s**: %s\n", i+1, escapeMarkdown(l.Code), escapeMarkdown(l.Summary))
			if l.Impact != "" {
				fmt.Fprintf(&b, "   - Impact: %s\n", escapeMarkdown(l.Impact))
			}
		}
		fmt.Fprintln(&b)
	}

	// ── Evidence Summary ──
	if doc.Evidence != nil {
		fmt.Fprintln(&b, "## Evidence Summary")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| Resource | Status |")
		fmt.Fprintln(&b, "|----------|--------|")
		fmt.Fprintf(&b, "| Deployments | %d/%d healthy |\n", doc.Evidence.Deployments.Healthy, doc.Evidence.Deployments.Total)
		fmt.Fprintf(&b, "| StatefulSets | %d/%d healthy |\n", doc.Evidence.StatefulSets.Healthy, doc.Evidence.StatefulSets.Total)
		fmt.Fprintf(&b, "| DaemonSets | %d/%d healthy |\n", doc.Evidence.DaemonSets.Healthy, doc.Evidence.DaemonSets.Total)
		fmt.Fprintf(&b, "| Nodes | %d/%d healthy |\n", doc.Evidence.Nodes.Healthy, doc.Evidence.Nodes.Total)
		fmt.Fprintf(&b, "| PVCs | %d/%d healthy |\n", doc.Evidence.PVCs.Healthy, doc.Evidence.PVCs.Total)
		fmt.Fprintf(&b, "| CRDs | %d |\n", doc.Evidence.CRDs.Total)
		fmt.Fprintf(&b, "| Deprecated APIs | %d found |\n", doc.Evidence.DeprecatedAPIs)
		fmt.Fprintln(&b)

		if len(doc.Evidence.Components) > 0 {
			fmt.Fprintln(&b, "### Detected Components")
			fmt.Fprintln(&b)
			fmt.Fprintln(&b, "| Component | Version | Status |")
			fmt.Fprintln(&b, "|-----------|---------|--------|")
			for _, c := range doc.Evidence.Components {
				fmt.Fprintf(&b, "| %s | %s | %s |\n",
					escapeMarkdown(c.Name),
					escapeMarkdown(defaultString(c.Version, "unknown")),
					escapeMarkdown(c.Status))
			}
			fmt.Fprintln(&b)
		}
	}

	// ── Recommended Upgrade Plan ──
	if doc.UpgradePlan != nil && len(doc.UpgradePlan.Steps) > 0 {
		fmt.Fprintln(&b, "## Recommended Upgrade Plan")
		fmt.Fprintln(&b)
		mdWritePlanSteps(&b, doc.UpgradePlan.Steps)
		fmt.Fprintln(&b)
	}

	// ── Post-Upgrade Validation ──
	if doc.UpgradePlan != nil && len(doc.UpgradePlan.ValidationSteps) > 0 {
		fmt.Fprintln(&b, "## Post-Upgrade Validation")
		fmt.Fprintln(&b)
		mdWritePlanSteps(&b, doc.UpgradePlan.ValidationSteps)
		fmt.Fprintln(&b)
	}

	// ── Rollback Guidance ──
	if doc.UpgradePlan != nil && doc.UpgradePlan.RollbackGuidance != "" {
		fmt.Fprintln(&b, "## Rollback Guidance")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, doc.UpgradePlan.RollbackGuidance)
		fmt.Fprintln(&b)
	}

	// ── Footer ──
	fmt.Fprintln(&b, "---")
	if doc.Decision != "" {
		fmt.Fprintf(&b, "**Overall Decision:** %s\n\n", mdDecisionDisplay(doc.Decision))
	}
	fmt.Fprintf(&b, "*Generated: %s | ID: %s*\n",
		doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"),
		escapeMarkdown(doc.AssessmentID))

	return []byte(b.String()), nil
}

func mdDecisionDisplay(d recommendation.Decision) string {
	switch d {
	case recommendation.DecisionGo:
		return "🟢 PROCEED WITH UPGRADE"
	case recommendation.DecisionCaution:
		return "🟡 PROCEED WITH CAUTION"
	case recommendation.DecisionStop:
		return "🔴 DO NOT PROCEED"
	default:
		return string(d)
	}
}

func mdWriteFinding(b *strings.Builder, num int, f recommendation.Finding) {
	fmt.Fprintf(b, "%d. **[%s]** %s\n", num, f.Category, escapeMarkdown(f.Summary))
	if f.Resource != nil {
		ns := defaultString(f.Resource.Namespace, "-")
		fmt.Fprintf(b, "   - Resource: `%s/%s`", escapeMarkdown(ns), escapeMarkdown(f.Resource.Name))
		if f.Resource.Kind != "" {
			fmt.Fprintf(b, " (%s)", escapeMarkdown(f.Resource.Kind))
		}
		fmt.Fprintln(b)
	}
	if f.Impact != nil {
		fmt.Fprintf(b, "   - Impact: **%s** — %s\n", f.Impact.Level, escapeMarkdown(f.Impact.Explanation))
	}
	if f.Action != nil {
		fmt.Fprintf(b, "   - Action: %s\n", escapeMarkdown(f.Action.Description))
		if f.Action.Command != "" {
			fmt.Fprintf(b, "     ```\n     %s\n     ```\n", f.Action.Command)
		}
	} else if f.Remediation != "" {
		fmt.Fprintf(b, "   - Remediation: %s\n", escapeMarkdown(f.Remediation))
	}
	if f.IfIgnored != "" {
		fmt.Fprintf(b, "   - If ignored: %s\n", escapeMarkdown(f.IfIgnored))
	}
}

func mdWritePlanSteps(b *strings.Builder, steps []plan.PlanStep) {
	for _, s := range steps {
		fmt.Fprintf(b, "**%d. %s**\n", s.Order, escapeMarkdown(s.Description))
		if s.Command != "" {
			fmt.Fprintf(b, "```bash\n%s\n```\n", s.Command)
		}
		if s.Expected != "" {
			fmt.Fprintf(b, "- Expected: %s\n", escapeMarkdown(s.Expected))
		}
		if s.EstimatedTime > 0 {
			fmt.Fprintf(b, "- Est. time: %s\n", mdFormatDuration(s.EstimatedTime))
		}
		fmt.Fprintln(b)
	}
}

func mdFormatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%d hour%s %d minute%s", h, mdPlural(h), m, mdPlural(m))
	case h > 0:
		return fmt.Sprintf("%d hour%s", h, mdPlural(h))
	default:
		return fmt.Sprintf("%d minute%s", m, mdPlural(m))
	}
}

func mdPlural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func mdFindingsBySeverity(findings []recommendation.Finding, sev recommendation.FindingSeverity) []recommendation.Finding {
	var out []recommendation.Finding
	for _, f := range findings {
		if f.Severity == sev {
			out = append(out, f)
		}
	}
	return out
}

func escapeMarkdown(in string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(in)
}
