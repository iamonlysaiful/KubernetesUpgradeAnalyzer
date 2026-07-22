# Domain and schema plan

Status: Phase 0 contracts in progress
Last updated: 2026-07-22

## 1. Goal

Define stable, versioned data contracts before collectors, analyzers, engines, or renderers are implemented. Domain packages remain independent of Cobra, Kubernetes client types, Azure CLI output types, kubent types, and rendering libraries.

## 2. Contract order

1. Scalar types: Kubernetes version/range, timestamp, duration, confidence, severity, status, readiness, risk, provider source, and sanitized resource reference.
2. `ClusterSnapshot`: immutable normalized collection input plus limitations.
3. `ComponentDetection`: product/version/method/evidence/confidence without compatibility conclusions.
4. `Finding`: stable rule/finding IDs, category, status, severity, resources, candidate stages, evidence, remediation, and limitations.
5. `ProviderEvidence`: identity, source, capture time, offered versions, node-pool/control-plane facts, and provenance.
6. `UpgradeCandidate`, `UpgradeStage`, and `UpgradePath`.
7. `Assessment`: metadata, inventory, detections, findings, decision trace, destination, path, readiness, risk, and actions.
8. Catalog manifest, Kubernetes API rules, component compatibility, AKS policies, and sources.

## 3. Schema artifacts

Phase 0 proposes:

```text
schemas/
  assessment/v1.json
  cluster-snapshot/v1.json
  provider-evidence/aks-v1.json
  catalog/v1.json
  fixtures/<schema>/valid/
  fixtures/<schema>/invalid/
```

Accepted Phase 0 schema artifacts are now tracked under `schemas/`:

- `schemas/assessment/v1.json`
- `schemas/cluster-snapshot/v1.json`
- `schemas/provider-evidence/aks-v1.json`
- `schemas/catalog/v1.json`
- representative valid and invalid fixtures under `schemas/fixtures/`

Go domain structs and JSON Schemas must be tested for agreement. Schemas use explicit `schemaVersion` values and reject unknown major versions.

## 4. Compatibility policy

- Additive optional fields may remain within schema major v1.
- New required fields, changed meaning/type, or removed fields require a new major version.
- Readers reject unsupported majors and tolerate documented optional additions.
- Renderers consume canonical `Assessment`, never raw collector/provider output.
- Stable rule IDs are not reused for different semantics.
- Ordering rules are specified for deterministic output.

## 5. Required examples

- Minimal and complete healthy assessment.
- Partial RBAC and incomplete provider evidence.
- Removed API at an intermediate stage.
- Compatible, incompatible, conditional, unknown, stale, and conflicting component evidence.
- Redacted and unredacted equivalent decisions.
- Malformed versions, timestamps, unknown schema majors, duplicate finding IDs, invalid paths, and corrupt catalog sources.

## 6. Decisions to record during Phase 0

Each contract documentation change specifies required/optional fields, null versus omission, enum extension behavior, stable sorting keys, size limits, timestamp format, identifier/redaction rules, validation errors, and migration expectations.

## 7. Exit criteria

- Every schema has valid/invalid fixtures and automated validation design.
- Domain dependency direction is documented and reviewable.
- The `1.30 → 1.33.12` scenario can be represented without implementation-specific fields.
- Unknown evidence is structurally distinct from pass/fail and cannot be omitted silently.

## 8. Phase 0 contract decisions

- Schema identifiers use explicit string constants such as `kua.assessment.v1`.
- Initial Kubernetes version patterns accept `1.30` through `1.33` only.
- Status enums include `UNKNOWN` and `INCONCLUSIVE` wherever evidence can be
  absent, stale, ambiguous, or tool-limited.
- Component detection records separate detection confidence from compatibility
  findings. A known product with an unknown version is a valid detection, but it
  cannot create a `PASS` compatibility finding.
- AKS provider evidence permits only the read-only `az aks get-upgrades` command
  shape and stores sanitized aliases, not raw subscription/resource identifiers.
- The assessment contract records sequential upgrade stages so destination
  `1.33.12` from current `1.30.0` can be represented without claiming a single
  direct provider operation.
- JSON Schema covers structure. Later domain validators cover cross-record rules
  such as unique IDs, catalog source cross-references, duplicate finding IDs, and
  unknown-evidence decision constraints.
