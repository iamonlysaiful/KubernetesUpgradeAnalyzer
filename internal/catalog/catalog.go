package catalog

const SchemaVersion = "kua.catalog.v1"

type Bundle struct {
	SchemaVersion  string      `json:"schemaVersion"`
	CatalogVersion string      `json:"catalogVersion"`
	CreatedAt      string      `json:"createdAt"`
	ReviewAfter    string      `json:"reviewAfter,omitempty"`
	Kubernetes     Kubernetes  `json:"kubernetes"`
	Providers      []Provider  `json:"providers"`
	Components     []Component `json:"components"`
	Sources        []Source    `json:"sources"`
	Source         SourceKind  `json:"-"`
	ChecksumSHA256 string      `json:"-"`
}

type SourceKind string

const (
	SourceEmbedded SourceKind = "EMBEDDED"
	SourceFile     SourceKind = "FILE"
)

type Kubernetes struct {
	ValidatedRange ValidatedRange `json:"validatedRange"`
	APIRules       []APIRule      `json:"apiRules"`
}

type ValidatedRange struct {
	MinMinor string `json:"minMinor"`
	MaxMinor string `json:"maxMinor"`
}

type APIRule struct {
	RuleID      string `json:"ruleId"`
	APIVersion  string `json:"apiVersion"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	RemovedIn   string `json:"removedIn"`
	Replacement string `json:"replacement,omitempty"`
	SourceID    string `json:"sourceId"`
}

type Provider struct {
	Provider                string `json:"provider"`
	SequentialMinorUpgrades bool   `json:"sequentialMinorUpgrades"`
	CandidateSource         string `json:"candidateSource"`
	SourceID                string `json:"sourceId"`
}

type Component struct {
	ProductID           string   `json:"productId"`
	Aliases             []string `json:"aliases,omitempty"`
	ProductVersionRange string   `json:"productVersionRange"`
	KubernetesRange     string   `json:"kubernetesRange"`
	Status              string   `json:"status"`
	Conditions          []string `json:"conditions,omitempty"`
	Confidence          string   `json:"confidence"`
	ExpiresAt           string   `json:"expiresAt,omitempty"`
	SourceID            string   `json:"sourceId"`
}

type Source struct {
	SourceID    string `json:"sourceId"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	RetrievedAt string `json:"retrievedAt"`
	Claim       string `json:"claim"`
}
