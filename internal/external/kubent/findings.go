package kubent

import "sort"

type FindingStatus string

const (
	FindingFail    FindingStatus = "FAIL"
	FindingWarn    FindingStatus = "WARN"
	FindingPass    FindingStatus = "PASS"
	FindingUnknown FindingStatus = "UNKNOWN"
)

type ResourceRef struct {
	Kind      string
	Namespace string
	Name      string
}

type Finding struct {
	AnalyzerVersion string
	TargetVersion   string
	Status          FindingStatus
	Resource        ResourceRef
	APIVersion      string
	Kind            string
	Replacement     string
	RemovedIn       string
	Limitations     []Limitation
}

type Limitation struct {
	Code    string
	Summary string
}

func NormalizeFindings(report Report, analyzerVersion string, targetVersion string, coverage CoverageResult) []Finding {
	if coverage.Status != CoverageVerified {
		return []Finding{{
			AnalyzerVersion: analyzerVersion,
			TargetVersion:   targetVersion,
			Status:          FindingUnknown,
			Limitations: []Limitation{{
				Code:    "TARGET_COVERAGE_UNVERIFIED",
				Summary: "kubent target-rule coverage is not verified",
			}},
		}}
	}
	if len(report.DeprecatedAPIs) == 0 {
		return []Finding{{
			AnalyzerVersion: analyzerVersion,
			TargetVersion:   targetVersion,
			Status:          FindingPass,
		}}
	}

	findings := make([]Finding, 0, len(report.DeprecatedAPIs))
	for _, api := range report.DeprecatedAPIs {
		status := FindingWarn
		if api.Deleted {
			status = FindingFail
		}
		findings = append(findings, Finding{
			AnalyzerVersion: analyzerVersion,
			TargetVersion:   targetVersion,
			Status:          status,
			Resource:        ResourceRef{Kind: api.Kind, Namespace: api.Namespace, Name: api.Name},
			APIVersion:      api.APIVersion,
			Kind:            api.Kind,
			Replacement:     api.ReplaceWith,
			RemovedIn:       api.Since,
		})
	}
	sort.SliceStable(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]
		if left.Status != right.Status {
			return findingStatusRank(left.Status) < findingStatusRank(right.Status)
		}
		if left.Resource.Namespace != right.Resource.Namespace {
			return left.Resource.Namespace < right.Resource.Namespace
		}
		if left.Resource.Kind != right.Resource.Kind {
			return left.Resource.Kind < right.Resource.Kind
		}
		if left.Resource.Name != right.Resource.Name {
			return left.Resource.Name < right.Resource.Name
		}
		return left.APIVersion < right.APIVersion
	})
	return findings
}

func findingStatusRank(status FindingStatus) int {
	switch status {
	case FindingFail:
		return 0
	case FindingWarn:
		return 1
	case FindingUnknown:
		return 2
	case FindingPass:
		return 3
	default:
		return 4
	}
}
