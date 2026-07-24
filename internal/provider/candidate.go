package provider

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// SemanticVersion represents a parsed Kubernetes version.
type SemanticVersion struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Raw        string
}

// ParseVersion parses a Kubernetes version string.
func ParseVersion(v string) (SemanticVersion, error) {
	sv := SemanticVersion{Raw: v}

	// Remove leading 'v' if present
	v = strings.TrimPrefix(v, "v")

	// Split on + or - for prerelease/metadata
	var base string
	if idx := strings.IndexAny(v, "+-"); idx >= 0 {
		base = v[:idx]
		sv.Prerelease = v[idx:]
	} else {
		base = v
	}

	parts := strings.Split(base, ".")
	if len(parts) < 2 {
		return sv, fmt.Errorf("invalid version format: %s", sv.Raw)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return sv, fmt.Errorf("invalid major version: %s", sv.Raw)
	}
	sv.Major = major

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return sv, fmt.Errorf("invalid minor version: %s", sv.Raw)
	}
	sv.Minor = minor

	if len(parts) >= 3 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return sv, fmt.Errorf("invalid patch version: %s", sv.Raw)
		}
		sv.Patch = patch
	}

	return sv, nil
}

// String returns the version as a string.
func (sv SemanticVersion) String() string {
	if sv.Raw != "" {
		return sv.Raw
	}
	base := fmt.Sprintf("%d.%d.%d", sv.Major, sv.Minor, sv.Patch)
	if sv.Prerelease != "" {
		return base + sv.Prerelease
	}
	return base
}

// MinorString returns just the major.minor portion.
func (sv SemanticVersion) MinorString() string {
	return fmt.Sprintf("%d.%d", sv.Major, sv.Minor)
}

