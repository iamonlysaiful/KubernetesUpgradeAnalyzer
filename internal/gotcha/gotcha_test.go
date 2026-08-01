package gotcha

import "testing"

func TestScanPath_VersionRange(t *testing.T) {
	// 1.30 → 1.35: CrossesMinor=29 is NOT > 30, so the sidecar gotcha must be absent.
	got := ScanPath(30, 35, false)
	codes := codeSet(got)

	if codes["SIDECAR_CONTAINERS_GA_1_29"] {
		t.Error("SIDECAR_CONTAINERS_GA_1_29 (CrossesMinor=29) must not appear when fromMinor=30")
	}
	if !codes["SA_TOKEN_LEGACY_CLEANUP_1_31"] {
		t.Error("missing SA_TOKEN_LEGACY_CLEANUP_1_31 for 1.30→1.35 path")
	}
	if !codes["FLOWCONTROL_V1BETA3_REMOVED_1_32"] {
		t.Error("missing FLOWCONTROL_V1BETA3_REMOVED_1_32 for 1.30→1.35 path")
	}
	// AKS-only entries must be absent for non-AKS clusters.
	if codes["AKS_LTS_1_32"] {
		t.Error("AKS_LTS_1_32 must not appear for non-AKS cluster")
	}
	if codes["AKS_NODE_POOL_UPGRADE_SEPARATE"] {
		t.Error("AKS_NODE_POOL_UPGRADE_SEPARATE must not appear for non-AKS cluster")
	}
}

func TestScanPath_NarrowRange(t *testing.T) {
	// 1.34 → 1.35: no catalog entries cross minor 35, so only AlwaysApply AKS
	// entries appear for AKS, and nothing for non-AKS.
	nonAKS := ScanPath(34, 35, false)
	if len(nonAKS) != 0 {
		t.Errorf("non-AKS 1.34→1.35: expected 0 gotchas, got %d: %v", len(nonAKS), codeSet(nonAKS))
	}

	aks := ScanPath(34, 35, true)
	codes := codeSet(aks)
	if !codes["AKS_NODE_POOL_UPGRADE_SEPARATE"] {
		t.Error("AKS_NODE_POOL_UPGRADE_SEPARATE (AlwaysApply) must appear for any AKS upgrade")
	}
}

func TestScanPath_AKSIncludesAKSOnlyGotchas(t *testing.T) {
	got := ScanPath(30, 35, true)
	codes := codeSet(got)

	if !codes["AKS_LTS_1_32"] {
		t.Error("missing AKS_LTS_1_32 for AKS 1.30→1.35 path")
	}
	if !codes["AKS_NODE_POOL_UPGRADE_SEPARATE"] {
		t.Error("missing AKS_NODE_POOL_UPGRADE_SEPARATE (AlwaysApply) for AKS upgrade")
	}
}

func TestScanPath_CrossesMinorBoundary(t *testing.T) {
	// 1.29 → 1.29 (same version): no crossing, zero gotchas.
	if got := ScanPath(29, 29, true); len(got) != 0 {
		// AlwaysApply AKS entries still appear; subtract those.
		for _, g := range got {
			if !g.AlwaysApply {
				t.Errorf("unexpected non-AlwaysApply gotcha for same-version scan: %s", g.Code)
			}
		}
	}

	// Exactly at boundary: CrossesMinor=29 with from=28, to=29 — should appear.
	got := ScanPath(28, 29, false)
	codes := codeSet(got)
	if !codes["SIDECAR_CONTAINERS_GA_1_29"] {
		t.Error("SIDECAR_CONTAINERS_GA_1_29 must appear when crossing 1.29 boundary (from=28, to=29)")
	}
}

// codeSet returns a set of Code values from a gotcha slice.
func codeSet(gs []Gotcha) map[string]bool {
	m := make(map[string]bool, len(gs))
	for _, g := range gs {
		m[g.Code] = true
	}
	return m
}
