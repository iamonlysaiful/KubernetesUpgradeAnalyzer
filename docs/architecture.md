# Architecture

Status: Proposed for implementation  
Last updated: 2026-07-22

## 1. Architectural style

KUA is a single Go CLI with a ports-and-adapters structure. Collection and external tools sit behind interfaces; analyzers consume a normalized snapshot; the recommendation engine consumes findings rather than raw Kubernetes clients. This makes deterministic fixture testing and future provider adapters possible without coupling core policy to AKS.

## 2. Logical flow

```text
CLI/config
   |
   v
Preflight ----> capability/evidence limitations
   |
   v
Live Kubernetes collectors ----> normalized ClusterSnapshot
External kubent adapter --------> API findings
AKS evidence resolver ----------> provider candidates/rules
  |-- local Azure CLI (auto/azure)
  `-- exported JSON (file/fallback)
Embedded catalog ---------------> API/component/provider knowledge
                                   |
                                   v
        inventory + detectors + health + compatibility analyzers
                                   |
                                   v
                         normalized Finding set
                                   |
                                   v
                         recommendation engine
                                   |
                                   v
                       canonical Assessment JSON
                          /       |       \
                     console   Markdown   HTML
```

## 3. Modules and responsibilities

### 3.1 `cmd/kua`

Cobra command wiring only. It parses flags, loads configuration, builds dependencies, and maps domain results to exit codes. Business policy does not live here.

### 3.2 Configuration

Precedence: command flags, environment variables, config file, defaults. The effective non-sensitive configuration is captured in report metadata. Provider source defaults to `auto`; kubeconfig follows standard client-go resolution and the current context is used unless explicitly selected. The selected context and intended provider-source behavior must be shown before collection.

### 3.3 Preflight

Validates kubeconfig/context, Kubernetes reachability, required discovery/RBAC access, `kubent` presence/version for relevant commands, embedded/override catalog integrity, output path, provider mode, Azure CLI availability/authentication where applicable, and optional provider evidence. Missing optional capabilities become explicit limitations; missing required capabilities fail safely. KUA never initiates interactive Azure authentication.

### 3.4 Collectors

Collectors use typed `client-go` clients where practical and discovery/dynamic clients where required. They return normalized domain records rather than Kubernetes API objects. Collection supports pagination, bounded concurrency, context cancellation, and partial-result diagnostics.

The collector never requests Secret contents. Event collection is bounded by namespace/resource association and configurable lookback.

### 3.5 Snapshot

`ClusterSnapshot` is the immutable input to internal analysis. It contains schema version, collection metadata, sanitized cluster identity, Kubernetes version, resource summaries, workload/container image metadata, node status, events, CRD metadata, and evidence limitations. It is serializable for sanitized fixtures, but raw export is not an MVP user feature until privacy review.

### 3.6 Detectors

Each component detector implements a common contract:

```go
type Detector interface {
    ID() string
    Detect(context.Context, ClusterSnapshot) []ComponentDetection
}
```

A detection contains product ID, normalized version when known, installation method (`helm`, `operator`, `manifest`, `managed`, `unknown`), namespace, evidence, and confidence. Detectors must tolerate incomplete RBAC and must not infer compatibility.

### 3.7 Analyzers

- Inventory analyzer summarizes scope and collection gaps.
- Health analyzer emits normalized health findings.
- API analyzer adapter translates kubent output into findings.
- Component compatibility analyzer joins detections with catalog entries.
- Provider analyzer evaluates AKS evidence and version/path constraints.

Analyzers do not print, mutate the cluster, or directly determine the overall recommendation.

### 3.8 Findings

All analyzers emit a shared structure:

- stable finding ID and rule ID;
- category and severity;
- status: `PASS`, `WARN`, `FAIL`, `UNKNOWN`, or `SKIPPED`;
- summary and remediation;
- affected resource references;
- candidate versions affected;
- evidence and provenance;
- confidence and limitations.

### 3.9 Recommendation engine

The engine evaluates candidates and paths using approved rules from `recommendation-model.md`. It is pure and deterministic: no I/O, wall-clock lookup, provider calls, or rendering. It returns an assessment plus a decision trace.

### 3.10 Reports

Canonical assessment JSON is the only renderer input. Console, Markdown, and HTML are views of the same domain result. Renderers cannot recalculate readiness or risk.

## 4. Provider boundary

The provider-neutral interface returns provider identity confidence, candidate versions, upgrade edges, support status, and evidence metadata. The AKS adapter is first. Future EKS/GKE/OpenShift adapters implement the same contract.

AKS evidence resolution modes are:

1. `auto` (default): detect AKS, resolve cluster identity, and invoke the local authenticated Azure CLI; if that fails, use a supplied JSON evidence file; otherwise mark exact provider availability `UNKNOWN` and continue independent analysis;
2. `azure`: require Azure CLI evidence; failure makes provider analysis inconclusive;
3. `file`: require a user-supplied `az aks get-upgrades` JSON export;
4. `offline`: make no Azure/provider network call, optionally consume a supplied evidence file, and qualify missing availability;
5. `none`: skip provider-specific analysis.

The Azure adapter invokes `az` directly without a shell and permits only documented read operations. It uses existing authentication, accepts explicit subscription/resource-group/cluster overrides, and never changes the active subscription or initiates login. Kubeconfig provides Kubernetes API connectivity but does not itself contain Azure Resource Manager upgrade availability. AKS identity inferred from context names, API hostname, or node `providerID` is evidence with confidence, not a guaranteed mapping.

The embedded catalog establishes provider policy and component compatibility knowledge but never claims cluster-specific patch availability.

## 5. Suggested repository structure

```text
cmd/kua/
internal/app/
internal/config/
internal/domain/
internal/collector/kubernetes/
internal/detector/
internal/analyzer/{inventory,health,api,component,provider}/
internal/external/kubent/
internal/external/azurecli/
internal/provider/{core,aks}/
internal/catalog/
internal/engine/
internal/report/{console,json,markdown,html}/
internal/sanitize/
schemas/
catalog/
docs/
examples/
testdata/
```

Keep packages internal until a demonstrated external-consumer requirement justifies `pkg/`.

## 6. Dependency direction

CLI and adapters depend inward on application/domain contracts. Domain and engine packages do not import Cobra, Viper, Kubernetes clients, kubent integration, report templates, or cloud SDKs. Catalog parsing depends on domain types, not provider implementations.

## 7. Failure and partial evidence

- Kubernetes authentication, cluster reachability, corrupt catalog, or incompatible schema: command error; no readiness claim.
- In `auto` mode, missing/expired Azure authentication or unresolved AKS identity: continue independent analysis, emit an actionable limitation, and set exact provider availability to `UNKNOWN`; do not run `az login`.
- In `azure` or `file` mode, failure of the required evidence source makes provider analysis inconclusive.
- Forbidden optional resources: continue with `UNKNOWN` findings and lower confidence.
- kubent unavailable for full analysis: `INCONCLUSIVE`, unless the user selected a command that does not require API analysis.
- Detector ambiguity or unparseable version: `UNKNOWN`, never `PASS`.
- Report rendering failure: analysis may be retained in memory/JSON where safely possible, but the command reports failure.

## 8. Performance targets

Initial design targets a cluster with 5,000 pods and 100 namespaces in under two minutes, excluding kubent runtime and network latency, with bounded memory and API concurrency. These are provisional until benchmark fixtures exist.

## 9. Observability

Use structured `slog` logging to stderr; reports go to stdout or a file. Default logging excludes container environment values, command credentials, resource specs, annotations unless allowlisted, and all Secret data. Debug logs remain sanitized.
