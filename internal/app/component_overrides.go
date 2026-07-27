package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/report"
)

const componentOverridesSchemaVersion = "kua.component-overrides.v1"

type componentOverridesFile struct {
	SchemaVersion string                   `json:"schemaVersion"`
	Components    []componentOverrideEntry `json:"components"`
}

func runComponentOverrides(cfg Config, stdout io.Writer, stderr io.Writer) int {
	if cfg.InputPath == "" {
		appErr := UsageError("missing --input for component-overrides")
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	if cfg.OutputPath == "" {
		appErr := UsageError("missing --output for component-overrides")
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	data, err := os.ReadFile(cfg.InputPath)
	if err != nil {
		appErr := ExecutionError("read assessment input failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	var doc report.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		appErr := ExecutionError("parse assessment input failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	if doc.ComponentVersionOverrides == nil || len(doc.ComponentVersionOverrides.Components) == 0 {
		appErr := ExecutionError("assessment does not contain componentVersionOverrides", nil)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	override := componentOverridesFile{
		SchemaVersion: componentOverridesSchemaVersion,
	}
	for _, request := range doc.ComponentVersionOverrides.Components {
		versions := uniqueSorted(request.ObservedVersions)
		if len(versions) == 0 {
			versions = []string{"<fill-version>"}
		}
		override.Components = append(override.Components, componentOverrideEntry{
			ID:       request.ID,
			Versions: versions,
			Evidence: "user-confirmed",
		})
	}
	sort.SliceStable(override.Components, func(i, j int) bool {
		return override.Components[i].ID < override.Components[j].ID
	})
	content, err := json.MarshalIndent(override, "", "  ")
	if err != nil {
		appErr := ExecutionError("render component overrides failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	content = append(content, '\n')
	if err := writeNewFile(cfg.OutputPath, content); err != nil {
		appErr := ExecutionError("write component overrides failed: "+err.Error(), err)
		fmt.Fprintln(stderr, appErr.Message)
		return appErr.Code
	}
	fmt.Fprintf(stdout, "component overrides written to %s\n", cfg.OutputPath)
	return ExitReady
}

func writeNewFile(path string, content []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	return err
}

type componentOverrideEntry struct {
	ID       string   `json:"id"`
	Versions []string `json:"versions"`
	Evidence string   `json:"evidence"`
}

func loadComponentOverrides(path string) (map[string][]string, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file componentOverridesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if file.SchemaVersion != componentOverridesSchemaVersion {
		return nil, fmt.Errorf("schemaVersion must be %s", componentOverridesSchemaVersion)
	}
	out := map[string][]string{}
	for _, entry := range file.Components {
		id := strings.TrimSpace(entry.ID)
		if id == "" {
			return nil, fmt.Errorf("component id is required")
		}
		for _, version := range entry.Versions {
			version = strings.TrimSpace(version)
			if version == "" || strings.Contains(version, "<") || strings.Contains(version, ">") {
				continue
			}
			if version == components.UnknownVersion {
				continue
			}
			out[id] = append(out[id], version)
		}
		out[id] = uniqueSorted(out[id])
	}
	return out, nil
}

func applyComponentOverrides(detections []components.Detection, overrides map[string][]string) []components.Detection {
	if len(overrides) == 0 {
		return detections
	}
	names := componentNames(detections)
	filtered := make([]components.Detection, 0, len(detections))
	for _, detection := range detections {
		if _, ok := overrides[detection.ComponentID]; ok && (detection.Status == components.StatusUnknown || detection.Version == components.UnknownVersion) {
			continue
		}
		filtered = append(filtered, detection)
	}
	for id, versions := range overrides {
		name := names[id]
		if name == "" {
			name = id
		}
		for _, version := range versions {
			filtered = append(filtered, components.Detection{
				ComponentID: id,
				Name:        name,
				Version:     version,
				Confidence:  components.ConfidenceMedium,
				Status:      components.StatusFound,
			})
		}
	}
	components.SortDetections(filtered)
	return filtered
}

func componentNames(detections []components.Detection) map[string]string {
	names := map[string]string{}
	for _, detection := range detections {
		if detection.ComponentID != "" && detection.Name != "" {
			names[detection.ComponentID] = detection.Name
		}
	}
	return names
}

func buildComponentOverrideTemplate(detections []components.Detection, cfg Config) *report.ComponentVersionOverrideTemplate {
	requestsByID := map[string]report.ComponentVersionOverrideRequest{}
	for _, detection := range detections {
		if detection.Status != components.StatusUnknown && detection.Version != components.UnknownVersion {
			continue
		}
		req := requestsByID[detection.ComponentID]
		req.ID = detection.ComponentID
		req.Name = detection.Name
		req.Evidence = "user-confirmed"
		req.Reason = "missing or ambiguous component version evidence"
		req.Versions = []string{"<fill-version>"}
		requestsByID[detection.ComponentID] = req
	}
	for _, detection := range detections {
		if detection.Status != components.StatusFound || detection.Version == "" || detection.Version == components.UnknownVersion {
			continue
		}
		req, ok := requestsByID[detection.ComponentID]
		if !ok {
			continue
		}
		req.ObservedVersions = append(req.ObservedVersions, detection.Version)
		requestsByID[detection.ComponentID] = req
	}
	if len(requestsByID) == 0 {
		return nil
	}
	requests := make([]report.ComponentVersionOverrideRequest, 0, len(requestsByID))
	for _, req := range requestsByID {
		req.ObservedVersions = uniqueSorted(req.ObservedVersions)
		requests = append(requests, req)
	}
	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].ID < requests[j].ID
	})
	outputPath := "component-overrides.json"
	return &report.ComponentVersionOverrideTemplate{
		SchemaVersion: componentOverridesSchemaVersion,
		OutputPath:    outputPath,
		RerunCommand:  rerunCommand(cfg, outputPath),
		Components:    requests,
	}
}

func rerunCommand(cfg Config, overridesPath string) string {
	parts := []string{"kua", "analyze"}
	if cfg.Context != "" {
		parts = append(parts, "--context", rerunValue(cfg.Context, "<context>", cfg.Redacted))
	}
	if cfg.Kubeconfig != "" {
		parts = append(parts, "--kubeconfig", rerunValue(cfg.Kubeconfig, "<kubeconfig>", cfg.Redacted))
	}
	if cfg.ProviderSource != "" {
		parts = append(parts, "--provider-source", shellQuote(cfg.ProviderSource))
	}
	if cfg.ProviderEvidence != "" {
		parts = append(parts, "--provider-evidence", rerunValue(cfg.ProviderEvidence, "<provider-evidence>", cfg.Redacted))
	}
	if cfg.Subscription != "" {
		parts = append(parts, "--subscription", "<subscription>")
	}
	if cfg.ResourceGroup != "" {
		parts = append(parts, "--resource-group", "<resource-group>")
	}
	if cfg.ClusterName != "" {
		parts = append(parts, "--cluster-name", "<cluster-name>")
	}
	if cfg.TargetVersion != "" {
		parts = append(parts, "--target-version", shellQuote(cfg.TargetVersion))
	}
	if cfg.Redacted {
		parts = append(parts, "--redacted")
	}
	parts = append(parts, "--component-overrides", shellQuote(overridesPath))
	return strings.Join(parts, " ")
}

func rerunValue(value string, placeholder string, redacted bool) string {
	if redacted {
		return placeholder
	}
	return shellQuote(value)
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.ContainsAny(value, " \t\n'\"$`\\") {
		return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
	}
	return value
}

func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
