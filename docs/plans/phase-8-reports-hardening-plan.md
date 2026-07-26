# Phase 8 reports and hardening plan

Status: Proposed plan for approval
Last updated: 2026-07-27

## 1. Scope

Phase 8 implements report rendering and hardening gates as defined in the roadmap,
CLI/report contract, testing strategy, and security/privacy requirements.

To keep delivery near the MVP PR target now that PR #30 is merged, Phase 8 is
compressed into two PRs:

| PR | Content |
| --- | --- |
| 31 | This plan document |
| 32 | Report renderers + hardening checks + Phase 8 closeout |

## 2. Deliverables

Phase 8 implementation must deliver:

- canonical JSON report rendering;
- deterministic console rendering;
- deterministic Markdown rendering;
- self-contained HTML rendering with no remote assets;
- redacted output mode with stable assessment-local aliases;
- hostile-input escaping validation for Markdown/HTML renderers;
- golden fixtures for report outputs;
- hardening checks integrated into local CI contract.

## 3. Report contracts

Report outputs must follow `docs/cli-and-reports.md`:

- stdout carries user-facing output;
- stderr carries logs/diagnostics;
- JSON mode emits only JSON on stdout;
- file output uses atomic write and does not overwrite existing files unless a
  future approved `--force` contract is added;
- rendered outputs preserve findings/readiness/risk/decision logic exactly.

## 4. Renderer package layout

Planned package structure:

```text
internal/
  report/
    json.go
    console.go
    markdown.go
    html.go
    redaction.go
    writer.go
    types.go
    testdata/
```

## 5. Redaction mode

Phase 8 adds MVP redacted mode for all report formats:

- redacts cluster, subscription, resource group, namespace, workload, node,
  registry, and sensitive event identifiers;
- uses stable per-assessment aliases (deterministic for same assessment);
- preserves version data, finding IDs, counts, readiness, risk, and destination/path;
- clearly marks report as redacted.

## 6. HTML safety requirements

HTML renderer must enforce:

- fully self-contained output (no external JS/CSS/fonts);
- escaping of all untrusted evidence fields;
- no executable user-provided markup;
- deterministic section ordering and stable IDs;
- safe handling of malformed/hostile strings.

## 7. Hardening gates

Phase 8 extends quality gates with explicit checks from testing/security docs:

- golden report fixtures for JSON/console/Markdown/HTML;
- hostile-input rendering tests (script/style/HTML injection payloads);
- redaction equivalence tests (decisions unchanged, identifiers redacted);
- race checks for concurrent rendering paths where applicable;
- dependency and license checks aligned with current contract;
- deterministic output verification across repeated runs.

## 8. Test matrix

Required test groups:

| Group | Cases |
| --- | --- |
| Renderer fidelity | JSON/console/Markdown/HTML golden snapshots |
| Redaction | alias stability, identifier replacement, logic equivalence |
| Hostile input | escaped script tags, malformed markdown/html payloads |
| Determinism | repeated render of same input is byte-equivalent |
| File safety | atomic writes, no overwrite without explicit approval |
| Exit behavior | report command does not require cluster access when input JSON is supplied |

## 9. Out of scope for Phase 8

Phase 8 does not include:

- live AKS staging validation;
- release artifact publication/signing;
- tag/release creation;
- policy changes to recommendation engine;
- additional provider adapters.

These remain in Phase 9 and require separate explicit approval.

## 10. Exit criteria

Phase 8 is complete when:

- all four report formats are implemented and deterministic;
- redacted mode is available for all report formats;
- HTML safety and hostile-input tests pass;
- extended hardening checks pass in local CI;
- Phase 8 closeout record is added with evidence;
- no live cluster/provider mutation or unapproved outbound calls occur.
