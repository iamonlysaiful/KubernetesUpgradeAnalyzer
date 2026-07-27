package recommendation

import (
	"testing"

	"github.com/iamonlysaiful/KubernetesUpgradeAnalyzer/internal/components"
)

func TestAggregateComponentsDeduplicatesUnknownVersions(t *testing.T) {
	aggregator := NewAggregator()

	findings := aggregator.AggregateComponents([]components.Detection{
		{ComponentID: "coredns", Name: "CoreDNS", Version: "UNKNOWN", Status: components.StatusUnknown},
		{ComponentID: "coredns", Name: "CoreDNS", Version: "UNKNOWN", Status: components.StatusUnknown},
		{ComponentID: "emqx", Name: "EMQX", Version: "UNKNOWN", Status: components.StatusUnknown},
	}, 34)

	if len(findings) != 2 {
		t.Fatalf("findings = %#v, want two component unknown findings", findings)
	}
	if findings[0].ID != "COMPONENT_coredns_UNKNOWN" || findings[1].ID != "COMPONENT_emqx_UNKNOWN" {
		t.Fatalf("findings = %#v, want coredns and emqx unknown findings", findings)
	}
}

func TestAggregateComponentsEmitsVerdictPerObservedVersion(t *testing.T) {
	aggregator := NewAggregator()

	findings := aggregator.AggregateComponents([]components.Detection{
		{ComponentID: "emqx", Name: "EMQX", Version: "5.8.1", Status: components.StatusFound},
		{ComponentID: "emqx", Name: "EMQX", Version: "5.8.2", Status: components.StatusFound},
		{ComponentID: "emqx", Name: "EMQX", Version: "5.8.2", Status: components.StatusFound},
	}, 34)

	if len(findings) != 2 {
		t.Fatalf("findings = %#v, want two component version verdicts", findings)
	}
	if findings[0].ID != "COMPONENT_emqx_5.8.1_OBSERVED" || findings[1].ID != "COMPONENT_emqx_5.8.2_OBSERVED" {
		t.Fatalf("findings = %#v, want one finding per observed version", findings)
	}
	for _, finding := range findings {
		if finding.Severity != SeverityInfo {
			t.Fatalf("finding = %#v, want observed-version INFO verdict", finding)
		}
	}
}
