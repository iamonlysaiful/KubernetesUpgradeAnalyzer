# Gate B P2-02 validation record

Status: Draft; no live execution approved
Last updated: 2026-07-23

This record is the audit trail for the proposed Gate B smoke test in
`docs/plans/gate-b-smoke-test-plan.md`. It must be completed before, during, and
after any approved live run. Do not fill this with guessed values.

## 1. Approval

| Field | Value |
| --- | --- |
| Approval timestamp | Pending |
| Approver | Pending |
| Approved kube context | Pending |
| Kubeconfig source | Pending |
| Approved command | Pending |
| Output retention approved | Pending |
| Sanitized fixture proposal approved | Pending |
| Expected cluster identity | Pending |
| Stop instruction if context differs | Pending |

## 2. Pre-run confirmation

| Check | Result |
| --- | --- |
| Working tree clean before run | Pending |
| AppleDouble sidecars removed before run | Pending |
| `scripts/ci-local.sh` before run | Pending |
| Approved context confirmed | Pending |
| Command matches approval exactly | Pending |

## 3. Execution

| Field | Value |
| --- | --- |
| Started at | Pending |
| Completed at | Pending |
| Exit code | Pending |
| Stdout handling | Pending |
| Stderr handling | Pending |
| Raw output local path, if approved | Pending |

## 4. Scope verification

| Approved read | Observed status |
| --- | --- |
| Server version discovery | Pending |
| API discovery | Pending |
| SelfSubjectAccessReview checks | Pending |
| List namespaces | Pending |
| List nodes | Pending |

| Prohibited read/action | Observed status |
| --- | --- |
| Secrets | Pending |
| ConfigMap contents | Pending |
| Pods/workloads | Pending |
| Storage/networking/CRDs/events | Pending |
| Watch | Pending |
| Mutation verbs | Pending |
| Azure CLI/kubent/provider calls | Pending |

## 5. Output review

| Check | Result |
| --- | --- |
| Stdout is JSON only | Pending |
| Snapshot subset validation passed | Pending |
| Contains `PARTIAL_INVENTORY_P2_02` limitation | Pending |
| No kubeconfig/token/certificate/password | Pending |
| No Secret payload | Pending |
| No raw provider ID value | Pending |
| No unexpected sensitive identifier | Pending |

## 6. Post-run checks

| Check | Result |
| --- | --- |
| `scripts/ci-local.sh` after run | Pending |
| `git diff --check` after run | Pending |
| `git fsck --full --strict` after cleanup | Pending |
| AppleDouble sidecars removed after run | Pending |
| Raw output cleanup or retention matches approval | Pending |

## 7. Decision

Gate B status for P2-02: Pending.

Decision notes:

- Pending.

Follow-up restrictions:

- Passing this record does not approve pods, workloads, storage, networking,
  CRDs, events, health, compatibility, provider, recommendation, or report
  collection.
