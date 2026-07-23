# Core inventory collector contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract governs P2-02: core cluster collectors and normalized cluster
snapshot construction. It depends on P2-01 preflight and does not open Gate B by
itself.

## 1. Scope

P2-02 adds the first normalized inventory snapshot path for:

- cluster/context metadata from the approved preflight result;
- Kubernetes server version metadata;
- namespaces;
- nodes;
- collection limitations.

P2-02 does not collect pods, workloads, storage, networking, CRDs, events,
Secrets, ConfigMap contents, provider evidence, kubent output, health findings,
component detections, recommendations, reports, or redacted sharing artifacts.
Those remain later work packages.

During P2-02, `kua inventory` may render a partial snapshot only if it is
explicitly labeled as partial/core inventory. It must not imply that workload,
storage, networking, CRD, event, health, API-compatibility, component, provider,
or recommendation analysis has happened.

P2-02 updates `kua inventory --format=json` to run preflight first and then emit
the partial/core `ClusterSnapshot` JSON only when required preflight evidence
passes. If required preflight evidence fails, the command returns
`INCONCLUSIVE` and does not render a snapshot that could be mistaken for
complete inventory.

`--format=console` may remain a conservative preflight/core summary during this
package. Full human inventory rendering remains a later report/output package.

## 2. Snapshot shape

The emitted snapshot follows `schemas/cluster-snapshot/v1.json`.

JSON snapshot output must emit only JSON on stdout. Diagnostics, collection
errors, and limitations summaries outside the snapshot go to stderr.

Because the schema requires future inventory groups, P2-02 must populate
out-of-scope groups as empty arrays:

- `workloads`;
- `storage`;
- `networking`;
- `crds`;
- `events`.

Empty arrays mean "not collected in this work package" only when paired with an
explicit limitation. They must not be interpreted as proof that the cluster has
no matching resources.

## 3. Collection behavior

Approved P2-02 read-only collection operations, when separately authorized for a
named live context, are:

- `list` namespaces;
- `list` nodes;
- server version discovery already covered by P2-01.

Secrets are excluded. ConfigMap contents are excluded. `watch` is excluded.
Mutation verbs are excluded. Object payloads must be reduced to the fields
needed by the snapshot schema.

## 4. Required fields

Namespace records include:

- kind;
- name;
- optional API version;
- stable UID alias when redaction or fixture stability requires it.

Node records include:

- ref;
- kubelet version;
- provider ID presence as a boolean;
- node pool only when safely derivable from non-secret labels/provider metadata;
- conditions reduced to type, status, and reason.

Provider ID values, node labels, annotations, allocatable/capacity details, and
taints are not collected in P2-02 unless a later contract explicitly adds them.

## 5. Limitations

Every partial or denied collection path records a limitation. Required examples:

- `PARTIAL_INVENTORY_P2_02` when future inventory groups are intentionally empty;
- `NAMESPACE_COLLECTION_FAILED` when namespace collection cannot complete;
- `NODE_COLLECTION_FAILED` when node collection cannot complete;
- `OPTIONAL_FIELD_OMITTED` when a non-required optional field is intentionally
  not captured.

Required evidence denial or collection failure returns an inconclusive result for
the affected command. Unknown, denied, missing, or intentionally skipped evidence
must never be rendered as `PASS`.

Namespace or node collection failure prevents snapshot emission because those
collections are required for P2-02. The command records the failure as an error
path instead of rendering partial data that cannot satisfy the P2-02 contract.

## 6. Test boundary

Automated tests must use fake clients and sanitized fixtures only. Live cluster
smoke testing remains blocked until the user approves Gate B for a named context
and read-only operation.

P2-02 snapshot fixtures must be deterministic and safe to commit:

- generated from fake Kubernetes clients or hand-written sanitized examples;
- no real cluster names, node names, provider IDs, UIDs, labels, annotations,
  IPs, image paths, event messages, or kubeconfig-derived values;
- stable timestamps and snapshot IDs;
- sorted namespaces, nodes, and node conditions;
- JSON matching `schemas/cluster-snapshot/v1.json`;
- explicit `PARTIAL_INVENTORY_P2_02` limitation.

At least one P2-02 golden fixture must cover the generated partial/core snapshot
shape used by `kua inventory --format=json`.

## 7. Snapshot validation

P2-02 adds a dependency-free validator for the generated partial/core
`ClusterSnapshot` subset. The validator is not a full JSON Schema implementation
and must not be described as one. It checks the P2-02 fields most likely to
drift from `schemas/cluster-snapshot/v1.json`:

- schema version, snapshot ID, and RFC3339 capture timestamp;
- required cluster identity, provider, and context fields;
- supported provider type and confidence enums;
- supported Kubernetes version range;
- namespace and node required fields;
- node condition status enums;
- required future inventory groups are non-nil arrays;
- limitation code, severity, and summary.

Full draft 2020-12 JSON Schema validation remains deferred until a focused
dependency/tooling assessment approves a validator.
