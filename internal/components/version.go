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

func hasImageTag(value string) bool {
	colonIndex := strings.LastIndex(value, ":")
	if colonIndex < 0 || colonIndex == len(value)-1 {
		return false
	}
	slashIndex := strings.LastIndex(value, "/")
	return colonIndex > slashIndex
}
