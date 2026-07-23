package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEmbeddedCatalog(t *testing.T) {
	bundle, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded returned error: %v", err)
	}

	if bundle.Source != SourceEmbedded {
		t.Fatalf("Source = %q, want %q", bundle.Source, SourceEmbedded)
	}
	if bundle.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %q, want %q", bundle.SchemaVersion, SchemaVersion)
	}
	if bundle.ChecksumSHA256 == "" {
		t.Fatalf("ChecksumSHA256 is empty")
	}
	if len(bundle.Sources) == 0 {
		t.Fatalf("Sources should not be empty")
	}
}

func TestLoadFileCatalog(t *testing.T) {
	path := filepath.Join("..", "..", "schemas", "fixtures", "catalog", "valid", "minimal-catalog.json")
	bundle, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if bundle.Source != SourceFile {
		t.Fatalf("Source = %q, want %q", bundle.Source, SourceFile)
	}
	if bundle.CatalogVersion != "0.1.0" {
		t.Fatalf("CatalogVersion = %q, want 0.1.0", bundle.CatalogVersion)
	}
}

func TestLoadRejectsMissingSourceReference(t *testing.T) {
	path := filepath.Join("..", "..", "schemas", "fixtures", "catalog", "invalid", "missing-source.json")
	_, err := LoadFile(path)
	if err == nil {
		t.Fatalf("LoadFile returned nil error")
	}
	if !strings.Contains(err.Error(), "components[0].sourceId must reference a source") {
		t.Fatalf("error = %q, want missing source diagnostic", err.Error())
	}
}

func TestValidateRejectsDuplicateComponentAliases(t *testing.T) {
	bundle := validBundle()
	bundle.Components = append(bundle.Components, Component{
		ProductID:           "other",
		Aliases:             []string{"ingress-nginx"},
		ProductVersionRange: "1.0.x",
		KubernetesRange:     ">=1.30 <=1.33",
		Status:              "unknown",
		Confidence:          "unknown",
		SourceID:            "src-one",
	})

	err := Validate(bundle)
	if err == nil {
		t.Fatalf("Validate returned nil error")
	}
	if !strings.Contains(err.Error(), "duplicates") {
		t.Fatalf("error = %q, want duplicate diagnostic", err.Error())
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	_, err := LoadBytes([]byte(`{"schemaVersion":`), SourceEmbedded)
	if err == nil {
		t.Fatalf("LoadBytes returned nil error")
	}
	if !strings.Contains(err.Error(), "decode catalog") {
		t.Fatalf("error = %q, want decode catalog", err.Error())
	}
}

func TestLoadFileReturnsReadError(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatalf("LoadFile returned nil error")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("error = %v, want not exist", err)
	}
}

func validBundle() Bundle {
	return Bundle{
		SchemaVersion:  SchemaVersion,
		CatalogVersion: "0.1.0",
		CreatedAt:      "2026-07-23T00:00:00Z",
		Kubernetes: Kubernetes{
			ValidatedRange: ValidatedRange{MinMinor: "1.30", MaxMinor: "1.33"},
		},
		Components: []Component{
			{
				ProductID:           "nginx-ingress",
				Aliases:             []string{"ingress-nginx"},
				ProductVersionRange: "1.12.x",
				KubernetesRange:     ">=1.30 <=1.33",
				Status:              "supported",
				Confidence:          "authoritative",
				SourceID:            "src-one",
			},
		},
		Sources: []Source{
			{
				SourceID:    "src-one",
				Title:       "Source One",
				URL:         "https://example.com/source-one",
				RetrievedAt: "2026-07-23T00:00:00Z",
				Claim:       "Fixture source.",
			},
		},
	}
}
