# Phase 10: Advisor Model Transformation

Status: Approved  
Last updated: 2026-07-31

## 1. Objective

Transform KUA from a **Kubernetes Upgrade Analyzer** (fact reporter) into a
**Kubernetes Upgrade Advisor** (decision engine). The core change is replacing
global `INCONCLUSIVE` / `UNKNOWN` states with confidence-based recommendations
that localize uncertainty and provide actionable guidance.

## 2. Problem statement

Current output:

```
UPGRADE ASSESSMENT  ·  INCONCLUSIVE
Current:     1.34.9
Destination: 1.35.6  (risk: UNKNOWN)

EVIDENCE GAPS (2)
[1] TARGET_COVERAGE_UNVERIFIED: kubent target-rule coverage is not verified
[2] FILE_EVIDENCE_SOURCE: Evidence loaded from file
```

User reaction: *"Why did I run this tool?"*

The tool reports facts but does not answer the user's actual question:

> **Can I upgrade tonight?**

## 3. Target output

```
═══════════════════════════════════════════════════════════════════
 KUBERNETES UPGRADE ADVISOR
═══════════════════════════════════════════════════════════════════

 Recommendation:     🟢 PROCEED WITH STAGING UPGRADE
 Confidence:         94%

 Current Version:    1.34.9
 Target Version:     1.35.6
 Upgrade Path:       1.34.9 → 1.35.6
 Estimated Time:     18–25 minutes

───────────────────────────────────────────────────────────────────
 EVIDENCE SUMMARY
───────────────────────────────────────────────────────────────────
   ✔ 124 Deployments analyzed (all available)
   ✔ 3 StatefulSets analyzed (all ready)
   ✔ 16 CRDs (no deprecated APIs found)
   ✔ 4 Nodes (all Ready, no pressure)
   ✔ 12 PVCs (all bound)
   ✔ Components: CoreDNS ✔ EMQX ✔ NGINX ✔ Fluent Bit ✔

───────────────────────────────────────────────────────────────────
 BLOCKERS (0)
───────────────────────────────────────────────────────────────────
   None — no issues blocking upgrade.

───────────────────────────────────────────────────────────────────
 WARNINGS (1)  — review before upgrading
───────────────────────────────────────────────────────────────────
 [1] kubent rule coverage not verified for 1.35
     Impact:      Low — only affects deprecated API detection
     Action:      Run `kubent --target-version 1.35.6` with latest rules
     Effort:      2 minutes
     If ignored:  Potential undetected deprecated APIs; unlikely to
                  cause upgrade failure but may cause post-upgrade issues

───────────────────────────────────────────────────────────────────
 RECOMMENDED UPGRADE PLAN
───────────────────────────────────────────────────────────────────
   1. ☐ Take AKS backup or etcd snapshot
   2. ☐ Upgrade control plane to 1.35.6
   3. ☐ Upgrade node pools (rolling)
   4. ☐ Wait for all nodes Ready
   5. ☐ Verify StatefulSets healthy
   6. ☐ Run application smoke tests
   7. ☐ Monitor for 30 minutes

───────────────────────────────────────────────────────────────────
 POST-UPGRADE VALIDATION
───────────────────────────────────────────────────────────────────
   ☐ All nodes Ready (kubectl get nodes)
   ☐ All DaemonSets at desired count
   ☐ All Deployments available
   ☐ All StatefulSets ready
   ☐ No CrashLoopBackOff pods
   ☐ Ingress responding

───────────────────────────────────────────────────────────────────
 Overall Decision:   GO
 Generated:          2026-07-31T10:30:00Z
═══════════════════════════════════════════════════════════════════
```

## 4. Key design changes

### 4.1 Confidence scoring replaces binary pass/fail

Current model: one `UNKNOWN` anywhere → global `INCONCLUSIVE`.

New model: weighted confidence score with localized uncertainty.

```
Confidence = Σ (factor_weight × factor_confidence)
```

| Factor | Weight | 100% confidence | Reduced confidence |
|--------|--------|-----------------|-------------------|
| API Compatibility | 0.25 | No removed APIs in use | kubent coverage gap: 50% |
| Component Compatibility | 0.20 | All detected components verified | Unknown version: 70% per component |
| Cluster Health | 0.20 | All nodes Ready, workloads available | Warnings present: 80% |
| Provider Evidence | 0.15 | Fresh AKS CLI response | File evidence: 80%; unavailable: 50% |
| Storage Health | 0.10 | All PVCs bound, no pressure | Unbound PVC: 0% (blocker) |
| Analysis Coverage | 0.10 | kubent rules verified | Not verified: 50% |

Example calculation:

