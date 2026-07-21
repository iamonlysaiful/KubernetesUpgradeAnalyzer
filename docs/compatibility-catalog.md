# Embedded compatibility catalog

Status: Proposed for implementation  
Last updated: 2026-07-22

## 1. Why the catalog exists

The live cluster can reveal installed products and versions, but it cannot reliably state which Kubernetes versions those products support. KUA therefore needs bundled knowledge describing API removals, component support claims, Kubernetes release metadata, and provider policies. Runtime internet searching would be non-deterministic and unsafe for compatibility decisions.

## 2. Distribution model

- Store human-reviewable YAML source under `catalog/` and embed the validated catalog into the Go binary with `go:embed` for every KUA release.
- Ship catalog checksums and, for separately distributed catalog bundles, signatures.
- Give the catalog its own semantic version and schema version.
- Print KUA version, catalog version, creation time, and age in every report.
- Permit an explicit local catalog override for development or organization policy, subject to the same schema/integrity validation.
- Do not search the internet, scrape vendor pages, or automatically download updates during assessment.
- A future explicit `kua catalog check|update|status` workflow may download only signed catalog releases from an approved KUA distribution location.

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

The embedded catalog is the default. A local override never merges silently: the report identifies the selected catalog source, version, checksum, and creation time.

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

## 6. Maintenance workflow

Catalog content is curated and version controlled. Maintainers periodically review official component/vendor documentation and submit focused catalog changes with source evidence and tests. Scheduled automation may detect new releases or changed source pages and open a proposed change, but it cannot promote a compatibility claim or publish a catalog without human review.

The preferred lifecycle is:

1. automation or a maintainer identifies a possible upstream change;
2. official sources are reviewed and captured with retrieval/review dates;
3. catalog records and fixtures are updated;
4. schema, overlap, provenance, and recommendation regression tests pass;
5. a reviewer approves the compatibility interpretation;
6. a new catalog version is published and embedded in the next KUA release, optionally also as a signed standalone bundle.

Unknown installed component versions, absent records, unbounded vendor claims, stale/conflicting evidence, and unsupported catalog schemas produce `UNKNOWN` or a qualified warning—never `PASS`.

## 7. Candidate resolution

Candidate patch availability is provider- and sometimes region-specific. The bundled catalog may establish policy and known release facts but cannot guarantee that a patch is currently offered to a particular AKS cluster. A report may call `1.33.12` the recommended destination only when candidate evidence supports it; otherwise it may recommend minor `1.33` with exact patch `UNKNOWN` or qualify the result.

Provider evidence must include capture time, cluster/provider identifiers in sanitized form, available upgrades, and source method. Stale thresholds are configurable policy and visible in the report.

## 8. Integrity and schema

Startup validates schema compatibility, unique IDs, valid semantic ranges, source presence, checksums, and cross-references. A corrupt or unsupported catalog is fatal for compatibility/recommendation commands.
