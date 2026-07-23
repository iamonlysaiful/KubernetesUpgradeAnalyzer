package components

import "testing"

func TestNormalizeVersionFindsImageTag(t *testing.T) {
	version, confidence, status := NormalizeVersion("registry.k8s.io/coredns/coredns:v1.11.3")

	if version != "1.11.3" {
		t.Fatalf("version = %q, want 1.11.3", version)
	}
	if confidence != ConfidenceHigh {
		t.Fatalf("confidence = %q, want %q", confidence, ConfidenceHigh)
	}
	if status != StatusFound {
		t.Fatalf("status = %q, want %q", status, StatusFound)
	}
}

func TestNormalizeVersionDropsDigestAfterTag(t *testing.T) {
	version, confidence, status := NormalizeVersion("example/component:2.0.1@sha256:abcdef")

	if version != "2.0.1" || confidence != ConfidenceHigh || status != StatusFound {
		t.Fatalf("NormalizeVersion returned %q/%q/%q, want 2.0.1/HIGH/FOUND", version, confidence, status)
	}
}

func TestNormalizeVersionReturnsUnknownForAmbiguousValues(t *testing.T) {
	for _, value := range []string{"", "unknown", "latest", "example/component", "registry:5000/component", "example/component:latest", "example/component@sha256:abcdef"} {
		version, confidence, status := NormalizeVersion(value)
		if version != UnknownVersion || confidence != ConfidenceUnknown || status != StatusUnknown {
			t.Fatalf("NormalizeVersion(%q) = %q/%q/%q, want UNKNOWN/UNKNOWN/UNKNOWN", value, version, confidence, status)
		}
	}
}
