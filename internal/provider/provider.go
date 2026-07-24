// Package provider defines the provider-neutral evidence interface for upgrade availability.
package provider

import (
	"context"
	"time"
)

// ProviderType identifies the managed Kubernetes provider.
type ProviderType string

const (
	// ProviderAKS is Azure Kubernetes Service.
	ProviderAKS ProviderType = "AKS"
	// ProviderUnknown indicates the provider could not be determined.
	ProviderUnknown ProviderType = "UNKNOWN"
)

// Confidence indicates how reliably the provider identity was determined.
type Confidence string

const (
	// ConfidenceHigh indicates explicit user input or definitive API evidence.
	ConfidenceHigh Confidence = "HIGH"
	// ConfidenceMedium indicates strong but not definitive signals.
	ConfidenceMedium Confidence = "MEDIUM"
	// ConfidenceLow indicates weak or ambiguous signals.
	ConfidenceLow Confidence = "LOW"
	// ConfidenceUnknown indicates no matching signals were found.
	ConfidenceUnknown Confidence = "UNKNOWN"
)

// SourceMode specifies how provider evidence should be obtained.
type SourceMode string

const (
	// SourceModeAuto detects AKS, invokes CLI, falls back to file, then UNKNOWN.
	SourceModeAuto SourceMode = "auto"
	// SourceModeAzure requires Azure CLI; failure is inconclusive.
	SourceModeAzure SourceMode = "azure"
	// SourceModeFile requires user-supplied JSON evidence file.
	SourceModeFile SourceMode = "file"
	// SourceModeOffline prohibits provider network calls; optional file.
	SourceModeOffline SourceMode = "offline"
	// SourceModeNone skips provider analysis entirely.
	SourceModeNone SourceMode = "none"
)

// EvidenceMethod indicates how the evidence was obtained.
type EvidenceMethod string

const (
	// MethodAzureCLI indicates live Azure CLI invocation.
	MethodAzureCLI EvidenceMethod = "AZURE_CLI"
	// MethodAzureCLIExport indicates exported Azure CLI output file.
	MethodAzureCLIExport EvidenceMethod = "AZURE_CLI_EXPORT"
	// MethodCatalog indicates catalog-based evidence.
	MethodCatalog EvidenceMethod = "CATALOG"
	// MethodUserFile indicates user-provided evidence file.
	MethodUserFile EvidenceMethod = "USER_FILE"
)

// SupportPlan indicates the AKS support tier for a version.
type SupportPlan string

const (
	// SupportStandard is the standard AKS support plan.
	SupportStandard SupportPlan = "STANDARD"
	// SupportLTS is the long-term support plan.
	SupportLTS SupportPlan = "LTS"
	// SupportUnknown indicates the support plan is not known.
	SupportUnknown SupportPlan = "UNKNOWN"
)

// LimitationSeverity indicates the severity of a limitation.
type LimitationSeverity string

const (
	// SeverityInfo is informational.
	SeverityInfo LimitationSeverity = "INFO"
	// SeverityWarn indicates a warning that may affect recommendations.
	SeverityWarn LimitationSeverity = "WARN"
	// SeverityError indicates an error that blocks provider analysis.
	SeverityError LimitationSeverity = "ERROR"
)

// Provider is the provider-neutral evidence interface.
type Provider interface {
	// Identity returns the detected provider and confidence.
	Identity() (ProviderType, Confidence)

	// Evidence retrieves upgrade availability for the cluster.
	Evidence(ctx context.Context, opts EvidenceOptions) (*ProviderEvidence, error)
}

// EvidenceOptions configures how evidence is retrieved.
type EvidenceOptions struct {
	// Mode specifies the evidence source mode.
	Mode SourceMode
	// FilePath is the path to a JSON evidence file (for file/offline modes).
	FilePath string
	// Subscription is the explicit Azure subscription (overrides detection).
	Subscription string
	// ResourceGroup is the explicit Azure resource group (overrides detection).
	ResourceGroup string
	// ClusterName is the explicit cluster name (overrides detection).
	ClusterName string
	// AllowPreview includes preview versions in candidates.
	AllowPreview bool
	// Timeout is the maximum time for CLI invocation.
	Timeout time.Duration
}

