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
	// Build a coverage limitation to attach when rules aren't fully verified for
	// this target. When unverified, we still return real kubent findings so that
	// actual deprecated API hits are never silently discarded.
	var coverageLim *Limitation
	if coverage.Status != CoverageVerified {
		coverageLim = &Limitation{
			Code:    "TARGET_COVERAGE_UNVERIFIED",
			Summary: "kubent target-rule coverage is not verified for " + targetVersion,
		}
	}

	if len(report.DeprecatedAPIs) == 0 {
		status := FindingPass
		var limitations []Limitation
		if coverageLim != nil {
			status = FindingUnknown
			limitations = []Limitation{*coverageLim}
		}
		return []Finding{{
			AnalyzerVersion: analyzerVersion,
			TargetVersion:   targetVersion,
			Status:          status,
			Limitations:     limitations,
		}}
	}

	findings := make([]Finding, 0, len(report.DeprecatedAPIs))
	for _, api := range report.DeprecatedAPIs {
		status := FindingWarn
		if api.Deleted {
			status = FindingFail
		}
		var limitations []Limitation
		if coverageLim != nil {
			limitations = []Limitation{*coverageLim}
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
			Limitations:     limitations,
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