// Compare compares two versions. Returns -1 if sv < other, 0 if equal, 1 if sv > other.
func (sv SemanticVersion) Compare(other SemanticVersion) int {
	if sv.Major != other.Major {
		if sv.Major < other.Major {
			return -1
		}
		return 1
	}
	if sv.Minor != other.Minor {
		if sv.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if sv.Patch != other.Patch {
		if sv.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// UpgradePath represents a sequential upgrade path from source to destination.
type UpgradePath struct {
	// Source is the starting version.
	Source SemanticVersion
	// Destination is the target version.
	Destination SemanticVersion
	// Steps contains each intermediate version in order.
	Steps []PathStep
	// IsValid indicates if all steps are provider-validated.
	IsValid bool
	// Limitations lists any path construction limitations.
	Limitations []Limitation
}

// PathStep represents one step in an upgrade path.
type PathStep struct {
	// From is the source version for this step.
	From SemanticVersion
	// To is the target version for this step.
	To SemanticVersion
	// IsProviderValid indicates if the provider confirms this upgrade is available.
	IsProviderValid bool
	// SupportPlan is the support tier of the target version.
	SupportPlan SupportPlan
	// IsPreview indicates if the target is a preview version.
	IsPreview bool
}

// CandidateSet contains available upgrade versions organized by minor.
type CandidateSet struct {
	// Current is the cluster's current version.
	Current SemanticVersion
	// ByMinor maps minor version to available patches, sorted descending.
	ByMinor map[int][]UpgradeOption
	// AllVersions contains all available versions.
	AllVersions []UpgradeOption
}

// BuildCandidateSet creates a CandidateSet from provider evidence.
func BuildCandidateSet(evidence *ProviderEvidence, allowPreview bool) (*CandidateSet, error) {
	current, err := ParseVersion(evidence.CurrentVersion)
	if err != nil {
		return nil, fmt.Errorf("parse current version: %w", err)
	}

	cs := &CandidateSet{
		Current:     current,
		ByMinor:     make(map[int][]UpgradeOption),
		AllVersions: make([]UpgradeOption, 0, len(evidence.AvailableUpgrades)),
	}

	for _, opt := range evidence.AvailableUpgrades {
		// Skip preview unless allowed
		if opt.IsPreview && !allowPreview {
			continue
		}

		sv, err := ParseVersion(opt.Version)
		if err != nil {
			continue // Skip unparseable versions
		}

		cs.AllVersions = append(cs.AllVersions, opt)
		cs.ByMinor[sv.Minor] = append(cs.ByMinor[sv.Minor], opt)
	}

	// Sort each minor's versions by patch descending (highest first)
	for minor := range cs.ByMinor {
		opts := cs.ByMinor[minor]
		sort.Slice(opts, func(i, j int) bool {
			vi, _ := ParseVersion(opts[i].Version)
			vj, _ := ParseVersion(opts[j].Version)
			return vi.Compare(vj) > 0 // Descending
		})
		cs.ByMinor[minor] = opts
	}

	return cs, nil
}

// HighestPatchForMinor returns the highest available patch for a minor version.
func (cs *CandidateSet) HighestPatchForMinor(minor int) (UpgradeOption, bool) {
	opts, ok := cs.ByMinor[minor]
	if !ok || len(opts) == 0 {
		return UpgradeOption{}, false
	}
	return opts[0], true // Already sorted descending
}

// HasMinor returns true if any version of the given minor is available.
func (cs *CandidateSet) HasMinor(minor int) bool {
	opts, ok := cs.ByMinor[minor]
	return ok && len(opts) > 0
}

// BuildSequentialPath constructs an upgrade path from current to destination.
// AKS requires sequential minor version upgrades (no skipping minors).
func BuildSequentialPath(candidates *CandidateSet, destination SemanticVersion) (*UpgradePath, error) {
	path := &UpgradePath{
		Source:      candidates.Current,
		Destination: destination,
		Steps:       make([]PathStep, 0),
		IsValid:     true,
	}

	// Validate destination is higher than current
	if destination.Compare(candidates.Current) <= 0 {
		return nil, fmt.Errorf("destination %s must be higher than current %s",
			destination.String(), candidates.Current.String())
	}

	// Check for minor skip
	minorDiff := destination.Minor - candidates.Current.Minor
	if minorDiff < 0 {
		return nil, fmt.Errorf("cannot downgrade from %s to %s",
			candidates.Current.String(), destination.String())
	}

	currentVersion := candidates.Current

	// Build sequential path through each minor
	for minor := candidates.Current.Minor + 1; minor <= destination.Minor; minor++ {
		var targetOpt UpgradeOption
		var targetVersion SemanticVersion
		var found bool

		if minor == destination.Minor {
			// For the final minor, try to use the exact destination patch
			opts, hasMinor := candidates.ByMinor[minor]
			if hasMinor {
				for _, opt := range opts {
					sv, _ := ParseVersion(opt.Version)
					if sv.Patch == destination.Patch {
						targetOpt = opt
						targetVersion = sv
						found = true
						break
					}
				}
				// If exact patch not found, use highest available
				if !found && len(opts) > 0 {
					targetOpt = opts[0]
					targetVersion, _ = ParseVersion(targetOpt.Version)
					found = true
					path.Limitations = append(path.Limitations, Limitation{
						Code:     "EXACT_PATCH_UNAVAILABLE",
						Severity: SeverityWarn,
						Summary:  fmt.Sprintf("Exact destination %s not available; using %s", destination.String(), targetVersion.String()),
					})
				}
			}
		} else {
			// For intermediate minors, use highest available patch
			targetOpt, found = candidates.HighestPatchForMinor(minor)
			if found {
				targetVersion, _ = ParseVersion(targetOpt.Version)
			}
		}

		if !found {
			path.IsValid = false
			path.Limitations = append(path.Limitations, Limitation{
				Code:     "MINOR_UNAVAILABLE",
				Severity: SeverityError,
				Summary:  fmt.Sprintf("No version available for minor %d.%d", candidates.Current.Major, minor),
			})
			// Continue to report full path gaps
			targetVersion = SemanticVersion{
				Major: candidates.Current.Major,
				Minor: minor,
				Patch: 0,
			}
		}

		step := PathStep{
			From:            currentVersion,
			To:              targetVersion,
			IsProviderValid: found,
			SupportPlan:     targetOpt.SupportPlan,
			IsPreview:       targetOpt.IsPreview,
		}
		path.Steps = append(path.Steps, step)

		currentVersion = targetVersion
	}

	// Handle same-minor patch upgrade
	if minorDiff == 0 && destination.Patch > candidates.Current.Patch {
		targetOpt, found := candidates.HighestPatchForMinor(destination.Minor)
		if found {
			targetVersion, _ := ParseVersion(targetOpt.Version)
			if targetVersion.Patch >= destination.Patch {
				step := PathStep{
					From:            candidates.Current,
					To:              targetVersion,
					IsProviderValid: true,
					SupportPlan:     targetOpt.SupportPlan,
					IsPreview:       targetOpt.IsPreview,
				}
				path.Steps = append(path.Steps, step)
			} else {
				path.IsValid = false
				path.Limitations = append(path.Limitations, Limitation{
					Code:     "PATCH_UNAVAILABLE",
					Severity: SeverityError,
					Summary:  fmt.Sprintf("Patch %s not available; highest is %s", destination.String(), targetVersion.String()),
				})
			}
		} else {
			path.IsValid = false
			path.Limitations = append(path.Limitations, Limitation{
				Code:     "NO_UPGRADES_AVAILABLE",
				Severity: SeverityError,
				Summary:  "No upgrade versions available from provider",
			})
		}
	}

	return path, nil
}
