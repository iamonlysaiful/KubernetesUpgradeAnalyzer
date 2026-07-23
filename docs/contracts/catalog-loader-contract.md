# Catalog loader contract

Status: Phase 4 contract artifact
Last updated: 2026-07-23

This contract governs the MVP embedded compatibility catalog loader and
validation foundation. It does not approve final recommendation decisions or
runtime internet access.

## 1. Scope

Phase 4 catalog loader work covers:

- embedding a curated catalog bundle in the KUA binary;
- loading an optional explicit local catalog override;
- validating catalog schema version, manifest metadata, source references, and
  checksums;
- exposing catalog metadata to later compatibility and report phases;
- returning explicit errors for corrupt, unsupported, or incomplete catalogs.

Phase 4 loader work does not evaluate live cluster compatibility, call provider
APIs, invoke kubent, download catalog updates, scrape vendor pages, or produce
upgrade recommendations.

## 2. Distribution rules

The default catalog is bundled with the binary. Runtime assessment must not
search the internet or automatically download compatibility data.

A local override is allowed only when the user explicitly provides a path. The
override must pass the same validation as the embedded catalog and must be
reported as the selected source by later output phases.

Future catalog update workflows may be designed separately, but they must use
approved signed KUA distribution locations and human-reviewed catalog releases.

## 3. Minimum metadata

Every catalog bundle must include:

- catalog version;
- schema version;
- generated or reviewed timestamp;
- source list;
- checksum metadata for bundled content;
- component records, even when the first MVP records are intentionally minimal.

Catalog version and schema version are separate from the KUA binary version.

## 4. Validation behavior

The loader must reject:

- unsupported schema versions;
- missing required manifest fields;
- malformed semantic versions or Kubernetes version ranges;
- missing source references for compatibility records;
- duplicate component IDs or aliases where ambiguity would affect detection;
- checksum mismatches for files covered by the manifest.

Unknown, stale, conflicting, or absent compatibility evidence is not a loader
failure by itself. Later compatibility evaluation must turn those cases into
`UNKNOWN`, not `PASS`.

## 5. Initial implementation boundary

The initial implementation may use a compact Go representation and curated test
fixtures before a larger YAML catalog tree is introduced. The public contract is
the loader behavior, validation guarantees, and no-runtime-internet boundary.

Any dependency added for YAML, semantic versioning, or signatures requires a
separate dependency assessment before implementation.
