# MVP implementation plan

Status: Approved plan; implementation in progress
Last updated: 2026-07-23

## 1. Delivery model

Implement one thin, testable capability at a time. Every work package follows the docs-first workflow and ends with its own verification evidence. Do not create a long-lived branch containing multiple unreviewed phases.

## 2. Work packages

| ID | Package | Depends on | Primary evidence |
| --- | --- | --- | --- |
| P0-01 | Lock domain enums and version rules | None | Contract review |
| P0-02 | Assessment and snapshot schemas | P0-01 | Valid/invalid fixtures |
| P0-03 | Catalog and provider-evidence schemas | P0-01 | Schema fixtures |
| P0-04 | RBAC, redaction, dependency, and threat contracts | P0-01 | Security review |
| P1-01 | Go module, CLI skeleton, build/version | Phase 0 | Cross-platform build |
| P1-02 | Config/logging/error/exit foundation | P1-01 | Unit/CLI tests |
| P1-03 | CI and quality gates | P1-01 | Passing workflow |
| P2-01 | Kubeconfig/context/preflight | P1 | Fake-client tests |
| P2-02 | Core cluster collectors | P2-01 | Snapshot fixtures |
| P2-03 | Workload/storage/network/CRD collectors | P2-02 | Inventory golden tests |
| P3-01 | Health rule framework | P2 | Rule unit tests |
| P3-02 | Health rules and event window | P3-01 | Boundary fixtures |
| P4-01 | Catalog loader/embed/validation | P0, P1 | Integrity tests |
| P4-02 | Detector framework/version normalization | P2, P4-01 | Detection contracts |
| P4-03 | First component detector cohort | P4-02 | Product fixtures |
| P5-01 | Kubent process adapter | P1, P2 | JSON/process fixtures |
| P5-02 | Rule-coverage validation and MVP go/no-go | P5-01 | Coverage report |
| P6-01 | Provider interface and AKS identity | P2 | Confidence fixtures |
| P6-02 | Azure CLI and file evidence adapters | P6-01 | Mode/fallback tests |
| P6-03 | Candidate/path construction | P6-02 | Graph fixtures |
| P7-01 | Finding aggregation | P3–P6 | Deterministic tests |
| P7-02 | Readiness/risk/destination engine | P7-01 | Decision matrix |
| P8-01 | JSON/console renderers | P7 | Schema/golden tests |
| P8-02 | Markdown/HTML/redaction | P8-01 | Security/golden tests |
| P8-03 | Performance and hardening | P2–P8 | Bench/race/security evidence |
| P9-01 | Approved AKS staging validation | P8 | Sanitized validation record |
| P9-02 | Release candidate and MVP release | P9-01 | Signed artifacts/checksums/SBOM |

## 3. Review gates

- Gate A — contracts: approve Phase 0 artifacts before module initialization.
- Gate B — collection safety: approve RBAC and snapshot fields before live access.
- Gate C — compatibility validity: verify catalog sources and kubent target coverage before recommendation claims.
- Gate D — recommendation calibration: approve decision matrix and staging expectation before full reports.
- Gate E — release: approve sanitized validation evidence and artifacts before publication.

## 4. Branch and commit approach

Use short-lived `docs/`, `feature/`, `fix/`, or `chore/` branches. Each behavior change has an approved documentation commit followed by a separate implementation commit. Prefer work packages small enough to review independently; do not combine collectors, recommendation policy, and rendering in one change.

## 5. Dependency baseline

Planned direct dependencies are Go standard library, Cobra, Viper, `client-go`, and schema/testing/lint tooling selected during Phase 0 review. Each dependency requires version, license, purpose, alternatives, maintenance, and vulnerability assessment before addition. Azure SDK is not needed for MVP because provider access uses the installed Azure CLI.

P1-01 starts with the Go standard library only and is governed by
`docs/contracts/foundation-contract.md`. External modules are intentionally
deferred to focused dependency-assessment commits.

P1-02 remains standard-library only and is governed by
`docs/contracts/config-error-contract.md`. It validates common flags and error
mapping without reading config files, accessing Kubernetes, invoking providers,
or rendering reports.

P1-03 is governed by `docs/contracts/ci-quality-contract.md`. It introduces
GitHub Actions and local quality gates only; release automation, dependency
scanning, schema semantic validation, and live-system checks remain deferred.

P2-01 is governed by `docs/contracts/kube-preflight-contract.md`. It adds
client-go kubeconfig/context/preflight implementation with fake-client tests.
Live cluster execution still requires separate approval naming the context and
read-only operation.

P2-02 is governed by `docs/contracts/core-inventory-contract.md`. It adds the
first partial/core snapshot path for cluster metadata, server version,
namespaces, nodes, and explicit limitations. Workloads, storage, networking,
CRDs, events, health, compatibility, provider evidence, recommendations, and
reports remain out of scope.

P2-03 workload inventory is governed by
`docs/contracts/workload-inventory-contract.md`. It starts with fake-client-only
workload collection for Kubernetes controller resources and does not approve live
workload reads until Gate B is separately expanded.

Phase 2 closes the MVP inventory foundation with live core inventory only and
fake-client/golden coverage for the expanded inventory groups. Expanded live
workload, CRD, networking, storage, and event reads are deferred until a later
Gate B expansion is explicitly approved.

## 6. Stop conditions

Stop affected work when a schema or accepted document conflicts with implementation, live access would exceed approved read-only permissions, kubent lacks verified target rules, a catalog claim lacks adequate provenance, redaction leaks an identifier, deterministic fixtures disagree, or ExFAT metadata makes Git integrity fail.
