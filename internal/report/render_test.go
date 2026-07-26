package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
)

func sampleDocument() Document {
	return Document{
		SchemaVersion: "kua.assessment.v1",
		AssessmentID:  "assess-123",
		GeneratedAt:   time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC),
		Current:       "1.30.0",
		Destination:   "1.33.12",
		Readiness:     recommendation.ReadinessReady,
		Risk:          recommendation.RiskLow,
		Path: []recommendation.UpgradeStage{
			{From: "1.30.0", To: "1.31.4", IsProviderValid: true},
			{From: "1.31.4", To: "1.32.2", IsProviderValid: true},
			{From: "1.32.2", To: "1.33.12", IsProviderValid: true},
		},
		Findings: []recommendation.Finding{
			{
				ID:       "API_DEPRECATED",
				Severity: recommendation.SeverityWarning,
				Summary:  "api batch/v1beta1 is deprecated <script>alert(1)</script>",
				Resource: &recommendation.ResourceRef{Namespace: "prod", Name: "deploy-a", Kind: "Deployment"},
			},
		},
		Limitations: []recommendation.Limitation{{Code: "INFO_ONLY", Summary: "none"}},
	}
}

func TestRenderDeterministic(t *testing.T) {
	doc := sampleDocument()
	formats := []RenderFormat{FormatJSON, FormatConsole, FormatMarkdown, FormatHTML}
	for _, f := range formats {
		a, err := Render(doc, RenderOptions{Format: f})
		if err != nil {
			t.Fatalf("first render %s failed: %v", f, err)
		}
		b, err := Render(doc, RenderOptions{Format: f})
		if err != nil {
			t.Fatalf("second render %s failed: %v", f, err)
		}
		if string(a) != string(b) {
			t.Fatalf("render not deterministic for %s", f)
		}
	}
}

func TestRedactionPreservesDecisionAndRedactsIdentifiers(t *testing.T) {
	doc := sampleDocument()
	out, err := Render(doc, RenderOptions{Format: FormatJSON, Redacted: true})
	if err != nil {
		t.Fatalf("redacted render failed: %v", err)
	}

	s := string(out)
	if strings.Contains(s, "prod") || strings.Contains(s, "deploy-a") {
		t.Fatalf("raw identifiers leaked in redacted output: %s", s)
	}
	if !strings.Contains(s, "\"readiness\": \"READY\"") || !strings.Contains(s, "\"risk\": \"LOW\"") {
		t.Fatalf("decision fields changed unexpectedly: %s", s)
	}
}

func TestHostileInputEscapedInMarkdownAndHTML(t *testing.T) {
	doc := sampleDocument()

	md, err := Render(doc, RenderOptions{Format: FormatMarkdown})
	if err != nil {
		t.Fatalf("markdown render failed: %v", err)
	}
	html, err := Render(doc, RenderOptions{Format: FormatHTML})
	if err != nil {
		t.Fatalf("html render failed: %v", err)
	}

	mds := string(md)
	htmls := string(html)
	if strings.Contains(mds, "<script>") {
		t.Fatalf("markdown contains unescaped script tag: %s", mds)
	}
	if strings.Contains(htmls, "<script>alert(1)</script>") {
		t.Fatalf("html contains executable script payload: %s", htmls)
	}
	if strings.Contains(htmls, "https://") || strings.Contains(htmls, "http://") {
		t.Fatalf("html must be self-contained without remote assets: %s", htmls)
	}
}

func TestWriteAtomicNoOverwrite(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "report.json")

	if err := WriteAtomic(out, []byte("one")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if err := WriteAtomic(out, []byte("two")); err == nil {
		t.Fatalf("expected overwrite protection error")
	}

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output failed: %v", err)
	}
	if string(b) != "one" {
		t.Fatalf("file content changed unexpectedly: %q", string(b))
	}
}

func TestRenderUnsupportedFormat(t *testing.T) {
	_, err := Render(sampleDocument(), RenderOptions{Format: "xml"})
	if err == nil {
		t.Fatalf("expected unsupported format error")
	}
}
