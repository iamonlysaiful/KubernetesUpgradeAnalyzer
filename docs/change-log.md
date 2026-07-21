# Documentation change log

This log records material scope and architecture changes. Git remains the detailed history.

## 2026-07-22 — Local-first provider and catalog lifecycle

- Superseded offline-by-default behavior with default `auto` AKS evidence through the local authenticated Azure CLI.
- Retained explicit `azure`, `file`, `offline`, and `none` modes and JSON evidence fallback.
- Clarified that kubeconfig supplies Kubernetes access but not Azure upgrade offerings.
- Defined repository YAML plus `go:embed` as the bundled catalog model.
- Established curated/manual review, automation-assisted proposals, optional future signed updates, and no runtime web searching or scraping.
- Reaffirmed that unknown or insufficient component compatibility evidence cannot produce `PASS`.

## 2026-07-21 — Initial architecture baseline

- Established docs-first, user-approved governance.
- Confirmed AKS as the first provider while retaining provider-neutral interfaces.
- Limited MVP analysis to live clusters.
- Selected an installed kubent binary adapter for MVP and deferred native analysis.
- Defined a bundled offline compatibility catalog proposal.
- Separated recommended destination from sequential provider-valid upgrade stages.
- Added product, architecture, recommendation, CLI/report, security, testing, process, roadmap, and open-question documents.
