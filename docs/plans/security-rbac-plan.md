# Security and RBAC plan

Status: Phase 0 contract in progress
Last updated: 2026-07-22

## 1. Security objectives

- Read-only Kubernetes and Azure behavior.
- No Secret payload collection.
- Least privilege and explicit partial-evidence reporting.
- No shell interpolation, implicit login, telemetry, runtime scraping, or catalog search.
- Safe local and redacted shareable reports.

## 2. Kubernetes permission design

Create an auditable resource/verb matrix before writing the ClusterRole. Expected verbs are `get` and `list`; `watch` is excluded unless separately justified. Candidate resources cover discovery plus nodes, namespaces, workloads, pods, services, ingresses, PVCs, StorageClasses, CSIDrivers, CRDs, and events. Secrets are explicitly absent.

The preflight maps each permission to required or optional evidence. Required denial prevents the affected command; optional denial emits `UNKNOWN` limitations. Normal setup never requests `cluster-admin`.

## 3. Kubent boundary

- Support/test `0.7.3` initially.
- Invoke directly with JSON output, selected kubeconfig/context/target, bounded timeout/output, and `--helm3=false`.
- Never enable Helm collectors because they may read Helm releases from Secrets or ConfigMaps.
- Validate version and target-rule coverage before interpreting empty output as no findings.
- Redact captured stderr and never log credentials or full manifests.

## 4. Azure CLI boundary

- Allowlist read-only identity/account inspection required to resolve the target and `az aks get-upgrades` only.
- Use explicit subscription/resource-group/cluster arguments when resolved; never execute `az account set`.
- Never initiate `az login`, browser/device flow, token export, write commands, extensions, or arbitrary query supplied as shell text.
- Invoke without a shell, bound runtime/output, parse JSON, sanitize diagnostics, and record CLI version/evidence time.
- `offline` mode must prove no Azure CLI process is invoked.

## 5. Data classification and redaction

Classify fields as public product metadata, operational identifiers, sensitive operational data, or prohibited secrets. Redacted mode replaces operational identifiers with stable assessment-local aliases. Tests cover nested evidence, errors, events, URLs, image registries, IDs, rendered HTML/Markdown, and logs.

## 6. Supply-chain baseline

Pin modules and tools, retain Go checksums, review licenses, enable dependency and secret scanning, generate an SBOM for releases, publish artifact checksums/provenance, and document reproducible build inputs. Catalog sources/checksums are part of release provenance.

## 7. Verification and incident behavior

Required checks include malicious fixture inputs, output/path traversal, symlink/overwrite policy, command-argument injection, oversized output, cancellation/timeouts, RBAC denial, redaction leaks, unsupported schemas, and catalog tampering. A suspected secret leak stops artifact sharing and release work until the owner reviews containment and rotation needs.

## 8. Exit criteria

- Exact RBAC matrix and manifest approved.
- Azure and kubent argument allowlists approved and tested.
- Redaction field matrix and prohibited-data tests approved.
- Threat cases map to prevention/detection behavior.
- No planned MVP behavior requires cluster/provider mutation or Secret reads.

The Phase 0 contract artifact is `docs/contracts/security-rbac-contract.md`.
