# Security and privacy

Status: Proposed for implementation  
Last updated: 2026-07-21

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
- Reports may contain operational identifiers by default only when required for remediation; a future redacted report mode should be designed before external sharing.

## 5. Offline guarantee

Default execution performs no DNS lookup, HTTP request, telemetry, update check, or cloud SDK call. Kubernetes API access is local/explicit cluster access and is not treated as general internet permission. External `kubent` invocation must receive flags/environment that prevent updates or telemetry where supported; its behavior/version must be documented and tested.

## 6. External process safety

Resolve kubent from an explicit path or trusted PATH, validate its version, pass arguments without a shell, bound runtime/output, capture stdout/stderr separately, and redact diagnostics. Never interpolate resource data into a shell command.

## 7. Supply chain

Pin Go modules, commit checksums, scan dependencies/licenses, produce reproducible release metadata, verify embedded catalog checksums, and publish release checksums/signatures when release work is approved.

## 8. Threat responses

- Malformed catalog/provider evidence: reject with no recommendation.
- Oversized API/process output: enforce limits and mark assessment incomplete.
- HTML/script injection: escape at renderer boundaries.
- Symlink/output overwrite risk: validate paths and use safe atomic creation.
- Sensitive log discovery: treat as a security defect; stop sharing artifacts and notify the user.

