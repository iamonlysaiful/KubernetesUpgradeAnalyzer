// Package aks implements the AKS provider adapter.
package aks

import (
	"regexp"
	"strings"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

// IdentitySignals contains signals used to detect AKS identity.
type IdentitySignals struct {
	// ExplicitSubscription is set when --subscription flag is provided.
	ExplicitSubscription string
	// ExplicitResourceGroup is set when --resource-group flag is provided.
	ExplicitResourceGroup string
	// ExplicitClusterName is set when --cluster-name flag is provided.
	ExplicitClusterName string
	// NodeProviderIDs contains spec.providerID from cluster nodes.
	NodeProviderIDs []string
	// APIServerHostname is the Kubernetes API server hostname.
	APIServerHostname string
	// ContextName is the kubeconfig context name.
	ContextName string
}

// IdentityResult contains the detected AKS identity.
type IdentityResult struct {
	// Provider is the detected provider type.
	Provider provider.ProviderType
	// Confidence indicates how reliably identity was determined.
	Confidence provider.Confidence
	// Subscription is the detected or explicit subscription ID.
	Subscription string
	// ResourceGroup is the detected or explicit resource group.
	ResourceGroup string
	// ClusterName is the detected or explicit cluster name.
	ClusterName string
	// DetectionMethod describes how identity was determined.
	DetectionMethod string
}

var (
	// azureProviderIDPattern matches node providerID for AKS nodes.
	// Example: azure:///subscriptions/SUB_ID/resourceGroups/RG/providers/Microsoft.Compute/...
	azureProviderIDPattern = regexp.MustCompile(
		`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/`,
	)

	// azmk8sHostnamePattern matches AKS API server hostnames.
	// Example: my-cluster-dns-abc123.hcp.eastus.azmk8s.io
	azmk8sHostnamePattern = regexp.MustCompile(`\.azmk8s\.io$`)

	// aksContextPattern matches common AKS context naming patterns.
	// Examples: my-cluster-aks, aks-my-cluster, my-aks-cluster
	aksContextPattern = regexp.MustCompile(`(?i)[-_]aks[-_]|^aks[-_]|[-_]aks$`)
)

// DetectIdentity determines AKS identity from available signals.
func DetectIdentity(signals IdentitySignals) IdentityResult {
	result := IdentityResult{
		Provider:   provider.ProviderUnknown,
		Confidence: provider.ConfidenceUnknown,
	}

	// Priority 1: Explicit flags (HIGH confidence)
	if signals.ExplicitSubscription != "" &&
		signals.ExplicitResourceGroup != "" &&
		signals.ExplicitClusterName != "" {
		result.Provider = provider.ProviderAKS
		result.Confidence = provider.ConfidenceHigh
		result.Subscription = signals.ExplicitSubscription
		result.ResourceGroup = signals.ExplicitResourceGroup
		result.ClusterName = signals.ExplicitClusterName
		result.DetectionMethod = "explicit_flags"
		return result
	}

	// Priority 2: Node providerID (HIGH confidence)
	for _, providerID := range signals.NodeProviderIDs {
		if matches := azureProviderIDPattern.FindStringSubmatch(providerID); matches != nil {
			result.Provider = provider.ProviderAKS
			result.Confidence = provider.ConfidenceHigh
			result.Subscription = matches[1]
			result.ResourceGroup = matches[2]
			result.DetectionMethod = "node_provider_id"
			// ClusterName cannot be reliably extracted from providerID alone
			// Use explicit flag or other signals
			if signals.ExplicitClusterName != "" {
				result.ClusterName = signals.ExplicitClusterName
			}
			return result
		}
	}

	// Priority 3: API server hostname (MEDIUM confidence)
	if azmk8sHostnamePattern.MatchString(signals.APIServerHostname) {
		result.Provider = provider.ProviderAKS
		result.Confidence = provider.ConfidenceMedium
		result.DetectionMethod = "api_hostname"
		// Extract cluster name from hostname if possible
		// Format: <cluster>-<dns-prefix>.<location>.azmk8s.io
		parts := strings.Split(signals.APIServerHostname, ".")
		if len(parts) >= 3 {
			// First part might contain cluster info
			if signals.ExplicitClusterName != "" {
				result.ClusterName = signals.ExplicitClusterName
			}
		}
		// Fill explicit overrides
		if signals.ExplicitSubscription != "" {
			result.Subscription = signals.ExplicitSubscription
		}
		if signals.ExplicitResourceGroup != "" {
			result.ResourceGroup = signals.ExplicitResourceGroup
		}
		return result
	}

	// Priority 4: Context name pattern (LOW confidence)
	if aksContextPattern.MatchString(signals.ContextName) {
		result.Provider = provider.ProviderAKS
		result.Confidence = provider.ConfidenceLow
		result.DetectionMethod = "context_name_pattern"
		// Fill explicit overrides
		if signals.ExplicitSubscription != "" {
			result.Subscription = signals.ExplicitSubscription
		}
		if signals.ExplicitResourceGroup != "" {
			result.ResourceGroup = signals.ExplicitResourceGroup
		}
		if signals.ExplicitClusterName != "" {
			result.ClusterName = signals.ExplicitClusterName
		}
		return result
	}

	// No AKS signals detected
	return result
}

// HasRequiredIdentity returns true if the identity has all required fields for Azure CLI.
func (r IdentityResult) HasRequiredIdentity() bool {
	return r.Subscription != "" && r.ResourceGroup != "" && r.ClusterName != ""
}

// RedactedSubscription returns a redacted subscription identifier.
func (r IdentityResult) RedactedSubscription() string {
	if r.Subscription == "" {
		return ""
	}
	if len(r.Subscription) <= 8 {
		return "sub-****"
	}
	return "sub-" + r.Subscription[:4] + "****"
}

// RedactedResourceGroup returns a redacted resource group identifier.
func (r IdentityResult) RedactedResourceGroup() string {
	if r.ResourceGroup == "" {
		return ""
	}
	if len(r.ResourceGroup) <= 4 {
		return "rg-****"
	}
	return "rg-" + r.ResourceGroup[:4] + "****"
}

// RedactedClusterName returns a redacted cluster name.
func (r IdentityResult) RedactedClusterName() string {
	if r.ClusterName == "" {
		return ""
	}
	if len(r.ClusterName) <= 4 {
		return "cluster-****"
	}
	return "cluster-" + r.ClusterName[:4] + "****"
}
