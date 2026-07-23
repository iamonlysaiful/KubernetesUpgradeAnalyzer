# API compatibility contract

Status: Phase 5 contract artifact
Last updated: 2026-07-24

This contract governs the MVP Kubernetes API compatibility foundation. It uses a
controlled external kubent adapter first and records a go/no-go decision before
any recommendation logic can rely on API compatibility findings.

## 1. Scope

Phase 5 API compatibility covers:

- a process adapter for an installed kubent binary;
- kubent version discovery and validation for the approved MVP version;
- controlled argument construction without shell interpolation;
- JSON output parsing into internal API compatibility findings;
- target-version rule coverage validation for assessed Kubernetes stages;
- negative-path handling for missing tools, wrong versions, malformed output,
  execution errors, timeouts, oversized output, and missing target rules.

Phase 5 does not produce final readiness, risk, destination, provider evidence,
component compatibility decisions, reports, or live-cluster validation records.

## 2. Approved kubent boundary

The MVP external adapter targets kubent `0.7.3`.

Invocation must:

- pass arguments without a shell;
- request JSON output;
- disable Helm collection with `--helm3=false`;
- avoid reading Kubernetes Secrets or ConfigMap contents through Helm release
  collection;
- bound runtime and output size;
- capture stdout and stderr separately;
- treat stderr and execution diagnostics as untrusted input;
- redact diagnostics before user-facing output.

No Phase 5 implementation may run kubent against a live cluster without a
separate user approval naming the context and exact command.

## 3. Output model

Normalized API compatibility evidence must include:

- analyzer name and version;
- target Kubernetes version or minor;
- finding status: `FAIL`, `WARN`, `PASS`, or `UNKNOWN`;
- affected resource reference when available;
- removed or deprecated API version/kind;
- replacement when reported;
- source rule or target coverage evidence when available;
- limitations for incomplete or inconclusive evidence.

Missing evidence, missing rules, unsupported targets, empty unverified output,
and malformed kubent data must produce `UNKNOWN` or an explicit error. They must
not produce `PASS`.

## 4. Target coverage

KUA must verify kubent rule coverage for each assessed target stage before that
stage's API compatibility result can be trusted.

If a target stage lacks verified rule coverage, that stage is `INCONCLUSIVE` for
API compatibility, even when kubent returns no findings.

## 5. Test strategy

Phase 5 tests use process fakes and static JSON fixtures. Tests must cover:

- successful version validation;
- missing kubent binary;
- wrong kubent version;
- malformed JSON;
- nonzero exit code;
- timeout or bounded execution failure;
- oversized output;
- missing target coverage;
- empty output with verified coverage;
- finding normalization for removed/deprecated APIs.

Live kubent execution remains deferred until an explicit validation plan and
user approval exist.

## 6. Native analyzer decision

Phase 5 must end with a documented go/no-go decision:

- continue with kubent adapter for MVP when `1.30` through `1.33` target
  coverage is verified; or
- add a minimal native API analyzer before MVP if kubent coverage is inadequate.
