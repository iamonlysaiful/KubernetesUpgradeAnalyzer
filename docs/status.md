# Project status

Last updated: 2026-07-24

This page summarizes the current implementation phase, review gates, and
publication state. Detailed history remains in Git and `docs/change-log.md`.

## Gate status

| Gate | Status | Evidence |
| --- | --- | --- |
| Gate A - contracts | Complete | Phase 0 contracts, schemas, fixtures, security/RBAC, and dependency rules are documented and merged. |
| Gate B - collection safety | Passed for P2-02 | P2-02 core inventory has fake-client/golden coverage and a locally approved live smoke-test record. Expanded live inventory is deferred and still requires separate Gate B expansion approval. |
| Gate C - compatibility validity | Complete for Phase 5 | Catalog foundation is merged; kubent target-rule coverage validated for 1.30-1.33; go/no-go decision is GO for MVP. |
| Gate D - recommendation calibration | Not started | Recommendation matrix and staging expectations are future Phase 7 work. |
| Gate E - release | Not started | Release validation, artifacts, SBOM, and publication are Phase 9 work. |

## Phase status

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 0 - Design freeze and contracts | Complete | Merged to `main`. |
| Phase 1 - CLI foundation | Complete | Merged to `main`, including local CI and GitHub Actions. |
| Phase 2 - Kubernetes preflight and inventory | Complete | Fake-client inventory foundation is merged; live core inventory is verified; expanded live inventory is deferred. |
| Phase 3 - Health analysis | Complete | Health foundation and internal rules are merged; no expanded live collection was introduced. |
| Phase 4 - Component detection and catalog | Complete | Catalog loader, detector framework, initial cohort, and closeout are merged; compatibility decisions remain deferred. |
| Phase 5 - API compatibility | Complete | Kubent adapter foundation, target-rule coverage for 1.30-1.33, go/no-go decision GO, and closeout are merged. |
| Phase 6 - AKS provider evidence | In progress | Provider plan proposed; implementation pending. |
| Phase 7+ | Not started | Blocked on earlier phase outputs and review gates. |

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

Phase 4 catalog and component detection is merged:

- embedded catalog loader contract;
- component detection contract;
- compressed PR strategy to keep MVP delivery near 30 ±2 PRs.
- embedded placeholder catalog;
- explicit local file loader;
- source-reference and duplicate-alias validation;
- checksum capture for loaded catalog bytes;
- detection result model;
- detector runner;
- deterministic ordering;
- version normalization that preserves `UNKNOWN` for ambiguous evidence;
- NGINX Ingress, CoreDNS, Metrics Server, Azure Disk CSI, Azure File CSI,
  Fluent Bit, and EMQX workload-backed detectors;
- Phase 4 closeout record;
- no compatibility decisions, provider evidence, kubent, recommendations,
  runtime internet access, or expanded live reads.

`docs/phase-5-api-compat-plan` started Phase 5:

- API compatibility and kubent adapter contract;
- target-rule coverage validation plan;
- go/no-go decision requirement for kubent coverage across `1.30` through
  `1.33`;
- no live kubent execution without separate explicit approval.

P5-02 kubent adapter foundation is merged:

- process runner abstraction;
- kubent `0.7.3` version validation;
- shell-free argument construction;
- bounded stdout/stderr handling;
- JSON parsing fixtures;
- no live kubent execution.

P5-03 kubent coverage decision is merged:

- target-rule coverage validation for `1.30` through `1.33`;
- normalized API findings;
- go/no-go helper returns GO for MVP;
- Phase 5 closeout record;
- no live kubent execution.

`docs/plans/phase-6-provider-plan` starts Phase 6:

- provider interface and AKS identity detection;
- Azure CLI adapter for allowlisted `az aks get-upgrades`;
- file evidence adapter;
- candidate and sequential path construction;
- no live Azure CLI execution without separate explicit approval.

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
