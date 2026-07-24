package aks

import (
	"context"
	"fmt"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

// AKSProvider implements the Provider interface for Azure Kubernetes Service.
type AKSProvider struct {
	identity    IdentityResult
	cliAdapter  *CLIAdapter
	fileAdapter *FileAdapter
}

// NewAKSProvider creates a new AKS provider with the given identity signals.
func NewAKSProvider(signals IdentitySignals) *AKSProvider {
	return &AKSProvider{
		identity:    DetectIdentity(signals),
		cliAdapter:  NewCLIAdapter(),
		fileAdapter: NewFileAdapter(),
	}
}

// Identity returns the detected provider and confidence.
func (p *AKSProvider) Identity() (provider.ProviderType, provider.Confidence) {
	return p.identity.Provider, p.identity.Confidence
}

// Evidence retrieves upgrade availability for the cluster.
func (p *AKSProvider) Evidence(ctx context.Context, opts provider.EvidenceOptions) (*provider.ProviderEvidence, error) {
	switch opts.Mode {
	case provider.SourceModeNone:
		return nil, nil

	case provider.SourceModeFile:
		if opts.FilePath == "" {
			return nil, fmt.Errorf("file path required for file mode")
		}
		return p.fileAdapter.LoadEvidence(opts.FilePath, opts.Mode)

	case provider.SourceModeOffline:
		// Offline mode: try file if provided, otherwise return UNKNOWN evidence
		if opts.FilePath != "" {
			evidence, err := p.fileAdapter.LoadEvidence(opts.FilePath, opts.Mode)
			if err == nil {
				return evidence, nil
			}
			// File failed in offline mode - return evidence with limitation
		}
		return p.unknownEvidence(opts.Mode, "offline mode without evidence file")

	case provider.SourceModeAzure:
		// Azure mode: require CLI success
		return p.cliEvidence(ctx, opts)

	case provider.SourceModeAuto:
		fallthrough
	default:
		// Auto mode: try CLI -> file -> UNKNOWN
		return p.autoEvidence(ctx, opts)
	}
}

// cliEvidence retrieves evidence from Azure CLI.
func (p *AKSProvider) cliEvidence(ctx context.Context, opts provider.EvidenceOptions) (*provider.ProviderEvidence, error) {
	// Apply explicit overrides
	identity := p.identity
	if opts.Subscription != "" {
		identity.Subscription = opts.Subscription
	}
	if opts.ResourceGroup != "" {
		identity.ResourceGroup = opts.ResourceGroup
	}
	if opts.ClusterName != "" {
		identity.ClusterName = opts.ClusterName
	}

	// Check if we have required identity
	if !identity.HasRequiredIdentity() {
		return nil, fmt.Errorf("%w: subscription=%q resource-group=%q cluster-name=%q",
			ErrMissingIdentity,
			identity.Subscription,
			identity.ResourceGroup,
			identity.ClusterName)
	}

	return p.cliAdapter.GetUpgrades(ctx, identity, opts.Timeout)
}

// autoEvidence implements the auto fallback chain.
func (p *AKSProvider) autoEvidence(ctx context.Context, opts provider.EvidenceOptions) (*provider.ProviderEvidence, error) {
	// Only try CLI if we detect AKS and have required identity
	if p.identity.Provider == provider.ProviderAKS {
		// Apply explicit overrides
		identity := p.identity
		if opts.Subscription != "" {
			identity.Subscription = opts.Subscription
		}
		if opts.ResourceGroup != "" {
			identity.ResourceGroup = opts.ResourceGroup
		}
		if opts.ClusterName != "" {
			identity.ClusterName = opts.ClusterName
		}

		if identity.HasRequiredIdentity() && p.cliAdapter.IsAvailable(ctx) {
			evidence, err := p.cliAdapter.GetUpgrades(ctx, identity, opts.Timeout)
			if err == nil {
				return evidence, nil
			}
			// CLI failed, continue to fallback
		}
	}

	// Try file if provided
	if opts.FilePath != "" {
		evidence, err := p.fileAdapter.LoadEvidence(opts.FilePath, opts.Mode)
		if err == nil {
			return evidence, nil
		}
	}

	// Fall back to UNKNOWN evidence
	return p.unknownEvidence(opts.Mode, "auto mode: no CLI access or evidence file")
}

// unknownEvidence returns evidence with UNKNOWN availability.
func (p *AKSProvider) unknownEvidence(mode provider.SourceMode, reason string) (*provider.ProviderEvidence, error) {
	return &provider.ProviderEvidence{
		SchemaVersion: "kua.provider-evidence.aks.v1",
		EvidenceID:    "unknown",
		Source: provider.EvidenceSource{
			Mode:   mode,
			Method: provider.MethodCatalog,
		},
		Cluster: provider.ClusterIdentity{
			Provider:           p.identity.Provider,
			IdentityConfidence: provider.ConfidenceUnknown,
		},
		AvailableUpgrades: []provider.UpgradeOption{},
		Limitations: []provider.Limitation{
			{
				Code:     "PROVIDER_EVIDENCE_UNAVAILABLE",
				Severity: provider.SeverityWarn,
				Summary:  reason,
			},
		},
	}, nil
}
