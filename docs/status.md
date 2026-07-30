# Project status

Last updated: 2026-07-27

This page summarizes the current implementation phase, review gates, and
publication state. Detailed history remains in Git and `docs/change-log.md`.

## Gate status

| Gate | Status | Evidence |
| --- | --- | --- |
| Gate A - contracts | Complete | Phase 0 contracts, schemas, fixtures, security/RBAC, and dependency rules are documented and merged. |
| Gate B - collection safety | Passed for P2-02 | P2-02 core inventory has fake-client/golden coverage and a locally approved live smoke-test record. Expanded live inventory is deferred and still requires separate Gate B expansion approval. |
| Gate C - compatibility validity | Complete for Phase 5 | Catalog foundation is merged; kubent target-rule coverage validated for 1.30-1.33; go/no-go decision is GO for MVP. |
| Gate D - recommendation calibration | Complete for Phase 8 | Recommendation matrix outputs, deterministic path evaluation, and report rendering/redaction foundation are merged. |
| Gate E - release | Blocked pending publication approval | End-to-end `kua analyze` wiring and sanitized AKS staging validation have passed locally. Publication still requires a fresh post-helper release candidate and explicit owner approval. |

## Phase status

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 0 - Design freeze and contracts | Complete | Merged to `main`. |
| Phase 1 - CLI foundation | Complete | Merged to `main`, including local CI and GitHub Actions. |
| Phase 2 - Kubernetes preflight and inventory | Complete | Fake-client inventory foundation is merged; live core inventory is verified; expanded live inventory is deferred. |
| Phase 3 - Health analysis | Complete | Health foundation and internal rules are merged; no expanded live collection was introduced. |
| Phase 4 - Component detection and catalog | Complete | Catalog loader, detector framework, initial cohort, and closeout are merged; compatibility decisions remain deferred. |
| Phase 5 - API compatibility | Complete | Kubent adapter foundation, target-rule coverage for 1.30-1.33, go/no-go decision GO, and closeout are merged. |
| Phase 6 - AKS provider evidence | Complete | Provider interface, AKS identity/CLI/file adapters, candidate/path construction, and closeout are merged; no live CLI execution. |
| Phase 7 - Recommendation engine | Complete | Recommendation engine, finding aggregation, policy evaluation, AKS 1.30 validation case, and closeout are merged. |
| Phase 8 - Reports and hardening | Complete | Multi-format report rendering, redaction mode, hardening checks, and closeout are merged. |
| Phase 8.5 - End-to-end CLI recovery | Complete pending merge of helper | `kua analyze`, `health`, `compatibility`, `report`, live analysis inventory, kubent, AKS provider evidence, advertised-edge handling, component multi-version verdicts, redacted preflight errors, and component override workflow are implemented. The helper command is on `feature/component-overrides-helper` pending merge. |
| Phase 9 - Controlled staging validation and MVP release | Blocked pending fresh RC/publication approval | Sanitized staging validation passed locally. `0.1.0-rc.2` was generated before the component override helper, so a fresh RC should be generated after helper merge before any tag/release publication. |

## Current branch focus

`feature/interactive-cli-ux` (in progress, not yet merged):

- Interactive `kua analyze` confirmation: detects current kubectl context and
  server URL; prompts `[y/N]` before running; requires `--yes` in CI/scripts.
- Interactive component version prompt: unknown component versions are asked
  inline after first pass; re-analysis runs with supplied answers.
- Console renderer rewrite: operator-focused grouped output with BLOCKERS /
  WARNINGS / EVIDENCE GAPS sections and per-finding remediation hints.
- Snapshot validator changed to format-only version check; catalog
  `validatedRange` is now strictly informational.
- Interactive target-version prompt: when no `--target-version` and provider
  upgrades are unavailable, operator is asked for destination version; missing
  AKS auth prints the corrective `az aks get-upgrades` command.
- `scripts/install-local.sh` added for local pre-publish verification.

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

