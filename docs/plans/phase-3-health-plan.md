# Phase 3 health analysis plan

Status: Draft Phase 3 plan
Last updated: 2026-07-23

This plan starts Phase 3 after Phase 2 closeout. It is based on verified Phase 2
facts: live core inventory is approved, expanded live inventory is deferred, and
expanded inventory groups are available through fake-client/golden fixtures.

## 1. Goal

Build a deterministic health analysis foundation that can later feed the
recommendation engine.

The phase should identify operational blockers, warnings, and unknowns without
claiming final upgrade readiness.

## 2. Delivery slices

To target an MVP in roughly 30 PRs, Phase 3 is planned as four PRs:

| PR | Branch | Scope | Evidence |
| --- | --- | --- | --- |
| P3-01 | `docs/phase-3-health-plan` | Health contract, rule list, compressed PR strategy | Docs review |
| P3-02 | `feature/health-rule-foundation` | Finding types, severities/statuses, rule runner, deterministic clocks | Unit tests |
| P3-03 | `feature/health-node-workload-rules` | Node readiness/pressure/skew and workload availability rules | Fixture/rule tests |
| P3-04 | `feature/health-storage-event-rules` | Storage unknown checks, event warning-window checks, Phase 3 closeout | Fixture/rule tests and closeout docs |

If one implementation PR becomes too dense, split it rather than weakening
tests.

## 3. Live access boundary

Phase 3 does not expand live Kubernetes collection. Rules that require expanded
inventory groups use synthetic fixtures until Gate B is separately expanded.

The only verified live inventory remains P2-02 core namespace/node data.

## 4. Initial API shape

Planned package:

```text
internal/health
```

Initial concepts:

- `Finding`
- `Severity`
- `Status`
- `Rule`
- `Runner`
- `Options`
- injected clock for event-window tests

The package consumes `inventory.Snapshot` and returns ordered findings. It does
not call Kubernetes clients.

## 5. Determinism

Findings are sorted by:

1. severity rank;
2. rule ID;
3. resource namespace;
4. resource kind;
5. resource name;
6. summary.

Tests must cover stable ordering and `UNKNOWN` outcomes for missing evidence.

## 6. Stop conditions

Stop affected work if:

- a rule would require unapproved live expanded inventory;
- a missing evidence path would be treated as pass;
- a finding needs raw event messages or secret-like values;
- a health rule duplicates compatibility or provider recommendation policy;
- golden fixtures become nondeterministic.

## 7. Exit criteria

Phase 3 is complete when:

- finding model and runner are tested;
- node and workload rules are tested;
- storage/event rules are tested;
- `UNKNOWN` evidence behavior is tested;
- Phase 3 closeout records the remaining gaps and confirms Phase 4 can start.
