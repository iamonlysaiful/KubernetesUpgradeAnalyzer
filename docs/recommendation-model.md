# Recommendation model

Status: Proposed for implementation  
Last updated: 2026-07-21

## 1. Design goals

Recommendations must be conservative, deterministic, explainable, and based on recorded evidence. The engine selects a destination and a sequence of valid transitions; it does not execute upgrades.

## 2. Terminology

- **Candidate**: a possible destination Kubernetes version.
- **Destination**: the highest suitable candidate selected by policy.
- **Upgrade stage**: one provider-supported transition, normally one minor version.
- **Path**: ordered stages from current version to destination.
- **Blocker**: evidence that makes a candidate/path unsafe or unsupported.
- **Limitation**: missing or uncertain evidence.

## 3. Candidate evaluation order

1. Normalize current control-plane and node-pool versions.
2. Obtain candidates from fresh provider evidence when available; otherwise use qualified catalog evidence.
3. Build only provider-valid edges. For supported AKS clusters, minor versions are sequential.
4. Evaluate removed/deprecated APIs at every stage.
5. Evaluate component support at every stage.
6. Evaluate health and operational prerequisites.
7. Reject blocked candidates and paths.
8. Choose the highest candidate allowed by configured policy and evidence confidence.
9. Return the destination plus every required stage and its findings.

## 4. Readiness states

| State | Meaning |
| --- | --- |
| `READY` | No blockers, no material warnings, and required evidence is sufficient |
| `READY_WITH_WARNINGS` | No blockers; warnings require review or post-upgrade validation |
| `NOT_READY` | One or more blockers prevent the proposed path |
| `INCONCLUSIVE` | Required evidence is absent, stale, failed, or ambiguous |

## 5. Initial blocker policy

Blockers include:

- an API removed at any proposed stage and still used by a live object;
- explicit component incompatibility with any stage;
- unavailable or invalid provider transition;
- unready node, unresolved node pressure, unbound required PVC, unavailable critical workload, or persistent fatal pod condition when policy classifies it as upgrade unsafe;
- failed required analyzer or corrupt compatibility catalog.

Deprecations scheduled after the destination are warnings. Unknown component versions, missing compatibility records, incomplete RBAC, and stale provider evidence generally make the relevant claim inconclusive rather than compatible.

## 6. Risk model

Risk is rule-based, not an unexplained arithmetic score:

- `HIGH`: any blocker or explicitly unsupported condition.
- `UNKNOWN`: required evidence is insufficient to bound risk.
- `MEDIUM`: no blocker, but material warnings, conditional support, or operational concerns exist.
- `LOW`: all required analyzers pass with adequate evidence and only informational findings remain.

The report includes the rules that determined the result. Future weighted scoring requires a separate approved ADR and calibration dataset.

## 7. AKS 1.30 validation case

Given sanitized evidence:

- current version `1.30.0`;
- provider evidence supports destination `1.33.12` and intervening upgrades;
- no deprecated/removed API blockers;
- NGINX Ingress `1.12.1`, EMQX `5.8.8`, Fluent Bit `4.0.3`, managed CoreDNS/Metrics Server, and Azure CSI components have sufficient compatible evidence;
- healthy workloads and storage;

Expected output:

- destination: `1.33.12`;
- path: `1.30.x → 1.31.x → 1.32.x → 1.33.12`, with exact intermediate patches determined by provider evidence;
- readiness: `READY`;
- risk: `LOW`;
- recommendation: execute each supported stage and perform post-stage/application smoke tests.

This does not assert that AKS supports one direct `1.30 → 1.33.12` operation.

## 8. Authoritative upgrade constraints

- Kubernetes version-skew policy states that kube-apiserver minor versions must not be skipped: <https://kubernetes.io/releases/version-skew-policy/>
- AKS states that supported cluster upgrades must proceed sequentially by minor version: <https://learn.microsoft.com/azure/aks/upgrade-aks-control-plane>
- AKS support/version policy and available versions remain provider-controlled: <https://learn.microsoft.com/azure/aks/supported-kubernetes-versions>

