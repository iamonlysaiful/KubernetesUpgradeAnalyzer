# Gate B P2-02 validation record

Status: Passed for P2-02 core inventory
Last updated: 2026-07-23

This record is the audit trail for the proposed Gate B smoke test in
`docs/plans/gate-b-smoke-test-plan.md`. It must be completed before, during, and
after any approved live run. Do not fill this with guessed values.

## 1. Approval

| Field | Value |
| --- | --- |
| Approval timestamp | 2026-07-23 |
| Approver | Project owner |
| Approved kube context | Approved staging context; exact name withheld from public record |
| Kubeconfig source | Default local kubeconfig |
| Approved command | `kua --context <approved-staging-context> --format=json inventory` |
| Output retention approved | Yes; local ignored path only; never commit raw output |
| Sanitized fixture proposal approved | No, not for this run |
| Expected cluster identity | Staging AKS cluster; exact name withheld from public record |
| Stop instruction if context differs | Stop immediately |

## 2. Pre-run confirmation

| Check | Result |
| --- | --- |
| Working tree clean before run | Passed |
| AppleDouble sidecars removed before run | Passed |
| `scripts/ci-local.sh` before run | Passed |
| Approved context confirmed | Passed; exact context name withheld from public record |
| Command matches approval exactly | Passed |

## 3. Execution

| Field | Value |
| --- | --- |
| Started at | 2026-07-23 |
| Completed at | 2026-07-23 |
| Exit code | `0` on approved run with normal network access |
| Stdout handling | Raw JSON retained locally under ignored `local/`; not committed |
| Stderr handling | Empty on successful run |
| Raw output local path, if approved | Ignored local `local/gate-b/` path; exact path withheld from public record |

Note: an initial sandboxed attempt failed before collection because DNS
resolution of the Kubernetes API endpoint was unavailable in the sandboxed
network environment. The same approved command was rerun with normal network
access and succeeded.

## 4. Scope verification

| Approved read | Observed status |
| --- | --- |
| Server version discovery | Passed |
| API discovery | Passed |
| SelfSubjectAccessReview checks | Passed |
| List namespaces | Passed |
| List nodes | Passed |

| Prohibited read/action | Observed status |
| --- | --- |
| Secrets | Not observed |
| ConfigMap contents | Not observed |
| Pods/workloads | Not observed; generated snapshot reported zero workloads |
| Storage/networking/CRDs/events | Not observed; generated snapshot reported zero for these groups |
| Watch | Not observed |
| Mutation verbs | Not observed |
| Azure CLI/kubent/provider calls | Not observed |

## 5. Output review

| Check | Result |
| --- | --- |
| Stdout is JSON only | Passed |
| Snapshot subset validation passed | Passed |
| Contains `PARTIAL_INVENTORY_P2_02` limitation | Passed |
| Namespace count | `8` |
| Node count | `1` |
| Out-of-scope inventory counts | `0` workloads, storage, networking, CRDs, and events |
| No kubeconfig/token/certificate/password | Passed |
| No Secret payload | Passed |
| No raw provider ID value | Passed; only `providerIdPresent` boolean field name observed |
| No unexpected sensitive identifier | Passed by limited review; raw output retained locally only |

## 6. Post-run checks

| Check | Result |
| --- | --- |
| `scripts/ci-local.sh` after run | Passed |
| `git diff --check` after run | Passed |
| `git fsck --full --strict` after cleanup | Passed with only accepted known dangling blobs |
| AppleDouble sidecars removed after run | Passed |
| Raw output cleanup or retention matches approval | Passed; retained only under ignored local path |

## 7. Decision

Gate B status for P2-02: Passed.

Decision notes:

- The P2-02 live smoke test succeeded for the approved staging context using
  only the approved read-only scope.
- The result validates the current namespace/node core inventory path only.
- Raw output was not committed and must not be committed later.

Follow-up restrictions:

- Passing this record does not approve pods, workloads, storage, networking,
  CRDs, events, health, compatibility, provider, recommendation, or report
  collection.
