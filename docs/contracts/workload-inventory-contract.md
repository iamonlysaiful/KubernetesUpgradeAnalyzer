# Workload inventory collector contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract governs the workload portion of P2-03. It depends on P2-02 core
inventory and the Gate B result for namespace/node collection. It does not
approve live workload collection by itself.

## 1. Scope

This package adds fake-client-first workload snapshot collection for:

- Deployments;
- DaemonSets;
- StatefulSets;
- ReplicaSets;
- Jobs;
- CronJobs.

The collected workload records populate `inventory.workloads` in
`schemas/cluster-snapshot/v1.json`.

This package may add an internal snapshot assembly path that includes workloads
for fake-client and golden-fixture tests. `kua inventory --format=json` must not
emit live workload records until Gate B is separately expanded for workload
reads.

This package does not collect pods, pod logs, Secret data, ConfigMap contents,
environment variables, volumes, service account tokens, workload annotations,
raw UIDs, owner references, storage, networking, CRDs, events, health findings,
compatibility findings, provider evidence, recommendations, or reports.

## 2. Required fields

Each workload record includes:

- `ref.apiVersion`;
- `ref.kind`;
- `ref.namespace`;
- `ref.name`;
- `desiredReplicas`;
- `readyReplicas`;
- `critical`;
- `containers`.

Container records include only:

- container name;
- image string;
- image tag when safely parseable from the image string.

Container env vars, envFrom refs, commands, args, ports, resources, probes,
volume mounts, and security contexts are out of scope for this package.

## 3. Replica semantics

Replica fields are normalized as:

- Deployments: `spec.replicas` defaulting to `1`, `status.readyReplicas`;
- StatefulSets: `spec.replicas` defaulting to `1`, `status.readyReplicas`;
- ReplicaSets: `spec.replicas` defaulting to `1`, `status.readyReplicas`;
- DaemonSets: `status.desiredNumberScheduled`,
  `status.numberReady`;
- Jobs: `spec.parallelism` defaulting to `1`, `status.succeeded`;
- CronJobs: `1` desired replica when not suspended, `0` when suspended, and `0`
  ready replicas.

These values are inventory summaries only. Health interpretation comes later.

## 4. Critical workload value

`critical` remains `UNKNOWN` in this package. Configured or labeled critical
workload policy belongs to later health/recommendation packages.

## 5. Determinism and safety

Workloads are sorted by:

1. namespace;
2. kind;
3. name.

Containers are sorted by name. Unknown or unparsable image tags leave `imageTag`
empty rather than guessing.

Fixtures must use sanitized names and images only. Real workload names, private
registry paths, raw images from a live cluster, labels, annotations, and UIDs
must not be committed.

## 6. Limitations

When workload collection is intentionally absent or not yet approved for live
use, KUA must not imply that the cluster has no workloads. The P2-02
`PARTIAL_INVENTORY_P2_02` limitation is replaced or supplemented by a P2-03
limitation that states which inventory groups remain intentionally uncollected.

Collection failure for any required workload API group makes the affected command
fail safely instead of emitting misleading partial workload data.

## 7. Gate B expansion

P2-02 Gate B passed only for namespace/node collection. Live workload collection
requires a separate Gate B expansion approval naming the context and allowing the
specific read-only workload API operations. Until then, automated tests use fake
clients only.

## 8. Fixture and validation expectations

At least one P2-03 golden fixture must include representative sanitized workload
records for every supported controller kind. The fixture must continue to use
empty arrays for storage, networking, CRDs, and events until those collectors are
implemented and approved.

The dependency-free snapshot subset validator must be expanded to cover workload
required fields, `critical` enum values, non-negative replica counts, and
container required fields before any workload snapshot path is considered
complete.