```
API Compatibility:      0.25 × 1.00 = 0.250
Component Compat:       0.20 × 1.00 = 0.200
Cluster Health:         0.20 × 1.00 = 0.200
Provider Evidence:      0.15 × 0.80 = 0.120  (file evidence)
Storage Health:         0.10 × 1.00 = 0.100
Analysis Coverage:      0.10 × 0.50 = 0.050  (kubent unverified)
─────────────────────────────────────────────
Total Confidence:                     0.920 = 92%
```

### 4.2 Traffic light decision model

Replace 4 readiness states with 3 actionable decisions:

| Decision | Criteria | User action |
|----------|----------|-------------|
| 🟢 **GO** | Confidence ≥ 90%, zero blockers | Proceed with upgrade |
| 🟡 **GO WITH CAUTION** | Confidence 70–89%, zero blockers | Review warnings, then proceed |
| 🔴 **DO NOT PROCEED** | Any blocker OR confidence < 70% | Resolve blockers first |

**Approved thresholds (OQ-009):** These are the approved defaults. Risk profiles
(conservative/balanced/aggressive) are included in Phase 10 MVP to allow
per-environment customization.

The old states map as follows:

| Old state | New decision |
|-----------|--------------|
| `READY` | 🟢 GO |
| `READY_WITH_WARNINGS` | 🟡 GO WITH CAUTION |
| `NOT_READY` | 🔴 DO NOT PROCEED |
| `INCONCLUSIVE` | 🟡 or 🔴 depending on confidence |

### 4.3 Localized unknowns with impact assessment

Current: global `INCONCLUSIVE` blocks entire decision.

New: each unknown becomes a localized finding with:

- **Impact level**: Low / Medium / High
- **Explanation**: What this affects
- **Action**: Specific command or step to resolve
- **Effort**: Time estimate
- **If ignored**: Consequence of proceeding anyway

Example:

```
⚠ EMQX version supplied manually
  Impact:      Low — affects component compatibility confidence only
  Action:      Verify running EMQX version after upgrade
  Effort:      5 minutes
  If ignored:  Possible unexpected EMQX behavior; unlikely to block upgrade
```

### 4.4 Mandatory action per finding

Every finding (blocker, warning, info) must include:

```go
type Finding struct {
    // ... existing fields ...
    
    Impact      FindingImpact  `json:"impact"`
    Action      ActionItem     `json:"action"`
    IfIgnored   string         `json:"ifIgnored,omitempty"`
}

type FindingImpact struct {
    Level       ImpactLevel `json:"level"`       // Low, Medium, High
    Explanation string      `json:"explanation"`
}

type ActionItem struct {
    Description string `json:"description"`
    Command     string `json:"command,omitempty"`
    Effort      string `json:"effort,omitempty"` // "2 minutes", "30 minutes"
}
```

### 4.5 Evidence summary ("Why do I trust this?")

New section showing what was actually analyzed:

```go
type EvidenceSummary struct {
    Deployments      CountSummary `json:"deployments"`
    StatefulSets     CountSummary `json:"statefulSets"`
    DaemonSets       CountSummary `json:"daemonSets"`
    Nodes            CountSummary `json:"nodes"`
    PVCs             CountSummary `json:"pvcs"`
    CRDs             CountSummary `json:"crds"`
    Components       []string     `json:"components"`
    DeprecatedAPIs   int          `json:"deprecatedAPIs"`
    AnalysisTime     time.Time    `json:"analysisTime"`
}

type CountSummary struct {
    Total   int `json:"total"`
    Healthy int `json:"healthy"`
}
```

Console output:

```
EVIDENCE SUMMARY
  ✔ 124 Deployments analyzed (all available)
  ✔ 3 StatefulSets analyzed (all ready)
  ✔ 16 CRDs (no deprecated APIs found)
  ✔ 4 Nodes (all Ready, no pressure)
  ✔ 12 PVCs (all bound)
  ✔ Components: CoreDNS ✔ EMQX ✔ NGINX ✔ Fluent Bit ✔
```

### 4.6 Upgrade plan generator

New package `internal/plan` generates a contextual upgrade checklist:

```go
type UpgradePlan struct {
    Steps            []PlanStep     `json:"steps"`
    EstimatedTime    Duration       `json:"estimatedTime"`
    ValidationSteps  []PlanStep     `json:"validationSteps"`
    RollbackGuidance string         `json:"rollbackGuidance,omitempty"`
}

type PlanStep struct {
    Order       int    `json:"order"`
    Description string `json:"description"`
    Command     string `json:"command,omitempty"`
    Expected    string `json:"expected,omitempty"`
}
```

Default AKS upgrade plan:

1. Take AKS backup or etcd snapshot
2. Upgrade control plane to target version
3. Upgrade system node pool
4. Upgrade user node pools (one at a time recommended)
5. Wait for all nodes Ready
6. Verify StatefulSets healthy
7. Run application smoke tests
8. Monitor for 30 minutes

