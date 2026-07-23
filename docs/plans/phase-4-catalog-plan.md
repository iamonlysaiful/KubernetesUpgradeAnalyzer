# Phase 4 catalog and component detection plan

Status: Draft Phase 4 plan
Last updated: 2026-07-23

This plan starts Phase 4 after Phase 3 closeout. It is based on verified facts:
Phase 3 internal health rules are merged, live inventory remains core-only, and
expanded inventory groups are available through fake-client/golden fixtures.

## 1. Goal

Build the catalog and component detection foundation needed before API
compatibility, provider evidence, and recommendation phases.

Phase 4 should identify component evidence and catalog limitations without
claiming upgrade readiness.

## 2. Delivery slices

To keep the MVP near the current 30 ±2 PR target, Phase 4 is planned as four PRs:

| PR | Branch | Scope | Evidence |
| --- | --- | --- | --- |
| P4-01 | `docs/phase-4-catalog-plan` | Catalog loader and detection contracts, PR strategy | Docs review |
| P4-02 | `feature/catalog-loader-foundation` | Embedded catalog representation, loader, validation errors, checksums | Unit tests |
| P4-03 | `feature/component-detection-foundation` | Detection result model, detector runner, version normalization | Unit tests |
| P4-04 | `feature/component-detector-cohort` | Initial detector cohort and Phase 4 closeout | Fixture/detector tests and closeout docs |

If P4-04 becomes too dense, split the detector cohort rather than reducing
fixtures or weakening `UNKNOWN` behavior.

## 3. Runtime data boundary

Phase 4 does not add runtime internet search, catalog scraping, provider calls,
or expanded live Kubernetes reads.

The embedded catalog is the default source. A user-supplied local catalog
override may be designed in the loader foundation, but it must be explicit,
validated, and deterministic.

## 4. Initial package shape

Planned packages:

```text
internal/catalog
internal/components
```

Initial concepts:

- catalog `Manifest`, `Source`, `Component`, `CompatibilityRecord`;
- catalog `Loader`, `Bundle`, and validation errors;
- component `Detection`, `Detector`, `Runner`, `Confidence`, and `Status`;
- deterministic version normalization helpers.

## 5. Stop conditions

Stop affected work if:

- a catalog record would require an unsupported compatibility claim;
- runtime code would search the internet or auto-download data;
- unknown component versions could be interpreted as pass;
- a detector requires unapproved live expanded inventory;
- a new dependency is needed before its assessment is approved;
- source provenance cannot be represented in the catalog.

## 6. Exit criteria

Phase 4 is complete when:

- embedded catalog loading and validation are tested;
- local override behavior is either implemented or explicitly deferred;
- detector framework and deterministic ordering are tested;
- initial detector cohort is tested against synthetic/fake-client fixtures;
- unknown/stale/missing evidence behavior is tested;
- Phase 4 closeout records remaining gaps and confirms Phase 5 or Phase 6 can
  start.
