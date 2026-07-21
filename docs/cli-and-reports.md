# CLI and report contracts

Status: Proposed for implementation  
Last updated: 2026-07-21

## 1. Common flags

Proposed common flags include `--context`, `--kubeconfig`, `--namespace`, `--timeout`, `--config`, `--output`, `--format`, `--catalog`, `--provider-evidence`, `--log-level`, and `--network-mode=offline`.

No command mutates the cluster. An online mode is reserved and must fail as unsupported until separately approved and implemented.

## 2. Command behavior

- `kua analyze`: preflight, collect, run all required analyzers, recommend, render.
- `kua inventory`: collect and render inventory; kubent is not required.
- `kua health`: collect health inputs and render findings; kubent is not required.
- `kua compatibility`: run component and API compatibility; kubent is required in MVP.
- `kua report --input assessment.json`: render canonical JSON without cluster access.
- `kua version`: show build, Go, schema, and embedded catalog versions.

## 3. Output discipline

Human output goes to stdout. Logs and diagnostics go to stderr. JSON mode emits only JSON on stdout. File writes should be atomic and must not overwrite an existing file unless an approved `--force` contract is added.

## 4. Canonical assessment

Top-level fields include:

- `schemaVersion`, `assessmentId`, `generatedAt`;
- tool/catalog/kubent versions;
- sanitized cluster and provider metadata;
- collection scope and limitations;
- inventory and component detections;
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

Exact automation semantics must be locked with CLI contract tests before implementation is considered stable.

## 6. HTML safety

HTML is self-contained, escapes all evidence, uses no remote scripts/fonts/assets, and contains no executable user-provided markup. Markdown escapes or safely fences untrusted content.

