# Documentation change log

This log records material scope and architecture changes. Git remains the detailed history.

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
