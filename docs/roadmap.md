# Roadmap

Status: Proposed sequencing; implementation requires approval per phase  
Last updated: 2026-07-21

## Phase 0: foundation

- Approve architecture and open questions.
- Initialize Go module and CI quality gates.
- Define domain models, JSON schema, catalog schema, and sanitized fixtures.
- Establish security/RBAC baseline.

Exit: approved contracts and passing skeleton quality gates.

## Phase 1A: inventory and health

- Live Kubernetes preflight and read-only collectors.
- Inventory and health analyzers.
- Console and JSON renderers.

Exit: deterministic inventory/health assessment with partial-RBAC handling.

## Phase 1B: compatibility

- Component detector framework and initial detector set.
- Version normalization and bundled component catalog.
- Controlled installed-kubent adapter.
- API and component findings.

Exit: compatibility report with evidence/confidence and no silent unknown passes.

## Phase 1C: AKS recommendations and reports

- AKS policy/provider evidence adapter.
- Candidate/path and recommendation engine.
- Markdown and self-contained HTML.
- Sanitized `1.30 → 1.33.12` destination validation scenario.

Exit: all MVP acceptance criteria pass.

## Phase 2

- Native API analyzer, initially in comparison mode with kubent.
- Helm/local manifest analysis.
- CRD and operator compatibility depth.
- Upgrade simulation.
- EKS, GKE, and OpenShift adapters.
- AKS best-practice checks, dashboard, and GitHub Action integration.

## Phase 3

- Historical comparison and drift.
- Continuous monitoring.
- Security posture and cost impact.
- AI-assisted explanations subject to offline/privacy design.
- VS Code extension.

