# Project status

Last updated: 2026-08-04

This page summarizes the current implementation phase, review gates, and
publication state. Detailed history remains in Git and `docs/change-log.md`.

## Gate status

| Gate | Status | Evidence |
| --- | --- | --- |
| Gate A - contracts | Complete | Phase 0 contracts, schemas, fixtures, security/RBAC, and dependency rules are documented and merged. |
| Gate B - collection safety | Passed for P2-02 | Core inventory has fake-client/golden coverage and locally approved live smoke-test record. |
| Gate C - compatibility validity | Complete | Catalog foundation merged; kubent coverage validated for 1.30-1.34; go/no-go decision GO. |
| Gate D - recommendation calibration | Complete | Recommendation matrix, path evaluation, and report rendering foundation merged. |
| Gate E - release | Blocked | End-to-end `kua analyze` validated locally. Phase 10 complete; publication blocked pending owner approval. |

## Phase status

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 0 - Design freeze and contracts | Complete | Merged to `main`. |
| Phase 1 - CLI foundation | Complete | Merged to `main`. |
| Phase 2 - Kubernetes preflight and inventory | Complete | Core and expanded inventory merged; live core verified. |
| Phase 3 - Health analysis | Complete | Health rules merged. |
| Phase 4 - Component detection and catalog | Complete | Catalog loader, detector framework, initial cohort merged. |
| Phase 5 - API compatibility | Complete | Kubent adapter, coverage for 1.30-1.34, go/no-go GO merged. |
| Phase 6 - AKS provider evidence | Complete | Provider interface, AKS CLI/file adapters, path construction merged. |
| Phase 7 - Recommendation engine | Complete | Recommendation engine, policy evaluation merged. |
| Phase 8 - Reports and hardening | Complete | Multi-format rendering, redaction mode merged. |
| Phase 8.5 - End-to-end CLI recovery | Complete | Interactive `kua analyze` with component/target version prompts, console UX rewrite, and local install script merged to `main`. |
| Phase 9 - Staging validation and MVP release | Blocked | Staging validation passed locally. Blocked pending Phase 10 completion and owner approval. |
| Phase 10 - Advisor model transformation | Complete | Confidence scoring, traffic-light decisions, evidence summary, upgrade plan, version-specific gotchas, pre-flight/day-of modes. See ADR-0006. |

## Current branch focus

Phase 10 is complete. All 8 sub-phases are implemented and merged or pending final PR merge:

Transform KUA from "Kubernetes Upgrade Analyzer" to "Kubernetes Upgrade Advisor":

- **Phase 10.1**: Confidence model foundation — weighted scoring replaces binary pass/fail ✅ Merged
- **Phase 10.2**: Traffic light decisions — 🟢 GO / 🟡 CAUTION / 🔴 STOP ✅ Merged
- **Phase 10.3**: Evidence summary — "Why do I trust this?" ✅ Merged
- **Phase 10.4**: Upgrade plan generator — step-by-step checklist with time estimates ✅ Merged
- **Phase 10.5**: Finding enhancement — mandatory actions, impact, consequences ✅ Merged
- **Phase 10.6**: Console renderer overhaul — advisor output format ✅ Merged
- **Phase 10.7**: Version-specific gotchas — proactive breaking change warnings ✅ Merged
- **Phase 10.8**: Pre-flight/day-of modes — separate preparation from execution ✅ Merged (PR pending final merge)

Approved decisions:
- OQ-008: Start with heuristic weights; calibrate from real outcomes
- OQ-009: GO ≥90%, CAUTION 70–89%, STOP <70%; risk profiles in MVP
- OQ-010: Clean break to Schema 2.0.0; no legacy output mode

## What's merged to main

All phases through 8.5 and Phase 10 are complete and merged:

- Full `kua analyze` with interactive confirmation, component version prompts,
  target version prompts, and console report rendering
- Animated progress spinner during analysis phases
- Confidence scoring, traffic-light decisions (🟢/🟡/🔴), and evidence summary
- Step-by-step upgrade plan with time estimates, post-upgrade validation, and rollback guidance
- Version-specific gotcha warnings for known breaking changes in the upgrade path
- Pre-flight and day-of analysis modes with file-based cache handoff
- `kua health`, `kua compatibility`, `kua report` commands
- JSON/console/Markdown/HTML output formats with content parity across all formats
- Auto-save `assessment.json` on analyze for seamless `kua report` workflow
- Redacted output mode for sharing
- Component override file and helper workflows
- AKS provider evidence via CLI or file
- kubent API compatibility adapter
- Format-only version validation (supports any 1.X.Y cluster)
- `scripts/install-local.sh` for local binary installation

## Current quality evidence

Latest local checks:

```text
scripts/ci-local.sh
git diff --check
git fsck --full --strict
```

All checks must pass before implementation commits.
