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
	SchemaVersion             string                            `json:"schemaVersion"`
	AssessmentID              string                            `json:"assessmentId"`
	GeneratedAt               time.Time                         `json:"generatedAt"`
	Redacted                  bool                              `json:"redacted"`
	Current                   string                            `json:"currentVersion"`
	Destination               string                            `json:"destination,omitempty"`
	Readiness                 recommendation.ReadinessState     `json:"readiness"`
	Risk                      recommendation.RiskLevel          `json:"risk"`
	Path                      []recommendation.UpgradeStage     `json:"path,omitempty"`
	Findings                  []recommendation.Finding          `json:"findings"`
	Limitations               []recommendation.Limitation       `json:"limitations"`
	ComponentVersionOverrides *ComponentVersionOverrideTemplate `json:"componentVersionOverrides,omitempty"`
}

// ComponentVersionOverrideTemplate tells operators how to provide missing component versions.
type ComponentVersionOverrideTemplate struct {
	SchemaVersion string                            `json:"schemaVersion"`
	OutputPath    string                            `json:"outputPath"`
	RerunCommand  string                            `json:"rerunCommand"`
	Components    []ComponentVersionOverrideRequest `json:"components"`
}

// ComponentVersionOverrideRequest is one component needing operator version input.
type ComponentVersionOverrideRequest struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ObservedVersions []string `json:"observedVersions,omitempty"`
	Versions         []string `json:"versions"`
	Evidence         string   `json:"evidence"`
	Reason           string   `json:"reason"`
}

// RenderOptions controls report rendering behavior.
type RenderOptions struct {
	Format   RenderFormat
	Redacted bool
}
