# Project status

Last updated: 2026-07-23

This page summarizes the current implementation phase, review gates, and
publication state. Detailed history remains in Git and `docs/change-log.md`.

## Gate status

| Gate | Status | Evidence |
| --- | --- | --- |
| Gate A - contracts | Complete | Phase 0 contracts, schemas, fixtures, security/RBAC, and dependency rules are documented and merged. |
| Gate B - collection safety | Passed for P2-02 | P2-02 core inventory has fake-client/golden coverage and a locally approved live smoke-test record. Expanded live inventory is deferred and still requires separate Gate B expansion approval. |
| Gate C - compatibility validity | Not started | Catalog source validation and kubent target-rule coverage are future Phase 4/5 work. |
| Gate D - recommendation calibration | Not started | Recommendation matrix and staging expectations are future Phase 7 work. |
| Gate E - release | Not started | Release validation, artifacts, SBOM, and publication are Phase 9 work. |

## Phase status

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 0 - Design freeze and contracts | Complete | Merged to `main`. |
| Phase 1 - CLI foundation | Complete | Merged to `main`, including local CI and GitHub Actions. |
| Phase 2 - Kubernetes preflight and inventory | Complete | Fake-client inventory foundation is merged; live core inventory is verified; expanded live inventory is deferred. |
| Phase 3 - Health analysis | Complete | Health foundation and internal rules are merged; no expanded live collection was introduced. |
| Phase 4 - Component detection and catalog | In progress | Catalog loader foundation is being implemented without runtime internet access. |
| Phase 5+ | Not started | Blocked on earlier phase outputs and review gates. |

## Current branch focus

P2-01 and P2-02 are merged to `main`:

- kubeconfig/current or explicit context preflight;
- core namespace and node inventory collection;
- `kua inventory --format=json` partial snapshot wiring after required preflight
  succeeds;
- subset validation and command-level golden output coverage;
- Gate B live smoke-test plan and passed validation record;
- explicit limitations for intentionally empty future inventory groups.

P2-03 workload inventory is merged:

- workload collector contract;
- fake-client collector tests for Deployments, DaemonSets, StatefulSets,
  ReplicaSets, Jobs, and CronJobs;
- fake-client workload snapshot fixture path and subset validation;
- no live workload collection until separate Gate B expansion approval.

P2-03 CRD inventory is merged:

- CRD collector contract;
- fake-client CRD collector tests;
- fake-client CRD snapshot fixture path and subset validation;
- no live CRD collection until separate Gate B expansion approval.

P2-03 networking inventory is merged:

- networking collector contract;
- fake-client Service and Ingress collector tests;
- fake-client networking snapshot fixture path and subset validation;
- no live networking collection until separate Gate B expansion approval.

P2-03 storage inventory is merged:

- storage collector contract;
- fake-client PVC, PV, and StorageClass collector tests;
- fake-client storage snapshot fixture path and subset validation;
- no live storage collection until separate Gate B expansion approval.

P2-03 event inventory is merged:

- events collector contract;
- fake-client event collector tests;
- fake-client full P2-03 snapshot fixture path and subset validation;
- no live event collection until separate Gate B expansion approval.

Phase 2 consolidation is merged:

- inventory snapshot assembly contract;
- explicit fake-client collection options;
- preserved golden fixture coverage;
- no live expanded inventory collection until separate Gate B expansion approval.

Phase 2 closeout is merged:

- Phase 2 foundation is complete for MVP continuation;
- `kua inventory --format=json` remains live core inventory only;
- expanded live inventory is deferred until a later approved Gate B expansion.

Phase 3 health analysis is merged:

- health analysis contract;
- internal finding runner;
- node, workload, storage, and event health rules;
- Phase 3 closeout record;
- no expanded live inventory collection.

`docs/phase-4-catalog-plan` starts Phase 4:

- embedded catalog loader contract;
- component detection contract;
- compressed PR strategy to keep MVP delivery near 30 ±2 PRs.

P4-02 catalog loader foundation is in progress:

- embedded placeholder catalog;
- explicit local file loader;
- source-reference and duplicate-alias validation;
- checksum capture for loaded catalog bytes;
- no runtime internet access or catalog download behavior.

Current live `kua inventory` behavior remains partial/core inventory only.
Workloads, storage, networking, CRDs, events, health, compatibility, provider
evidence, recommendations, and reports are not emitted from live collection yet.

## Current quality evidence

Latest local checks:

```text
scripts/ci-local.sh
git diff --check
git fsck --full --strict
```

`git fsck` reports only the accepted known dangling blobs. Any AppleDouble
`._*` files must be removed before publication after a verified recovery point.
