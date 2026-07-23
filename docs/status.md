# Project status

Last updated: 2026-07-23

This page summarizes the current implementation phase, review gates, and
publication state. Detailed history remains in Git and `docs/change-log.md`.

## Gate status

| Gate | Status | Evidence |
| --- | --- | --- |
| Gate A - contracts | Complete | Phase 0 contracts, schemas, fixtures, security/RBAC, and dependency rules are documented and merged. |
| Gate B - collection safety | Passed for P2-02 | P2-02 core inventory has fake-client/golden coverage and a locally approved live smoke-test record. Later collectors still require separate Gate B expansion approval. |
| Gate C - compatibility validity | Not started | Catalog source validation and kubent target-rule coverage are future Phase 4/5 work. |
| Gate D - recommendation calibration | Not started | Recommendation matrix and staging expectations are future Phase 7 work. |
| Gate E - release | Not started | Release validation, artifacts, SBOM, and publication are Phase 9 work. |

## Phase status

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 0 - Design freeze and contracts | Complete | Merged to `main`. |
| Phase 1 - CLI foundation | Complete | Merged to `main`, including local CI and GitHub Actions. |
| Phase 2 - Kubernetes preflight and inventory | In progress | P2-03 workload inventory has started on `feature/kube-workload-collectors`; Gate B passed for namespace/node collection only. |
| Phase 3+ | Not started | Blocked on earlier phase outputs and review gates. |

## Current branch focus

P2-01 and P2-02 are merged to `main`:

- kubeconfig/current or explicit context preflight;
- core namespace and node inventory collection;
- `kua inventory --format=json` partial snapshot wiring after required preflight
  succeeds;
- subset validation and command-level golden output coverage;
- Gate B live smoke-test plan and passed validation record;
- explicit limitations for intentionally empty future inventory groups.

`feature/kube-workload-collectors` starts P2-03 workload inventory:

- workload collector contract;
- fake-client tests first;
- no live workload collection until separate Gate B expansion approval.

Current `kua inventory` behavior remains partial/core inventory only. Workloads,
storage, networking, CRDs, events, health, compatibility, provider evidence,
recommendations, and reports are not collected yet.

## Current quality evidence

Latest local checks:

```text
scripts/ci-local.sh
git diff --check
git fsck --full --strict
```

`git fsck` reports only the accepted known dangling blobs. Any AppleDouble
`._*` files must be removed before publication after a verified recovery point.
