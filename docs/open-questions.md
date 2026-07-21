# Open questions

Status: Open  
Owner: User unless noted  
Last updated: 2026-07-21

These questions do not prevent completing the architecture baseline, but affected implementation must wait for approval.

| ID | Question | Why it matters | Proposed default | Needed by |
| --- | --- | --- | --- | --- |
| OQ-001 | How will MVP obtain exact AKS offered upgrades: user-supplied `az aks get-upgrades` snapshot, live opt-in Azure query, or both? | Exact patches and region/cluster availability cannot be guaranteed by a static offline catalog. | Sanitized user-supplied snapshot; no live Azure call | Phase 1C |
| OQ-002 | Which kubent versions and output mode are supported? | Parsing human output is brittle; machine-readable contract/version pinning is needed. | Pin a tested minimum/range and require structured output if available | Phase 1B |
| OQ-003 | What is the health stability window and which workloads are “critical”? | A transient restart should not equal a persistent blocker. | Configurable event lookback plus conservative built-in rules | Phase 1A |
| OQ-004 | How should component evidence be maintained and reviewed? | Compatibility claims age quickly and vary in quality. | Official sources, review date, catalog version, explicit unknowns | Phase 0/1B |
| OQ-005 | Should reports expose real resource names by default or offer redaction profiles in MVP? | Names help remediation but may be sensitive when reports are shared. | Local reports show names; fixtures always sanitize; design redacted export before sharing | Phase 1A |
| OQ-006 | What licenses and public repository identity should be used? | Required before an open-source release. | Apache-2.0 proposed, not accepted | Before public release |
| OQ-007 | Should `READY_WITH_WARNINGS` return exit `0`? | CI users may want warnings to fail policy. | Exit `0`, with a future configurable strict mode | Phase 0 |

