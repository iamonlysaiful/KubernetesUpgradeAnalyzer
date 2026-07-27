package app

import (
	"os"
	"reflect"
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
