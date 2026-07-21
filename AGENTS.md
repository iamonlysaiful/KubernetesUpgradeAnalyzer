# KubeUpgrade Advisor Agent Rules

This file is the mandatory operating contract for every human or automated agent working in this repository. It applies to the entire repository unless a deeper `AGENTS.md` adds stricter rules.

## 1. Authority and approvals

- The user/project owner is the final authority for scope, architecture, requirements, risk acceptance, and implementation.
- Explicit user approval is required before any state-changing action, including source changes, dependency changes, generated artifacts, destructive commands, releases, deployments, or external writes.
- A user request that clearly asks for a particular change counts as approval for that stated scope only. Do not expand it by assumption.
- Read-only inspection, analysis, and verification may proceed without separate approval when needed to answer the user.
- If requirements are ambiguous or conflict with the source-of-truth documents, stop and ask the user. Do not silently choose a new product direction.

## 2. Documentation is the source of truth

- Treat the approved files in `docs/` as the product and engineering source of truth.
- Before proposing or making a change, inspect `AGENTS.md`, `docs/README.md`, and every document relevant to the affected area.
- Every change must be traceable to an approved requirement, architecture decision, roadmap item, or change record.
- Code, tests, examples, CLI help, schemas, and reports must agree with the documentation. A discrepancy is a defect; do not resolve it by silently changing either side.
- Record assumptions and unresolved decisions explicitly. Never present an unapproved proposal as an accepted decision.

## 3. Required docs-first change workflow

For every product, architecture, behavior, interface, dependency, or operational change:

1. Read the governing documentation and inspect the current implementation.
2. Identify the required documentation changes and affected artifacts.
3. Present the proposed documentation change, impact, risks, and open questions to the user.
4. Obtain explicit user approval.
5. Update the relevant source-of-truth documents first.
6. Validate internal links, terminology, requirements, and decision status.
7. Commit the documentation update as its own focused commit.
8. Only then begin implementation.
9. Add or update tests and supporting material with the implementation.
10. Verify the result against the approved documentation.
11. Commit the implementation as a separate focused commit.
12. Report commit identifiers, validation performed, known limitations, and any follow-up decisions.

Do not combine the governing documentation change and its implementation into one commit. Documentation-only corrections that do not alter requirements or behavior may use one documentation commit after user approval.

## 4. Git and change discipline

- Keep commits small, cohesive, and reviewable.
- Use imperative commit subjects and explain why in the body when the reason is not obvious.
- Do not rewrite, squash, amend, reset, force-push, or delete user history without explicit approval.
- Preserve unrelated user changes. Never discard or overwrite work that is not part of the approved scope.
- Do not claim a commit exists unless Git confirms it.
- Do not commit secrets, kubeconfigs, credentials, cluster dumps, identifiable production data, or local machine artifacts.

## 5. Safety and Kubernetes constraints

- KUA itself must remain read-only against Kubernetes and provider APIs unless a future, separately approved design explicitly changes that boundary.
- Prefer least-privilege access and document every required permission.
- Never collect Kubernetes Secrets or secret payloads. Redact tokens, certificates, registry credentials, environment values, and sensitive annotations from logs, fixtures, and reports.
- Offline is the default. No runtime outbound network access is permitted unless the user explicitly enables an approved feature that documents destinations, data sent, failure behavior, and auditability.
- Never run a command against a live cluster unless the user explicitly approves the cluster and operation. Confirm the active context before any approved cluster command.
- Use synthetic or sanitized fixtures for tests. Real cluster data requires explicit approval and documented sanitization.

## 6. Engineering quality gates

- Follow the architecture, contracts, compatibility-data rules, and testing strategy documented under `docs/`.
- New behavior requires appropriate automated tests. Bug fixes require a regression test when feasible.
- Errors and recommendations must be explainable and actionable; avoid opaque scores without evidence.
- Deterministic offline runs over the same input and catalog must produce equivalent findings and recommendations.
- Treat malformed, missing, stale, or conflicting compatibility data as explicit evidence limitations, not as silent compatibility success.
- Run formatting, unit tests, static analysis, and relevant integration tests before declaring implementation complete. Report any check not run and why.

## 7. Decision and change records

- Record significant architectural decisions as ADRs in `docs/decisions/`.
- Track unresolved product choices in `docs/open-questions.md` with an owner and status.
- Update `docs/roadmap.md` only after user approval; roadmap placement is not implementation authorization.
- Material scope changes require an entry in `docs/change-log.md` that links to the relevant requirement or ADR.

## 8. Communication requirements

- Lead with outcomes, risks, and decisions needed from the user.
- Clearly label facts, proposals, assumptions, and accepted decisions.
- Before implementation, summarize what will change, what will not change, and how it will be verified.
- If blocked, state the exact blocker and the smallest decision or authorization needed.

