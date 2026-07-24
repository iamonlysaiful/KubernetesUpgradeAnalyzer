package aks

import (
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/provider"
)

func TestDetectIdentity_ExplicitFlags(t *testing.T) {
	signals := IdentitySignals{
		ExplicitSubscription:  "sub-123",
		ExplicitResourceGroup: "rg-test",
		ExplicitClusterName:   "my-cluster",
	}

	result := DetectIdentity(signals)

	if result.Provider != provider.ProviderAKS {
		t.Errorf("expected ProviderAKS, got %v", result.Provider)
	}
	if result.Confidence != provider.ConfidenceHigh {
		t.Errorf("expected ConfidenceHigh, got %v", result.Confidence)
	}
	if result.Subscription != "sub-123" {
		t.Errorf("expected sub-123, got %v", result.Subscription)
	}
	if result.ResourceGroup != "rg-test" {
		t.Errorf("expected rg-test, got %v", result.ResourceGroup)
	}
	if result.ClusterName != "my-cluster" {
		t.Errorf("expected my-cluster, got %v", result.ClusterName)
	}
	if result.DetectionMethod != "explicit_flags" {
		t.Errorf("expected explicit_flags, got %v", result.DetectionMethod)
	}
}

func TestDetectIdentity_NodeProviderID(t *testing.T) {
	signals := IdentitySignals{
		NodeProviderIDs: []string{
			"azure:///subscriptions/abc-123-def/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachineScaleSets/aks-nodepool1",
		},
		ExplicitClusterName: "test-cluster",
	}

	result := DetectIdentity(signals)

	if result.Provider != provider.ProviderAKS {
		t.Errorf("expected ProviderAKS, got %v", result.Provider)
	}
	if result.Confidence != provider.ConfidenceHigh {
		t.Errorf("expected ConfidenceHigh, got %v", result.Confidence)
	}
	if result.Subscription != "abc-123-def" {
		t.Errorf("expected abc-123-def, got %v", result.Subscription)
	}
	if result.ResourceGroup != "my-rg" {
		t.Errorf("expected my-rg, got %v", result.ResourceGroup)
	}
	if result.ClusterName != "test-cluster" {
		t.Errorf("expected test-cluster, got %v", result.ClusterName)
	}
	if result.DetectionMethod != "node_provider_id" {
		t.Errorf("expected node_provider_id, got %v", result.DetectionMethod)
	}
}

func TestDetectIdentity_APIHostname(t *testing.T) {
	signals := IdentitySignals{
		APIServerHostname: "my-cluster-abc123.hcp.eastus.azmk8s.io",
	}

	result := DetectIdentity(signals)

	if result.Provider != provider.ProviderAKS {
		t.Errorf("expected ProviderAKS, got %v", result.Provider)
	}
	if result.Confidence != provider.ConfidenceMedium {
		t.Errorf("expected ConfidenceMedium, got %v", result.Confidence)
	}
	if result.DetectionMethod != "api_hostname" {
		t.Errorf("expected api_hostname, got %v", result.DetectionMethod)
	}
}

func TestDetectIdentity_ContextNamePattern(t *testing.T) {
	tests := []struct {
		name        string
		contextName string
		wantAKS     bool
	}{
		{"suffix aks", "my-cluster-aks", true},
		{"prefix aks", "aks-my-cluster", true},
		{"middle aks", "my-aks-cluster", true},
		{"underscore aks", "my_aks_cluster", true},
		{"no aks", "my-cluster", false},
		{"eks not aks", "my-eks-cluster", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signals := IdentitySignals{
				ContextName: tt.contextName,
			}

			result := DetectIdentity(signals)

			if tt.wantAKS {
				if result.Provider != provider.ProviderAKS {
					t.Errorf("expected ProviderAKS for context %q, got %v", tt.contextName, result.Provider)
				}
				if result.Confidence != provider.ConfidenceLow {
					t.Errorf("expected ConfidenceLow, got %v", result.Confidence)
				}
			} else {
				if result.Provider != provider.ProviderUnknown {
					t.Errorf("expected ProviderUnknown for context %q, got %v", tt.contextName, result.Provider)
				}
			}
		})
	}
}

func TestDetectIdentity_NoSignals(t *testing.T) {
	signals := IdentitySignals{}

	result := DetectIdentity(signals)

	if result.Provider != provider.ProviderUnknown {
		t.Errorf("expected ProviderUnknown, got %v", result.Provider)
	}
	if result.Confidence != provider.ConfidenceUnknown {
		t.Errorf("expected ConfidenceUnknown, got %v", result.Confidence)
	}
}

func TestIdentityResult_HasRequiredIdentity(t *testing.T) {
	tests := []struct {
		name     string
		identity IdentityResult
		want     bool
	}{
		{
			name: "all fields",
			identity: IdentityResult{
				Subscription:  "sub",
				ResourceGroup: "rg",
				ClusterName:   "cluster",
			},
			want: true,
		},
		{
			name: "missing subscription",
			identity: IdentityResult{
				ResourceGroup: "rg",
				ClusterName:   "cluster",
			},
			want: false,
		},
		{
			name: "missing resource group",
			identity: IdentityResult{
				Subscription: "sub",
				ClusterName:  "cluster",
			},
			want: false,
		},
		{
			name: "missing cluster name",
			identity: IdentityResult{
				Subscription:  "sub",
				ResourceGroup: "rg",
			},
			want: false,
		},
		{
			name:     "all empty",
			identity: IdentityResult{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.identity.HasRequiredIdentity(); got != tt.want {
				t.Errorf("HasRequiredIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIdentityResult_Redaction(t *testing.T) {
	identity := IdentityResult{
		Subscription:  "12345678-1234-1234-1234-123456789abc",
		ResourceGroup: "my-resource-group",
		ClusterName:   "my-cluster-name",
	}

	if got := identity.RedactedSubscription(); got != "sub-1234****" {
		t.Errorf("RedactedSubscription() = %v", got)
	}
	if got := identity.RedactedResourceGroup(); got != "rg-my-r****" {
		t.Errorf("RedactedResourceGroup() = %v", got)
	}
	if got := identity.RedactedClusterName(); got != "cluster-my-c****" {
		t.Errorf("RedactedClusterName() = %v", got)
	}
}

func TestIdentityResult_RedactionShortValues(t *testing.T) {
	identity := IdentityResult{
		Subscription:  "sub",
		ResourceGroup: "rg",
		ClusterName:   "cl",
	}

	if got := identity.RedactedSubscription(); got != "sub-****" {
		t.Errorf("RedactedSubscription() = %v", got)
	}
	if got := identity.RedactedResourceGroup(); got != "rg-****" {
		t.Errorf("RedactedResourceGroup() = %v", got)
	}
	if got := identity.RedactedClusterName(); got != "cluster-****" {
		t.Errorf("RedactedClusterName() = %v", got)
	}
}
