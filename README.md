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
- Implementation progress: Phases 1-8.5 complete; Phase 10 (advisor model) 10.1-10.6 merged — confidence scoring, traffic-light decisions, evidence summary, upgrade plan, and advisor console output are live; 10.7 (version-specific gotchas) and 10.8 (pre-flight/day-of modes) pending; Phase 9 publication blocked pending Phase 10 completion and owner approval

Follow [`AGENTS.md`](AGENTS.md) and the docs-first approval workflow.

## Real-world setup and analysis runbook

This runbook is for operator-controlled AKS analysis. Do not run against
production first. KUA is read-only and does not mutate Kubernetes or Azure
resources.

### 1. Prerequisites

- Go `1.25+`
- `kubectl` configured with access to target cluster
- Azure CLI (`az`) authenticated to the subscription (for AKS upgrade info)
- `kubent` installed and available on `PATH`
- `jq` (optional, for inspecting JSON output)

### 2. Build and install locally

```bash
git clone <repo-url>
cd KubernetesUpgradeAnalyzer
scripts/ci-local.sh          # run tests
go build -o ./bin/kua ./cmd/kua
./bin/kua version

# Or install to /usr/local/bin:
scripts/install-local.sh
```

### 3. Quick start (interactive mode)

The simplest way to analyze a cluster:

```bash
kua analyze
```

KUA will:
1. Detect your current kubectl context and prompt for confirmation
2. Collect cluster inventory and run health/API/component checks
3. Prompt for any unknown component versions
4. Prompt for target version if AKS upgrade info is unavailable
5. Display the assessment result

### 4. Scripted/CI mode (non-interactive)

For automation, use `--yes` to skip prompts and provide all required flags:

```bash
az login
az account set --subscription "<SUBSCRIPTION_ID>"

./bin/kua analyze \
  --yes \
  --context "<AKS_CONTEXT>" \
  --format json \
  --redacted \
  --provider-source azure \
  --subscription "<SUBSCRIPTION_ID>" \
  --resource-group "<RESOURCE_GROUP>" \
  --cluster-name "<AKS_CLUSTER_NAME>" \
  > analyze.redacted.json
```

### 5. Assessment states

Console output leads with a confidence-based traffic-light decision:

| Decision | Confidence | Meaning |
|----------|------------|---------|
| 🟢 `GO` (PROCEED WITH UPGRADE) | ≥90% (balanced profile) | No blockers; confidence factors all strong |
| 🟡 `GO_WITH_CAUTION` (PROCEED WITH CAUTION) | 70-89% | No blockers, but review warnings/evidence gaps first |
| 🔴 `DO_NOT_PROCEED` (DO NOT PROCEED) | <70% or any blocker | Fix blockers or gather more evidence before upgrading |

Confidence is a weighted score (API compatibility, component compatibility,
cluster health, provider evidence, storage health, analysis coverage) that
drops whenever evidence is missing, ambiguous, or a provider/API check could
not run — see [`docs/recommendation-model.md`](docs/recommendation-model.md).
Each finding also carries a plain-language impact, action, and consequence
if ignored, and the console report includes an evidence summary and
step-by-step upgrade plan when a destination version is available.

The legacy `readiness`/`risk` fields (`READY` / `READY_WITH_WARNINGS` /
`NOT_READY` / `INCONCLUSIVE`, `LOW` / `MEDIUM` / `HIGH` / `UNKNOWN`) remain in
the JSON output for backward reference alongside `decision` and `confidence`.

### 6. Component version overrides

If component versions cannot be detected automatically, you have two options:

**Option A: Interactive prompts (default)**

When running `kua analyze` interactively, KUA prompts for unknown versions.

**Option B: Override file (for scripting)**

```bash
# Generate override template from a previous assessment:
./bin/kua component-overrides \
  --input analyze.redacted.json \
  --output component-overrides.json

# Edit component-overrides.json with correct versions, then rerun:
./bin/kua analyze \
  --yes \
  --context "<AKS_CONTEXT>" \
  --component-overrides component-overrides.json \
  > analyze.final.json
```

### 7. Offline provider evidence

If Azure CLI access is unavailable during analysis, pre-collect upgrade info:

```bash
# Collect AKS upgrade evidence separately:
az aks get-upgrades \
  --subscription "<SUBSCRIPTION_ID>" \
  --resource-group "<RESOURCE_GROUP>" \
  --name "<AKS_CLUSTER_NAME>" \
  -o json > ./local-output/aks-upgrades.json

# Run analysis with file-based provider evidence:
./bin/kua analyze \
  --provider-source file \
  --provider-evidence ./local-output/aks-upgrades.json
```

### 8. Output formats

```bash
kua analyze --format console   # Default: human-readable
kua analyze --format json      # Machine-readable
kua analyze --format markdown  # Documentation
kua analyze --format html      # Self-contained report

# Add --redacted to mask sensitive identifiers:
kua analyze --format json --redacted > report.json
```

## Development

See [`docs/`](docs/README.md) for architecture, contracts, and development process.

```bash
scripts/ci-local.sh            # Run all quality checks
go test -race ./...            # Run tests with race detector
```

## License

Apache License 2.0
