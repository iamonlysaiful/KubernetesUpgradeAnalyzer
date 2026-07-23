package preflight

type Status string

const (
	StatusPass    Status = "PASS"
	StatusFail    Status = "FAIL"
	StatusUnknown Status = "UNKNOWN"
)

type EvidenceClass string

const (
	EvidenceRequired EvidenceClass = "REQUIRED"
	EvidenceOptional EvidenceClass = "OPTIONAL"
)

type Result struct {
	Context          ContextSelection
	ServerVersion    string
	DiscoveryStatus  Status
	PermissionChecks []PermissionCheck
	Limitations      []Limitation
}

type PermissionCheck struct {
	Resource      string
	Verb          string
	EvidenceClass EvidenceClass
	Status        Status
	Reason        string
}

type Limitation struct {
	Code     string
	Severity string
	Summary  string
}

func (r Result) HasRequiredFailure() bool {
	if r.DiscoveryStatus == StatusFail {
		return true
	}
	for _, check := range r.PermissionChecks {
		if check.EvidenceClass == EvidenceRequired && check.Status == StatusFail {
			return true
		}
	}
	return false
}
