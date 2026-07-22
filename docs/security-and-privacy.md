# Security and privacy

Status: Proposed for implementation  
Last updated: 2026-07-22

## 1. Trust boundaries

Treat kubeconfig, Kubernetes API responses, image names/tags, labels, annotations, events, kubent output, provider evidence files, catalogs, configuration, and report paths as untrusted input.

## 2. Read-only guarantee

KUA uses GET/LIST/WATCH only; MVP should avoid WATCH unless justified. It never PATCHes, POSTs, PUTs, DELETEs, evicts, drains, scales, restarts, or upgrades resources. Kubernetes impersonation is not performed unless explicitly supplied through existing client configuration and approved later.

## 3. RBAC design

Before implementation, generate and review a least-privilege ClusterRole listing only required read verbs/resources. Secrets are excluded. The tool must report forbidden resources as evidence gaps. It must not ask for `cluster-admin` as its normal setup.

## 4. Data minimization

- Never request or serialize Secret objects.
- Do not collect ConfigMap contents by default.
- Do not collect environment variable values, volume secret references beyond sanitized type/presence, pod logs, exec output, or full resource specs.
- Sanitize cluster UID/name, namespaces, workload names, node names, registry hosts, subscription/resource-group IDs, and event messages in fixtures and support bundles.
- Reports may contain operational identifiers locally when required for remediation. MVP redacted mode is mandatory before external sharing and must use stable per-assessment aliases without changing findings or decisions.

## 5. Network and offline guarantees

Default `auto` mode may access the selected Kubernetes API and invoke the local Azure CLI for read-only AKS provider evidence. It performs no telemetry, arbitrary HTTP requests, update check, catalog download, vendor search, or page scraping. `--provider-source=offline` guarantees no Azure/provider network invocation; Kubernetes API access remains necessary for live analysis. External kubent invocation uses version `0.7.3`, JSON output, and `--helm3=false`; it must receive flags/environment that prevent updates or telemetry where supported, and its behavior must be tested.

Before invoking Azure CLI, KUA shows/logs the sanitized target identity and operation category. It uses the existing Azure authentication context, never initiates login, never changes the active subscription, never persists tokens, and never passes credentials on command arguments.

## 6. External process safety

Resolve kubent and Azure CLI from explicit paths or trusted PATH, validate versions, pass arguments without a shell, bound runtime/output, capture stdout/stderr separately, and redact diagnostics. Permit only an allowlisted read-only Azure command shape. Never interpolate resource data into a shell command.

## 7. Supply chain

Pin Go modules, commit checksums, scan dependencies/licenses, produce reproducible release metadata, verify embedded catalog checksums, and publish release checksums/signatures when release work is approved.

## 8. Threat responses

- Malformed catalog/provider evidence: reject with no recommendation.
- Oversized API/process output: enforce limits and mark assessment incomplete.
- HTML/script injection: escape at renderer boundaries.
- Symlink/output overwrite risk: validate paths and use safe atomic creation.
- Sensitive log discovery: treat as a security defect; stop sharing artifacts and notify the user.
