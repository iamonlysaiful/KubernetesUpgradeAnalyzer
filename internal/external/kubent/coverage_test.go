package kubent

import (
	"reflect"
	"testing"
)

func TestVerifyCoverageAcceptsDefaultTargetMinors(t *testing.T) {
	for _, target := range []string{"1.30", "1.31.4", "v1.32.0", "1.33.12"} {
		result := VerifyCoverage(target, DefaultCoveragePolicy())
		if result.Status != CoverageVerified {
			t.Fatalf("VerifyCoverage(%q) = %#v, want VERIFIED", target, result)
		}
		if result.RuleSet == "" {
			t.Fatalf("VerifyCoverage(%q) RuleSet is empty", target)
		}
	}
}

func TestVerifyCoverageRejectsMissingTargetMinor(t *testing.T) {
	result := VerifyCoverage("1.34.0", DefaultCoveragePolicy())

	if result.Status != CoverageUnverified {
		t.Fatalf("Status = %q, want %q", result.Status, CoverageUnverified)
	}
	if len(result.Limitations) != 1 || result.Limitations[0].Code != "TARGET_COVERAGE_MISSING" {
		t.Fatalf("Limitations = %#v, want TARGET_COVERAGE_MISSING", result.Limitations)
	}
}

func TestVerifyCoverageRejectsInvalidTarget(t *testing.T) {
	result := VerifyCoverage("not-a-version", DefaultCoveragePolicy())

	if result.Status != CoverageInconclusive {
		t.Fatalf("Status = %q, want %q", result.Status, CoverageInconclusive)
	}
	if len(result.Limitations) != 1 || result.Limitations[0].Code != "TARGET_VERSION_INVALID" {
		t.Fatalf("Limitations = %#v, want TARGET_VERSION_INVALID", result.Limitations)
	}
}

func TestNormalizeFindingsRequiresVerifiedCoverage(t *testing.T) {
	findings := NormalizeFindings(Report{}, SupportedVersion, "1.34.0", VerifyCoverage("1.34.0", DefaultCoveragePolicy()))

	if len(findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(findings))
	}
	if findings[0].Status != FindingUnknown {
		t.Fatalf("Status = %q, want %q", findings[0].Status, FindingUnknown)
	}
	if len(findings[0].Limitations) != 1 || findings[0].Limitations[0].Code != "TARGET_COVERAGE_UNVERIFIED" {
		t.Fatalf("Limitations = %#v, want TARGET_COVERAGE_UNVERIFIED", findings[0].Limitations)
	}
}

func TestNormalizeFindingsReturnsPassOnlyWhenCoverageVerifiedAndNoFindings(t *testing.T) {
	findings := NormalizeFindings(Report{}, SupportedVersion, "1.33.0", VerifyCoverage("1.33.0", DefaultCoveragePolicy()))

	if len(findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(findings))
	}
	if findings[0].Status != FindingPass {
		t.Fatalf("Status = %q, want %q", findings[0].Status, FindingPass)
	}
}

func TestNormalizeFindingsMapsDeletedAndDeprecatedAPIs(t *testing.T) {
	report := Report{DeprecatedAPIs: []DeprecatedAPI{
		{Name: "old", Namespace: "default", Kind: "Ingress", APIVersion: "extensions/v1beta1", ReplaceWith: "networking.k8s.io/v1", Since: "1.22", Deleted: true},
		{Name: "warn", Namespace: "alpha", Kind: "CronJob", APIVersion: "batch/v1beta1", ReplaceWith: "batch/v1", Since: "1.25", Deleted: false},
	}}

	findings := NormalizeFindings(report, SupportedVersion, "1.33.0", VerifyCoverage("1.33.0", DefaultCoveragePolicy()))

	got := findingKeys(findings)
	want := []string{
		"FAIL|default|Ingress|old|extensions/v1beta1",
		"WARN|alpha|CronJob|warn|batch/v1beta1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings = %#v, want %#v", got, want)
	}
}

func TestDecideKubentMVP(t *testing.T) {
	goDecision := DecideKubentMVP(VerifyAllTargets([]string{"1.30", "1.31", "1.32", "1.33"}, DefaultCoveragePolicy()))
	if goDecision.Status != DecisionGo {
		t.Fatalf("goDecision = %#v, want GO", goDecision)
	}

	noGoDecision := DecideKubentMVP(VerifyAllTargets([]string{"1.30", "1.34"}, DefaultCoveragePolicy()))
	if noGoDecision.Status != DecisionNoGo {
		t.Fatalf("noGoDecision = %#v, want NO_GO", noGoDecision)
	}
	if len(noGoDecision.Limitations) != 1 {
		t.Fatalf("limitations = %#v, want 1", noGoDecision.Limitations)
	}
}

func findingKeys(findings []Finding) []string {
	keys := make([]string, 0, len(findings))
	for _, finding := range findings {
		keys = append(keys, string(finding.Status)+"|"+finding.Resource.Namespace+"|"+finding.Resource.Kind+"|"+finding.Resource.Name+"|"+finding.APIVersion)
	}
	return keys
}
