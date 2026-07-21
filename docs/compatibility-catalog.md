# Offline compatibility catalog

Status: Proposed for implementation  
Last updated: 2026-07-21

## 1. Why the catalog exists

An offline analyzer cannot ask vendor websites which versions are compatible at runtime. It therefore needs bundled knowledge describing API removals, component support claims, Kubernetes release metadata, and provider policies. This is the context behind the earlier compatibility-data question.

## 2. Distribution model

- Ship a signed or checksum-verified catalog with every KUA release.
- Give the catalog its own semantic version and schema version.
- Print KUA version, catalog version, creation time, and age in every report.
- Do not automatically download updates in MVP.
- A future explicit `kua catalog update` may be designed separately with provenance and network controls.

## 3. Proposed layout

```text
catalog/
  manifest.yaml
  kubernetes/releases.yaml
  kubernetes/apis.yaml
  providers/aks.yaml
  components/<product>.yaml
  sources.yaml
schemas/catalog-v1.json
```

## 4. Component entry

Each product record contains:

- stable product ID and aliases;
- product version range;
- supported Kubernetes range or explicitly documented compatibility statements;
- status (`supported`, `unsupported`, `conditional`, `unknown`);
- conditions and notes;
- authoritative source URL, title, retrieval date, and applicable version/date;
- confidence (`authoritative`, `inferred`, `community`, `unknown`);
- expiry/review date;
- optional detector hints, kept separate from compatibility policy.

“Modern Kubernetes versions” is not a machine-enforceable support range. Such vendor wording must yield `UNKNOWN` or `conditional` unless bounded evidence exists.

## 5. Provenance rules

1. Prefer official vendor release notes, support matrices, and documentation.
2. Record an exact claim and source; do not rely on undocumented memory.
3. Inferences must be labeled and cannot produce an unconditional `PASS` by themselves.
4. Conflicting sources select the most specific, current, authoritative statement and retain the conflict as a limitation.
5. Expired evidence lowers confidence; it does not silently disappear.
6. Every catalog change requires review, schema validation, fixtures, and its own docs-first approved change.

## 6. Candidate resolution

Candidate patch availability is provider- and sometimes region-specific. The bundled catalog may establish policy and known release facts but cannot guarantee that a patch is currently offered to a particular AKS cluster. A report may call `1.33.12` the recommended destination only when candidate evidence supports it; otherwise it may recommend minor `1.33` with exact patch `UNKNOWN` or qualify the result.

Provider evidence must include capture time, cluster/provider identifiers in sanitized form, available upgrades, and source method. Stale thresholds are configurable policy and visible in the report.

## 7. Integrity and schema

Startup validates schema compatibility, unique IDs, valid semantic ranges, source presence, checksums, and cross-references. A corrupt or unsupported catalog is fatal for compatibility/recommendation commands.

