# Phase 9 release candidate record

Status: Draft record
Last updated: 2026-07-27

This record tracks MVP release-candidate artifacts and final publication gates.

## 1. Release candidate metadata

- Candidate version:
- Git commit:
- Build date/time (UTC):
- Operator:

## 2. Artifact set

- Linux amd64 binary:
- Linux arm64 binary:
- macOS amd64 binary:
- macOS arm64 binary:
- SHA256SUMS:
- SBOM file:
- Provenance file:
- Release notes file:

## 3. Validation gates

- `scripts/ci-local.sh` pass:
- Race checks pass:
- `git diff --check` pass:
- `git fsck --full --strict` pass:
- No AppleDouble files in `.git`:
- Deterministic render checks pass:
- Workflow artifact upload pass:

## 3.1 Release-candidate command checklist

1. `scripts/release-candidate.sh <candidate-version>`
2. `scripts/ci-local.sh`
3. `go test -race ./internal/recommendation ./internal/report`
4. `git diff --check`
5. `git fsck --full --strict`
6. AppleDouble cleanup verification

Record results:

- Command 1 result:
- Command 2 result:
- Command 3 result:
- Command 4 result:
- Command 5 result:
- Command 6 result:

## 4. Publication controls

- Owner approval for tag/release publication (yes/no):
- Approved tag name:
- Approved release channel:
- Rollback/deprecation instructions confirmed (yes/no):

## 5. Notes and limitations

- Known limitations:
- Deferred work:
- Follow-up actions:

## 6. Publication hold point

Do not create tags, releases, or publish artifacts until owner explicitly
approves publication for the recorded candidate version.
