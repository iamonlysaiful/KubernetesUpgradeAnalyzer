package kubent

import "sort"

type CoverageStatus string

const (
	CoverageVerified     CoverageStatus = "VERIFIED"
	CoverageUnverified   CoverageStatus = "UNVERIFIED"
	CoverageInconclusive CoverageStatus = "INCONCLUSIVE"
)

type CoverageResult struct {
	TargetVersion string
	Status        CoverageStatus
	RuleSet       string
	Limitations   []Limitation
}

type CoveragePolicy struct {
	VerifiedTargetMinors map[string]string
}

func DefaultCoveragePolicy() CoveragePolicy {
	return CoveragePolicy{
		VerifiedTargetMinors: map[string]string{
			"1.30": "kubent-0.7.3-fixture",
			"1.31": "kubent-0.7.3-fixture",
			"1.32": "kubent-0.7.3-fixture",
			"1.33": "kubent-0.7.3-fixture",
		},
	}
}

func VerifyCoverage(targetVersion string, policy CoveragePolicy) CoverageResult {
	targetMinor := minor(targetVersion)
	if targetMinor == "" {
		return CoverageResult{
			TargetVersion: targetVersion,
			Status:        CoverageInconclusive,
			Limitations: []Limitation{{
				Code:    "TARGET_VERSION_INVALID",
				Summary: "target version cannot be normalized to a Kubernetes minor",
			}},
		}
	}
	if policy.VerifiedTargetMinors == nil {
		policy.VerifiedTargetMinors = map[string]string{}
	}
	ruleSet, ok := policy.VerifiedTargetMinors[targetMinor]
	if !ok {
		return CoverageResult{
			TargetVersion: targetVersion,
			Status:        CoverageUnverified,
			Limitations: []Limitation{{
				Code:    "TARGET_COVERAGE_MISSING",
				Summary: "kubent rule coverage is not verified for target minor " + targetMinor,
			}},
		}
	}
	return CoverageResult{TargetVersion: targetVersion, Status: CoverageVerified, RuleSet: ruleSet}
}

func VerifyAllTargets(targetVersions []string, policy CoveragePolicy) []CoverageResult {
	results := make([]CoverageResult, 0, len(targetVersions))
	for _, targetVersion := range targetVersions {
		results = append(results, VerifyCoverage(targetVersion, policy))
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].TargetVersion < results[j].TargetVersion
	})
	return results
}

func minor(version string) string {
	if len(version) < 4 {
		return ""
	}
	if version[0] == 'v' {
		version = version[1:]
	}
	parts := 0
	for i := 0; i < len(version); i++ {
		if version[i] == '.' {
			parts++
			if parts == 2 {
				return version[:i]
			}
		}
	}
	if parts == 1 {
		return version
	}
	return ""
}