`docs/plans/phase-6-provider-plan` started Phase 6:

- provider interface and AKS identity detection;
- Azure CLI adapter for allowlisted `az aks get-upgrades`;
- file evidence adapter;
- candidate and sequential path construction;
- no live Azure CLI execution without separate explicit approval.

P6 provider foundation is merged:

- provider-neutral interface with Identity and Evidence methods;
- AKS identity detection with HIGH/MEDIUM/LOW/UNKNOWN confidence;
- Azure CLI adapter with mutating command rejection;
- file evidence adapter for exported JSON;
- candidate set and sequential path construction;
- Phase 6 closeout record;
- no live Azure CLI execution.

`docs/plans/phase-7-recommendation-plan` started Phase 7:

- recommendation engine contract and policy;
- finding aggregation across health, API, component, and provider evidence;
- deterministic readiness/risk outputs;
- AKS 1.30 -> 1.33 validation case;
- no live cluster access.

P7 recommendation engine is merged:

- recommendation engine implementation;
- readiness/risk policy evaluator;
- sequential destination/path construction;
- Phase 7 closeout record;
- no live cluster access.

`docs/plans/phase-8-reports-hardening-plan` started Phase 8:

- JSON/console/Markdown/HTML renderer scope;
- redacted mode requirements;
- hostile-input and determinism hardening checks.

P8 reports and hardening foundation is merged:

- `internal/report` package with deterministic multi-format renderers;
- self-contained HTML output and hostile-input escaping checks;
- redaction aliases for sensitive identifiers;
- atomic no-overwrite file writer;
- targeted race checks in CI/local quality gates;
- Phase 8 closeout record.

Current live `kua inventory` behavior remains partial/core inventory only, by
contract. `kua analyze` now uses the expanded MVP read-only metadata groups
needed for assessment: workloads, storage, networking, CRDs, and events.

Phase 8.5 recovery work merged to `main`:

- `kua analyze` renders a canonical assessment document and uses the
  recommendation policy once provider evidence and kubent API compatibility are
  available;
- `kua health` and `kua compatibility` expose filtered views of the same
  assessment document;
- `kua report --input <assessment.json>` renders saved assessment documents
  without cluster/provider access;
- live analysis collection includes the approved MVP read-only inventory groups;
- kubent is invoked through the controlled adapter with JSON output and Helm
  collection disabled;
- AKS provider evidence supports `auto`, `azure`, `file`, `offline`, and `none`
  source modes;
- provider-advertised AKS upgrade edges are followed when AKS no longer exposes
  intermediate lower minors, with a provider-direct multi-minor warning and at
  least `MEDIUM` risk unless future evidence proves lower-risk behavior;
- component detection preserves each confidently observed component version and
  emits per-version verdicts instead of hiding known versions behind one
  component-level `UNKNOWN`.

Phase 8.5 validation result:

- approved sanitized AKS staging validation passed locally;
- redacted output preserved versions, decisions, counts, and remediation;
- subscription, resource group, and cluster-name leak check returned zero leaks;
- after generated component overrides, the assessment returned
  `READY_WITH_WARNINGS` / `MEDIUM`, destination `1.34.9`, provider-valid
  `1.30.0 -> 1.34.9`, zero blockers, zero unknown findings, and only the
  provider-direct multi-minor warning;
- the component override helper command is implemented on
  `feature/component-overrides-helper` and must be merged before the next
  release candidate.

Approved hybrid AKS behavior follows provider-advertised upgrade edges. When AKS
offers only a direct higher target such as `1.34.x` and no lower intermediate
minors, KUA reports a provider-direct warning and at least `MEDIUM` risk instead
of inventing unavailable sequential stages.

## Current quality evidence

Latest local checks:

```text
scripts/ci-local.sh
git diff --check
git fsck --full --strict
```

`git fsck` reports only accepted known dangling objects. Any AppleDouble
`._*` files must be removed before publication after a verified recovery point.
