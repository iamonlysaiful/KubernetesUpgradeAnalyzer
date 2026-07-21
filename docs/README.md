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
| [`open-questions.md`](open-questions.md) | Decisions still requiring owner approval |
| [`change-log.md`](change-log.md) | Material documentation and scope changes |

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
