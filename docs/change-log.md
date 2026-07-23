# Documentation change log

This log records material scope and architecture changes. Git remains the detailed history.

## 2026-07-23 — Phase 2 CRD inventory contract

- Added the P2-03 CRD inventory contract.
- Scoped CRD collection to sanitized `CustomResourceDefinition` metadata
  references only.
- Reaffirmed that live CRD reads require separate Gate B expansion approval.

## 2026-07-23 — Workload snapshot fixture path

- Added a fake-client-only workload snapshot path for P2-03 fixtures without
  changing live `kua inventory --format=json` behavior.
- Added a sanitized golden cluster snapshot fixture covering supported workload
  controller kinds.
- Expanded subset validation for workload refs, replica counts, criticality
  values, and container required fields.

## 2026-07-23 — Workload snapshot integration contract

- Clarified that P2-03 may assemble workload snapshots for fake-client and
  golden-fixture tests only.
- Required live `kua inventory --format=json` to withhold workload records until
  Gate B is separately expanded for workload reads.
- Added P2-03 workload fixture and subset-validator expectations.

## 2026-07-23 — Phase 2 workload inventory contract

- Added the P2-03 workload inventory contract.
- Scoped workload collection to controller summaries and container image strings
  using fake clients first.
- Reaffirmed that P2-02 Gate B covers namespace/node collection only and live
  workload reads require separate Gate B expansion approval.

## 2026-07-23 — Gate B P2-02 smoke-test result

- Recorded a sanitized Gate B P2-02 smoke-test result.
- Marked Gate B as passed for namespace/node core inventory only.
- Kept raw live output local-only under ignored paths and excluded cluster,
  namespace, and node identifiers from the public record.
- Reaffirmed that later collectors require separate Gate B expansion approval.

## 2026-07-23 — Gate B validation record template

- Added a draft Gate B P2-02 validation record.
- Recorded the approval, pre-run, execution, scope verification, output review,
  post-run, and decision fields required before any live smoke-test result can
  be accepted.
- Reaffirmed that the record is not approval to run against a live cluster.

## 2026-07-23 — Gate B smoke-test plan

- Added a proposed Gate B smoke-test plan for P2-02 core inventory.
- Defined required user approval fields, allowed command shapes, approved
  read-only Kubernetes operations, stop conditions, artifact handling, and
  success criteria.
- Reaffirmed that the plan is not approval to access a live cluster.

## 2026-07-23 — Inventory command golden output

- Required a command-level golden test for `kua inventory --format=json`.
- Clarified that the CLI stdout contract must be protected separately from the
  collector package fixture.

## 2026-07-23 — Inventory JSON validation gate

- Required `kua inventory --format=json` to validate generated partial/core
  snapshots before writing them to stdout.
- Required validation failures to stop snapshot emission and report diagnostics
  on stderr.

## 2026-07-23 — Core inventory subset validator

- Added the P2-02 contract for a dependency-free generated snapshot validator.
- Scoped validation to the generated partial/core `ClusterSnapshot` fields
  needed to prevent schema drift in P2-02.
- Deferred full JSON Schema draft 2020-12 validation until a focused validator
  dependency or tooling assessment is approved.

## 2026-07-23 — Core inventory fixture contract

- Added deterministic P2-02 snapshot fixture requirements for fake-client
  namespace/node collection.
- Required fixture output to avoid real cluster identifiers and to include the
  explicit partial-inventory limitation.
- Established a golden fixture for `kua inventory --format=json` partial/core
  snapshot output.

## 2026-07-23 — Inventory JSON partial snapshot handoff

- Updated the P2-02 contract so `kua inventory --format=json` may emit a
  partial/core `ClusterSnapshot` after required preflight checks pass.
- Kept console rendering conservative and deferred full human inventory
  rendering to a later output package.
- Required namespace/node collection failures to stop snapshot emission instead
  of producing misleading partial data.

## 2026-07-23 — Phase 2 core inventory contract

- Added the P2-02 contract for partial/core snapshot collection.
- Scoped P2-02 to cluster metadata, server version, namespaces, nodes, and
  explicit limitations only.
- Required future inventory groups to remain empty with limitations rather than
  implying absence of resources.
- Reaffirmed that live collection remains blocked until Gate B approval names a
  context and read-only operation.

## 2026-07-23 — Inventory preflight JSON contract

- Added the P2-01 `kua inventory --format=json` contract for deterministic
  machine-readable preflight output.
- Required both console and JSON formats to explicitly identify P2-01 output as
  preflight-only.
- Reaffirmed that unknown, denied, or incomplete preflight evidence must not be
  rendered as `PASS`.

## 2026-07-23 — Project gate status

- Added `docs/status.md` to track current gate status, phase status, active
  branch focus, and latest quality evidence.
- Recorded Gate A as complete, Phase 1 as complete, and Gate B as not yet open
  because live cluster access still requires explicit approval.
