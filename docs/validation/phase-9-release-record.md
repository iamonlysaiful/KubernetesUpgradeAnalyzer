# Phase 9 release candidate record

Status: Prior release candidate superseded; provider validation incomplete
Last updated: 2026-07-27

This record tracks MVP release-candidate artifacts and final publication gates.

## 1. Release candidate metadata

- Candidate version: `0.1.0-rc.1` (superseded by merged Phase 8.5 recovery work)
- Git commit: `7528a7e2d5d5`
- Build date/time (UTC): 2026-07-27T03:27:18Z
- Operator: local maintainer through Codex-assisted run

## 2. Artifact set

- Linux amd64 binary: `artifacts/release-candidate/0.1.0-rc.1/bin/kua_0.1.0-rc.1_linux_amd64`
- Linux arm64 binary: `artifacts/release-candidate/0.1.0-rc.1/bin/kua_0.1.0-rc.1_linux_arm64`
- macOS amd64 binary: `artifacts/release-candidate/0.1.0-rc.1/bin/kua_0.1.0-rc.1_darwin_amd64`
- macOS arm64 binary: `artifacts/release-candidate/0.1.0-rc.1/bin/kua_0.1.0-rc.1_darwin_arm64`
- SHA256SUMS: `artifacts/release-candidate/0.1.0-rc.1/SHA256SUMS`
- SBOM file: `artifacts/release-candidate/0.1.0-rc.1/SBOM-go-modules.json`
- Provenance file: `artifacts/release-candidate/0.1.0-rc.1/provenance.json`
- Release notes file: `artifacts/release-candidate/0.1.0-rc.1/RELEASE_NOTES.md`

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

- Known limitations: the recorded candidate was built before merged Phase 8.5
  recovery work for end-to-end `kua analyze`, expanded live analysis inventory,
  kubent wiring, AKS advertised-edge handling, and component multi-version
  verdicts.
- Deferred work: rerun read-only AKS staging validation with the merged CLI,
  review redacted output, then generate a fresh release candidate.
- Follow-up actions: complete the Phase 8.5 validation gate with sanitized
  evidence policy before requesting explicit publication approval.

## 6. Publication hold point

Do not create tags, releases, or publish artifacts until owner explicitly
approves publication for the recorded candidate version.
