package kubent

type Report struct {
	DeprecatedAPIs []DeprecatedAPI `json:"DeprecatedAPIs"`
}

type DeprecatedAPI struct {
	Name        string `json:"Name"`
	Namespace   string `json:"Namespace"`
	Kind        string `json:"Kind"`
	APIVersion  string `json:"APIVersion"`
	ReplaceWith string `json:"ReplaceWith"`
	Since       string `json:"Since"`
	Deleted     bool   `json:"Deleted"`
}
