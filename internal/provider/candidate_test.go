package provider

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"full version", "1.30.4", 1, 30, 4, false},
		{"with v prefix", "v1.31.2", 1, 31, 2, false},
		{"major.minor only", "1.32", 1, 32, 0, false},
		{"with prerelease", "1.33.0-preview", 1, 33, 0, false},
		{"with build metadata", "1.30.5+aks", 1, 30, 5, false},
		{"invalid", "invalid", 0, 0, 0, true},
		{"empty", "", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if sv.Major != tt.major || sv.Minor != tt.minor || sv.Patch != tt.patch {
				t.Errorf("ParseVersion(%q) = %d.%d.%d, want %d.%d.%d",
					tt.input, sv.Major, sv.Minor, sv.Patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

func TestSemanticVersion_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal", "1.30.4", "1.30.4", 0},
		{"a major less", "1.30.4", "2.30.4", -1},
		{"a major greater", "2.30.4", "1.30.4", 1},
		{"a minor less", "1.30.4", "1.31.4", -1},
		{"a minor greater", "1.31.4", "1.30.4", 1},
		{"a patch less", "1.30.4", "1.30.5", -1},
		{"a patch greater", "1.30.5", "1.30.4", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, _ := ParseVersion(tt.a)
			b, _ := ParseVersion(tt.b)
			if got := a.Compare(b); got != tt.want {
				t.Errorf("Compare(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSemanticVersion_MinorString(t *testing.T) {
	sv, _ := ParseVersion("1.30.4")
	if got := sv.MinorString(); got != "1.30" {
		t.Errorf("MinorString() = %v, want 1.30", got)
	}
}

func TestBuildCandidateSet(t *testing.T) {
	evidence := &ProviderEvidence{
		CurrentVersion: "1.30.4",
		AvailableUpgrades: []UpgradeOption{
			{Version: "1.30.5", IsPreview: false},
			{Version: "1.30.6", IsPreview: false},
			{Version: "1.31.0", IsPreview: false},
			{Version: "1.31.1", IsPreview: false},
			{Version: "1.32.0", IsPreview: false},
			{Version: "1.33.0", IsPreview: true},
		},
	}

	// Without preview
	cs, err := BuildCandidateSet(evidence, false)
	if err != nil {
		t.Fatalf("BuildCandidateSet: %v", err)
	}

	if cs.Current.Minor != 30 {
		t.Errorf("Current.Minor = %d, want 30", cs.Current.Minor)
	}

	// Should have 5 versions (excluding preview 1.33.0)
	if len(cs.AllVersions) != 5 {
		t.Errorf("AllVersions count = %d, want 5", len(cs.AllVersions))
	}

	// Check ByMinor
	if len(cs.ByMinor[30]) != 2 {
		t.Errorf("ByMinor[30] count = %d, want 2", len(cs.ByMinor[30]))
	}
	if len(cs.ByMinor[31]) != 2 {
		t.Errorf("ByMinor[31] count = %d, want 2", len(cs.ByMinor[31]))
	}
	if len(cs.ByMinor[32]) != 1 {
		t.Errorf("ByMinor[32] count = %d, want 1", len(cs.ByMinor[32]))
	}
	if cs.HasMinor(33) {
		t.Error("should not have minor 33 without preview")
	}

	// With preview
	csWithPreview, err := BuildCandidateSet(evidence, true)
	if err != nil {
		t.Fatalf("BuildCandidateSet with preview: %v", err)
	}
	if !csWithPreview.HasMinor(33) {
		t.Error("should have minor 33 with preview")
	}
}

func TestCandidateSet_HighestPatchForMinor(t *testing.T) {
	evidence := &ProviderEvidence{
		CurrentVersion: "1.30.4",
		AvailableUpgrades: []UpgradeOption{
			{Version: "1.31.0", IsPreview: false},
			{Version: "1.31.2", IsPreview: false},
			{Version: "1.31.1", IsPreview: false},
		},
	}

	cs, _ := BuildCandidateSet(evidence, false)

	highest, ok := cs.HighestPatchForMinor(31)
	if !ok {
		t.Fatal("expected to find minor 31")
	}
	if highest.Version != "1.31.2" {
		t.Errorf("HighestPatchForMinor(31) = %v, want 1.31.2", highest.Version)
	}

	_, ok = cs.HighestPatchForMinor(32)
	if ok {
		t.Error("should not find minor 32")
	}
}

func TestBuildSequentialPath(t *testing.T) {
	evidence := &ProviderEvidence{
		CurrentVersion: "1.30.4",
		AvailableUpgrades: []UpgradeOption{
			{Version: "1.30.6", IsPreview: false},
			{Version: "1.31.2", IsPreview: false},
			{Version: "1.32.1", IsPreview: false},
			{Version: "1.33.0", IsPreview: false},
		},
	}

	cs, _ := BuildCandidateSet(evidence, false)
	dest, _ := ParseVersion("1.33.0")

	path, err := BuildSequentialPath(cs, dest)
	if err != nil {
		t.Fatalf("BuildSequentialPath: %v", err)
	}

	if !path.IsValid {
		t.Errorf("expected valid path, limitations: %v", path.Limitations)
	}

	if len(path.Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(path.Steps))
	}

	// Verify sequential: 1.30.4 -> 1.31.2 -> 1.32.1 -> 1.33.0
	expectedSteps := []struct {
		from string
		to   string
	}{
		{"1.30.4", "1.31.2"},
		{"1.31.2", "1.32.1"},
		{"1.32.1", "1.33.0"},
	}

	for i, exp := range expectedSteps {
		if path.Steps[i].From.String() != exp.from {
			t.Errorf("step %d from = %v, want %v", i, path.Steps[i].From.String(), exp.from)
		}
		if path.Steps[i].To.String() != exp.to {
			t.Errorf("step %d to = %v, want %v", i, path.Steps[i].To.String(), exp.to)
		}
	}
}

func TestBuildSequentialPath_MissingMinor(t *testing.T) {
	evidence := &ProviderEvidence{
		CurrentVersion: "1.30.4",
		AvailableUpgrades: []UpgradeOption{
			{Version: "1.31.2", IsPreview: false},
			// Missing 1.32
			{Version: "1.33.0", IsPreview: false},
		},
	}

	cs, _ := BuildCandidateSet(evidence, false)
	dest, _ := ParseVersion("1.33.0")

	path, err := BuildSequentialPath(cs, dest)
	if err != nil {
		t.Fatalf("BuildSequentialPath: %v", err)
	}

	if path.IsValid {
		t.Error("expected invalid path due to missing minor 32")
	}

	// Should have limitation about missing minor
	var hasMissingMinor bool
	for _, lim := range path.Limitations {
		if lim.Code == "MINOR_UNAVAILABLE" {
			hasMissingMinor = true
			break
		}
	}
	if !hasMissingMinor {
		t.Error("expected MINOR_UNAVAILABLE limitation")
	}
}

func TestBuildSequentialPath_DowngradeError(t *testing.T) {
	evidence := &ProviderEvidence{
		CurrentVersion: "1.32.0",
		AvailableUpgrades: []UpgradeOption{
			{Version: "1.31.2", IsPreview: false},
		},
	}

	cs, _ := BuildCandidateSet(evidence, false)
	dest, _ := ParseVersion("1.31.0")

	_, err := BuildSequentialPath(cs, dest)
	if err == nil {
		t.Error("expected error for downgrade attempt")
	}
}

func TestBuildSequentialPath_SameMinorPatch(t *testing.T) {
	evidence := &ProviderEvidence{
		CurrentVersion: "1.30.4",
		AvailableUpgrades: []UpgradeOption{
			{Version: "1.30.5", IsPreview: false},
			{Version: "1.30.6", IsPreview: false},
		},
	}

	cs, _ := BuildCandidateSet(evidence, false)
	dest, _ := ParseVersion("1.30.6")

	path, err := BuildSequentialPath(cs, dest)
	if err != nil {
		t.Fatalf("BuildSequentialPath: %v", err)
	}

	if !path.IsValid {
		t.Error("expected valid path for same-minor patch upgrade")
	}

	if len(path.Steps) != 1 {
		t.Errorf("expected 1 step for patch upgrade, got %d", len(path.Steps))
	}
}
