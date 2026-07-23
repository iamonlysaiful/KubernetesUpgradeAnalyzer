package inventory

type Snapshot struct {
	SchemaVersion string       `json:"schemaVersion"`
	SnapshotID    string       `json:"snapshotId"`
	CapturedAt    string       `json:"capturedAt"`
	Cluster       Cluster      `json:"cluster"`
	Kubernetes    Kubernetes   `json:"kubernetes"`
	Inventory     Inventory    `json:"inventory"`
	Limitations   []Limitation `json:"limitations"`
}

type Cluster struct {
	Identity ResourceRef `json:"identity"`
	Provider Provider    `json:"provider"`
	Context  Context     `json:"context"`
}

type Provider struct {
	Type       string `json:"type"`
	Confidence string `json:"confidence"`
}

type Context struct {
	Name             string `json:"name"`
	KubeconfigSource string `json:"kubeconfigSource,omitempty"`
}

type Kubernetes struct {
	ServerVersion string `json:"serverVersion"`
}

type Inventory struct {
	Namespaces []ResourceRef `json:"namespaces"`
	Nodes      []Node        `json:"nodes"`
	Workloads  []Workload    `json:"workloads"`
	Storage    []ResourceRef `json:"storage"`
	Networking []ResourceRef `json:"networking"`
	CRDs       []ResourceRef `json:"crds"`
	Events     []Event       `json:"events"`
}

type ResourceRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace,omitempty"`
	Name       string `json:"name"`
	UIDAlias   string `json:"uidAlias,omitempty"`
}

type Node struct {
	Ref               ResourceRef `json:"ref"`
	KubeletVersion    string      `json:"kubeletVersion"`
	ProviderIDPresent bool        `json:"providerIdPresent"`
	NodePool          string      `json:"nodePool,omitempty"`
	Conditions        []Condition `json:"conditions"`
}

type Condition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type Workload struct {
	Ref             ResourceRef `json:"ref"`
	DesiredReplicas int         `json:"desiredReplicas"`
	ReadyReplicas   int         `json:"readyReplicas"`
	Containers      []Container `json:"containers"`
}

type Container struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	ImageTag string `json:"imageTag,omitempty"`
}

type Event struct {
	Ref        ResourceRef `json:"ref"`
	Type       string      `json:"type"`
	Reason     string      `json:"reason"`
	LastSeenAt string      `json:"lastSeenAt"`
}

type Limitation struct {
	Code      string        `json:"code"`
	Severity  string        `json:"severity"`
	Summary   string        `json:"summary"`
	Resources []ResourceRef `json:"resources,omitempty"`
}
