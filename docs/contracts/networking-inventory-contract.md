# Networking inventory collector contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract governs the networking portion of P2-03. It depends on P2-02 core
inventory and the earlier P2-03 workload/CRD fixture paths. It does not approve
live networking collection by itself.

## 1. Scope

This package adds fake-client-first networking snapshot collection for:

- Services;
- Ingresses.

The collected networking records populate `inventory.networking` in
`schemas/cluster-snapshot/v1.json`.

This package may add an internal snapshot assembly path that includes networking
resources for fake-client and golden-fixture tests. `kua inventory --format=json`
must not emit live networking records until Gate B is separately expanded for
networking reads.

This package does not collect EndpointSlices, Endpoints, Gateway API resources,
load balancer IPs, hostnames, Service account tokens, annotations, labels, raw
UIDs, cloud provider IDs, storage, events, health findings, compatibility
findings, provider evidence, recommendations, or reports.

## 2. Required fields

Each networking record is represented as a `ResourceRef` and includes:

- `apiVersion`;
- `kind`;
- `namespace`;
- `name`.

Supported `kind` values are:

- `Service` with `apiVersion` `v1`;
- `Ingress` with `apiVersion` `networking.k8s.io/v1`.

The MVP snapshot schema currently models networking resources as `ResourceRef`
values only. Service type, ports, selectors, cluster IPs, external IPs, ingress
hosts, TLS settings, ingress class, backend details, and status are intentionally
deferred until a schema expansion is approved.

## 3. Determinism and safety

Networking resources are sorted by:

1. namespace;
2. kind;
3. name.

Fixtures must use sanitized names only. Real namespaces, service names, ingress
hostnames, private DNS zones, IP addresses, cloud load balancer identifiers,
annotations, labels, and UIDs must not be committed.

## 4. Limitations

When networking collection is intentionally absent or not yet approved for live
use, KUA must not imply that the cluster has no networking resources. P2-03
fixture paths that include networking must use a limitation that names the
inventory groups still intentionally uncollected.

Collection failure for any required networking API group makes the affected
command fail safely instead of emitting misleading partial networking data.

## 5. Gate B expansion

P2-02 Gate B passed only for namespace/node collection. Live workload, CRD, and
networking collection require separate Gate B expansion approval naming the
context and allowing the specific read-only API operations. Until then,
automated tests use fake clients only.

## 6. Fixture and validation expectations

At least one P2-03 golden fixture must include representative sanitized Service
and Ingress refs. The fixture must continue to use empty arrays for storage and
events until those collectors are implemented and approved.

The dependency-free snapshot subset validator must be expanded to cover
networking required fields and allowed kind/apiVersion pairs before any
networking snapshot path is considered complete.
