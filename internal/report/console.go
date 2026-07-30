package report

import (
	"fmt"
	"strings"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
)

const consoleWidth = 56

// RenderConsole renders a focused, operator-friendly console report.
func RenderConsole(doc Document) ([]byte, error) {
	var b strings.Builder

	// ── header ──────────────────────────────────────────────
	fmt.Fprintln(&b, strings.Repeat("═", consoleWidth))
	fmt.Fprintf(&b, " UPGRADE ASSESSMENT  ·  %s\n", doc.Readiness)
	fmt.Fprintln(&b, strings.Repeat("═", consoleWidth))

	// ── summary ─────────────────────────────────────────────
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

	// ── blockers ────────────────────────────────────────────
	blockers := findingsBySeverity(doc.Findings, recommendation.SeverityBlocker)
	if len(blockers) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "── BLOCKERS (%d)  must fix before upgrading %s\n",
			len(blockers), strings.Repeat("─", max(0, consoleWidth-43-len(fmt.Sprint(len(blockers))))))
		for i, f := range blockers {
			fmt.Fprintf(&b, " [%d] [%s] %s\n", i+1, f.Category, f.Summary)
			if f.Resource != nil {
				ns := defaultString(f.Resource.Namespace, "-")
				fmt.Fprintf(&b, "     %s/%s", ns, f.Resource.Name)
				if f.Resource.Kind != "" {
					fmt.Fprintf(&b, " (%s)", f.Resource.Kind)
				}
				fmt.Fprintln(&b)
			}
			if f.Remediation != "" {
				fmt.Fprintf(&b, "     → %s\n", f.Remediation)
			}
		}
	}

	// ── warnings ────────────────────────────────────────────
	warnings := findingsBySeverity(doc.Findings, recommendation.SeverityWarning)
	if len(warnings) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintf(&b, "── WARNINGS (%d) %s\n",
			len(warnings), strings.Repeat("─", max(0, consoleWidth-16-len(fmt.Sprint(len(warnings))))))
		for i, f := range warnings {
			fmt.Fprintf(&b, " [%d] [%s] %s\n", i+1, f.Category, f.Summary)
			if f.Resource != nil {
				ns := defaultString(f.Resource.Namespace, "-")
				fmt.Fprintf(&b, "     %s/%s\n", ns, f.Resource.Name)
			}
			if f.Remediation != "" {
				fmt.Fprintf(&b, "     → %s\n", f.Remediation)
			}
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
