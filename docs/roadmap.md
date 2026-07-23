# Roadmap

Status: Approved sequencing; implementation requires approval per phase
Last updated: 2026-07-22

## Phase 0 — Design freeze and contracts

- Approve product defaults and resolve blocking questions.
- Define domain types and schemas for snapshots, findings, detections, assessments, catalogs, and AKS provider evidence.
- Define least-privilege RBAC, Azure CLI allowlist, redaction contract, dependency policy, and sanitized fixtures.
- Convert phases into issue-sized work packages with acceptance tests and commit boundaries.

Exit: all contracts have versioning rules and representative valid/invalid fixtures; no blocking open question remains.

## Phase 1 — CLI foundation

- Initialize module `github.com/iamonlysaiful/KubernetesUpgradeAnalyzer` and binary `kua`.
- Add Cobra wiring, configuration, `slog`, build/version metadata, exit-code mapping, and dependency injection.
- Add formatting, unit-test, vet, lint, schema-validation, and GitHub Actions gates.
- Add Apache-2.0 license after implementation approval.

Exit: skeleton builds and tests on Linux/macOS amd64/arm64; unimplemented commands fail clearly.

## Phase 2 — Kubernetes preflight and inventory

- Resolve kubeconfig/current or explicit context and show selected context.
- Check discovery, version, permissions, and required/optional resources.
- Implement read-only collectors and normalized sanitized snapshot.
- Render deterministic console/JSON inventory with partial-RBAC limitations.

Exit: inventory acceptance fixtures and approved live read-only core smoke test
pass. Expanded live inventory may be deferred when fake-client collector
coverage is complete and downstream MVP features do not yet consume expanded
live records.

## Phase 3 — Health analysis

- Implement current-state node, pod, workload, DaemonSet, PVC, and event rules.
- Default events to a configurable 30-minute warning window.
- Support explicitly configured/labeled critical workloads.
- Initial implementation uses the Phase 2 fake-client inventory foundation;
  expanded live inventory remains deferred until a separate Gate B expansion.

Exit: every blocker, warning, unknown transition, and time-window boundary has direct tests.

## Phase 4 — Component detection and catalog

- Implement embedded catalog/schema/checksum and validated local overrides.
- Implement detector/version/confidence contracts.
- Start with NGINX Ingress, CoreDNS, Metrics Server, Azure Disk/File CSI, Fluent Bit, and EMQX; expand only after framework validation.

Exit: known, incompatible, conditional, ambiguous, stale, and missing evidence cases pass regression tests; unknown never passes.

## Phase 5 — API compatibility

- Implement controlled kubent `0.7.3` adapter using JSON and `--helm3=false`.
- Verify target-rule coverage for every assessed stage.
- Treat missing tool, malformed output, execution failure, or missing rules as inconclusive.

Exit: adapter contract and negative-path fixtures pass; a documented go/no-go decision confirms whether kubent covers `1.30`–`1.33` or a minimal native analyzer must enter MVP.

## Phase 6 — AKS provider evidence

- Implement `auto`, `azure`, `file`, `offline`, and `none` modes.
- Detect AKS identity with confidence and accept explicit overrides.
- Parse allowlisted `az aks get-upgrades` and exported JSON evidence.
- Build candidate versions and sequential provider-valid edges.

Exit: all source/fallback/authentication/offline cases pass without provider mutation.

## Phase 7 — Recommendation engine

- Evaluate APIs, components, provider availability, health, and evidence sufficiency per candidate and stage.
- Produce deterministic readiness, risk, decision trace, remediation, destination, and sequential path.
- Validate sanitized `1.30.0 → 1.33.12` destination through intermediate minors.

Exit: every MVP recommendation acceptance criterion passes.

## Phase 8 — Reports and hardening

- Complete JSON, console, Markdown, and self-contained HTML renderers.
- Add stable-alias redacted mode.
- Add golden, hostile-input, concurrency, performance, race, dependency, license, and security checks.

Exit: release-candidate quality gates and privacy review pass.

## Phase 9 — Controlled staging validation and MVP release

- Run separately approved read-only steps against the named AKS staging context.
- Compare results with Azure CLI evidence and manual platform-engineering review.
- Sanitize only approved fixture material and record product gaps.
- Produce versioned artifacts, checksums, SBOM, provenance, release notes, and rollback instructions.

Exit: owner approves validation evidence and publication of the MVP release.

## Later phases

- Native API analyzer, initially compared with kubent.
- Helm/local manifest, CRD, and operator compatibility depth.
- Upgrade simulation; EKS, GKE, OpenShift, and vanilla provider adapters.
- AKS best-practice checks, dashboard, and GitHub Action integration.
- Historical comparison, drift, monitoring, security posture, cost impact, offline-safe AI explanations, and VS Code extension.
