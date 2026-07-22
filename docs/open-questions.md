# Open questions

Status: Open  
Owner: User unless noted  
Last updated: 2026-07-22

No product questions currently block Phase 0 implementation planning. New uncertainties discovered during implementation must be recorded here before affected work continues.

| ID | Question | Owner | Needed by |
| --- | --- | --- | --- |
| — | None currently | — | — |

## Resolved

| ID | Resolution | Date |
| --- | --- | --- |
| OQ-001 | Default `auto` mode uses local authenticated Azure CLI, exported JSON is the fallback/explicit file source, and strict offline mode remains available. | 2026-07-22 |
| OQ-004 | Embed a curated, versioned catalog; allow validated local overrides; use automation only to propose reviewed updates; never search/scrape at assessment runtime. | 2026-07-22 |
| OQ-002 | Support kubent `0.7.3` JSON output initially, disable Helm collection, and require verified rule coverage for every target stage. | 2026-07-22 |
| OQ-003 | Current state determines blockers; events default to a configurable 30-minute warning window; critical workloads require explicit configuration/labels. | 2026-07-22 |
| OQ-005 | Local reports show actionable names; MVP includes stable-alias redacted output for sharing. | 2026-07-22 |
| OQ-006 | Use Apache-2.0 and module path `github.com/iamonlysaiful/KubernetesUpgradeAnalyzer`; binary is `kua`. | 2026-07-22 |
| OQ-007 | `READY_WITH_WARNINGS` returns `0`; strict warning failure is deferred. | 2026-07-22 |
