# Testing strategy

Status: Proposed for implementation  
Last updated: 2026-07-21

## 1. Test layers

1. Unit tests for normalization, detectors, catalog queries, findings, path selection, risk, redaction, and rendering helpers.
2. Contract tests for detector/analyzer/provider interfaces and canonical JSON schema.
3. Golden tests for console, JSON, Markdown, and HTML reports with deterministic fixtures.
4. Integration tests using fake Kubernetes clients/discovery and controlled kubent process fixtures.
5. End-to-end tests against an ephemeral local cluster only after separately approved dependencies and execution.
6. Sanitized AKS fixture validation; no routine CI access to a real AKS cluster.

## 2. Required fixture families

- Healthy AKS-like 1.30 snapshot targeting 1.33.12 via sequential stages.
- Removed API at first, intermediate, and destination stages.
- Deprecated but not yet removed API.
- Known compatible, incompatible, conditional, unknown, and ambiguous component versions.
- Helm, operator, raw-manifest, and managed component evidence.
- Partial RBAC and API discovery failures.
- Unready nodes, pressure, CrashLoopBackOff, ImagePullBackOff, Pending, unbound PVC, unavailable workloads, and noisy events.
- Mixed control-plane/node-pool versions.
- Stale/missing/conflicting provider evidence.
- Malformed/hostile catalog, kubent output, image labels, events, and HTML content.

## 3. Determinism

Inject clock and ID generators. Sort maps/resources/findings before rendering. Golden comparisons normalize only explicitly volatile metadata. Same snapshot, configuration, binaries, and catalog must yield the same recommendation.

## 4. Quality gates

Before each implementation commit is complete:

- `gofmt` is clean;
- unit and relevant integration tests pass;
- `go vet` passes;
- approved `golangci-lint` configuration passes;
- schemas and catalog validate;
- security/redaction tests pass;
- race tests run for concurrent collector/analysis code where feasible;
- docs and CLI examples match behavior.

## 5. Coverage philosophy

Prioritize decision branches and safety boundaries over a headline percentage. Every blocker rule, unknown-evidence transition, path-selection rule, redaction boundary, and exit code requires direct tests.

