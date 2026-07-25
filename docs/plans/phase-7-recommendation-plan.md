# Phase 7 recommendation engine plan

Status: Proposed plan for approval
Last updated: 2026-07-25

## 1. Scope

Phase 7 implements the recommendation engine as defined in the roadmap and
`docs/recommendation-model.md`. To stay within the MVP target of 30 ±2 PRs
(currently at 27), this phase consolidates P7-01 and P7-02 into two PRs:

| PR | Content |
|----|---------|
| 28 | This plan document |
| 29 | Recommendation engine + validation fixtures + closeout |

## 2. Recommendation engine contract

The engine produces:

- **Destination**: highest suitable candidate version
- **Path**: sequential upgrade stages from current to destination
- **Readiness**: READY, READY_WITH_WARNINGS, NOT_READY, or INCONCLUSIVE
- **Risk**: LOW, MEDIUM, HIGH, or UNKNOWN
- **Decision trace**: rules and evidence that determined the result
- **Remediation**: actionable steps for blockers and warnings

```go
// Recommendation contains the complete upgrade recommendation.
type Recommendation struct {
    // Destination is the recommended target version.
    Destination string
    // Path is the sequential upgrade stages.
    Path []UpgradeStage
    // Readiness indicates overall upgrade readiness.
    Readiness ReadinessState
    // Risk indicates the risk level.
    Risk RiskLevel
    // Findings contains all blockers, warnings, and info.
    Findings []Finding
    // Limitations lists evidence gaps.
    Limitations []Limitation
    // GeneratedAt is when the recommendation was produced.
    GeneratedAt time.Time
}

type ReadinessState string

const (
    ReadinessReady            ReadinessState = "READY"
    ReadinessReadyWithWarnings ReadinessState = "READY_WITH_WARNINGS"
    ReadinessNotReady         ReadinessState = "NOT_READY"
    ReadinessInconclusive     ReadinessState = "INCONCLUSIVE"
)

type RiskLevel string

const (
    RiskLow     RiskLevel = "LOW"
    RiskMedium  RiskLevel = "MEDIUM"
    RiskHigh    RiskLevel = "HIGH"
    RiskUnknown RiskLevel = "UNKNOWN"
)
```

## 3. Finding aggregation

The engine aggregates findings from:

| Source | Finding types |
|--------|---------------|
| Health rules | Node pressure, workload unavailability, PVC issues, events |
| API compatibility | Removed APIs, deprecated APIs |
| Component detection | Version compatibility, unknown versions |
| Provider evidence | Upgrade availability, path validity |

Findings are classified as:

| Severity | Effect |
|----------|--------|
| BLOCKER | Prevents upgrade, sets NOT_READY |
| WARNING | Allows upgrade with review, sets READY_WITH_WARNINGS |
| INFO | Informational, no readiness impact |
| UNKNOWN | Insufficient evidence, may set INCONCLUSIVE |

## 4. Blocker policy

Per `docs/recommendation-model.md`, blockers include:

- API removed at any proposed stage and still used by a live object
- Explicit component incompatibility with any stage
- Unavailable or invalid provider transition
- Unready node, unresolved node pressure, unbound required PVC
- Unavailable critical workload or persistent fatal pod condition
- Failed required analyzer or corrupt compatibility catalog

## 5. Risk determination

Risk is rule-based, not scored:

| Risk | Condition |
|------|-----------|
| HIGH | Any blocker or explicitly unsupported condition |
| UNKNOWN | Required evidence is insufficient |
| MEDIUM | No blocker, but material warnings or conditional support |
| LOW | All analyzers pass with adequate evidence |

## 6. Engine inputs

```go
// RecommendationInput contains all data needed for recommendation.
type RecommendationInput struct {
    // ClusterSnapshot is the inventory snapshot.
    ClusterSnapshot *inventory.Snapshot
    // HealthFindings are from the health analyzer.
    HealthFindings []health.Finding
    // APIFindings are from the API compatibility analyzer.
    APIFindings []kubent.APIFinding
    // ComponentDetections are from the component detector.
    ComponentDetections []components.Detection
    // ProviderEvidence is from the provider adapter.
    ProviderEvidence *provider.ProviderEvidence
    // Catalog is the compatibility catalog.
    Catalog *catalog.Catalog
    // Options configures the recommendation.
    Options RecommendationOptions
}

type RecommendationOptions struct {
    // TargetVersion is an explicit destination (optional).
    TargetVersion string
    // AllowPreview includes preview versions.
    AllowPreview bool
    // MaxMinorSkip limits the destination (default: 3 minors).
    MaxMinorSkip int
}
```

## 7. Stage evaluation

For each stage in the path:

1. Check provider evidence confirms the transition is available
2. Check no removed APIs are still in use
3. Check no component has explicit incompatibility
4. Collect warnings for deprecated APIs and conditional support
5. Aggregate findings for the stage

## 8. Destination selection

1. Build candidate set from provider evidence
2. Filter by policy (preview, max skip)
3. Evaluate each candidate from highest to lowest
4. Select the highest candidate without blockers
5. If all candidates have blockers, select none and return NOT_READY
6. If evidence is insufficient, return INCONCLUSIVE

## 9. Test coverage

Required test cases:

| Category | Cases |
|----------|-------|
| Readiness | READY, READY_WITH_WARNINGS, NOT_READY, INCONCLUSIVE |
| Risk | LOW, MEDIUM, HIGH, UNKNOWN |
| Blockers | API removed, component incompatible, provider invalid |
| Warnings | API deprecated, conditional support, health concerns |
| Path | Single hop, multi-hop, no upgrades available |
| Evidence gaps | Missing health, missing API, missing provider |

## 10. Validation case

Per `docs/recommendation-model.md`, the AKS 1.30 validation case:

- Current: `1.30.0`
- Provider evidence: destination `1.33.12` available
- No API blockers
- Components compatible
- Healthy workloads

Expected:
- Destination: `1.33.12`
- Path: `1.30.x → 1.31.x → 1.32.x → 1.33.12`
- Readiness: `READY`
- Risk: `LOW`

## 11. Package structure

```text
internal/
  recommendation/
    engine.go         # Main recommendation engine
    aggregator.go     # Finding aggregation
    policy.go         # Blocker and risk policy
    types.go          # Recommendation types
    engine_test.go    # Engine tests
    testdata/         # Test fixtures
```

## 12. Exit criteria

Phase 7 is complete when:

- recommendation engine produces deterministic output
- all readiness states are correctly determined
- all risk levels are correctly assigned
- blocker policy matches specification
- AKS 1.30 → 1.33 validation case passes
- no live cluster access occurred

## 13. Deferred scope

The following remain out of scope for Phase 7:

- CLI command integration
- Report rendering (Phase 8)
- Live cluster validation (Phase 9)
- Weighted risk scoring (requires separate ADR)
