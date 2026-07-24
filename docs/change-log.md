# Documentation change log

This log records material scope and architecture changes. Git remains the detailed history.

## 2026-07-23 — Phase 3 health analysis plan

- Added the Phase 3 health analysis contract.
- Added the Phase 3 health implementation plan with a compressed PR strategy to
  keep MVP delivery near 30 total PRs.
- Confirmed Phase 3 starts from fake-client inventory fixtures while expanded
  live inventory remains deferred.

## 2026-07-23 — Health rule foundation

- Added the internal health finding model, rule interface, runner, default
  options, and deterministic clock injection.
- Added stable finding ordering by severity, rule ID, resource namespace, kind,
  name, and summary.
- Kept Phase 3 health rules, CLI commands, and live Kubernetes reads out of this
  foundation slice.

## 2026-07-23 — Node and workload health rules

- Added node readiness, node pressure, node kubelet minor skew, and workload
  unavailable health rules over normalized inventory snapshots.
- Added `UNKNOWN` findings for missing node readiness, kubelet, and server
  version evidence instead of treating absent evidence as pass.
- Kept the rules internal and fixture-driven; no CLI command or expanded live
  Kubernetes read was introduced.

## 2026-07-23 — Storage and event health rules

- Added storage evidence unknown, warning event, and unknown event type health
  rules over normalized inventory snapshots.
- Added the Phase 3 closeout record confirming health analysis remains internal
  and fixture-driven, with expanded live inventory still deferred.
- Kept user-facing health/analyze commands, compatibility checks, kubent,
  provider evidence, and final recommendations out of Phase 3.

## 2026-07-23 — Phase 4 catalog and detection plan

- Added the Phase 4 embedded catalog loader contract.
- Added the Phase 4 component detection contract.
- Added the Phase 4 implementation plan with a compressed PR strategy and
  explicit no-runtime-internet catalog boundary.
- Reaffirmed that unknown component versions and absent compatibility evidence
  must produce `UNKNOWN`, never `PASS`.

## 2026-07-23 — Catalog loader foundation

- Added the internal catalog bundle model, embedded placeholder catalog, explicit
  file loader, checksum capture, and validation errors.
- Added validation for schema version, semantic catalog version, timestamps,
  source references, duplicate component IDs/aliases, and known enum values.
- Kept runtime internet search, catalog downloads, compatibility decisions, and
  recommendations out of scope.

## 2026-07-24 — Component detection foundation

- Added the internal component detection result model, detector interface,
  runner, deterministic sorting, and sanitized resource references.
- Added version normalization helpers that treat missing, `latest`, `unknown`,
  and digest-only image evidence as `UNKNOWN`.
- Kept detector cohort logic, compatibility decisions, live reads, and runtime
  network access out of scope.

## 2026-07-24 — Initial component detector cohort

- Added workload-backed detectors for NGINX Ingress, CoreDNS, Metrics Server,
  Azure Disk CSI, Azure File CSI, Fluent Bit, and EMQX.
- Added tests proving ambiguous, `latest`, and conflicting component version
  evidence remains `UNKNOWN`.
- Added the Phase 4 closeout record and kept compatibility decisions, provider
  evidence, kubent, recommendations, runtime internet access, and expanded live
  reads deferred.

## 2026-07-24 — Phase 5 API compatibility plan

- Added the Phase 5 API compatibility and kubent adapter contract.
- Added the Phase 5 implementation plan with controlled kubent execution,
  target-rule coverage validation, and a required go/no-go decision.
- Reaffirmed that missing, malformed, empty, or unverified kubent evidence must
  produce `UNKNOWN` or `INCONCLUSIVE`, never `PASS`.
- Reaffirmed that live kubent execution requires separate explicit approval.

## 2026-07-24 — Kubent adapter foundation

- Added the internal kubent adapter foundation with process-runner abstraction,
  controlled argument construction, version validation, output bounds, and JSON
  parsing.
- Added process-fake tests for supported and unsupported versions, malformed
  output, nonzero exit, oversized output, and shell-free argument construction.
- Kept target-rule coverage validation, normalized API findings, live kubent
  execution, and go/no-go decisions out of this foundation slice.

## 2026-07-24 — Kubent coverage decision

- Added target-rule coverage validation for Kubernetes `1.30` through `1.33`
  fixture policy.
- Added normalized API compatibility findings that only return `PASS` when
  target coverage is verified and kubent reports no deprecated APIs.
- Added the Phase 5 closeout and kubent go/no-go record while keeping live
  kubent execution deferred.

## 2026-07-23 — Phase 2 closeout draft

- Added the Phase 2 closeout record.
- Marked the Phase 2 MVP inventory foundation as ready to close with live core
  inventory only.
- Deferred expanded live workload, CRD, networking, storage, and event reads
  until a later explicit Gate B expansion.
- Marked Phase 3 health analysis as the recommended next implementation phase
  after closeout is reviewed and merged.

## 2026-07-23 — Inventory snapshot assembly consolidation contract

- Added the Phase 2 inventory snapshot assembly consolidation contract.
- Scoped consolidation to replacing long fake-client snapshot method names with
  explicit collection options.
- Reaffirmed that live `kua inventory --format=json` remains on core inventory
  until Gate B is separately expanded.
- Replaced long fake-client snapshot assembly methods with a single
  option-driven `CollectSnapshot` path while preserving golden fixtures.

## 2026-07-23 — Phase 2 events inventory contract

- Added the P2-03 events inventory contract.
- Scoped event collection to sanitized metadata and reason/type/timestamp fields
  only.
- Required raw event messages to stay out of snapshots and fixtures.
- Reaffirmed that live event reads require separate Gate B expansion approval.
- Added fake-client event collection, full P2-03 fake snapshot fixture coverage,
  and subset validation for event refs, type, reason, and timestamps.

## 2026-07-23 — Phase 2 storage inventory contract

- Added the P2-03 storage inventory contract.
- Scoped storage collection to sanitized PVC, PV, and StorageClass metadata
  references only.
- Reaffirmed that live storage reads require separate Gate B expansion approval.
- Added fake-client storage collection, snapshot fixture coverage, and subset
  validation for PVC, PV, and StorageClass refs.

## 2026-07-23 — Phase 2 networking inventory contract

- Added the P2-03 networking inventory contract.
- Scoped networking collection to sanitized Service and Ingress metadata
  references only.
- Reaffirmed that live networking reads require separate Gate B expansion
  approval.
- Added fake-client networking collection, snapshot fixture coverage, and subset
  validation for Service and Ingress refs.

## 2026-07-23 — Phase 2 CRD inventory contract

- Added the P2-03 CRD inventory contract.
- Scoped CRD collection to sanitized `CustomResourceDefinition` metadata
  references only.
- Reaffirmed that live CRD reads require separate Gate B expansion approval.
- Expanded the Kubernetes dependency assessment to cover the
  `k8s.io/apiextensions-apiserver` module for CRD fake-client tests.
- Added fake-client CRD collection, snapshot fixture coverage, and subset
  validation for CRD refs.

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
