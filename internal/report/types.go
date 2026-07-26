package report

import (
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
)

// RenderFormat identifies the supported report output format.
type RenderFormat string

const (
	FormatJSON     RenderFormat = "json"
	FormatConsole  RenderFormat = "console"
	FormatMarkdown RenderFormat = "markdown"
	FormatHTML     RenderFormat = "html"
)

// Document is the report model rendered to output formats.
type Document struct {
	SchemaVersion string                        `json:"schemaVersion"`
	AssessmentID  string                        `json:"assessmentId"`
	GeneratedAt   time.Time                     `json:"generatedAt"`
	Redacted      bool                          `json:"redacted"`
	Current       string                        `json:"currentVersion"`
	Destination   string                        `json:"destination,omitempty"`
	Readiness     recommendation.ReadinessState `json:"readiness"`
	Risk          recommendation.RiskLevel      `json:"risk"`
	Path          []recommendation.UpgradeStage `json:"path,omitempty"`
	Findings      []recommendation.Finding      `json:"findings"`
	Limitations   []recommendation.Limitation   `json:"limitations"`
}

// RenderOptions controls report rendering behavior.
type RenderOptions struct {
	Format   RenderFormat
	Redacted bool
}
