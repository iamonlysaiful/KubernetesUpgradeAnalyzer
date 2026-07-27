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
