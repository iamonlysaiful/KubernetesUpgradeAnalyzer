# KubeUpgrade Advisor

KubeUpgrade Advisor (KUA) is a planned open-source, local-first, read-only CLI for assessing whether a live Kubernetes cluster is ready to upgrade. It will combine inventory, deprecated API analysis, component compatibility, provider constraints, health checks, and explainable recommendation logic. For AKS, default `auto` mode may use the locally installed and already authenticated Azure CLI; explicit offline operation remains supported.

Core MVP foundations are implemented. The end-to-end AKS `kua analyze` flow is
validated locally with redacted output; release publication still requires
explicit owner approval.

## Current status

- Product direction: approved
- Initial provider: AKS
- Analysis target: live clusters
- Deprecated API adapter: installed `kubent` binary for MVP
- Native API analyzer: planned
- Architecture baseline: documented under [`docs/`](docs/README.md)
- Implementation progress: Phases 1-8 complete, Phase 8.5 validation in
  progress, Phase 9 publication blocked pending owner approval

Follow [`AGENTS.md`](AGENTS.md) and the docs-first approval workflow.

## Real-world setup and analysis runbook

This runbook is for operator-controlled AKS analysis. Do not run against
production first. KUA is read-only and does not mutate Kubernetes or Azure
resources.

### 1. Prerequisites

- Go `1.25+`
- `kubectl`
- `jq`
- Azure CLI (`az`) authenticated to the approved subscription
- `kubent` installed and available on `PATH`

### 2. Build and validate locally

```bash
git clone <repo-url>
cd KubernetesUpgradeAnalyzer
scripts/ci-local.sh
go build -o ./bin/kua ./cmd/kua
./bin/kua version
```

### 3. Pre-approval requirements for live staging runs

Before live commands, complete and approve:

- [`docs/validation/phase-9-approval-record.md`](docs/validation/phase-9-approval-record.md)
- approved staging context name
- approved read-only command list
- approved raw/sanitized output paths

No live command should be run outside the approved list.

### 4. Analyze any AKS cluster (read-only)

```bash
az login
az account set --subscription "<SUBSCRIPTION_ID>"

kubectl config get-contexts
kubectl cluster-info --context "<AKS_CONTEXT>"

./bin/kua analyze \
  --context "<AKS_CONTEXT>" \
  --format json \
  --redacted \
  --provider-source azure \
  --subscription "<SUBSCRIPTION_ID>" \
  --resource-group "<RESOURCE_GROUP>" \
  --cluster-name "<AKS_CLUSTER_NAME>" \
  > analyze.redacted.json
```

Expected successful assessment states:

- `READY` / `LOW`: no blockers, no material warnings.
- `READY_WITH_WARNINGS` / `MEDIUM`: no blockers, but review warnings before
  upgrade.
- `NOT_READY` / `HIGH`: blockers must be fixed before upgrade.
- `INCONCLUSIVE` / `UNKNOWN`: required evidence is missing or failed.

Store raw or cluster-specific outputs locally only. Do not commit raw outputs,
kubeconfigs, subscription IDs, resource groups, cluster names, node names,
namespaces, workload names, registry hosts, or secrets.

### 5. Fill component version overrides when requested

If component versions are missing, ambiguous, or confusing, JSON output includes
a `componentVersionOverrides` helper:

```bash
jq '.componentVersionOverrides' analyze.redacted.json
```

Generate `component-overrides.json` from the assessment:

```bash
./bin/kua component-overrides \
  --input analyze.redacted.json \
  --output component-overrides.json
```

Review the generated file. KUA uses observed versions when present and leaves a
`<fill-version>` placeholder only when no usable version was observed.

Then rerun:

```bash
./bin/kua analyze \
  --context "<AKS_CONTEXT>" \
  --format json \
  --redacted \
  --provider-source azure \
  --subscription "<SUBSCRIPTION_ID>" \
  --resource-group "<RESOURCE_GROUP>" \
  --cluster-name "<AKS_CLUSTER_NAME>" \
  --component-overrides component-overrides.json \
  > analyze.final.redacted.json
```

Overrides are local operator evidence. They can resolve missing component-version
evidence, but they do not suppress API, provider, health, or explicit
incompatibility findings.

### 6. Optional provider evidence file mode

If Azure CLI access is unavailable during analysis, collect AKS upgrade evidence
separately and run file mode:

```bash
az aks get-upgrades \
  --subscription "<SUBSCRIPTION_ID>" \
  --resource-group "<RESOURCE_GROUP>" \
  --name "<AKS_CLUSTER_NAME>" \
  -o json > ./local-output/aks-upgrades.raw.json

./bin/kua analyze \
  --context "<AKS_CONTEXT>" \
  --format json \
  --redacted \
  --provider-source file \
  --provider-evidence ./local-output/aks-upgrades.raw.json \
  > analyze.redacted.json
```

`local-output/` is ignored by Git and is intended for private local evidence.

### 7. Generate release-candidate artifacts

```bash
scripts/release-candidate.sh 0.1.0-rc.2
```

This creates artifacts under `artifacts/release-candidate/<version>/` including:

- Linux/macOS binaries (`amd64`, `arm64`)
- `SHA256SUMS`
- `SBOM-go-modules.json`
- `provenance.json`
- `RELEASE_NOTES.md`

### 8. Validation gates before publication

```bash
scripts/ci-local.sh
go test -race ./internal/recommendation ./internal/report
git diff --check
git fsck --full --strict
```

Also ensure no AppleDouble sidecars are present before publication.

### 9. Publication control

Release publication (tag/GitHub release/artifact publication) requires explicit owner approval after Phase 9 records are complete.
