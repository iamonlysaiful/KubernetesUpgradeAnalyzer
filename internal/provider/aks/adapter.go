package aks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

const (
	// AllowedCommand is the only Azure CLI command permitted for provider evidence.
	AllowedCommand = "az aks get-upgrades"

	// DefaultTimeout is the default timeout for CLI invocation.
	DefaultTimeout = 30 * time.Second

	// MaxOutputSize is the maximum output size to read from CLI.
	MaxOutputSize = 10 * 1024 * 1024 // 10MB
)

var (
	// ErrMutatingCommand is returned when a mutating command is attempted.
	ErrMutatingCommand = errors.New("mutating Azure CLI commands are not permitted")
	// ErrAuthenticationRequired is returned when Azure CLI is not authenticated.
	ErrAuthenticationRequired = errors.New("Azure CLI authentication required")
	// ErrCLINotFound is returned when Azure CLI is not installed.
	ErrCLINotFound = errors.New("Azure CLI not found")
	// ErrMissingIdentity is returned when required identity fields are missing.
	ErrMissingIdentity = errors.New("missing required AKS identity: subscription, resource-group, and cluster-name required")
	// ErrInvalidOutput is returned when CLI output cannot be parsed.
	ErrInvalidOutput = errors.New("invalid Azure CLI output")
)

// mutatingCommands are Azure CLI commands that modify state (never permitted).
var mutatingCommands = []string{
	"az aks upgrade",
	"az aks delete",
	"az aks create",
	"az aks update",
	"az aks stop",
	"az aks start",
	"az aks scale",
	"az aks nodepool add",
	"az aks nodepool delete",
	"az aks nodepool upgrade",
	"az aks nodepool scale",
	"az login",
	"az account set",
	"az extension",
}

// CLIAdapter executes Azure CLI commands for provider evidence.
type CLIAdapter struct {
	// AzPath is the path to the az binary. Defaults to "az".
	AzPath string
}

// NewCLIAdapter creates a new CLI adapter.
func NewCLIAdapter() *CLIAdapter {
	return &CLIAdapter{
		AzPath: "az",
	}
}

// GetUpgrades retrieves upgrade availability from Azure CLI.
func (a *CLIAdapter) GetUpgrades(ctx context.Context, identity IdentityResult, timeout time.Duration) (*provider.ProviderEvidence, error) {
	if !identity.HasRequiredIdentity() {
		return nil, ErrMissingIdentity
	}

	if timeout == 0 {
		timeout = DefaultTimeout
	}

	args := []string{
		"aks", "get-upgrades",
		"--subscription", identity.Subscription,
		"--resource-group", identity.ResourceGroup,
		"--name", identity.ClusterName,
		"--output", "json",
	}

	// Validate this is an allowed command (defense in depth)
	fullCmd := "az " + strings.Join(args[:2], " ")
	if err := validateAllowedCommand(fullCmd); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.AzPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	capturedAt := time.Now().UTC()

	err := cmd.Run()
	if err != nil {
		// Check for common error patterns
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not logged in") ||
			strings.Contains(stderrStr, "AADSTS") ||
			strings.Contains(stderrStr, "az login") {
			return nil, fmt.Errorf("%w: %s", ErrAuthenticationRequired, stderrStr)
		}
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrCLINotFound
		}
		return nil, fmt.Errorf("az aks get-upgrades failed: %w: %s", err, stderrStr)
	}

	// Parse the CLI output
	evidence, err := ParseCLIOutput(stdout.Bytes(), identity, capturedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOutput, err)
	}

	evidence.Source.Mode = provider.SourceModeAzure
	evidence.Source.Method = provider.MethodAzureCLI
	evidence.Source.Command = AllowedCommand

	// Try to get Azure CLI version
	if version, err := a.GetVersion(ctx); err == nil {
		evidence.Source.AzureCLIVersion = version
	}

	return evidence, nil
}

// GetVersion returns the Azure CLI version.
func (a *CLIAdapter) GetVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.AzPath, "version", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var versionInfo struct {
		AzureCLI string `json:"azure-cli"`
	}
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		return "", err
	}

	return versionInfo.AzureCLI, nil
}

