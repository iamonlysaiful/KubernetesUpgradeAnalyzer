# Storage inventory collector contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract governs the storage portion of P2-03. It depends on P2-02 core
inventory and the earlier P2-03 workload, CRD, and networking fixture paths. It
does not approve live storage collection by itself.

## 1. Scope

This package adds fake-client-first storage snapshot collection for:

- PersistentVolumeClaims;
- PersistentVolumes;
- StorageClasses.

The collected storage records populate `inventory.storage` in
`schemas/cluster-snapshot/v1.json`.

This package may add an internal snapshot assembly path that includes storage
resources for fake-client and golden-fixture tests. `kua inventory --format=json`
must not emit live storage records until Gate B is separately expanded for
storage reads.

This package does not collect Secret data, CSI node/pod volume mounts,
PersistentVolume claim refs, volume handles, Azure disk/file IDs, storage
account names, annotations, labels, raw UIDs, events, health findings,
compatibility findings, provider evidence, recommendations, or reports.

## 2. Required fields

Each storage record is represented as a `ResourceRef` and includes:

- `apiVersion`;
- `kind`;
- `name`;
- `namespace` only for namespaced resources.

Supported `kind` values are:

- `PersistentVolumeClaim` with `apiVersion` `v1` and non-empty `namespace`;
- `PersistentVolume` with `apiVersion` `v1` and empty `namespace`;
- `StorageClass` with `apiVersion` `storage.k8s.io/v1` and empty `namespace`.

The MVP snapshot schema currently models storage resources as `ResourceRef`
values only. Capacity, access modes, reclaim policy, volume mode, provisioner,
parameters, binding state, selected node, and claim references are intentionally
deferred until a schema expansion is approved.

## 3. Determinism and safety

Storage resources are sorted by:

1. namespace, with cluster-scoped resources first;
2. kind;
3. name.

Fixtures must use sanitized names only. Real PVC/PV/StorageClass names, cloud
volume identifiers, storage account names, annotations, labels, claim refs,
driver parameters, and UIDs must not be committed.

## 4. Limitations

When storage collection is intentionally absent or not yet approved for live use,
KUA must not imply that the cluster has no storage resources. P2-03 fixture paths
that include storage must use a limitation that names the inventory groups still
intentionally uncollected.

Collection failure for any required storage API group makes the affected command
fail safely instead of emitting misleading partial storage data.

## 5. Gate B expansion

P2-02 Gate B passed only for namespace/node collection. Live workload, CRD,
networking, and storage collection require separate Gate B expansion approval
naming the context and allowing the specific read-only API operations. Until
then, automated tests use fake clients only.

## 6. Fixture and validation expectations

At least one P2-03 golden fixture must include representative sanitized PVC, PV,
and StorageClass refs. The fixture must continue to use an empty events array
until event collection is implemented and approved.

The dependency-free snapshot subset validator must be expanded to cover storage
required fields and allowed kind/apiVersion/namespace combinations before any
storage snapshot path is considered complete.
