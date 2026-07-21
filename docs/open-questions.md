# Open questions

Status: Open  
Owner: User unless noted  
Last updated: 2026-07-22

These questions do not prevent completing the architecture baseline, but affected implementation must wait for approval.

| ID | Question | Why it matters | Proposed default | Needed by |
| --- | --- | --- | --- | --- |
| OQ-002 | Which kubent versions and output mode are supported? | Parsing human output is brittle; machine-readable contract/version pinning is needed. | Pin a tested minimum/range and require structured output if available | Phase 1B |
| OQ-003 | What is the health stability window and which workloads are “critical”? | A transient restart should not equal a persistent blocker. | Configurable event lookback plus conservative built-in rules | Phase 1A |
| OQ-005 | Should reports expose real resource names by default or offer redaction profiles in MVP? | Names help remediation but may be sensitive when reports are shared. | Local reports show names; fixtures always sanitize; design redacted export before sharing | Phase 1A |
| OQ-006 | What licenses and public repository identity should be used? | Required before an open-source release. | Apache-2.0 proposed, not accepted | Before public release |
| OQ-007 | Should `READY_WITH_WARNINGS` return exit `0`? | CI users may want warnings to fail policy. | Exit `0`, with a future configurable strict mode | Phase 0 |

## Resolved

| ID | Resolution | Date |
| --- | --- | --- |
| OQ-001 | Default `auto` mode uses local authenticated Azure CLI, exported JSON is the fallback/explicit file source, and strict offline mode remains available. | 2026-07-22 |
| OQ-004 | Embed a curated, versioned catalog; allow validated local overrides; use automation only to propose reviewed updates; never search/scrape at assessment runtime. | 2026-07-22 |
