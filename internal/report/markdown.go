package report

import (
	"fmt"
	"strings"
)

// RenderMarkdown renders a deterministic Markdown report.
func RenderMarkdown(doc Document) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# KUA Assessment %s\n\n", escapeMarkdown(doc.AssessmentID))
	fmt.Fprintf(&b, "- Generated: `%s`\n", doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(&b, "- Current: `%s`\n", escapeMarkdown(doc.Current))
	fmt.Fprintf(&b, "- Destination: `%s`\n", escapeMarkdown(defaultString(doc.Destination, "n/a")))
	fmt.Fprintf(&b, "- Readiness: `%s`\n", doc.Readiness)
	fmt.Fprintf(&b, "- Risk: `%s`\n", doc.Risk)
	fmt.Fprintf(&b, "- Redacted: `%t`\n\n", doc.Redacted)

	fmt.Fprintln(&b, "## Path")
	if len(doc.Path) == 0 {
		fmt.Fprintln(&b, "No upgrade path available.")
		fmt.Fprintln(&b)
	} else {
		for i, step := range doc.Path {
			fmt.Fprintf(&b, "%d. `%s -> %s` (providerValid=%t)\n", i+1, escapeMarkdown(step.From), escapeMarkdown(step.To), step.IsProviderValid)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "## Findings")
	if len(doc.Findings) == 0 {
		fmt.Fprintln(&b, "No findings.")
		fmt.Fprintln(&b)
	} else {
		for _, f := range doc.Findings {
			fmt.Fprintf(&b, "- **[%s] %s**: %s", f.Severity, escapeMarkdown(f.ID), escapeMarkdown(f.Summary))
			if f.Resource != nil {
				fmt.Fprintf(&b, " (`%s/%s`)", escapeMarkdown(defaultString(f.Resource.Namespace, "-")), escapeMarkdown(defaultString(f.Resource.Name, "-")))
			}
			fmt.Fprintln(&b)
			if f.Detail != "" {
				fmt.Fprintf(&b, "  - Detail: `%s`\n", escapeMarkdown(f.Detail))
			}
			if f.Remediation != "" {
				fmt.Fprintf(&b, "  - Remediation: %s\n", escapeMarkdown(f.Remediation))
			}
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "## Limitations")
	if len(doc.Limitations) == 0 {
		fmt.Fprintln(&b, "No limitations.")
		fmt.Fprintln(&b)
	} else {
		for _, l := range doc.Limitations {
			fmt.Fprintf(&b, "- `%s`: %s\n", escapeMarkdown(l.Code), escapeMarkdown(l.Summary))
		}
	}

	return []byte(b.String()), nil
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
