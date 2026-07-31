# ADR-0006: Advisor Model with Confidence Scoring

Status: Accepted  
Date: 2026-07-31  
Deciders: User (project owner)

## Context

KUA MVP produces assessments with four readiness states: `READY`,
`READY_WITH_WARNINGS`, `NOT_READY`, and `INCONCLUSIVE`. The `INCONCLUSIVE`
state occurs when any required evidence is absent or ambiguous.

In practice, this creates a frustrating user experience:

```
UPGRADE ASSESSMENT  ·  INCONCLUSIVE
Destination: 1.35.6  (risk: UNKNOWN)
```

The user's actual question is: *"Can I upgrade tonight?"*

The current output does not answer this question. Instead, it reports facts
and uncertainty without providing a recommendation. One missing piece of
evidence (e.g., kubent target-rule coverage) makes the entire assessment
`INCONCLUSIVE`, even when 95% of the analysis passed cleanly.

## Decision

Transform KUA from an **analyzer** (fact reporter) into an **advisor**
(decision engine) by implementing:

1. **Confidence scoring**: Replace binary pass/fail with weighted confidence
   percentage. Each evidence factor contributes to overall confidence based
   on its weight and completeness. Initial weights are heuristic (OQ-008):
   - API Compatibility: 25%
   - Component Compatibility: 20%
   - Cluster Health: 20%
   - Provider Evidence: 15%
   - Storage Health: 10%
   - Analysis Coverage: 10%

2. **Traffic light decisions**: Replace four readiness states with three
   actionable decisions (OQ-009):
   - 🟢 GO (confidence ≥ 90%, no blockers)
   - 🟡 GO WITH CAUTION (confidence 70–89%, no blockers)
   - 🔴 DO NOT PROCEED (any blocker or confidence < 70%)

3. **Localized unknowns**: Unknown evidence reduces confidence for that
   specific factor rather than blocking the entire assessment. Each unknown
   includes impact level, action to resolve, and consequence if ignored.

4. **Mandatory actions**: Every finding must include a concrete action item
   with effort estimate and consequence of ignoring.

5. **Upgrade plan generation**: Produce a step-by-step upgrade checklist
   with time estimates, post-upgrade validation steps, and rollback guidance.

6. **Clean schema break**: JSON output moves to Schema 2.0.0 without backward
   compatibility mode (OQ-010). Old `readiness` and `risk` fields are replaced
   by `decision` and `confidence`.

6. **Evidence summary**: Show what was analyzed ("Why do I trust this?") to
   build user confidence in the recommendation.

## Rationale

### Why confidence scoring?

Binary pass/fail treats all evidence equally. In reality:

- Missing kubent coverage (10% weight) should not block an assessment where
  API compatibility, component compatibility, cluster health, and provider
  evidence (90% weight combined) all pass.

- A confidence score of "92%" with explanation is more actionable than
  "INCONCLUSIVE" with no guidance.

### Why traffic light?

- Users make GO/NO-GO decisions. Four states with `INCONCLUSIVE` force users
  to interpret the tool's uncertainty rather than getting a recommendation.

- 🟢/🟡/🔴 maps directly to change management approval workflows.

### Why localized unknowns?

Global `INCONCLUSIVE` treats all unknowns as equally severe. In practice:

- Unknown kubent coverage: Low impact (API analysis may be incomplete)
- Unknown component version: Medium impact (compatibility unverified)
- Unknown node health: High impact (upgrade may fail)

Localization allows appropriate confidence reduction per factor.

### Why mandatory actions?

"kubent rule coverage not verified" is not actionable. Users need:

- What to do: "Run kubent --target-version 1.35.6"
- How long: "2 minutes"
- What if I don't: "Potential undetected deprecated APIs"

### Why upgrade plan?

The user's goal is not "assess the cluster" but "upgrade the cluster safely."
A generated checklist with time estimates supports the actual workflow.

## Consequences

### Positive

- Users get actionable GO/CAUTION/STOP recommendations
- Uncertainty is quantified and explained, not blocking
- Findings include concrete next steps
- Upgrade workflow is supported end-to-end
- Compliance teams can audit decision rationale

### Negative

- Confidence weights are initially heuristic; calibration requires real data
- Schema version bump may affect existing JSON consumers
- More complex codebase

### Neutral

- Old readiness/risk fields preserved for backward compatibility
- `--legacy-output` flag available for old format

## Alternatives considered

### Keep INCONCLUSIVE but add guidance

Add "Next steps" to INCONCLUSIVE output without changing the model.

Rejected: Does not solve the core problem. Users still see "INCONCLUSIVE"
and must interpret it themselves.

### Require all evidence before producing a recommendation

Only produce GO/NO-GO when all evidence is complete.

Rejected: Real clusters always have gaps. This would make the tool unusable
for practical purposes.

### Probabilistic model with Bayesian inference

Use a proper probabilistic model to compute upgrade success probability.

Rejected for MVP: Requires training data from actual upgrades. Can be
considered for future calibration phase.

## Implementation

See [phase-10-advisor-model-plan.md](../plans/phase-10-advisor-model-plan.md)
for detailed implementation plan.
