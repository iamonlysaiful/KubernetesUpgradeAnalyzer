# Phase 5 API compatibility plan

Status: Draft Phase 5 plan
Last updated: 2026-07-24

This plan starts Phase 5 after Phase 4 closeout. It is based on verified facts:
Phase 4 catalog and component detection foundations are merged, live inventory
remains core-only, and kubent integration is still deferred.

## 1. Goal

Build a controlled API compatibility foundation using kubent for MVP, then
record whether kubent coverage is sufficient for Kubernetes `1.30` through
`1.33`.

Phase 5 should produce normalized API compatibility evidence and explicit
limitations. It must not claim upgrade readiness.

## 2. Delivery slices

To stay near the current 30 ±2 PR target, Phase 5 is planned as three PRs:

| PR | Branch | Scope | Evidence |
| --- | --- | --- | --- |
| P5-01 | `docs/phase-5-api-compat-plan` | API compatibility contract, kubent adapter plan, Phase 4 status cleanup | Docs review |
| P5-02 | `feature/kubent-adapter-foundation` | Process runner interface, version validation, bounded execution, JSON parsing fixtures | Unit/process-fake tests |
| P5-03 | `feature/kubent-coverage-decision` | Target-rule coverage validation, normalized API findings, Phase 5 go/no-go and closeout | Fixture tests and decision record |

If kubent output coverage is more complex than expected, split P5-03 rather than
weakening negative-path tests.

## 3. Live and process boundary

Phase 5 implementation tests use fakes and fixtures. Running kubent against a
real cluster requires a separate approval naming:

- kube context;
- command;
- target version;
- raw output handling;
- cleanup and validation expectations.

No Phase 5 docs or implementation commit is approval for live kubent execution.

## 4. Initial package shape

Planned package:

```text
internal/external/kubent
```

Initial concepts:

- `Binary` or `Runner` abstraction;
- version probe result;
- bounded command options;
- parsed kubent report;
- normalized API finding;
- coverage validator;
- go/no-go decision record.

## 5. Stop conditions

Stop affected work if:

- kubent would need to read Helm Secrets or ConfigMaps;
- implementation requires shell interpolation;
- empty or missing kubent evidence could become pass;
- target coverage cannot be verified;
- live execution would be needed without explicit approval;
- output fixtures expose cluster identifiers or sensitive data.

## 6. Exit criteria

Phase 5 is complete when:

- kubent version validation is tested;
- process execution is bounded and shell-free;
- JSON parsing and malformed-output paths are tested;
- target coverage validation is tested;
- normalized API findings preserve `UNKNOWN`/`INCONCLUSIVE` where evidence is
  incomplete;
- Phase 5 closeout records whether kubent remains acceptable for MVP or a
  minimal native analyzer must be added.
