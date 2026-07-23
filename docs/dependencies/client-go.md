# Dependency assessment: Kubernetes client modules

Status: Approved for Phase 2 implementation
Last updated: 2026-07-23

## 1. Purpose

KUA needs Kubernetes client libraries for kubeconfig loading, REST client
construction, discovery, fake-client tests, and read-only authorization checks.

## 2. Packages

Planned direct modules:

- `k8s.io/client-go`
- `k8s.io/apimachinery`
- `k8s.io/api`
- `k8s.io/apiextensions-apiserver`

The initial version line should match the approved Kubernetes validation range
and use a currently maintained minor compatible with Kubernetes `1.30` through
`1.33`. The implementation starts with the latest available `v0.33.x` line
unless `go get` resolution or vulnerability review identifies a blocker.

`k8s.io/apiextensions-apiserver` is added during P2-03 for typed
CustomResourceDefinition clients and fake-client tests. It must stay on the same
minor line as the other Kubernetes modules.

## 3. License

The Kubernetes Go modules are Apache-2.0 licensed, compatible with KUA's planned
Apache-2.0 distribution.

## 4. Alternatives considered

- Shelling out to `kubectl`: rejected because it weakens structured error
  handling, testability, and argument safety.
- Hand-written kubeconfig parsing and raw HTTP: rejected because it would
  duplicate mature client-go behavior and increase authentication risk.
- Delaying Kubernetes libraries: rejected because Phase 2 requires real
  kubeconfig and discovery semantics.

## 5. Risk and controls

- Pin versions in `go.mod`/`go.sum`.
- Use fake clients and synthetic fixtures in tests.
- Do not commit kubeconfig files or cluster output.
- Do not use write, patch, delete, exec, log, or Secret-reading APIs.
- Run `go test`, `go vet`, and local CI after adding modules.

## 6. Approval boundary

This approval covers adding Kubernetes client modules for Phase 2 preflight and
inventory collectors only. Provider SDKs, kubent integration modules, schema
validators, lint tools, and release tooling still require separate assessment
and approval.
