# Product requirements

Status: Accepted architecture baseline  
Product owner: User  
Last updated: 2026-07-22

## 1. Vision

KUA helps platform engineers decide whether a live Kubernetes cluster can be upgraded safely, recommends a supported destination version, identifies the required upgrade stages, and explains the evidence and uncertainty behind that recommendation.

## 2. Principles

1. Local-first provider enrichment by default, with a strict explicit offline mode.
2. Read-only access to Kubernetes and provider systems.
3. No Kubernetes Secret payload collection.
4. Explainable, deterministic recommendations.
5. AKS-first implementation with provider-neutral core contracts.
6. Evidence limitations reduce confidence; they never silently become a pass.
7. Machine-readable output is stable enough for automation.

## 3. MVP scope

### 3.1 Live cluster inventory

Collect only metadata and status needed for assessment:

- server Kubernetes version and discovery information;
- nodes and node pools when identifiable;
- namespaces;
- Deployments, StatefulSets, DaemonSets, Jobs, and CronJobs;
- Services and Ingresses;
- PersistentVolumeClaims, StorageClasses, and CSIDrivers;
- CustomResourceDefinitions;
- non-secret workload metadata needed for component detection and health.

Secret objects and secret-backed values are out of scope.

### 3.2 Component detection

Detect, where evidence permits: NGINX Ingress, Metrics Server, CoreDNS, Azure Disk CSI, Azure File CSI, Fluent Bit, Fluentd, Prometheus, Grafana, Loki, Tempo, OpenTelemetry Collector, EMQX, cert-manager, Argo CD, Istio, Linkerd, KEDA, External Secrets, and Velero.

Detection must not depend solely on Helm. It may use a weighted set of evidence such as labels, annotations, image repositories/tags, workload names, namespaces, CRDs, API groups, and operator-owned resources. Every result includes evidence and confidence. Ambiguous evidence produces `UNKNOWN`, not a guessed version.

### 3.3 API compatibility

For MVP, invoke an installed `kubent` binary through a controlled adapter. Report deprecated APIs, removed APIs, the affected objects, removal versions, and analyzer limitations. A future native analyzer replaces or supplements this adapter behind the same internal contract.

### 3.4 Health assessment

Assess node readiness and pressure, pod phases and waiting reasons, workload availability, DaemonSet coverage, PVC binding, and relevant Warning events. Health findings must distinguish blockers from warnings and include resource references without secret data.

### 3.5 Compatibility and provider assessment

- Evaluate detected component versions against a bundled, versioned compatibility catalog.
- Evaluate API removals for each candidate Kubernetes minor version.
- Read the current kubeconfig context by default, following standard kubeconfig resolution, and permit explicit kubeconfig/context selection.
- For detected AKS clusters, default provider source `auto` invokes the locally installed, already authenticated Azure CLI to retrieve exact upgrade availability when the cluster identity can be resolved.
- Never initiate `az login`, open a browser, change subscriptions, or mutate Azure resources. Azure failure falls back to supplied provider evidence when available; otherwise provider availability becomes `UNKNOWN` while independent Kubernetes analysis continues.
- Support `azure`, `file`, `offline`, and `none` provider-source modes in addition to `auto`.
- Accept user-supplied JSON exported from `az aks get-upgrades` as the offline provider-evidence source.

### 3.6 Recommendation

Return:

- current version;
- recommended destination version;
- sequential, provider-valid upgrade stages;
- overall readiness: `READY`, `READY_WITH_WARNINGS`, `NOT_READY`, or `INCONCLUSIVE`;
- risk: `LOW`, `MEDIUM`, `HIGH`, or `UNKNOWN`;
- blockers, warnings, assumptions, evidence age, and remediation.

KUA may recommend destination `1.33.12` for a cluster at `1.30.x` when evidence supports it, but it must not describe `1.30 → 1.33.12` as one supported AKS operation. It must display intervening minor stages.

### 3.7 Output

MVP outputs: console, JSON, Markdown, and self-contained HTML. PDF and interactive dashboards are later phases.

## 4. CLI commands

- `kua analyze`: full assessment.
- `kua inventory`: inventory only.
- `kua health`: health only.
- `kua compatibility`: API and component compatibility.
- `kua report`: render a previously saved canonical JSON assessment.
- `kua version`: CLI and embedded catalog versions.

## 5. Explicit non-goals for MVP

- Mutating or upgrading a cluster.
- Local manifest, Helm chart, Git repository, or CI analysis.
- Native deprecated API analysis.
- Live EKS, GKE, OpenShift, or generic provider-specific recommendations.
- Automatic catalog downloads.
- Runtime internet searches or vendor-page scraping for compatibility claims.
- AI-generated recommendations.
- Security posture, cost, drift, or historical monitoring.

## 6. MVP acceptance criteria

1. A sanitized AKS fixture representing Kubernetes `1.30.0` can produce destination `1.33.12`, sequential stages, `READY`, and `LOW` risk when all stated evidence passes.
2. A removed API blocks every candidate where it is unavailable and identifies the responsible object.
3. A component with unknown version or missing compatibility evidence cannot produce an unconditional compatibility pass.
4. A serious health condition produces a blocker according to the approved recommendation rules.
5. `--provider-source=offline` causes no provider network connection attempt; catalog analysis remains local.
6. JSON output validates against its versioned schema and is deterministic after volatile timestamps are normalized.
7. Kubernetes Secret contents never appear in logs, evidence, fixtures, or reports.
8. In `auto` mode, unavailable Azure CLI/authentication degrades provider availability to `UNKNOWN` without turning unrelated findings into failures.
