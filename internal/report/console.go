package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/plan"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
)

const consoleWidth = 56

// decisionDisplay maps a traffic light decision to its emoji and label.
func decisionDisplay(d recommendation.Decision) string {
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

// RenderConsole renders a focused, operator-friendly console report.
func RenderConsole(doc Document) ([]byte, error) {
	var b strings.Builder

	// ── header ──────────────────────────────────────────────
	fmt.Fprintln(&b, strings.Repeat("═", consoleWidth))
	fmt.Fprintln(&b, " KUBERNETES UPGRADE ADVISOR")
	fmt.Fprintln(&b, strings.Repeat("═", consoleWidth))

	// ── decision and confidence ──────────────────────────────
	if doc.Decision != "" {
		fmt.Fprintf(&b, " Recommendation:  %s\n", decisionDisplay(doc.Decision))
		if doc.Confidence != nil {
			fmt.Fprintf(&b, " Confidence:      %d%%\n", doc.Confidence.Percentage)
		}
	} else {
		fmt.Fprintf(&b, " Readiness:       %s\n", doc.Readiness)
	}

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, " Current:     %s\n", doc.Current)
	dest := defaultString(doc.Destination, "n/a")
	fmt.Fprintf(&b, " Destination: %s  (risk: %s)\n", dest, doc.Risk)
	if len(doc.Path) > 0 {
		steps := make([]string, 0, len(doc.Path)+1)
		steps = append(steps, doc.Path[0].From)
		for _, stage := range doc.Path {
			steps = append(steps, stage.To)
		}
		fmt.Fprintf(&b, " Path:        %s\n", strings.Join(steps, " → "))
	}
	if doc.UpgradePlan != nil && doc.UpgradePlan.EstimatedTime > 0 {
		fmt.Fprintf(&b, " Est. Time:   %s\n", doc.UpgradePlan.EstimatedTime.Round(time.Minute))
	}

	// ── blockers ────────────────────────────────────────────
	blockers := findingsBySeverity(doc.Findings, recommendation.SeverityBlocker)
	if len(blockers) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "── BLOCKERS (%d)  must fix before upgrading %s\n",
			len(blockers), strings.Repeat("─", max(0, consoleWidth-43-len(fmt.Sprint(len(blockers))))))
		for i, f := range blockers {
			writeFinding(&b, i, f)
		}
	}

	// ── warnings ────────────────────────────────────────────
	warnings := findingsBySeverity(doc.Findings, recommendation.SeverityWarning)
	if len(warnings) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "── WARNINGS (%d) %s\n",
			len(warnings), strings.Repeat("─", max(0, consoleWidth-16-len(fmt.Sprint(len(warnings))))))
		for i, f := range warnings {
			writeFinding(&b, i, f)
		}
	}

	// ── evidence gaps ───────────────────────────────────────
	if len(doc.Limitations) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "── EVIDENCE GAPS (%d) %s\n",
			len(doc.Limitations), strings.Repeat("─", max(0, consoleWidth-21-len(fmt.Sprint(len(doc.Limitations))))))
		for i, l := range doc.Limitations {
			fmt.Fprintf(&b, " [%d] %s: %s\n", i+1, l.Code, l.Summary)
			if l.Impact != "" {
				fmt.Fprintf(&b, "     %s\n", l.Impact)
			}
		}
	}

	// ── all-clear ───────────────────────────────────────────
	if len(blockers) == 0 && len(warnings) == 0 && len(doc.Limitations) == 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, " No blockers, warnings, or evidence gaps found.")
	}

	// ── evidence summary ─────────────────────────────────────
	if doc.Evidence != nil {
		writeEvidenceSummary(&b, doc.Evidence)
	}

	// ── recommended upgrade plan ──────────────────────────────
	if doc.UpgradePlan != nil && len(doc.UpgradePlan.Steps) > 0 {
		writePlanSteps(&b, "RECOMMENDED UPGRADE PLAN", doc.UpgradePlan.Steps)
	}

	// ── post-upgrade validation ───────────────────────────────
	if doc.UpgradePlan != nil && len(doc.UpgradePlan.ValidationSteps) > 0 {
		writePlanSteps(&b, "POST-UPGRADE VALIDATION", doc.UpgradePlan.ValidationSteps)
	}

	// ── rollback guidance ─────────────────────────────────────
	if doc.UpgradePlan != nil && doc.UpgradePlan.RollbackGuidance != "" {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "── ROLLBACK GUIDANCE %s\n", strings.Repeat("─", max(0, consoleWidth-21)))
		fmt.Fprintf(&b, " %s\n", doc.UpgradePlan.RollbackGuidance)
	}

	// ── component version input hint (non-interactive) ──────
	if doc.ComponentVersionOverrides != nil && len(doc.ComponentVersionOverrides.Components) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, " Component versions needed:")
		for _, c := range doc.ComponentVersionOverrides.Components {
			fmt.Fprintf(&b, "   • %s (%s): %s\n", c.Name, c.ID, c.Reason)
		}
		fmt.Fprintf(&b, "   Run: %s\n", doc.ComponentVersionOverrides.RerunCommand)
	}

	// ── footer ──────────────────────────────────────────────
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, strings.Repeat("─", consoleWidth))
	fmt.Fprintf(&b, " Generated: %s  |  ID: %s\n",
		doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"),
		doc.AssessmentID)
	fmt.Fprintln(&b, strings.Repeat("═", consoleWidth))

	return []byte(b.String()), nil
}

