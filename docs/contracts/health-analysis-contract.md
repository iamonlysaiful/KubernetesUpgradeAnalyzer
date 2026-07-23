# Health analysis contract

Status: Phase 3 contract artifact
Last updated: 2026-07-23

This contract governs the MVP health analysis foundation. It consumes normalized
inventory snapshots and produces deterministic health findings. It does not
approve new live Kubernetes reads.

## 1. Scope

Phase 3 health analysis covers:

- node readiness and pressure;
- node kubelet version skew against the observed server version;
- workload availability for supported controller summaries;
- DaemonSet coverage through workload readiness summaries;
- PVC/PV/StorageClass metadata health where represented by the snapshot schema;
- Warning and unknown-type events within an approved lookback window;
- explicitly configured or labeled critical workload policy when the later
  configuration surface exists.

Phase 3 does not evaluate Kubernetes API removals, component compatibility,
provider upgrade availability, kubent output, final readiness/risk, report
rendering, or redaction. Those belong to later phases.

## 2. Inputs

The rule engine consumes `inventory.Snapshot` values. Tests use synthetic
fixtures and fake-client generated snapshots only.

Current live `kua inventory --format=json` remains core inventory only. Health
rules that depend on expanded inventory groups must be tested against fixtures
until Gate B is separately expanded.

## 3. Finding model

Each health finding must include:

- stable rule ID;
- severity: `BLOCKER`, `WARNING`, or `INFO`;
- status: `FAIL`, `WARN`, `PASS`, or `UNKNOWN`;
- resource reference when applicable;
- concise summary;
- evidence details that do not contain secrets or raw event messages.

`UNKNOWN` is used when required evidence is absent. Missing evidence must not be
treated as `PASS`.

## 4. Initial rule set

The first MVP rule set is:

| Rule ID | Intent | Initial result behavior |
| --- | --- | --- |
| `health.node.notReady` | Node `Ready` is not true | `BLOCKER`/`FAIL` |
| `health.node.pressure` | Memory, disk, or PID pressure is true | `WARNING`/`WARN` |
| `health.node.kubeletSkew` | Node kubelet minor differs from server minor | `WARNING`/`WARN` initially |
| `health.workload.unavailable` | Ready replicas are lower than desired | `WARNING`/`WARN`, promoted later for critical workloads |
| `health.storage.pvcUnknown` | Storage evidence is absent where required | `UNKNOWN` |
| `health.event.warning` | Warning event in lookback window | `WARNING`/`WARN` |
| `health.event.unknownType` | Event type cannot be normalized | `INFO`/`UNKNOWN` |

Critical workload blocker policy remains deferred until configuration and label
contracts exist. Until then, workload availability is warning-level only.

## 5. Event window

The default event lookback is 30 minutes. A rule evaluates only events whose
`lastSeenAt` falls within the configured window relative to an injected clock.

Raw event messages are never required for Phase 3 rules.

## 6. Output boundary

Phase 3 may add internal health result structs and tests. User-facing `kua
health`, `kua analyze`, and final report rendering can be introduced only in
separately scoped implementation slices.

## 7. PR strategy

To keep the MVP near 30 total PRs, Phase 3 should be delivered in about four
reviewable PRs:

1. health plan and contract;
2. health finding/rule runner foundation;
3. node and workload rules;
4. storage and event rules plus Phase 3 closeout.

Splitting is still allowed if a PR becomes hard to review or validation risk
increases.
