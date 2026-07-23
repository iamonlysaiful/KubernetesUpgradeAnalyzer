# Component detection contract

Status: Phase 4 contract artifact
Last updated: 2026-07-23

This contract governs MVP component detection over normalized inventory
snapshots. It does not approve compatibility decisions or final upgrade
recommendations.

## 1. Scope

Phase 4 component detection covers:

- deterministic detectors that consume `inventory.Snapshot`;
- normalized detected component IDs, names, versions, and confidence;
- evidence references that avoid secrets and raw cluster identifiers;
- `UNKNOWN` outcomes for absent, ambiguous, or unsupported evidence;
- an initial detector cohort after the framework is validated.

Detectors must not call Kubernetes clients, provider CLIs, vendor websites, or
runtime network services. Live data still enters only through the approved
inventory snapshot boundary.

## 2. Detection result model

Each detection result must include:

- stable component ID when known;
- display name;
- version string when confidently detected;
- confidence: `HIGH`, `MEDIUM`, `LOW`, or `UNKNOWN`;
- status: `FOUND`, `NOT_FOUND`, or `UNKNOWN`;
- sanitized resource references used as evidence;
- limitations when evidence is partial or ambiguous.

Unknown component versions must produce `UNKNOWN`, never `PASS`.

## 3. Initial detector cohort

The first detector cohort may include:

- NGINX Ingress;
- CoreDNS;
- Metrics Server;
- Azure Disk CSI;
- Azure File CSI;
- Fluent Bit;
- EMQX.

The cohort may be split if fixtures or validation become too large for one PR.
Adding a detector does not imply a compatibility support claim exists. Detection
and compatibility policy remain separate.

## 4. Version extraction rules

Version extraction may use sanitized workload image tags, known component labels,
or catalog detector hints. If a version cannot be extracted confidently, the
detector must report the component with `UNKNOWN` version or `UNKNOWN` status as
appropriate.

Unbounded tags such as `latest`, mutable digests without tag context, missing
container images, or conflicting versions across replicas must not become a
confident supported result.

## 5. Determinism

Detection results must be sorted deterministically by component ID, namespace,
kind, resource name, and version. Equivalent snapshots must produce equivalent
results.

## 6. Live boundary

Phase 4 does not expand live Kubernetes collection. Detectors that require
workload, storage, or event evidence use synthetic or fake-client snapshots
until Gate B is separately expanded.
