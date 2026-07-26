package report

import (
	"bytes"
	"html/template"
)

var reportTemplate = template.Must(template.New("report").Funcs(template.FuncMap{
  "add": func(a, b int) int { return a + b },
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>KUA Assessment {{ .AssessmentID }}</title>
<style>
:root { --bg:#f5f6f8; --panel:#ffffff; --text:#1f2937; --muted:#6b7280; --line:#d1d5db; --accent:#0f766e; }
body { margin:0; font-family: ui-sans-serif, -apple-system, Segoe UI, sans-serif; background:var(--bg); color:var(--text); }
.container { max-width:960px; margin:24px auto; padding:0 16px; }
.card { background:var(--panel); border:1px solid var(--line); border-radius:10px; padding:16px; margin-bottom:16px; }
h1,h2 { margin:0 0 12px 0; }
.meta { color:var(--muted); font-size:14px; }
table { width:100%; border-collapse:collapse; }
th, td { text-align:left; padding:8px; border-bottom:1px solid var(--line); vertical-align:top; }
.badge { display:inline-block; padding:2px 8px; border:1px solid var(--line); border-radius:999px; font-size:12px; }
code { background:#eef2f7; padding:2px 4px; border-radius:4px; }
</style>
</head>
<body>
  <div class="container">
    <div class="card">
      <h1>KUA Assessment {{ .AssessmentID }}</h1>
      <div class="meta">Generated {{ .GeneratedAt }} • Redacted {{ .Redacted }}</div>
      <p><strong>Current:</strong> <code>{{ .Current }}</code> &nbsp; <strong>Destination:</strong> <code>{{ if .Destination }}{{ .Destination }}{{ else }}n/a{{ end }}</code></p>
      <p><span class="badge">Readiness {{ .Readiness }}</span> <span class="badge">Risk {{ .Risk }}</span></p>
    </div>

    <div class="card">
      <h2>Path</h2>
      <table>
        <thead><tr><th>#</th><th>From</th><th>To</th><th>Provider Valid</th></tr></thead>
        <tbody>
          {{ range $i, $s := .Path }}
          <tr><td>{{ add $i 1 }}</td><td><code>{{ $s.From }}</code></td><td><code>{{ $s.To }}</code></td><td>{{ $s.IsProviderValid }}</td></tr>
          {{ else }}
          <tr><td colspan="4">No upgrade path available.</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>

    <div class="card">
      <h2>Findings</h2>
      <table>
        <thead><tr><th>ID</th><th>Severity</th><th>Summary</th><th>Resource</th></tr></thead>
        <tbody>
          {{ range .Findings }}
          <tr>
            <td><code>{{ .ID }}</code></td>
            <td>{{ .Severity }}</td>
            <td>{{ .Summary }}</td>
            <td>{{ if .Resource }}{{ .Resource.Namespace }}/{{ .Resource.Name }}{{ else }}-{{ end }}</td>
          </tr>
          {{ else }}
          <tr><td colspan="4">No findings.</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>

    <div class="card">
      <h2>Limitations</h2>
      <table>
        <thead><tr><th>Code</th><th>Summary</th></tr></thead>
        <tbody>
          {{ range .Limitations }}
          <tr><td><code>{{ .Code }}</code></td><td>{{ .Summary }}</td></tr>
          {{ else }}
          <tr><td colspan="2">No limitations.</td></tr>
          {{ end }}
        </tbody>
      </table>
    </div>
  </div>
</body>
</html>
`))

// RenderHTML renders a self-contained, escaped HTML report.
func RenderHTML(doc Document) ([]byte, error) {
	var out bytes.Buffer
	if err := reportTemplate.Execute(&out, struct {
		AssessmentID string
		GeneratedAt  string
		Current      string
		Destination  string
		Readiness    string
		Risk         string
		Redacted     bool
		Path         any
		Findings     any
		Limitations  any
	}{
		AssessmentID: doc.AssessmentID,
		GeneratedAt:  doc.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z"),
		Current:      doc.Current,
		Destination:  doc.Destination,
		Readiness:    string(doc.Readiness),
		Risk:         string(doc.Risk),
		Redacted:     doc.Redacted,
		Path:         doc.Path,
		Findings:     doc.Findings,
		Limitations:  doc.Limitations,
	}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
