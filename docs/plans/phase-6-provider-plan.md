# Phase 6 provider evidence plan

Status: Proposed plan for approval
Last updated: 2026-07-24

## 1. Scope

Phase 6 implements the AKS provider evidence layer as defined in the roadmap.
To stay within the MVP target of 30 ±2 PRs (currently at 25), this phase
consolidates P6-01, P6-02, and P6-03 into two PRs:

| PR | Content |
|----|---------|
| 26 | This plan document |
| 27 | Provider foundation + AKS adapter + candidate construction + closeout |

## 2. Provider interface contract

The provider-neutral interface returns:

- provider identity and confidence level;
- current cluster version;
- available upgrade versions with support status;
- sequential upgrade path edges;
- evidence metadata and limitations.

```go
// Provider is the provider-neutral evidence interface.
type Provider interface {
    // Identity returns the detected provider and confidence.
    Identity() (ProviderType, Confidence)
    
    // Evidence retrieves upgrade availability for the cluster.
    Evidence(ctx context.Context, opts EvidenceOptions) (*ProviderEvidence, error)
}

type ProviderType string

const (
    ProviderAKS     ProviderType = "AKS"
    ProviderUnknown ProviderType = "UNKNOWN"
)

type Confidence string

const (
    ConfidenceHigh    Confidence = "HIGH"
    ConfidenceMedium  Confidence = "MEDIUM"
    ConfidenceLow     Confidence = "LOW"
    ConfidenceUnknown Confidence = "UNKNOWN"
)
```

## 3. AKS identity detection

AKS identity is inferred with confidence levels:

| Signal | Confidence |
|--------|------------|
| Explicit `--subscription`/`--resource-group`/`--cluster-name` flags | HIGH |
| Node `spec.providerID` contains `azure:///subscriptions/` | HIGH |
| API server hostname matches `*.azmk8s.io` | MEDIUM |
| Context name matches `*-aks-*` pattern | LOW |
| No matching signals | UNKNOWN |

Identity detection is evidence, not a guarantee. The adapter accepts explicit
overrides to resolve ambiguity.

## 4. Evidence source modes

As defined in ADR-0005 and architecture.md:

| Mode | Behavior |
|------|----------|
| `auto` | Detect AKS → invoke `az aks get-upgrades` → fall back to file → `UNKNOWN` |
| `azure` | Require Azure CLI; failure is inconclusive |
| `file` | Require user-supplied JSON export |
| `offline` | No provider network call; optional file; qualify missing availability |
| `none` | Skip provider analysis |

## 5. Azure CLI adapter

The AKS adapter:

- invokes `az` directly without a shell;
- permits only the allowlisted `az aks get-upgrades` command;
- uses existing authentication without initiating login;
- accepts explicit `--subscription`, `--resource-group`, `--cluster-name` overrides;
- never changes subscriptions or mutates resources;
- validates output against `schemas/provider-evidence/aks-v1.json`.

Command construction:

```text
az aks get-upgrades \
  --subscription <subscription> \
  --resource-group <resource-group> \
  --name <cluster-name> \
  --output json
```

## 6. File evidence adapter

The file adapter:

- accepts a path to a JSON file exported from `az aks get-upgrades`;
- validates against the schema;
- records `method: AZURE_CLI_EXPORT` or `method: USER_FILE` in evidence.

## 7. Candidate and path construction

From provider evidence, build:

1. **Candidate versions**: all `availableUpgrades` versions that are not preview
   (unless preview is explicitly allowed).

2. **Sequential path**: for destination `1.33.x` from current `1.30.x`, the
   path is `1.30 → 1.31 → 1.32 → 1.33`, selecting the highest available patch
   in each minor that the provider reports as available.

3. **Validation**: each edge in the path must be provider-valid (the target
   version appears in `availableUpgrades` from the source version's perspective,
   or the catalog defines AKS sequential policy).

## 8. Limitations and fallback

When provider evidence is unavailable or incomplete:

- `auto` mode: emit limitation, set exact availability to `UNKNOWN`, continue
  with independent Kubernetes/API/component analysis;
- `azure`/`file` mode: provider analysis is `INCONCLUSIVE`;
- `offline` mode: no provider call, qualify findings;
- `none` mode: skip provider findings entirely.

Missing node pool data, expired CLI auth, or unresolved identity become
explicit limitations with severity `WARN` or `ERROR`.

## 9. Package structure

```text
internal/
  provider/
    provider.go       # Interface and types
    evidence.go       # ProviderEvidence struct
    candidate.go      # Candidate/path construction
    aks/
      identity.go     # AKS identity detection
      adapter.go      # Azure CLI adapter
      file.go         # File evidence adapter
      parser.go       # az aks get-upgrades JSON parsing
```

## 10. Test coverage

Required test cases:

| Category | Cases |
|----------|-------|
| Identity detection | HIGH/MEDIUM/LOW/UNKNOWN signals, explicit overrides |
| Azure CLI | Success, auth failure, timeout, invalid output, mutating command rejection |
| File adapter | Valid JSON, invalid JSON, schema mismatch, missing file |
| Candidate construction | Single hop, multi-hop sequential, preview filtering, empty upgrades |
| Path validation | Valid sequential, skip-minor rejection, provider-invalid edge |
| Mode fallback | auto→file→UNKNOWN, azure failure, offline behavior |

All tests use fixtures and fakes. No live Azure CLI execution without separate
explicit approval.

## 11. Fixtures

Existing fixtures:

- `schemas/fixtures/provider-evidence/valid/aks-get-upgrades-1-30-to-1-33.json`
- `schemas/fixtures/provider-evidence/invalid/mutating-command.json`

Additional fixtures to add:

- `internal/provider/aks/testdata/az-output-valid.json`
- `internal/provider/aks/testdata/az-output-empty-upgrades.json`
- `internal/provider/aks/testdata/az-output-auth-error.json`

## 12. Exit criteria

Phase 6 is complete when:

- provider interface and types are implemented;
- AKS identity detection returns correct confidence levels;
- Azure CLI adapter invokes only allowlisted commands;
- file adapter validates against schema;
- candidate and sequential path construction work for `1.30 → 1.33`;
- all mode/fallback/authentication/offline cases pass without provider mutation;
- no live Azure CLI execution occurred.

## 13. Deferred scope

The following remain out of scope for Phase 6:

- recommendation engine integration (Phase 7);
- report rendering (Phase 8);
- live Azure CLI validation against approved context (Phase 9);
- EKS, GKE, OpenShift, or vanilla provider adapters (later phases).

## 14. Security boundaries

- Azure CLI is invoked shell-free with validated arguments only;
- only `az aks get-upgrades` is permitted; mutating commands are rejected;
- no credential capture, token export, or subscription change;
- evidence JSON is validated against schema before use;
- node `providerID` parsing does not expose subscription/resource-group names
  in logs without redaction.
