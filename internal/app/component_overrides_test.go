package app

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
)

func TestBuildComponentOverrideTemplateIncludesObservedVersionsAndRerun(t *testing.T) {
	template := buildComponentOverrideTemplate([]components.Detection{
		{ComponentID: "coredns", Name: "CoreDNS", Version: components.UnknownVersion, Status: components.StatusUnknown},
		{ComponentID: "coredns", Name: "CoreDNS", Version: "1.9.4-13", Status: components.StatusFound},
		{ComponentID: "coredns", Name: "CoreDNS", Version: "1.9.4-9", Status: components.StatusFound},
	}, Config{
		Context:        "ctx-stage",
		ProviderSource: "azure",
		Redacted:       true,
	})

	if template == nil {
		t.Fatal("template = nil, want override template")
	}
	if template.SchemaVersion != "kua.component-overrides.v1" {
		t.Fatalf("schemaVersion = %q, want kua.component-overrides.v1", template.SchemaVersion)
	}
	if len(template.Components) != 1 {
		t.Fatalf("components = %#v, want one component", template.Components)
	}
	gotVersions := template.Components[0].ObservedVersions
	wantVersions := []string{"1.9.4-13", "1.9.4-9"}
	if !reflect.DeepEqual(gotVersions, wantVersions) {
		t.Fatalf("observedVersions = %#v, want %#v", gotVersions, wantVersions)
	}
	if template.Components[0].Versions[0] != "<fill-version>" {
		t.Fatalf("versions = %#v, want placeholder", template.Components[0].Versions)
	}
	if template.RerunCommand != "kua analyze --context <context> --provider-source azure --redacted --component-overrides component-overrides.json" {
		t.Fatalf("rerunCommand = %q", template.RerunCommand)
	}
}

func TestRunComponentOverridesWritesObservedVersions(t *testing.T) {
	dir := t.TempDir()
	input := dir + "/assessment.json"
	output := dir + "/component-overrides.json"
	err := os.WriteFile(input, []byte(`{
  "schemaVersion": "kua.assessment.v1",
  "assessmentId": "assessment-test",
  "generatedAt": "2026-07-27T04:02:00Z",
  "redacted": true,
  "currentVersion": "1.30.0",
  "readiness": "READY_WITH_WARNINGS",
  "risk": "MEDIUM",
  "findings": [],
  "limitations": [],
  "componentVersionOverrides": {
    "schemaVersion": "kua.component-overrides.v1",
    "outputPath": "component-overrides.json",
    "rerunCommand": "kua analyze --component-overrides component-overrides.json",
    "components": [
      {"id": "coredns", "name": "CoreDNS", "observedVersions": ["1.9.4-9", "1.9.4-13"], "versions": ["<fill-version>"], "evidence": "user-confirmed", "reason": "missing"}
    ]
  }
}`), 0o600)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--input", input, "--output", output, "component-overrides"}, &stdout, &stderr, BuildInfo{})

	if code != ExitReady {
		t.Fatalf("Run(component-overrides) exit code = %d, want %d; stderr=%s", code, ExitReady, stderr.String())
	}
	if !strings.Contains(stdout.String(), "component overrides written") {
		t.Fatalf("stdout = %q, want success message", stdout.String())
	}
	var got componentOverridesFile
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("generated overrides JSON invalid: %v\n%s", err, string(data))
	}
	if got.SchemaVersion != componentOverridesSchemaVersion {
		t.Fatalf("schemaVersion = %q", got.SchemaVersion)
	}
	if len(got.Components) != 1 || got.Components[0].ID != "coredns" {
		t.Fatalf("components = %#v", got.Components)
	}
	wantVersions := []string{"1.9.4-13", "1.9.4-9"}
	if !reflect.DeepEqual(got.Components[0].Versions, wantVersions) {
		t.Fatalf("versions = %#v, want %#v", got.Components[0].Versions, wantVersions)
	}
}

func TestRunComponentOverridesRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	input := dir + "/assessment.json"
	output := dir + "/component-overrides.json"
	err := os.WriteFile(input, []byte(`{
  "schemaVersion": "kua.assessment.v1",
  "assessmentId": "assessment-test",
  "generatedAt": "2026-07-27T04:02:00Z",
  "redacted": true,
  "currentVersion": "1.30.0",
  "readiness": "READY_WITH_WARNINGS",
  "risk": "MEDIUM",
  "findings": [],
  "limitations": [],
  "componentVersionOverrides": {
    "schemaVersion": "kua.component-overrides.v1",
    "outputPath": "component-overrides.json",
    "rerunCommand": "kua analyze --component-overrides component-overrides.json",
    "components": [
      {"id": "coredns", "name": "CoreDNS", "versions": ["<fill-version>"], "evidence": "user-confirmed", "reason": "missing"}
    ]
  }
}`), 0o600)
	if err != nil {
		t.Fatalf("WriteFile input returned error: %v", err)
	}
	if err := os.WriteFile(output, []byte("keep"), 0o600); err != nil {
		t.Fatalf("WriteFile output returned error: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"--input", input, "--output", output, "component-overrides"}, &stdout, &stderr, BuildInfo{})

	if code != ExitExecution {
		t.Fatalf("Run(component-overrides overwrite) exit code = %d, want %d", code, ExitExecution)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(data) != "keep" {
		t.Fatalf("output overwritten: %q", string(data))
	}
}

func TestApplyComponentOverridesRemovesUnknownDetection(t *testing.T) {
	detections := applyComponentOverrides([]components.Detection{
		{ComponentID: "emqx", Name: "EMQX", Version: components.UnknownVersion, Status: components.StatusUnknown},
		{ComponentID: "emqx", Name: "EMQX", Version: "1.0.3", Status: components.StatusFound},
	}, map[string][]string{"emqx": []string{"2.0.0"}})

	got := detectionVersions(detections)
	want := []string{"1.0.3", "2.0.0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("versions = %#v, want %#v", got, want)
	}
	for _, detection := range detections {
		if detection.Status != components.StatusFound {
			t.Fatalf("detection = %#v, want FOUND", detection)
		}
	}
}

func TestLoadComponentOverridesIgnoresPlaceholders(t *testing.T) {
	path := t.TempDir() + "/component-overrides.json"
	err := os.WriteFile(path, []byte(`{
  "schemaVersion": "kua.component-overrides.v1",
  "components": [
    {"id": "coredns", "versions": ["<fill-version>", "1.9.4-13"], "evidence": "user-confirmed"}
  ]
}`), 0o600)
	if err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	overrides, err := loadComponentOverrides(path)
	if err != nil {
		t.Fatalf("loadComponentOverrides returned error: %v", err)
	}
	want := map[string][]string{"coredns": []string{"1.9.4-13"}}
	if !reflect.DeepEqual(overrides, want) {
		t.Fatalf("overrides = %#v, want %#v", overrides, want)
	}
}

func detectionVersions(detections []components.Detection) []string {
	versions := make([]string, 0, len(detections))
	for _, detection := range detections {
		versions = append(versions, detection.Version)
	}
	return versions
}
