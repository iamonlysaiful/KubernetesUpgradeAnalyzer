# Gate B smoke-test plan

Status: Proposed; not approved for execution
Last updated: 2026-07-23

This plan defines the minimum live read-only smoke test needed to open Gate B for
P2-02 core inventory. It is not approval to access a cluster. The user must
explicitly approve the named context and operation before any command touches a
live cluster.

## 1. Purpose

Gate B validates that the P2-02 live collection path stays within the approved
read-only boundary and can produce a partial/core `ClusterSnapshot` for one
approved staging context.

The smoke test covers:

- P2-01 preflight;
- P2-02 namespace collection;
- P2-02 node collection;
- generated partial/core snapshot validation;
- stdout/stderr discipline;
- cleanup and artifact handling.

It does not cover pods, workloads, storage, networking, CRDs, events, health,
kubent, Azure CLI, provider evidence, recommendations, reports, redaction, or
release validation.

## 2. Approval record required before execution

Before any live command, record explicit user approval containing:

- kube context name;
- kubeconfig source: default kubeconfig or explicit path;
- command to run;
- output path, if any;
- whether output may be retained locally;
- whether sanitized derived output may later be proposed as a fixture;
- expected identity of the cluster as understood by the user;
- stop/rollback instruction if the active context differs.

Approval must be for a staging/non-production cluster unless the user explicitly
states otherwise. Approval for this smoke test does not approve future broad
inventory, health, compatibility, provider, or recommendation collection.

## 3. Pre-command read-only confirmation

Immediately before running KUA against the approved context, confirm:

```text
kubectl config current-context
```

or the equivalent client-go resolved context shown by KUA.

If the current or selected context does not exactly match the approved context,
stop without running collection.

## 4. Approved command shape

Preferred command:

```text
kua --context <approved-context> --format=json inventory
```

Allowed variation when the user approves an explicit kubeconfig path:

```text
kua --kubeconfig <approved-kubeconfig> --context <approved-context> --format=json inventory
```

No `kubectl get`, `az`, `kubent`, shell pipelines, or helper commands are part
of this smoke test unless separately approved.

## 5. Approved Kubernetes reads

The live KUA path may perform only:

- server version discovery;
- API discovery;
- `SelfSubjectAccessReview` permission checks;
- `list namespaces`;
- `list nodes`.

Secrets are prohibited. ConfigMap contents are prohibited. `watch` is
prohibited. Mutation verbs are prohibited. Pods, workloads, storage, networking,
CRDs, and events are not collected in this smoke test.

## 6. Expected output

Successful output is a partial/core `ClusterSnapshot` JSON document containing:

- schema version;
- snapshot ID and capture timestamp;
- approved context name;
- kubeconfig source;
- Kubernetes server version;
- namespace refs;
- node refs, kubelet versions, provider ID presence, node pool when safely
  derivable, and node conditions;
- empty arrays for out-of-scope inventory groups;
- `PARTIAL_INVENTORY_P2_02` limitation.

Diagnostics go to stderr. JSON stdout must pass the P2-02 subset validator before
being written.

## 7. Stop conditions

Stop immediately if:

- active/selected context differs from approval;
- the command requests or appears to request Secrets, ConfigMaps, pods,
  workloads, storage, networking, CRDs, events, watch, or mutation verbs;
- stdout contains raw kubeconfig material, token, certificate, password, Secret,
  provider ID value, UID, IP address, or unexpected sensitive identifier;
- validation fails;
- Git integrity or AppleDouble metadata is unhealthy before publication;
- the user interrupts or changes approval.

## 8. Artifact handling

Raw live output must not be committed. If the user approves retaining local
output, store it under a local ignored path and record the path in the validation
notes. Any future fixture derived from the smoke test requires a separate
sanitization review and explicit approval before commit.

Cleanup of raw outputs, temporary files, and recovery archives follows the
recoverable cleanup policy. Do not delete recovery artifacts without separate
approval.

## 9. Gate B success criteria

Gate B may be marked open/passed for P2-02 only when:

- the approved command ran against the approved context;
- no out-of-scope Kubernetes reads occurred;
- stdout was valid partial/core snapshot JSON;
- stderr contained no sensitive data;
- raw output was handled according to approval;
- local CI still passes after the smoke test;
- `git fsck --full --strict` shows no corruption beyond accepted known dangling
  blobs.

Passing Gate B for P2-02 does not approve later workload, storage, networking,
CRD, event, health, compatibility, provider, or recommendation collectors.
