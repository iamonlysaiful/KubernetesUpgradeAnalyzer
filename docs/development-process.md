# Development process

Status: Accepted  
Last updated: 2026-07-22

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

## 7. Git hygiene and publication

- `main` is the primary branch. Use `feature/`, `fix/`, `docs/`, or `chore/` prefixes for short-lived branches.
- Review `git status`, the staged diff, and `git diff --check` before every commit.
- Publishing, pushing, changing upstreams, tagging, releasing, rewriting history, and force-pushing require explicit user approval.
- Before first publication, verify the remote URL and default branch, run applicable quality gates, confirm the working tree is clean, and run `git fsck`.
- Use a repository-local real author identity and verified email for human commits. Clearly identify approved automation commits.
- Never commit credentials, kubeconfigs, raw provider exports, unsanitized cluster data, local overrides, generated reports, or machine-specific files.
- macOS `._*` AppleDouble files must not be tracked. Prefer APFS for active Git working copies. On ExFAT, verify that AppleDouble files have not entered `.git`; `.gitignore` cannot protect Git's internal directory.
- If repository integrity checks report invalid refs or bad objects, stop publication. Diagnose and obtain approval before deleting metadata, repairing refs, recloning, or rewriting history.

## 8. Ignore policy

The root `.gitignore` is the canonical ignore policy. Keep it narrow enough that source, schemas, catalog records, fixtures, examples, and documentation remain visible. Secrets are ignored only as a last line of defense; contributors remain responsible for never creating or staging sensitive material in the repository.

## 9. Recoverable cleanup procedure

All cleanup, deletion, overwrite, repository repair, and other destructive operations follow this sequence:

1. Resolve and list the exact targets; do not rely on broad or unresolved paths.
2. Record the relevant pre-change state, checksums, references, or integrity results.
3. Create a backup, reversible copy, or trash-based recovery point in an approved safe location.
4. Verify that the recovery artifact exists and is readable; record its checksum when practical.
5. Remove or modify only the approved targets.
6. Run domain-appropriate integrity checks and compare critical state with the pre-change record.
7. Report what changed, validation results, recovery location, and restoration method.
8. Retain the recovery artifact until the user gives separate explicit approval to remove it.

Successful validation does not authorize backup deletion. If validation fails, stop, preserve evidence and the recovery artifact, and request direction before attempting additional repair or restoration.