// ProviderEvidence contains upgrade availability from a provider.
type ProviderEvidence struct {
	// SchemaVersion identifies the evidence schema.
	SchemaVersion string `json:"schemaVersion"`
	// EvidenceID is a unique identifier for this evidence capture.
	EvidenceID string `json:"evidenceId"`
	// CapturedAt is when the evidence was captured.
	CapturedAt time.Time `json:"capturedAt"`
	// Source describes how the evidence was obtained.
	Source EvidenceSource `json:"source"`
	// Cluster contains provider identity information.
	Cluster ClusterIdentity `json:"cluster"`
	// CurrentVersion is the cluster's current Kubernetes version.
	CurrentVersion string `json:"currentVersion"`
	// ControlPlaneVersion is the control plane version if different.
	ControlPlaneVersion string `json:"controlPlaneVersion,omitempty"`
	// NodePools contains per-node-pool version information.
	NodePools []NodePoolEvidence `json:"nodePools,omitempty"`
	// AvailableUpgrades lists versions the cluster can upgrade to.
	AvailableUpgrades []UpgradeOption `json:"availableUpgrades"`
	// Limitations lists any evidence collection limitations.
	Limitations []Limitation `json:"limitations"`
}

// EvidenceSource describes how evidence was obtained.
type EvidenceSource struct {
	// Mode is the evidence source mode used.
	Mode SourceMode `json:"mode"`
	// Method is how the evidence was actually obtained.
	Method EvidenceMethod `json:"method"`
	// Command is the CLI command used (if applicable).
	Command string `json:"command,omitempty"`
	// AzureCLIVersion is the Azure CLI version (if applicable).
	AzureCLIVersion string `json:"azureCliVersion,omitempty"`
}

// ClusterIdentity contains provider-specific cluster identification.
type ClusterIdentity struct {
	// Provider is the detected provider type.
	Provider ProviderType `json:"provider"`
	// SubscriptionAlias is a redacted subscription identifier.
	SubscriptionAlias string `json:"subscriptionAlias,omitempty"`
	// ResourceGroupAlias is a redacted resource group identifier.
	ResourceGroupAlias string `json:"resourceGroupAlias,omitempty"`
	// ClusterNameAlias is a redacted cluster name.
	ClusterNameAlias string `json:"clusterNameAlias,omitempty"`
	// Region is the cluster's Azure region.
	Region string `json:"region,omitempty"`
	// IdentityConfidence indicates how reliably identity was determined.
	IdentityConfidence Confidence `json:"identityConfidence"`
}

// NodePoolEvidence contains per-node-pool version information.
type NodePoolEvidence struct {
	// NameAlias is a redacted node pool name.
	NameAlias string `json:"nameAlias"`
	// CurrentVersion is the node pool's current version.
	CurrentVersion string `json:"currentVersion"`
	// AvailableUpgrades lists versions the node pool can upgrade to.
	AvailableUpgrades []UpgradeOption `json:"availableUpgrades,omitempty"`
}

// UpgradeOption describes an available upgrade version.
type UpgradeOption struct {
	// Version is the Kubernetes version.
	Version string `json:"version"`
	// IsPreview indicates if this is a preview version.
	IsPreview bool `json:"isPreview"`
	// SupportPlan is the AKS support tier.
	SupportPlan SupportPlan `json:"supportPlan"`
}

// Limitation describes a limitation in evidence collection.
type Limitation struct {
	// Code is a unique identifier for the limitation type.
	Code string `json:"code"`
	// Severity indicates the impact of this limitation.
	Severity LimitationSeverity `json:"severity"`
	// Summary is a human-readable description.
	Summary string `json:"summary"`
}
