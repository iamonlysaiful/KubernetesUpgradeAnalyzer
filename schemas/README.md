# KUA schema contracts

Status: Phase 0 contract artifact
Last updated: 2026-07-22

The files in this directory define versioned JSON contracts for KUA's canonical
data exchange surfaces. They are source-of-truth artifacts and must remain ahead
of implementation.

## Versioned schemas

| Schema | Version value | Purpose |
| --- | --- | --- |
| `assessment/v1.json` | `kua.assessment.v1` | Canonical renderer and automation output. |
| `cluster-snapshot/v1.json` | `kua.cluster-snapshot.v1` | Normalized collector input for analyzers. |
| `provider-evidence/aks-v1.json` | `kua.provider-evidence.aks.v1` | Sanitized AKS upgrade availability evidence. |
| `catalog/v1.json` | `kua.catalog.v1` | Embedded compatibility and provider policy catalog. |

## Contract rules

- `schemaVersion` is required and rejects unknown major versions.
- Timestamps use RFC 3339 date-time values.
- Kubernetes versions are initially limited to the approved `1.30` through
  `1.33` validation range.
- Unknown or ambiguous evidence must be represented explicitly with `UNKNOWN`,
  `INCONCLUSIVE`, `AMBIGUOUS`, or `UNPARSEABLE`; it must not be omitted to imply
  success.
- Renderers consume only `kua.assessment.v1`; they do not recalculate readiness,
  risk, destination, or upgrade stages.
- Catalog and provider records include provenance so claims can be audited.
- Local report identifiers may be unredacted, but shareable fixtures use stable
  aliases.

## Fixture classes

Fixtures under `schemas/fixtures/*/valid` should pass structural schema
validation. Fixtures under `schemas/fixtures/*/invalid` are split into two
groups:

- structural invalid examples, such as unsupported schema versions or disallowed
  command names;
- semantic invalid examples, such as an unknown component version represented as
  a compatibility pass. These may be structurally valid JSON and must fail later
  domain validation tests.

Phase 1 adds automated schema validation. Later phases add semantic validators
for cross-record rules that JSON Schema cannot express cleanly.
