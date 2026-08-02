package components

import "strings"

const UnknownVersion = "UNKNOWN"

func NormalizeVersion(value string) (string, Confidence, Status) {
	version := strings.TrimSpace(value)
	if version == "" || strings.EqualFold(version, "unknown") || strings.EqualFold(version, "latest") {
		return UnknownVersion, ConfidenceUnknown, StatusUnknown
	}
	if digestIndex := strings.Index(version, "@sha256:"); digestIndex >= 0 {
		if !hasImageTag(version[:digestIndex]) {
			return UnknownVersion, ConfidenceUnknown, StatusUnknown
		}
		version = version[:digestIndex]
	}
	if !hasImageTag(version) {
		return UnknownVersion, ConfidenceUnknown, StatusUnknown
	}
	version = version[strings.LastIndex(version, ":")+1:]
	version = strings.TrimPrefix(version, "v")
	if version == "" || strings.EqualFold(version, "latest") {
		return UnknownVersion, ConfidenceUnknown, StatusUnknown
	}
	return version, ConfidenceHigh, StatusFound
}

// NormalizeLabelVersion normalizes a plain version string sourced from a
// Kubernetes label (e.g. app.kubernetes.io/version: "5.8.8"). It strips a
// leading 'v' prefix and returns ConfidenceHigh when a non-empty result is
// found, which takes priority over image-tag extraction in detectors.
func NormalizeLabelVersion(value string) (string, Confidence, Status) {
	version := strings.TrimSpace(value)
	if version == "" || strings.EqualFold(version, "unknown") {
		return UnknownVersion, ConfidenceUnknown, StatusUnknown
	}
	version = strings.TrimPrefix(version, "v")
	if version == "" {
		return UnknownVersion, ConfidenceUnknown, StatusUnknown
	}
	return version, ConfidenceHigh, StatusFound
}

func hasImageTag(value string) bool {
	colonIndex := strings.LastIndex(value, ":")
	if colonIndex < 0 || colonIndex == len(value)-1 {
		return false
	}
	slashIndex := strings.LastIndex(value, "/")
	return colonIndex > slashIndex
}
