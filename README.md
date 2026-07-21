# KubeUpgrade Advisor

KubeUpgrade Advisor (KUA) is a planned open-source, local-first, read-only CLI for assessing whether a live Kubernetes cluster is ready to upgrade. It will combine inventory, deprecated API analysis, component compatibility, provider constraints, health checks, and explainable recommendation logic. For AKS, default `auto` mode may use the locally installed and already authenticated Azure CLI; explicit offline operation remains supported.

This repository is currently in the architecture phase. It intentionally contains no application implementation.

## Current status

- Product direction: approved
- Initial provider: AKS
- Analysis target: live clusters
- Deprecated API adapter: installed `kubent` binary for MVP
- Native API analyzer: planned
- Architecture baseline: documented under [`docs/`](docs/README.md)

No implementation work is authorized merely by the presence of this plan. Follow [`AGENTS.md`](AGENTS.md) and the docs-first approval workflow.