func writeFinding(b *strings.Builder, i int, f recommendation.Finding) {
	fmt.Fprintf(b, " [%d] [%s] %s\n", i+1, f.Category, f.Summary)
	if f.Resource != nil {
		ns := defaultString(f.Resource.Namespace, "-")
		fmt.Fprintf(b, "     %s/%s", ns, f.Resource.Name)
		if f.Resource.Kind != "" {
			fmt.Fprintf(b, " (%s)", f.Resource.Kind)
		}
		fmt.Fprintln(b)
	}
	if f.Impact != nil {
		fmt.Fprintf(b, "     Impact: %s — %s\n", f.Impact.Level, f.Impact.Explanation)
	}
	if f.Action != nil {
		fmt.Fprintf(b, "     Action: %s\n", f.Action.Description)
		if f.Action.Command != "" {
			fmt.Fprintf(b, "       $ %s\n", f.Action.Command)
		}
	} else if f.Remediation != "" {
		fmt.Fprintf(b, "     → %s\n", f.Remediation)
	}
	if f.IfIgnored != "" {
		fmt.Fprintf(b, "     If ignored: %s\n", f.IfIgnored)
	}
}

func writeEvidenceSummary(b *strings.Builder, e *recommendation.EvidenceSummary) {
	fmt.Fprintln(b)
	fmt.Fprintf(b, "── EVIDENCE SUMMARY %s\n", strings.Repeat("─", max(0, consoleWidth-20)))
	fmt.Fprintf(b, " Deployments:   %d/%d healthy\n", e.Deployments.Healthy, e.Deployments.Total)
	fmt.Fprintf(b, " StatefulSets:  %d/%d healthy\n", e.StatefulSets.Healthy, e.StatefulSets.Total)
	fmt.Fprintf(b, " DaemonSets:    %d/%d healthy\n", e.DaemonSets.Healthy, e.DaemonSets.Total)
	fmt.Fprintf(b, " Nodes:         %d/%d healthy\n", e.Nodes.Healthy, e.Nodes.Total)
	fmt.Fprintf(b, " PVCs:          %d/%d healthy\n", e.PVCs.Healthy, e.PVCs.Total)
	fmt.Fprintf(b, " CRDs:          %d\n", e.CRDs.Total)
	fmt.Fprintf(b, " Deprecated APIs found: %d\n", e.DeprecatedAPIs)
	if len(e.Components) > 0 {
		fmt.Fprintln(b, " Components:")
		for _, c := range e.Components {
			fmt.Fprintf(b, "   • %s %s — %s\n", c.Name, defaultString(c.Version, "unknown"), c.Status)
		}
	}
}

func writePlanSteps(b *strings.Builder, title string, steps []plan.PlanStep) {
	fmt.Fprintln(b)
	fmt.Fprintf(b, "── %s %s\n", title, strings.Repeat("─", max(0, consoleWidth-3-len(title))))
	for _, s := range steps {
		fmt.Fprintf(b, " [%d] %s\n", s.Order, s.Description)
		if s.Command != "" {
			fmt.Fprintf(b, "     $ %s\n", s.Command)
		}
		if s.Expected != "" {
			fmt.Fprintf(b, "     Expected: %s\n", s.Expected)
		}
		if s.EstimatedTime > 0 {
			fmt.Fprintf(b, "     Est. time: %s\n", s.EstimatedTime.Round(time.Second))
		}
	}
}

func findingsBySeverity(findings []recommendation.Finding, sev recommendation.FindingSeverity) []recommendation.Finding {
	var out []recommendation.Finding
	for _, f := range findings {
		if f.Severity == sev {
			out = append(out, f)
		}
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