Time estimate heuristic:

- Base: 15 minutes (control plane)
- +5 minutes per node pool
- +1 minute per 10 nodes

### 4.7 Post-upgrade validation checklist

Generated based on inventory:

```
POST-UPGRADE VALIDATION
  ☐ All nodes Ready (kubectl get nodes)
  ☐ All DaemonSets at desired count
  ☐ All Deployments available
  ☐ All StatefulSets ready
  ☐ No CrashLoopBackOff pods
  ☐ Ingress responding (if Ingress detected)
  ☐ [Custom] EMQX cluster healthy (if EMQX detected)
```

### 4.8 Pre-flight vs. day-of separation

Some checks can run days before upgrade; others must run at execution time:

| Pre-flight (anytime) | Day-of (at upgrade) |
|----------------------|---------------------|
| API compatibility | Node health |
| Component version checks | Pod status |
| Upgrade path validation | Event stream |
| RBAC requirements | PVC binding |
| Provider upgrade availability | Active alerts |

New flags:

- `kua analyze --preflight` — runs only pre-flight checks
- `kua analyze --day-of` — runs only day-of checks (requires prior preflight)
- `kua analyze` (default) — runs both

### 4.9 Rollback guidance

New section when risk is MEDIUM or higher:

```
ROLLBACK GUIDANCE
  Control plane: Cannot be rolled back after upgrade completes.
  Node pools: Can be recreated at previous version if needed.
  
  Recommended:
    • Take AKS backup before starting
    • Upgrade one node pool at a time
    • If node pool upgrade fails, create new pool at old version
    • Have application rollback plan ready
```

### 4.10 Risk appetite configuration

Support user-configurable thresholds:

```yaml
# ~/.kua/config.yaml or --risk-profile flag
riskProfile: balanced  # conservative | balanced | aggressive
```

| Profile | GO threshold | CAUTION threshold |
|---------|-------------|-------------------|
| Conservative | ≥95% | ≥85% |
| Balanced | ≥90% | ≥75% |
| Aggressive | ≥80% | ≥60% |

### 4.11 Version-specific gotchas

Proactive warnings for known breaking changes in upgrade path:

| Version | Gotcha |
|---------|--------|
| 1.25 | PodSecurityPolicy removed |
| 1.26 | CRI v1alpha2 removed |
| 1.29 | Flowcontrol v1beta2 removed |
| 1.32 | In-tree cloud providers removed |

When upgrade path crosses a gotcha version:

```
⚠ Version 1.25 removes PodSecurityPolicy
  Impact: High if PSP in use; Medium otherwise
  Action: Migrate to Pod Security Admission before upgrading past 1.24
  Effort: 1–4 hours depending on policy complexity
```

### 4.12 Deprecation timeline pressure

Surface AKS support deadlines:

```
⚠ Current version 1.30.x reaches end-of-support in 45 days
  Recommended: Complete upgrade to 1.33+ within 30 days
```

### 4.13 Audit trail for compliance

Include assessment metadata for regulated environments:

```json
{
  "assessmentId": "kua-2026-07-31-abc123",
  "decision": "GO",
  "confidence": 0.94,
  "decisionFactors": [...],
  "generatedBy": "kua v0.1.0",
  "timestamp": "2026-07-31T10:30:00Z",
  "clusterFingerprint": "sha256:...",
  "configHash": "sha256:..."
}
```

## 5. Implementation plan

### Phase 10.1: Confidence model foundation

1. Define `ConfidenceModel` and `ContributionFactor` types
2. Define `ImpactLevel` enum (Low, Medium, High)
3. Add `Impact`, `Action`, `IfIgnored` fields to `Finding`
4. Implement `CalculateConfidence()` function with weighted factors
5. Unit tests for confidence calculation edge cases

### Phase 10.2: Traffic light decision engine

1. Define `Decision` type (GO, GO_WITH_CAUTION, DO_NOT_PROCEED)
2. Implement `DetermineDecision()` using confidence thresholds
3. Add `riskProfile` configuration (conservative/balanced/aggressive)
4. Update `Policy.EvaluateReadiness()` to use new model
5. Backward-compatible: keep old `ReadinessState` as internal detail

### Phase 10.3: Evidence summary builder

1. Define `EvidenceSummary` type
2. Implement `BuildEvidenceSummary()` from inventory snapshot
3. Generate component compatibility status list
4. Add evidence summary to assessment document

### Phase 10.4: Upgrade plan generator

1. Create `internal/plan` package
2. Define `UpgradePlan` and `PlanStep` types
3. Implement AKS-specific plan generator
4. Implement time estimation heuristic
5. Implement post-upgrade validation checklist generator
6. Add rollback guidance for MEDIUM+ risk

