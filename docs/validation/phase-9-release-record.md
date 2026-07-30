# Phase 9 release candidate record

Status: RC.2 generated locally; superseded — helper merged, interactive CLI UX in progress
Last updated: 2026-07-30

This record tracks MVP release-candidate artifacts and final publication gates.

## 1. Release candidate metadata

- Candidate version: `0.1.0-rc.2` (local artifact only; superseded for publication by unmerged component override helper)
- Git commit: `492894c`
- Build date/time (UTC): 2026-07-27 after Phase 8.5 validation merge
- Operator: local maintainer through Codex-assisted run

## 2. Artifact set

- Linux amd64 binary: `artifacts/release-candidate/0.1.0-rc.2/bin/kua_0.1.0-rc.2_linux_amd64`
- Linux arm64 binary: `artifacts/release-candidate/0.1.0-rc.2/bin/kua_0.1.0-rc.2_linux_arm64`
- macOS amd64 binary: `artifacts/release-candidate/0.1.0-rc.2/bin/kua_0.1.0-rc.2_darwin_amd64`
- macOS arm64 binary: `artifacts/release-candidate/0.1.0-rc.2/bin/kua_0.1.0-rc.2_darwin_arm64`
- SHA256SUMS: `artifacts/release-candidate/0.1.0-rc.2/SHA256SUMS`
- SBOM file: `artifacts/release-candidate/0.1.0-rc.2/SBOM-go-modules.json`
- Provenance file: `artifacts/release-candidate/0.1.0-rc.2/provenance.json`
- Release notes file: `artifacts/release-candidate/0.1.0-rc.2/RELEASE_NOTES.md`

## 3. Validation gates

- `scripts/ci-local.sh` pass: yes
- Race checks pass: yes; `go test -race ./internal/recommendation ./internal/report`
- `git diff --check` pass: yes
- `git fsck --full --strict` pass: yes, except the four accepted known dangling blobs
- No AppleDouble files in `.git`: yes
- Deterministic render checks pass: covered by `scripts/ci-local.sh`
- Workflow artifact upload pass: not run locally

## 3.1 Release-candidate command checklist

1. `scripts/release-candidate.sh <candidate-version>`
2. `scripts/ci-local.sh`
3. `go test -race ./internal/recommendation ./internal/report`
4. `git diff --check`
5. `git fsck --full --strict`
6. AppleDouble cleanup verification

Record results:

- Command 1 result: pass; artifacts generated under ignored local path. Go reported sandbox-limited module stat-cache warnings, but build output and follow-up checks passed.
- Command 2 result: pass
- Command 3 result: pass
- Command 4 result: pass
- Command 5 result: pass, except the four accepted known dangling blobs
- Command 6 result: pass

## 4. Publication controls

- Owner approval for tag/release publication (yes/no): no
- Approved tag name: not approved
- Approved release channel: not approved
- Rollback/deprecation instructions confirmed (yes/no): documented, not publication-approved

## 5. Notes and limitations

- Known limitations: `0.1.0-rc.2` was generated before the
  `feature/component-overrides-helper` branch. That helper is now merged to
  `main` in PR #36. `feature/interactive-cli-ux` (interactive analyze UX,
  format-only version validation, target version prompt) is in progress and
  must merge before `0.1.0-rc.3` is generated.
- Deferred work: merge `feature/interactive-cli-ux`, generate `0.1.0-rc.3`,
  rerun release gates, then request explicit publication approval.
- Follow-up actions: no tag, GitHub release, or artifact publication is approved
  yet.

## 6. Publication hold point

Do not create tags, releases, or publish artifacts until owner explicitly
approves publication for the recorded candidate version.
