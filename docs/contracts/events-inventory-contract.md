# Events inventory collector contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract governs the events portion of P2-03. It depends on P2-02 core
inventory and the earlier P2-03 workload, CRD, networking, and storage fixture
paths. It does not approve live event collection by itself.

## 1. Scope

This package adds fake-client-first event snapshot collection from core
Kubernetes Events.

The collected event records populate `inventory.events` in
`schemas/cluster-snapshot/v1.json`.

This package may add an internal snapshot assembly path that includes events for
fake-client and golden-fixture tests. `kua inventory --format=json` must not emit
live event records until Gate B is separately expanded for event reads.

This package does not collect raw event messages, involved-object UIDs,
annotations, labels, managed fields, source hostnames, reporting instance IDs,
Secret data, health findings, compatibility findings, provider evidence,
recommendations, or reports.

## 2. Required fields

Each event record includes:

- `ref.apiVersion`;
- `ref.kind`;
- `ref.namespace` when the involved object is namespaced;
- `ref.name`;
- `type`;
- `reason`;
- `lastSeenAt`.

Supported `type` values are normalized to:

- `NORMAL`;
- `WARNING`;
- `UNKNOWN`.

`lastSeenAt` is the best available event timestamp, normalized to RFC3339 UTC.

## 3. Message handling

Raw event messages may contain workload names, image names, hostnames, volume
details, cloud resource IDs, or other sensitive operational context. This package
must not store raw event messages in fixtures or snapshot records.

The schema permits `messageAlias`, but this package leaves it empty. Stable
message aliasing belongs to a later redaction/reports package.

## 4. Determinism and safety

Events are sorted by:

1. `lastSeenAt`;
2. namespace;
3. kind;
4. name;
5. reason;
6. type.

Fixtures must use sanitized object names and reasons only. Real event messages,
private hostnames, image paths, storage details, cloud IDs, and UIDs must not be
committed.

## 5. Limitations

When event collection is intentionally absent or not yet approved for live use,
KUA must not imply that the cluster has no events. P2-03 fixture paths that
include events must use a limitation that states the inventory is still
fake-client-only until Gate B is expanded.

Collection failure for events makes the affected command fail safely instead of
emitting misleading partial event data.

## 6. Gate B expansion

P2-02 Gate B passed only for namespace/node collection. Live workload, CRD,
networking, storage, and event collection require separate Gate B expansion
approval naming the context and allowing the specific read-only API operations.
Until then, automated tests use fake clients only.

## 7. Fixture and validation expectations

At least one P2-03 golden fixture must include representative sanitized Normal,
Warning, and unknown-type event records.

The dependency-free snapshot subset validator must be expanded to cover event
required fields, type enum values, and RFC3339 `lastSeenAt` before any event
snapshot path is considered complete.
