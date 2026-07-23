# Project status

Last updated: 2026-07-23

This page summarizes the current implementation phase, review gates, and
publication state. Detailed history remains in Git and `docs/change-log.md`.

## Gate status

| Gate | Status | Evidence |
| --- | --- | --- |
| Gate A - contracts | Complete | Phase 0 contracts, schemas, fixtures, security/RBAC, and dependency rules are documented and merged. |
| Gate B - collection safety | Not open | P2-01 preflight is in progress; live cluster access is still blocked until the user approves a named context and read-only operation. |
| Gate C - compatibility validity | Not started | Catalog source validation and kubent target-rule coverage are future Phase 4/5 work. |
| Gate D - recommendation calibration | Not started | Recommendation matrix and staging expectations are future Phase 7 work. |
| Gate E - release | Not started | Release validation, artifacts, SBOM, and publication are Phase 9 work. |

## Phase status

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 0 - Design freeze and contracts | Complete | Merged to `main`. |
| Phase 1 - CLI foundation | Complete | Merged to `main`, including local CI and GitHub Actions. |
| Phase 2 - Kubernetes preflight and inventory | In progress | Active branch `feature/kube-preflight`; no live cluster access has been performed. |
| Phase 3+ | Not started | Blocked on earlier phase outputs and review gates. |

## Current branch focus

`feature/kube-preflight` contains P2-01 work:

- kubeconfig/current or explicit context resolution;
- client-go dependency assessment and module updates;
- normalized preflight result and limitations;
- fake-client Kubernetes checker tests;
- `kua inventory` preflight-only command wiring.

Current `kua inventory` behavior is preflight-only and must be treated as an
early Phase 2 safety check, not full inventory collection.

## Current quality evidence

Latest local checks:

```text
scripts/ci-local.sh
git diff --check
git fsck --full --strict
```

`git fsck` reports only the accepted known dangling blobs. Any AppleDouble
`._*` files must be removed before publication after a verified recovery point.
