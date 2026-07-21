# Development process

Status: Accepted  
Last updated: 2026-07-21

## 1. Change lifecycle

All contributors follow `AGENTS.md`:

1. establish an approved need;
2. inspect source-of-truth docs and code;
3. propose doc/ADR changes and impact;
4. obtain user approval;
5. update and commit docs;
6. implement with tests;
7. verify against docs;
8. commit implementation separately;
9. report results and residual risk.

Architecture documents can contain proposals, but a proposal is not implementation permission.

## 2. Definition of ready

A change is ready for implementation only when scope and non-goals are clear, relevant questions are resolved or explicitly deferred, acceptance tests are defined, security/privacy impact is assessed, and the documentation commit exists.

## 3. Definition of done

Implementation is done when behavior matches approved docs, tests and quality gates pass, sensitive data review passes, reports/errors are actionable, no unrelated changes are included, implementation has its own commit, and limitations are reported.

## 4. Proposed commit style

Use imperative subjects, for example:

- `Document component detection confidence rules`
- `Add Kubernetes inventory collector`
- `Block candidates with removed APIs`

Documentation commits precede and remain separate from implementation commits.

## 5. Dependency changes

Any new or upgraded dependency requires documented purpose, alternatives, license, maintenance/security assessment, version rationale, user approval, and checksum updates. The planned baseline is Go, Cobra, `slog`, `client-go`, and Viper; listing these does not approve installation yet.

## 6. Live-system work

Real cluster access always needs explicit approval naming the context and read-only operation. Record commands and sanitization approach. Never commit kubeconfigs or raw cluster output.

