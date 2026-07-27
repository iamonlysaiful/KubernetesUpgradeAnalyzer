# Phase 8.5 end-to-end CLI recovery plan

Status: Approved recovery plan
Last updated: 2026-07-27

This plan corrects the gap between the accepted MVP product goal and the
current implementation state. The MVP release must not be published until the
user-facing CLI can run an end-to-end AKS upgrade assessment.

## 1. Problem statement

The accepted MVP requires KUA to scan a live AKS cluster and produce an
explainable upgrade recommendation. Phases 3 through 8 implemented the internal
health, compatibility, provider, recommendation, and report packages, but the
CLI still exposes only live core inventory. `kua analyze`, `kua health`,
`kua compatibility`, and `kua report` are not yet wired as usable product flows.

Release-candidate artifact generation alone is therefore not sufficient for an
MVP release.

## 2. Release hold

Phase 9 publication is blocked until this recovery plan exits successfully.
Release-candidate binaries may be generated for local verification, but no tag,
GitHub release, published artifact, or public release claim is approved while
the end-to-end CLI remains incomplete.

## 3. Required MVP command behavior

### 3.1 `kua analyze`

`kua analyze` is the primary MVP command. It must:

- resolve kubeconfig and context using existing preflight rules;
- collect live inventory required for MVP analysis;
- run health rules over collected evidence;
- run component detection and catalog compatibility over collected evidence;
- run kubent API compatibility when permitted and available;
- collect AKS provider upgrade evidence through `auto`, `azure`, `file`,
  `offline`, or `none` source modes;
- build provider-valid candidate versions and sequential upgrade stages;
- produce readiness, risk, destination, blockers, warnings, limitations,
  assumptions, and remediation;
- render console, JSON, Markdown, and HTML outputs through the approved report
  package;
- support redacted output for shareable reports.

Missing optional evidence must produce `UNKNOWN` or `INCONCLUSIVE` findings and
limitations, never a silent pass.

### 3.2 `kua inventory`

`kua inventory` must either remain explicitly partial or be upgraded to expose
the same live inventory scope needed by `kua analyze`. It must not imply full
assessment coverage while future inventory groups are omitted.

### 3.3 `kua health`, `kua compatibility`, and `kua report`

These commands may be implemented as thin wrappers around the same assessment
pipeline, but they must no longer fail as unimplemented commands before the MVP
release.

## 4. Strict work packages

| ID | Package | Exit evidence |
| --- | --- | --- |
| P8.5-01 | CLI contract correction | Docs and tests define exact command behavior, flags, exit codes, redaction, and live-command boundaries. |
| P8.5-02 | Live inventory expansion gate | Read-only collectors required by `kua analyze` are wired for live AKS only after Gate B expansion approval. |
| P8.5-03 | Assessment pipeline assembly | One internal pipeline connects inventory, health, components, kubent, provider evidence, recommendations, and report documents. |
| P8.5-04 | User-facing command wiring | `analyze`, `health`, `compatibility`, `report`, and `inventory` use the pipeline or documented subsets. |
| P8.5-05 | AKS staging validation | Approved staging run proves the CLI output can produce a useful AKS upgrade analysis without secrets or private identifiers in committed material. |
| P8.5-06 | Release readiness reset | Phase 9 records are updated only after sanitized validation evidence passes and owner publication approval is requested separately. |

## 5. Acceptance criteria

1. `kua analyze --format json --provider-source=file --provider-evidence <file>`
   over a sanitized AKS fixture validates against the assessment schema.
2. `kua analyze` can evaluate a Kubernetes `1.30.x` AKS cluster and recommend a
   provider-valid staged path toward `1.33.x` when evidence supports it.
3. If Azure CLI authentication is unavailable in `auto` mode, provider evidence
   becomes `UNKNOWN` and the command reports the corrective action without
   running `az login`.
4. Kubent is invoked only through the controlled adapter, with Helm collection
   disabled, and absent kubent evidence yields `INCONCLUSIVE` API findings.
5. Live expanded Kubernetes reads are limited to the documented read-only fields
   and require explicit Gate B expansion approval before staging execution.
6. Console, JSON, Markdown, and HTML outputs render deterministically from the
   same canonical assessment.
7. Redacted output removes cluster, subscription, resource group, namespace,
   workload, node, registry, and event identifiers while preserving versions,
   counts, decisions, and remediation.
8. `scripts/ci-local.sh`, targeted race checks, schema/golden validation,
   `git diff --check`, and `git fsck --full --strict` pass before release
   publication is reconsidered.

## 6. Non-goals

- Mutating or upgrading clusters.
- Production validation before staging passes.
- Native deprecated API analysis.
- Internet catalog scraping or automatic catalog downloads.
- Publishing release artifacts before owner approval.

## 7. Stop conditions

Stop and return to docs if implementation would:

- collect secrets, kubeconfig material, secret-backed values, or raw provider
  identifiers for committed artifacts;
- require live cluster commands beyond the approved read-only command list;
- publish raw staging evidence, context names, subscription IDs, resource
  groups, cluster names, node names, namespaces, workload names, registry hosts,
  trace IDs, or authentication details;
- report `READY` or `LOW` risk when required evidence is missing,
  unauthenticated, stale, or contradictory;
- leave any MVP command unimplemented while claiming MVP release readiness.
