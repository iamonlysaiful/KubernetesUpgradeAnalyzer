# KubeUpgrade Advisor

KubeUpgrade Advisor (KUA) is a planned open-source, local-first, read-only CLI for assessing whether a live Kubernetes cluster is ready to upgrade. It will combine inventory, deprecated API analysis, component compatibility, provider constraints, health checks, and explainable recommendation logic. For AKS, default `auto` mode may use the locally installed and already authenticated Azure CLI; explicit offline operation remains supported.

Core MVP foundations are now implemented through Phase 9 release-candidate tooling. Live staging validation and release publication still require explicit owner approval.

## Current status

- Product direction: approved
- Initial provider: AKS
- Analysis target: live clusters
- Deprecated API adapter: installed `kubent` binary for MVP
- Native API analyzer: planned
- Architecture baseline: documented under [`docs/`](docs/README.md)
- Implementation progress: Phases 1-8 complete, Phase 9 in progress

Follow [`AGENTS.md`](AGENTS.md) and the docs-first approval workflow.

## Real-world setup and analysis runbook

This runbook is for controlled staging validation. Do not run against production first.

### 1. Prerequisites

- Go `1.25+`
- `kubectl`
- `jq`
- Azure CLI (`az`) authenticated to the approved subscription

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

### 4. Controlled staging command sequence (read-only)

```bash
# Verify active context is the approved staging context
kubectl config current-context

# Core inventory preflight snapshot
./bin/kua inventory --format json > ./local-output/inventory-core.json

# Provider evidence snapshot (AKS read-only)
az aks get-upgrades \
	--subscription <subscription> \
	--resource-group <resource-group> \
	--name <cluster-name> \
	-o json > ./local-output/aks-upgrades.raw.json
```

Store raw outputs locally only. Do not commit raw outputs.

### 5. Generate release-candidate artifacts

```bash
scripts/release-candidate.sh 0.1.0-rc.1
```

This creates artifacts under `artifacts/release-candidate/<version>/` including:

- Linux/macOS binaries (`amd64`, `arm64`)
- `SHA256SUMS`
- `SBOM-go-modules.json`
- `provenance.json`
- `RELEASE_NOTES.md`

### 6. Validation gates before publication

```bash
scripts/ci-local.sh
go test -race ./internal/recommendation ./internal/report
git diff --check
git fsck --full --strict
```

Also ensure no AppleDouble sidecars are present before publication.

### 7. Publication control

Release publication (tag/GitHub release/artifact publication) requires explicit owner approval after Phase 9 records are complete.