- Clarified that current `kua inventory` behavior is preflight-only during
  P2-01 and does not represent full inventory collection.

## 2026-07-23 — Phase-close archive cleanup policy

- Added a phase/branch closure checkpoint for reviewing and removing obsolete
  recovery archives and temporary artifacts.
- Required clean working tree, passing gates, Git integrity validation, and no
  active recovery need before archive deletion.
- Kept archive deletion under the same explicit recoverable cleanup discipline.

## 2026-07-23 — Phase 2 Kubernetes preflight contract

- Added the P2-01 contract for kubeconfig/context resolution, read-only
  discovery/RBAC preflight, limitations, and fake-client test boundaries.
- Added the Kubernetes client-go dependency assessment for P2-01 only.
- Reaffirmed that live cluster execution requires separate approval naming the
  context and read-only operation.
- Narrowed the kubeconfig ignore policy so source files named `kubeconfig.go`
  can be tracked while local kubeconfig credential files remain ignored.

## 2026-07-23 — Phase 1 CI quality contract

- Added the P1-03 contract for GitHub Actions, local quality gates, formatting,
  tests, vet, build, and JSON syntax checks.
- Deferred release automation, dependency scanning, schema semantic validation,
  lint tooling, race tests, SBOM generation, and live-system checks.

## 2026-07-22 — Phase 1 config and error contract

- Added the P1-02 contract for standard-library common flag parsing, runtime
  config defaults, log-level validation, command-error categories, and exit-code
  mapping.
- Confirmed P1-02 does not add Cobra, Viper, Kubernetes clients, Azure access,
  file config loading, report rendering, or live-system access.

## 2026-07-22 — Phase 1 foundation contract

- Added the foundation implementation contract for the Go module, `kua` binary,
  standard-library CLI skeleton, build metadata placeholders, and exit-code
  constants.
- Deferred Cobra, Viper, `client-go`, schema tooling, linting, CI, and release
  dependencies to later focused dependency-assessment changes.

## 2026-07-22 — Phase 0 schema contracts

- Added Phase 0 schema contract artifacts for assessment, cluster snapshot, AKS
  provider evidence, and catalog records.
- Added the Phase 0 security, RBAC, external-command, redaction, and dependency
  contract.
- Added valid and invalid schema fixtures, including the approved
  `1.30.0 -> 1.33.12` staged AKS recommendation scenario.
- Documented the split between structural JSON Schema validation and later
  semantic domain validation for cross-record recommendation rules.

## 2026-07-22 — MVP development planning baseline

- Expanded delivery into contract, foundation, inventory, health, component, API, AKS, recommendation, reporting, and release phases.
- Approved Apache-2.0, module/binary/platform/version-range defaults, exit behavior, health window, critical-workload policy, and redacted reports.
- Pinned the initial kubent contract to `0.7.3` JSON with Helm collection disabled and mandatory target-rule coverage validation.
- Added detailed implementation, domain/schema, security/RBAC, and validation/release plans.
- Resolved all currently recorded blocking product questions.

## 2026-07-22 — Recoverable cleanup policy

- Required an inventoried and verified recovery point before every destructive operation.
- Required post-cleanup integrity validation and recorded recovery instructions.
- Required separate user approval before deleting recovery artifacts, even after successful validation.

## 2026-07-22 — Git hygiene and publication rules

- Standardized the `main` branch and short-lived branch prefixes.
- Added pre-commit and pre-publication integrity checks.
- Required explicit approval for pushes, upstream changes, tags, releases, and history rewrites.
- Documented author identity, sensitive-file, and macOS/ExFAT AppleDouble safeguards.
- Established the root `.gitignore` as the canonical ignore policy.

## 2026-07-22 — Local-first provider and catalog lifecycle

- Superseded offline-by-default behavior with default `auto` AKS evidence through the local authenticated Azure CLI.
- Retained explicit `azure`, `file`, `offline`, and `none` modes and JSON evidence fallback.
- Clarified that kubeconfig supplies Kubernetes access but not Azure upgrade offerings.
- Defined repository YAML plus `go:embed` as the bundled catalog model.
- Established curated/manual review, automation-assisted proposals, optional future signed updates, and no runtime web searching or scraping.
- Reaffirmed that unknown or insufficient component compatibility evidence cannot produce `PASS`.

## 2026-07-21 — Initial architecture baseline

- Established docs-first, user-approved governance.
- Confirmed AKS as the first provider while retaining provider-neutral interfaces.
- Limited MVP analysis to live clusters.
- Selected an installed kubent binary adapter for MVP and deferred native analysis.
- Defined a bundled offline compatibility catalog proposal.
- Separated recommended destination from sequential provider-valid upgrade stages.
- Added product, architecture, recommendation, CLI/report, security, testing, process, roadmap, and open-question documents.
