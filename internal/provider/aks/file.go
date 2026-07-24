package aks

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

// FileAdapter reads provider evidence from a JSON file.
type FileAdapter struct{}

// NewFileAdapter creates a new file adapter.
func NewFileAdapter() *FileAdapter {
	return &FileAdapter{}
}

// LoadEvidence reads and validates a provider evidence file.
func (a *FileAdapter) LoadEvidence(filePath string, mode provider.SourceMode) (*provider.ProviderEvidence, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read evidence file: %w", err)
	}

	// Try to parse as KUA provider evidence format first
	var kuaEvidence provider.ProviderEvidence
	if err := json.Unmarshal(data, &kuaEvidence); err == nil {
		if kuaEvidence.SchemaVersion == "kua.provider-evidence.aks.v1" {
			// Update source based on mode
			kuaEvidence.Source.Mode = mode
			if kuaEvidence.Source.Method == "" {
				kuaEvidence.Source.Method = provider.MethodUserFile
			}
			return &kuaEvidence, nil
		}
	}

	// Try to parse as raw Azure CLI export
	identity := IdentityResult{
		Provider:   provider.ProviderAKS,
		Confidence: provider.ConfidenceMedium,
	}

	evidence, err := ParseCLIOutput(data, identity, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("parse evidence file: %w", err)
	}

	evidence.Source.Mode = mode
	evidence.Source.Method = provider.MethodAzureCLIExport

	// Add limitation about file source
	evidence.Limitations = append(evidence.Limitations, provider.Limitation{
		Code:     "FILE_EVIDENCE_SOURCE",
		Severity: provider.SeverityInfo,
		Summary:  "Evidence loaded from file; may not reflect current cluster state",
	})

	return evidence, nil
}

// ValidateEvidence checks that evidence contains required fields.
func (a *FileAdapter) ValidateEvidence(evidence *provider.ProviderEvidence) error {
	if evidence.SchemaVersion == "" {
		return fmt.Errorf("missing schemaVersion")
	}
	if evidence.CurrentVersion == "" {
		return fmt.Errorf("missing currentVersion")
	}
	if evidence.Cluster.Provider == "" {
		return fmt.Errorf("missing cluster.provider")
	}
	return nil
}