// IsAvailable checks if Azure CLI is installed and accessible.
func (a *CLIAdapter) IsAvailable(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, a.AzPath, "version")
	return cmd.Run() == nil
}

// validateAllowedCommand ensures only the allowlisted command is used.
func validateAllowedCommand(cmd string) error {
	// Check against mutating commands first
	cmdLower := strings.ToLower(cmd)
	for _, mutating := range mutatingCommands {
		if strings.HasPrefix(cmdLower, strings.ToLower(mutating)) {
			return fmt.Errorf("%w: %s", ErrMutatingCommand, mutating)
		}
	}

	// Verify it's the exact allowed command
	if !strings.HasPrefix(cmdLower, strings.ToLower(AllowedCommand)) {
		return fmt.Errorf("command not allowed: %s (only %s is permitted)", cmd, AllowedCommand)
	}

	return nil
}

// ParseCLIOutput parses the JSON output from az aks get-upgrades.
func ParseCLIOutput(data []byte, identity IdentityResult, capturedAt time.Time) (*provider.ProviderEvidence, error) {
	var raw struct {
		ControlPlaneProfile struct {
			KubernetesVersion string `json:"kubernetesVersion"`
			Upgrades          []struct {
				KubernetesVersion string `json:"kubernetesVersion"`
				IsPreview         *bool  `json:"isPreview"`
			} `json:"upgrades"`
		} `json:"controlPlaneProfile"`
		AgentPoolProfiles []struct {
			Name              string `json:"name"`
			KubernetesVersion string `json:"kubernetesVersion"`
			Upgrades          []struct {
				KubernetesVersion string `json:"kubernetesVersion"`
				IsPreview         *bool  `json:"isPreview"`
			} `json:"upgrades"`
		} `json:"agentPoolProfiles"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal CLI output: %w", err)
	}

	evidence := &provider.ProviderEvidence{
		SchemaVersion:       "kua.provider-evidence.aks.v1",
		EvidenceID:          fmt.Sprintf("aks-%d", capturedAt.Unix()),
		CapturedAt:          capturedAt,
		CurrentVersion:      raw.ControlPlaneProfile.KubernetesVersion,
		ControlPlaneVersion: raw.ControlPlaneProfile.KubernetesVersion,
		Cluster: provider.ClusterIdentity{
			Provider:           provider.ProviderAKS,
			SubscriptionAlias:  identity.RedactedSubscription(),
			ResourceGroupAlias: identity.RedactedResourceGroup(),
			ClusterNameAlias:   identity.RedactedClusterName(),
			IdentityConfidence: identity.Confidence,
		},
		AvailableUpgrades: make([]provider.UpgradeOption, 0),
		Limitations:       make([]provider.Limitation, 0),
	}

	// Parse control plane upgrades
	for _, upgrade := range raw.ControlPlaneProfile.Upgrades {
		isPreview := false
		if upgrade.IsPreview != nil {
			isPreview = *upgrade.IsPreview
		}
		evidence.AvailableUpgrades = append(evidence.AvailableUpgrades, provider.UpgradeOption{
			Version:     upgrade.KubernetesVersion,
			IsPreview:   isPreview,
			SupportPlan: provider.SupportUnknown, // AKS API doesn't include this
		})
	}

	// Parse node pool profiles
	for _, pool := range raw.AgentPoolProfiles {
		poolEvidence := provider.NodePoolEvidence{
			NameAlias:         "pool-" + pool.Name[:min(4, len(pool.Name))] + "****",
			CurrentVersion:    pool.KubernetesVersion,
			AvailableUpgrades: make([]provider.UpgradeOption, 0),
		}
		for _, upgrade := range pool.Upgrades {
			isPreview := false
			if upgrade.IsPreview != nil {
				isPreview = *upgrade.IsPreview
			}
			poolEvidence.AvailableUpgrades = append(poolEvidence.AvailableUpgrades, provider.UpgradeOption{
				Version:     upgrade.KubernetesVersion,
				IsPreview:   isPreview,
				SupportPlan: provider.SupportUnknown,
			})
		}
		evidence.NodePools = append(evidence.NodePools, poolEvidence)
	}

	if evidence.CurrentVersion == "" {
		return nil, errors.New("missing kubernetesVersion in CLI output")
	}

	return evidence, nil
}