### Phase 10.5: Finding enhancement

1. Add mandatory `Action` to all finding generators
2. Add `IfIgnored` consequence text
3. Add `Effort` estimates
4. Update all existing finding factories

### Phase 10.6: Console renderer overhaul

1. Rename "UPGRADE ASSESSMENT" to "KUBERNETES UPGRADE ADVISOR"
2. Add traffic light emoji (🟢🟡🔴)
3. Add confidence percentage display
4. Add estimated time display
5. Add EVIDENCE SUMMARY section
6. Add RECOMMENDED UPGRADE PLAN section
7. Add POST-UPGRADE VALIDATION section
8. Add ROLLBACK GUIDANCE section (when applicable)

### Phase 10.7: Version-specific gotchas

1. Define gotcha catalog (version → breaking change)
2. Implement path scanner for gotcha versions
3. Generate proactive warnings for upcoming gotchas
4. Add AKS support timeline warnings

### Phase 10.8: Pre-flight/day-of modes

1. Tag each analyzer as pre-flight or day-of
2. Add `--preflight` and `--day-of` flags
3. Implement selective analyzer execution
4. Cache pre-flight results for day-of reuse

## 6. Schema changes

### Assessment document additions

```yaml
# New top-level fields
decision: GO                    # GO | GO_WITH_CAUTION | DO_NOT_PROCEED
confidence: 0.94
confidenceFactors:              # Breakdown of confidence calculation
  - factor: API_COMPATIBILITY
    weight: 0.25
    confidence: 1.0
    evidence: "No deprecated APIs found"
  - factor: COMPONENT_COMPATIBILITY
    weight: 0.20
    confidence: 1.0
    evidence: "5 components verified"
  # ...

evidenceSummary:
  deployments: { total: 124, healthy: 124 }
  statefulSets: { total: 3, healthy: 3 }
  # ...

upgradePlan:
  steps:
    - order: 1
      description: "Take AKS backup"
      command: "az aks backup create ..."
  estimatedTime: "PT25M"
  validationSteps:
    - order: 1
      description: "Verify all nodes Ready"
      command: "kubectl get nodes"
  rollbackGuidance: "..."

# Finding additions
findings:
  - id: KUBENT_COVERAGE_UNVERIFIED
    severity: WARNING
    summary: "kubent rule coverage not verified for 1.35"
    impact:
      level: Low
      explanation: "Only affects deprecated API detection"
    action:
      description: "Run kubent with latest rules"
      command: "kubent --target-version 1.35.6"
      effort: "2 minutes"
    ifIgnored: "Potential undetected deprecated APIs"
```

## 7. Backward compatibility

**Decision (OQ-010):** Clean break to Schema 2.0.0. No `--legacy-output` flag.

- JSON schema version bumps from `1.0.0` to `2.0.0`
- Old `readiness` and `risk` fields removed; replaced by `decision` and `confidence`
- JSON consumers must update to new schema
- Release notes will document migration path
- Schema 1.x remains documented for reference but is not supported

## 8. Testing strategy

### Unit tests

- Confidence calculation with all factor combinations
- Decision thresholds at boundaries (89%, 90%, 91%)
- Risk profile switching
- Plan generation with various cluster sizes
- Evidence summary accuracy

### Integration tests

- Full assessment produces valid confidence score
- Traffic light matches expected for fixture scenarios
- Plan contains correct AKS commands

### Golden tests

- New console format matches approved golden files
- JSON schema validates against new schema version

## 9. Exit criteria

1. `kua analyze` produces confidence percentage, not just READY/INCONCLUSIVE
2. Output shows 🟢/🟡/🔴 decision with explanation
3. Every finding has Action and IfIgnored fields
4. Evidence summary shows what was analyzed
5. Upgrade plan with time estimate is generated
6. Post-upgrade validation checklist is generated
7. Rollback guidance appears for MEDIUM+ risk
8. JSON output validates against Schema 2.0.0
9. Documentation updated to reflect advisor model
10. Risk profile configuration (`--risk-profile`) works for all three profiles

## 10. Dependencies

- Phase 8.5 complete (current state)
- No external dependencies
- No new binaries required

## 11. Risks

| Risk | Mitigation |
|------|------------|
| Confidence weights are arbitrary | Start with documented heuristics (OQ-008); calibrate with real upgrade outcomes |
| Users expect 100% means guaranteed success | Clear disclaimer: confidence reflects evidence quality, not upgrade outcome |
| Breaking change for JSON consumers | Clean break to 2.0.0 (OQ-010); document migration path in release notes |

## 12. Future work (out of scope for Phase 10)

- Multi-cluster promotion workflow
- Integration with monitoring systems for day-of checks
- Machine learning calibration of confidence weights
- PDF report format
- Interactive dashboard
