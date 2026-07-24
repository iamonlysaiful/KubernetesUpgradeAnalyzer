package aks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

func TestParseCLIOutput_Valid(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "az-output-valid.json"))
	if err != nil {
		t.Fatalf("read test data: %v", err)
	}

	identity := IdentityResult{
		Provider:      provider.ProviderAKS,
		Confidence:    provider.ConfidenceHigh,
		Subscription:  "test-sub",
		ResourceGroup: "test-rg",
		ClusterName:   "test-cluster",
	}

	evidence, err := ParseCLIOutput(data, identity, fixedTime())
	if err != nil {
		t.Fatalf("ParseCLIOutput: %v", err)
	}

	if evidence.CurrentVersion != "1.30.4" {
		t.Errorf("CurrentVersion = %v, want 1.30.4", evidence.CurrentVersion)
	}

	if len(evidence.AvailableUpgrades) != 8 {
		t.Errorf("AvailableUpgrades count = %d, want 8", len(evidence.AvailableUpgrades))
	}

	// Check for preview version
	var hasPreview bool
	for _, u := range evidence.AvailableUpgrades {
		if u.Version == "1.33.0" && u.IsPreview {
			hasPreview = true
			break
		}
	}
	if !hasPreview {
		t.Error("expected 1.33.0 to be marked as preview")
	}

	if len(evidence.NodePools) != 2 {
		t.Errorf("NodePools count = %d, want 2", len(evidence.NodePools))
	}
}

func TestParseCLIOutput_EmptyUpgrades(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "az-output-empty-upgrades.json"))
	if err != nil {
		t.Fatalf("read test data: %v", err)
	}

	identity := IdentityResult{
		Provider:   provider.ProviderAKS,
		Confidence: provider.ConfidenceHigh,
	}

	evidence, err := ParseCLIOutput(data, identity, fixedTime())
	if err != nil {
		t.Fatalf("ParseCLIOutput: %v", err)
	}

	if evidence.CurrentVersion != "1.33.12" {
		t.Errorf("CurrentVersion = %v, want 1.33.12", evidence.CurrentVersion)
	}

	if len(evidence.AvailableUpgrades) != 0 {
		t.Errorf("AvailableUpgrades count = %d, want 0", len(evidence.AvailableUpgrades))
	}
}

func TestParseCLIOutput_Invalid(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"empty", ""},
		{"invalid json", "{invalid}"},
		{"missing version", `{"controlPlaneProfile":{"upgrades":[]}}`},
	}

	identity := IdentityResult{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCLIOutput([]byte(tt.data), identity, fixedTime())
			if err == nil {
				t.Error("expected error for invalid input")
			}
		})
	}
}

func TestValidateAllowedCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		wantErr bool
	}{
		{"allowed get-upgrades", "az aks get-upgrades", false},
		{"mutating upgrade", "az aks upgrade", true},
		{"mutating delete", "az aks delete", true},
		{"mutating create", "az aks create", true},
		{"mutating login", "az login", true},
		{"mutating account set", "az account set", true},
		{"unknown command", "az vm list", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAllowedCommand(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAllowedCommand(%q) error = %v, wantErr %v", tt.cmd, err, tt.wantErr)
			}
		})
	}
}
