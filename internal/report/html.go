package report

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/plan"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
)

var reportTemplate = template.Must(template.New("report").Funcs(template.FuncMap{
	"add": func(a, b int) int { return a + b },
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Kubernetes Upgrade Advisor</title>
<style>
:root { --bg:#f5f6f8; --panel:#ffffff; --text:#1f2937; --muted:#6b7280; --line:#d1d5db; --accent:#0f766e; --green:#16a34a; --yellow:#ca8a04; --red:#dc2626; }
body { margin:0; font-family: ui-sans-serif, -apple-system, Segoe UI, sans-serif; background:var(--bg); color:var(--text); }
.container { max-width:960px; margin:24px auto; padding:0 16px; }
.card { background:var(--panel); border:1px solid var(--line); border-radius:10px; padding:16px; margin-bottom:16px; }
h1,h2,h3 { margin:0 0 12px 0; }
.meta { color:var(--muted); font-size:14px; }
table { width:100%; border-collapse:collapse; }
th, td { text-align:left; padding:8px; border-bottom:1px solid var(--line); vertical-align:top; }
.badge { display:inline-block; padding:2px 8px; border:1px solid var(--line); border-radius:999px; font-size:12px; }
.decision-go { color:var(--green); font-weight:bold; }
.decision-caution { color:var(--yellow); font-weight:bold; }
.decision-stop { color:var(--red); font-weight:bold; }
code { background:#eef2f7; padding:2px 4px; border-radius:4px; font-size:13px; }
pre { background:#1e293b; color:#e2e8f0; padding:12px; border-radius:6px; overflow-x:auto; }
pre code { background:transparent; padding:0; }
.severity-blocker { color:var(--red); font-weight:bold; }
.severity-warning { color:var(--yellow); font-weight:bold; }
.severity-info { color:var(--muted); }
.step { margin-bottom:16px; padding:12px; background:#f8fafc; border-radius:6px; border-left:3px solid var(--accent); }
.step-number { font-weight:bold; color:var(--accent); }
.component-table th { background:#f1f5f9; }
</style>
</head>
<body>
  <div class="container">
    <div class="card">
      <h1>🔍 Kubernetes Upgrade Advisor</h1>
      {{ if .Decision }}
      <p class="{{ .DecisionClass }}">{{ .DecisionEmoji }} {{ .DecisionLabel }}</p>
      {{ if .Confidence }}<p><strong>Confidence:</strong> {{ .Confidence }}%</p>{{ end }}
      {{ else }}
      <p><strong>Readiness:</strong> {{ .Readiness }}</p>
      {{ end }}
      <table style="width:auto;">
        <tr><td><strong>Current</strong></td><td><code>{{ .Current }}</code></td></tr>
        <tr><td><strong>Destination</strong></td><td><code>{{ if .Destination }}{{ .Destination }}{{ else }}n/a{{ end }}</code></td></tr>
        {{ if .UpgradePath }}<tr><td><strong>Upgrade Path</strong></td><td>{{ .UpgradePath }}</td></tr>{{ end }}
        {{ if .EstimatedTime }}<tr><td><strong>Est. Time</strong></td><td>{{ .EstimatedTime }}</td></tr>{{ end }}
      </table>
      <div class="meta" style="margin-top:12px;">Generated {{ .GeneratedAt }} • ID: {{ .AssessmentID }}</div>
    </div>

    {{ if .HasBlockers }}
    <div class="card">
      <h2>🔴 Blockers ({{ .BlockerCount }})</h2>
      <p style="color:var(--muted);">Must fix before upgrading</p>
      {{ range .Blockers }}
      <div class="step">
        <p><span class="severity-blocker">[{{ .Category }}]</span> {{ .Summary }}</p>
        {{ if .Resource }}<p><code>{{ .ResourceDisplay }}</code></p>{{ end }}
        {{ if .Impact }}<p><strong>Impact:</strong> {{ .Impact.Level }} — {{ .Impact.Explanation }}</p>{{ end }}
        {{ if .Action }}<p><strong>Action:</strong> {{ .Action.Description }}</p>
        {{ if .Action.Command }}<pre><code>{{ .Action.Command }}</code></pre>{{ end }}
        {{ else if .Remediation }}<p><strong>Remediation:</strong> {{ .Remediation }}</p>{{ end }}
        {{ if .IfIgnored }}<p><strong>If ignored:</strong> {{ .IfIgnored }}</p>{{ end }}
      </div>
      {{ end }}
    </div>
    {{ else }}
    <div class="card">
      <h2>✅ Blockers (0)</h2>
      <p>None — no issues blocking upgrade.</p>
    </div>
    {{ end }}

    {{ if .HasWarnings }}
    <div class="card">
      <h2>⚠️ Warnings ({{ .WarningCount }})</h2>
      {{ range .Warnings }}
      <div class="step">
        <p><span class="severity-warning">[{{ .Category }}]</span> {{ .Summary }}</p>
        {{ if .Resource }}<p><code>{{ .ResourceDisplay }}</code></p>{{ end }}
        {{ if .Impact }}<p><strong>Impact:</strong> {{ .Impact.Level }} — {{ .Impact.Explanation }}</p>{{ end }}
        {{ if .Action }}<p><strong>Action:</strong> {{ .Action.Description }}</p>
        {{ if .Action.Command }}<pre><code>{{ .Action.Command }}</code></pre>{{ end }}
        {{ else if .Remediation }}<p><strong>Remediation:</strong> {{ .Remediation }}</p>{{ end }}
      </div>
      {{ end }}
    </div>
    {{ end }}

    {{ if .HasLimitations }}
    <div class="card">
      <h2>📋 Evidence Gaps ({{ .LimitationCount }})</h2>
      <table>
        <thead><tr><th>Code</th><th>Summary</th><th>Impact</th></tr></thead>
        <tbody>
          {{ range .Limitations }}
          <tr><td><code>{{ .Code }}</code></td><td>{{ .Summary }}</td><td>{{ .Impact }}</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>
    {{ end }}

    {{ if .Evidence }}
    <div class="card">
      <h2>📊 Evidence Summary</h2>
      <table class="component-table">
        <thead><tr><th>Resource</th><th>Status</th></tr></thead>
        <tbody>
          <tr><td>Deployments</td><td>{{ .Evidence.Deployments.Healthy }}/{{ .Evidence.Deployments.Total }} healthy</td></tr>
          <tr><td>StatefulSets</td><td>{{ .Evidence.StatefulSets.Healthy }}/{{ .Evidence.StatefulSets.Total }} healthy</td></tr>
          <tr><td>DaemonSets</td><td>{{ .Evidence.DaemonSets.Healthy }}/{{ .Evidence.DaemonSets.Total }} healthy</td></tr>
          <tr><td>Nodes</td><td>{{ .Evidence.Nodes.Healthy }}/{{ .Evidence.Nodes.Total }} healthy</td></tr>
          <tr><td>PVCs</td><td>{{ .Evidence.PVCs.Healthy }}/{{ .Evidence.PVCs.Total }} healthy</td></tr>
          <tr><td>CRDs</td><td>{{ .Evidence.CRDs.Total }}</td></tr>
          <tr><td>Deprecated APIs</td><td>{{ .Evidence.DeprecatedAPIs }} found</td></tr>
        </tbody>
      </table>
      {{ if .Evidence.Components }}
      <h3 style="margin-top:16px;">Detected Components</h3>
      <table class="component-table">
        <thead><tr><th>Component</th><th>Version</th><th>Status</th></tr></thead>
        <tbody>
          {{ range .Evidence.Components }}
          <tr><td>{{ .Name }}</td><td>{{ if .Version }}{{ .Version }}{{ else }}unknown{{ end }}</td><td>{{ .Status }}</td></tr>
          {{ end }}
        </tbody>
      </table>
      {{ end }}
    </div>
    {{ end }}

    {{ if .HasUpgradeSteps }}
    <div class="card">
      <h2>📝 Recommended Upgrade Plan</h2>
      {{ range .UpgradeSteps }}
      <div class="step">
        <p><span class="step-number">Step {{ .Order }}:</span> {{ .Description }}</p>
        {{ if .Command }}<pre><code>{{ .Command }}</code></pre>{{ end }}
        {{ if .Expected }}<p><strong>Expected:</strong> {{ .Expected }}</p>{{ end }}
        {{ if .EstimatedTime }}<p><strong>Est. time:</strong> {{ .EstimatedTime }}</p>{{ end }}
      </div>
      {{ end }}
    </div>
    {{ end }}

    {{ if .HasValidationSteps }}
    <div class="card">
      <h2>✔️ Post-Upgrade Validation</h2>
      {{ range .ValidationSteps }}
      <div class="step">
        <p><span class="step-number">Step {{ .Order }}:</span> {{ .Description }}</p>
        {{ if .Command }}<pre><code>{{ .Command }}</code></pre>{{ end }}
        {{ if .Expected }}<p><strong>Expected:</strong> {{ .Expected }}</p>{{ end }}
      </div>
      {{ end }}
    </div>
    {{ end }}

    {{ if .RollbackGuidance }}
    <div class="card">
      <h2>↩️ Rollback Guidance</h2>
      <p>{{ .RollbackGuidance }}</p>
    </div>
    {{ end }}

    <div class="card">
      {{ if .Decision }}
      <p class="{{ .DecisionClass }}"><strong>Overall Decision:</strong> {{ .DecisionEmoji }} {{ .DecisionLabel }}</p>
      {{ end }}
      <div class="meta">Generated {{ .GeneratedAt }} • ID: {{ .AssessmentID }}</div>
    </div>
  </div>
</body>
</html>
`))

type htmlFinding struct {
	Category        recommendation.FindingCategory
	Summary         string
	Resource        *recommendation.ResourceRef
	ResourceDisplay string
	Impact          *recommendation.FindingImpact
	Action          *recommendation.ActionItem
	Remediation     string
	IfIgnored       string
}

type htmlStep struct {
	Order         int
	Description   string
	Command       string
	Expected      string
	EstimatedTime string
}

// RenderHTML renders a self-contained, escaped HTML report.
func RenderHTML(doc Document) ([]byte, error) {
	blockers := htmlFindingsBySeverity(doc.Findings, recommendation.SeverityBlocker)
	warnings := htmlFindingsBySeverity(doc.Findings, recommendation.SeverityWarning)

	var upgradePath string
	if len(doc.Path) > 0 {
		steps := make([]string, 0, len(doc.Path)+1)
		steps = append(steps, doc.Path[0].From)
		for _, stage := range doc.Path {
			steps = append(steps, stage.To)
		}
		upgradePath = strings.Join(steps, " → ")
	} else if doc.Destination != "" {
		upgradePath = fmt.Sprintf("%s → %s", doc.Current, doc.Destination)
	}

	var estTime string
	if doc.UpgradePlan != nil && doc.UpgradePlan.EstimatedTime > 0 {
		estTime = htmlFormatDuration(doc.UpgradePlan.EstimatedTime)
	}

	var confidence *int
	if doc.Confidence != nil {
		confidence = &doc.Confidence.Percentage
	}

	decisionClass, decisionEmoji, decisionLabel := htmlDecisionDisplay(doc.Decision)

	var upgradeSteps, validationSteps []htmlStep
	var rollbackGuidance string
	if doc.UpgradePlan != nil {
		upgradeSteps = toHTMLSteps(doc.UpgradePlan.Steps)
		validationSteps = toHTMLSteps(doc.UpgradePlan.ValidationSteps)
		rollbackGuidance = doc.UpgradePlan.RollbackGuidance
	}

	var out bytes.Buffer
	if err := reportTemplate.Execute(&out, struct {
		AssessmentID      string
		GeneratedAt       string
		Current           string
		Destination       string
		Readiness         string
		Risk              string
		Redacted          bool
		Decision          string
		DecisionClass     string
		DecisionEmoji     string
		DecisionLabel     string
		Confidence        *int
		UpgradePath       string
		EstimatedTime     string
		Blockers          []htmlFinding
		HasBlockers       bool
		BlockerCount      int
		Warnings          []htmlFinding
		HasWarnings       bool
		WarningCount      int
		Limitations       []recommendation.Limitation
		HasLimitations    bool
		LimitationCount   int
		Evidence          *recommendation.EvidenceSummary
		UpgradeSteps      []htmlStep
		HasUpgradeSteps   bool
		ValidationSteps   []htmlStep
		HasValidationSteps bool
		RollbackGuidance  string
	}{
		AssessmentID:       doc.AssessmentID,
		GeneratedAt:        doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Current:            doc.Current,
		Destination:        doc.Destination,
		Readiness:          string(doc.Readiness),
		Risk:               string(doc.Risk),
		Redacted:           doc.Redacted,
		Decision:           string(doc.Decision),
		DecisionClass:      decisionClass,
		DecisionEmoji:      decisionEmoji,
		DecisionLabel:      decisionLabel,
		Confidence:         confidence,
		UpgradePath:        upgradePath,
		EstimatedTime:      estTime,
		Blockers:           blockers,
		HasBlockers:        len(blockers) > 0,
		BlockerCount:       len(blockers),
		Warnings:           warnings,
		HasWarnings:        len(warnings) > 0,
		WarningCount:       len(warnings),
		Limitations:        doc.Limitations,
		HasLimitations:     len(doc.Limitations) > 0,
		LimitationCount:    len(doc.Limitations),
		Evidence:           doc.Evidence,
		UpgradeSteps:       upgradeSteps,
		HasUpgradeSteps:    len(upgradeSteps) > 0,
		ValidationSteps:    validationSteps,
		HasValidationSteps: len(validationSteps) > 0,
		RollbackGuidance:   rollbackGuidance,
	}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func htmlDecisionDisplay(d recommendation.Decision) (class, emoji, label string) {
	switch d {
	case recommendation.DecisionGo:
		return "decision-go", "🟢", "PROCEED WITH UPGRADE"
	case recommendation.DecisionCaution:
		return "decision-caution", "🟡", "PROCEED WITH CAUTION"
	case recommendation.DecisionStop:
		return "decision-stop", "🔴", "DO NOT PROCEED"
	default:
		return "", "", string(d)
	}
}

func htmlFindingsBySeverity(findings []recommendation.Finding, sev recommendation.FindingSeverity) []htmlFinding {
	var out []htmlFinding
	for _, f := range findings {
		if f.Severity == sev {
			var resourceDisplay string
			if f.Resource != nil {
				ns := f.Resource.Namespace
				if ns == "" {
					ns = "-"
				}
				resourceDisplay = fmt.Sprintf("%s/%s", ns, f.Resource.Name)
				if f.Resource.Kind != "" {
					resourceDisplay += fmt.Sprintf(" (%s)", f.Resource.Kind)
				}
			}
			out = append(out, htmlFinding{
				Category:        f.Category,
				Summary:         f.Summary,
				Resource:        f.Resource,
				ResourceDisplay: resourceDisplay,
				Impact:          f.Impact,
				Action:          f.Action,
				Remediation:     f.Remediation,
				IfIgnored:       f.IfIgnored,
			})
		}
	}
	return out
}

func toHTMLSteps(steps []plan.PlanStep) []htmlStep {
	out := make([]htmlStep, 0, len(steps))
	for _, s := range steps {
		var est string
		if s.EstimatedTime > 0 {
			est = htmlFormatDuration(s.EstimatedTime)
		}
		out = append(out, htmlStep{
			Order:         s.Order,
			Description:   s.Description,
			Command:       s.Command,
			Expected:      s.Expected,
			EstimatedTime: est,
		})
	}
	return out
}

func htmlFormatDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%d hour%s %d minute%s", h, htmlPlural(h), m, htmlPlural(m))
	case h > 0:
		return fmt.Sprintf("%d hour%s", h, htmlPlural(h))
	default:
		return fmt.Sprintf("%d minute%s", m, htmlPlural(m))
	}
}

func htmlPlural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
