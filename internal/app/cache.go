package app

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/external/kubent"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/recommendation"
	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/report"
)

const preflightCacheSchema = "kua.preflight-cache.v1"

// PreflightCacheEntry holds pre-flight analysis results for day-of reuse.
type PreflightCacheEntry struct {
	SchemaVersion       string                     `json:"schemaVersion"`
	CachedAt            time.Time                  `json:"cachedAt"`
	ContextName         string                     `json:"contextName"`
	TargetVersion       string                     `json:"targetVersion"`
	APIFindings         []kubent.Finding           `json:"apiFindings"`
	APILimitation       recommendation.Limitation  `json:"apiLimitation"`
	ComponentDetections []components.Detection     `json:"componentDetections"`
	ProviderEvidence    *provider.ProviderEvidence `json:"providerEvidence,omitempty"`
	ProviderLimitation  recommendation.Limitation  `json:"providerLimitation"`
}

func savePreflightCache(path string, entry PreflightCacheEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal preflight cache: %w", err)
	}
	return report.WriteAtomic(path, data)
}

func loadPreflightCache(path string) (PreflightCacheEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PreflightCacheEntry{}, fmt.Errorf("read preflight cache %q: %w", path, err)
	}
	var entry PreflightCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return PreflightCacheEntry{}, fmt.Errorf("parse preflight cache %q: %w", path, err)
	}
	if entry.SchemaVersion != preflightCacheSchema {
		return PreflightCacheEntry{}, fmt.Errorf("preflight cache %q has unsupported schema %q; expected %q",
			path, entry.SchemaVersion, preflightCacheSchema)
	}
	return entry, nil
}
