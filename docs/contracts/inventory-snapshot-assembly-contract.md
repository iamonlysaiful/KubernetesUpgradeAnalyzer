# Inventory snapshot assembly contract

Status: Phase 2 consolidation contract
Last updated: 2026-07-23

This contract governs consolidation of the Phase 2 fake-client inventory snapshot
assembly paths. It is a refactor contract, not approval for new live reads.

## 1. Scope

The current P2-03 implementation has separate test-only methods for each
collector milestone. This package consolidates those paths into an explicit
collector option model so future inventory groups can be selected without adding
long method names.

The consolidation may replace methods such as:

- `CollectSnapshotWithWorkloads`;
- `CollectSnapshotWithWorkloadsAndCRDs`;
- `CollectSnapshotWithWorkloadsCRDsAndNetworking`;
- `CollectSnapshotWithWorkloadsCRDsNetworkingAndStorage`;
- `CollectSnapshotWithFullFakeInventory`.

with a single internal snapshot method that accepts approved collection options.

`CollectCore` remains the only method used by live `kua inventory --format=json`
until Gate B is separately expanded.

## 2. Required behavior

The consolidated assembly path must:

- preserve every existing golden fixture byte-for-byte;
- preserve existing validation behavior;
- preserve existing fake-client collector tests;
- keep live CLI behavior unchanged;
- make unsupported option combinations fail safely when required clients are
  missing;
- keep limitations explicit about intentionally uncollected inventory groups.

## 3. Option model

Options are boolean feature flags for inventory groups:

- workloads;
- storage;
- networking;
- CRDs;
- events.

Namespace and node collection remain mandatory for every snapshot path.

CRD collection requires an apiextensions client. Other currently implemented
groups use the Kubernetes client.

## 4. Safety boundaries

This refactor must not:

- add live workload, storage, networking, CRD, or event reads;
- alter `internal/app` inventory collector interfaces;
- add dependencies;
- change schemas or fixture semantics;
- copy raw event messages or sensitive resource details.

## 5. Verification expectations

The implementation must run:

```text
scripts/ci-local.sh
git diff --check
git fsck --full --strict
```

`git fsck` may continue to report only the accepted known dangling blobs.
