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

## 3. Validation gates

- `scripts/ci-local.sh` pass:
- Race checks pass:
- `git diff --check` pass:
- `git fsck --full --strict` pass:
- No AppleDouble files in `.git`:
- Deterministic render checks pass:

## 4. Publication controls

- Owner approval for tag/release publication (yes/no):
- Approved tag name:
- Approved release channel:
- Rollback/deprecation instructions confirmed (yes/no):

## 5. Notes and limitations

- Known limitations:
- Deferred work:
- Follow-up actions:
