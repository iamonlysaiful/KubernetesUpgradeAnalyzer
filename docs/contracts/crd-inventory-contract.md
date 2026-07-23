# CRD inventory collector contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract governs the CRD portion of P2-03. It depends on P2-02 core
inventory and the P2-03 workload fixture path. It does not approve live CRD
collection by itself.

## 1. Scope

This package adds fake-client-first CRD snapshot collection from Kubernetes
`apiextensions.k8s.io/v1` CustomResourceDefinition metadata.

The collected CRD records populate `inventory.crds` in
`schemas/cluster-snapshot/v1.json`.

This package may add an internal snapshot assembly path that includes CRDs for
fake-client and golden-fixture tests. `kua inventory --format=json` must not emit
live CRD records until Gate B is separately expanded for CRD reads.

This package does not list custom resources created from CRDs. It collects only
CRD definitions, not resource instances, object specs, Secret data, ConfigMap
contents, annotations, raw UIDs, storage, networking, events, health findings,
compatibility findings, provider evidence, recommendations, or reports.

## 2. Required fields

Each CRD record is represented as a `ResourceRef` and includes:

- `apiVersion`: `apiextensions.k8s.io/v1`;
- `kind`: `CustomResourceDefinition`;
- `name`: CRD name.

CRDs are cluster-scoped definitions, so `namespace` must remain empty.

The MVP snapshot schema currently models CRDs as `ResourceRef` values only.
Version lists, served/storage flags, group, scope, categories, and conversion
strategy are intentionally deferred until a schema expansion is approved.

## 3. Determinism and safety

CRDs are sorted by name.

Fixtures must use sanitized CRD names only. Real organization/product names from
live clusters must not be committed unless separately reviewed and approved as
safe public examples.

## 4. Limitations

When CRD collection is intentionally absent or not yet approved for live use, KUA
must not imply that the cluster has no CRDs. P2-03 fixture paths that include
CRDs must use a limitation that names the inventory groups still intentionally
uncollected.

Collection failure for CRDs makes the affected command fail safely instead of
emitting misleading partial CRD data.

## 5. Gate B expansion

P2-02 Gate B passed only for namespace/node collection. Live workload and CRD
collection require separate Gate B expansion approval naming the context and
allowing the specific read-only API operations. Until then, automated tests use
fake clients only.

## 6. Fixture and validation expectations

At least one P2-03 golden fixture must include representative sanitized CRD
records. The fixture must continue to use empty arrays for storage, networking,
and events until those collectors are implemented and approved.

The dependency-free snapshot subset validator must be expanded to cover CRD
required fields before any CRD snapshot path is considered complete.
