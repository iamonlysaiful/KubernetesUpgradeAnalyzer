package report

import (
	"fmt"
	"strings"
)

// RenderConsole renders a deterministic human-readable console report.
func RenderConsole(doc Document) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Assessment: %s\n", doc.AssessmentID)
	fmt.Fprintf(&b, "Generated: %s\n", doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(&b, "Current: %s\n", doc.Current)
	fmt.Fprintf(&b, "Destination: %s\n", defaultString(doc.Destination, "n/a"))
	fmt.Fprintf(&b, "Readiness: %s\n", doc.Readiness)
	fmt.Fprintf(&b, "Risk: %s\n", doc.Risk)
	fmt.Fprintf(&b, "Redacted: %t\n", doc.Redacted)

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Path (%d):\n", len(doc.Path))
	for i, step := range doc.Path {
		fmt.Fprintf(&b, "  %d. %s -> %s (providerValid=%t)\n", i+1, step.From, step.To, step.IsProviderValid)
	}

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Findings (%d):\n", len(doc.Findings))
	for i, f := range doc.Findings {
		line := fmt.Sprintf("  %d. [%s] %s: %s", i+1, f.Severity, f.ID, f.Summary)
		if f.Resource != nil {
			line += fmt.Sprintf(" (%s/%s)", defaultString(f.Resource.Namespace, "-"), defaultString(f.Resource.Name, "-"))
		}
		fmt.Fprintln(&b, line)
	}

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Limitations (%d):\n", len(doc.Limitations))
	for i, l := range doc.Limitations {
		fmt.Fprintf(&b, "  %d. %s: %s\n", i+1, l.Code, l.Summary)
	}

	if doc.ComponentVersionOverrides != nil && len(doc.ComponentVersionOverrides.Components) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "Component version input needed:")
		fmt.Fprintf(&b, "  Write template to: %s\n", doc.ComponentVersionOverrides.OutputPath)
		fmt.Fprintf(&b, "  Rerun: %s\n", doc.ComponentVersionOverrides.RerunCommand)
		for _, component := range doc.ComponentVersionOverrides.Components {
			fmt.Fprintf(&b, "  - %s (%s): %s\n", component.Name, component.ID, component.Reason)
		}
	}

	return []byte(b.String()), nil
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
