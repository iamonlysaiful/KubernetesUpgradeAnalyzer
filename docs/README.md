# Documentation index

The documents in this directory are the source of truth for KubeUpgrade Advisor.

## Governing documents

| Document | Purpose |
| --- | --- |
| [`product-requirements.md`](product-requirements.md) | Scope, actors, requirements, and acceptance criteria |
| [`architecture.md`](architecture.md) | Components, data flow, boundaries, and contracts |
| [`compatibility-catalog.md`](compatibility-catalog.md) | Offline compatibility data model and provenance rules |
| [`recommendation-model.md`](recommendation-model.md) | Findings, blockers, risk, and target selection |
| [`cli-and-reports.md`](cli-and-reports.md) | CLI surface, exit codes, and report contracts |
| [`security-and-privacy.md`](security-and-privacy.md) | Threat model, RBAC, offline behavior, and redaction |
| [`testing-strategy.md`](testing-strategy.md) | Test levels, fixtures, determinism, and release gates |
| [`development-process.md`](development-process.md) | Docs-first delivery and review workflow |
| [`roadmap.md`](roadmap.md) | Approved phases and delivery gates |
| [`status.md`](status.md) | Current phase, gate, branch, and quality status |
| [`open-questions.md`](open-questions.md) | Decisions still requiring owner approval |
| [`change-log.md`](change-log.md) | Material documentation and scope changes |

## Implementation plans

| Document | Purpose |
| --- | --- |
| [`plans/mvp-implementation-plan.md`](plans/mvp-implementation-plan.md) | Ordered MVP work packages, dependencies, and approval gates |
| [`plans/phase-3-health-plan.md`](plans/phase-3-health-plan.md) | Phase 3 health analysis plan |
| [`contracts/foundation-contract.md`](contracts/foundation-contract.md) | Phase 1 module, CLI skeleton, version, and dependency boundary |
| [`contracts/config-error-contract.md`](contracts/config-error-contract.md) | Phase 1 config, logging, command-error, and exit-code foundation |
| [`contracts/ci-quality-contract.md`](contracts/ci-quality-contract.md) | Phase 1 CI workflow and local quality gate contract |
| [`contracts/kube-preflight-contract.md`](contracts/kube-preflight-contract.md) | Phase 2 kubeconfig/context and read-only preflight contract |
| [`contracts/core-inventory-contract.md`](contracts/core-inventory-contract.md) | Phase 2 core inventory and partial snapshot contract |
| [`contracts/workload-inventory-contract.md`](contracts/workload-inventory-contract.md) | Phase 2 workload inventory contract |
| [`contracts/crd-inventory-contract.md`](contracts/crd-inventory-contract.md) | Phase 2 CRD inventory contract |
| [`contracts/networking-inventory-contract.md`](contracts/networking-inventory-contract.md) | Phase 2 networking inventory contract |
| [`contracts/storage-inventory-contract.md`](contracts/storage-inventory-contract.md) | Phase 2 storage inventory contract |
| [`contracts/events-inventory-contract.md`](contracts/events-inventory-contract.md) | Phase 2 events inventory contract |
| [`contracts/inventory-snapshot-assembly-contract.md`](contracts/inventory-snapshot-assembly-contract.md) | Phase 2 inventory snapshot assembly consolidation contract |
| [`contracts/health-analysis-contract.md`](contracts/health-analysis-contract.md) | Phase 3 health analysis contract |
| [`contracts/domain-and-schema-plan.md`](contracts/domain-and-schema-plan.md) | Domain types, schemas, compatibility, and versioning plan |
| [`contracts/security-rbac-contract.md`](contracts/security-rbac-contract.md) | Phase 0 read-only RBAC, external-command, redaction, and dependency contract |
| [`plans/security-rbac-plan.md`](plans/security-rbac-plan.md) | Least-privilege Kubernetes/Azure access and security validation |
| [`plans/gate-b-smoke-test-plan.md`](plans/gate-b-smoke-test-plan.md) | Proposed Gate B live read-only smoke-test approval plan |
| [`plans/validation-release-plan.md`](plans/validation-release-plan.md) | Staging validation, artifacts, release checks, and rollback |
| [`validation/gate-b-p2-02-record.md`](validation/gate-b-p2-02-record.md) | Draft Gate B P2-02 validation record |
| [`validation/phase-2-closeout.md`](validation/phase-2-closeout.md) | Draft Phase 2 closeout record |

## Schema contracts

| Path | Purpose |
| --- | --- |
| [`../schemas/README.md`](../schemas/README.md) | Schema contract index, versioning rules, and fixture classes |
| [`../schemas/assessment/v1.json`](../schemas/assessment/v1.json) | Canonical assessment JSON schema |
| [`../schemas/cluster-snapshot/v1.json`](../schemas/cluster-snapshot/v1.json) | Normalized cluster snapshot schema |
| [`../schemas/provider-evidence/aks-v1.json`](../schemas/provider-evidence/aks-v1.json) | AKS provider evidence schema |
| [`../schemas/catalog/v1.json`](../schemas/catalog/v1.json) | Embedded catalog schema |

## Dependency assessments

| Document | Purpose |
| --- | --- |
| [`dependencies/client-go.md`](dependencies/client-go.md) | Kubernetes client-go dependency assessment for P2-01 |

## Architecture decisions

- [`ADR-0001`](decisions/0001-offline-first.md): Original offline-first execution decision (superseded)
- [`ADR-0002`](decisions/0002-aks-first-provider.md): AKS-first provider scope
- [`ADR-0003`](decisions/0003-kubent-mvp-adapter.md): External kubent adapter for MVP
- [`ADR-0004`](decisions/0004-destination-and-upgrade-path.md): Separate destination from sequential path
- [`ADR-0005`](decisions/0005-local-first-provider-evidence.md): Local-first provider evidence and catalog lifecycle

## Document states

- **Accepted**: approved and binding.
- **Proposed**: designed but requires explicit approval before implementation.
- **Open**: a decision is still required.
- **Superseded**: retained for history and linked to its replacement.

When documents disagree, accepted ADRs take precedence for the decision they cover, then product requirements, then architecture and subordinate contracts. Any conflict must be brought to the user rather than resolved silently.
