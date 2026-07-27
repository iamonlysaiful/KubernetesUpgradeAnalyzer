# CLI and report contracts

Status: Accepted with Phase 8.5 recovery updates
Last updated: 2026-07-27

## 1. Common flags

Common flags include `--context`, `--kubeconfig`, `--config`, `--input`,
`--output`, `--format`, `--provider-source=auto`, `--provider-evidence`,
`--subscription`, `--resource-group`, `--cluster-name`, `--target-version`,
`--component-overrides`, `--redacted`, and `--log-level`. `--namespace`,
`--timeout`, and `--catalog` remain planned flags and must not be documented as
complete until wired.

Provider source values are `auto`, `azure`, `file`, `offline`, and `none`. No command mutates the cluster or Azure resources.

Standard kubeconfig resolution applies: explicit `--kubeconfig`, then normal client-go environment/default behavior; the current context is used unless `--context` is set. Multi-context batch analysis is not part of MVP.

## 2. Command behavior

- `kua analyze`: preflight, collect, run all required analyzers, recommend, render.
- `kua inventory`: collect and render inventory; kubent is not required.
- `kua health`: collect health inputs and render findings; kubent is not required.
- `kua compatibility`: run component and API compatibility; kubent is required in MVP.
- `kua report --input assessment.json`: render canonical JSON without cluster access.
- `kua component-overrides --input assessment.json --output component-overrides.json`: generate a local operator-fillable component override file from an assessment's `componentVersionOverrides` helper without cluster or provider access.
- `kua version`: show build, Go, schema, and embedded catalog versions.

During Phase 8.5, `kua analyze` is the recovery command. It must be usable and
must not fail as unimplemented. Until kubent wiring lands, it returns
`INCONCLUSIVE` with explicit limitations for missing API compatibility evidence.
That temporary behavior is valid only inside Phase 8.5 and must not be used to
claim MVP release readiness.

Phase 8.5 expands the `kua analyze` live inventory path to the accepted MVP
read-only metadata groups: workloads, storage, networking, CRDs, and events.
This expansion does not approve unsupervised live staging execution. Any run
against a real context still requires explicit command approval under the
staging record. `kua inventory` may remain core-only until its own output
contract is updated.

Phase 8.5 wires kubent API compatibility through the existing adapter. KUA may
invoke only the approved kubent command shape: JSON output, `--helm3=false`,
explicit target version, optional kubeconfig, and optional context. If no target
version can be determined, or kubent is missing, wrong-versioned, malformed, or
fails, API compatibility becomes `INCONCLUSIVE` with sanitized limitations.
Stderr and raw kubent diagnostics must not be copied into committed records.
When provider evidence and kubent API compatibility both pass and no health or
component blockers are present, `kua analyze` may return `READY` or
`READY_WITH_WARNINGS` according to the recommendation policy.

`kua health` and `kua compatibility` may initially render filtered views of the
same assessment pipeline, as long as omitted evidence is reported as a
limitation. `kua report --input <assessment.json>` renders a saved report
document without Kubernetes or provider access.

`auto` provider behavior for detected AKS clusters is: use the local authenticated Azure CLI, fall back to supplied JSON evidence, then continue with exact provider availability `UNKNOWN`. `azure` and `file` require their named source. `offline` prohibits provider network access but may consume a local evidence file. `none` skips provider-specific analysis.

KUA never starts `az login`; it reports the corrective command when authentication is missing or expired.

When component evidence is missing, ambiguous, or confusing, JSON assessment
output must include a `componentVersionOverrides` helper object. The helper must
list each affected component, observed versions when available, placeholder
versions to fill when no confident version exists, and a rerun command using
`--component-overrides <file>`. The placeholder is operator input only; KUA must
not infer support from it until the user supplies a filled overrides file.

`--component-overrides <file>` consumes a local JSON file with this shape:

```json
{
  "schemaVersion": "kua.component-overrides.v1",
  "components": [
    {
      "id": "coredns",
      "versions": ["1.9.4-13"],
      "evidence": "user-confirmed"
    }
  ]
}
```

Overrides may add or confirm component versions for the current assessment. They
must not suppress explicit incompatibility findings or provider/API/health
findings. Unknown component-version findings may be downgraded only when the
file supplies one or more non-placeholder versions for the matching component.

To avoid long manual JSON editing, `kua component-overrides --input
assessment.json --output component-overrides.json` writes an override file using
observed versions when present and placeholders only when no version was
observed. The command must refuse to overwrite an existing file and must not
contact Kubernetes or Azure.

## 3. Output discipline

Human output goes to stdout. Logs and diagnostics go to stderr. JSON mode emits only JSON on stdout. File writes should be atomic and must not overwrite an existing file unless an approved `--force` contract is added.

During P2-01, `kua inventory` is a preflight-only command. Its console and JSON
outputs must explicitly say they are preflight-only and must not imply that full
inventory collection has happened.

## 4. Canonical assessment

Top-level fields include:

- `schemaVersion`, `assessmentId`, `generatedAt`;
- tool/catalog/kubent versions;
- sanitized cluster and provider metadata;
- collection scope and limitations;
- inventory and component detections;
- component version override template when operator input is needed;
- findings;
- current version, candidates, destination, and staged path;
- readiness, risk, decision trace, and recommended actions.

Schema version changes follow compatibility policy: additive optional fields are minor; breaking changes require a new major schema and migration/rendering plan.

## 5. Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Assessment completed and is `READY` or `READY_WITH_WARNINGS` |
| `2` | Assessment completed and is `NOT_READY` |
| `3` | Assessment completed but is `INCONCLUSIVE` |
| `4` | Usage or configuration error |
| `5` | Collection, dependency, catalog, or internal execution error |

Exact automation semantics must be locked with CLI contract tests before
implementation is considered stable. During Phase 8.5, `INCONCLUSIVE` is the
expected exit for end-to-end analysis when required MVP evidence is still
unwired or unavailable.

`READY_WITH_WARNINGS` returns `0`. A future explicit strict mode may assign a nonzero result without changing the default contract.

## 6. HTML safety

HTML is self-contained, escapes all evidence, uses no remote scripts/fonts/assets, and contains no executable user-provided markup. Markdown escapes or safely fences untrusted content.

## 7. Redaction

All report formats support an MVP redacted mode for sharing. Redaction replaces cluster, subscription, resource group, namespace, workload, node, registry, and sensitive event identifiers with stable assessment-local aliases. It preserves finding IDs, versions, counts, readiness, risk, and decision logic. Local output remains unredacted by default for remediation value; reports state their redaction status.
