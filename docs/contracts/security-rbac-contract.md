# Security, RBAC, and dependency contract

Status: Phase 0 contract artifact
Last updated: 2026-07-22

This contract locks the MVP safety boundary before collectors, external process
adapters, renderers, or release tooling are implemented.

## 1. Kubernetes RBAC matrix

KUA uses Kubernetes read operations only. MVP collectors use `get` and `list`.
`watch` is not approved for MVP.

| API group | Resources | Verbs | Evidence class | Secret risk |
| --- | --- | --- | --- | --- |
| core | namespaces, nodes, pods, services, persistentvolumeclaims, events | get, list | required for live analysis | no Secret payloads |
| apps | deployments, daemonsets, statefulsets, replicasets | get, list | required for workload and health analysis | no Secret payloads |
| batch | jobs, cronjobs | get, list | required for workload and health analysis | no Secret payloads |
| networking.k8s.io | ingresses | get, list | optional networking evidence | no Secret payloads |
| storage.k8s.io | storageclasses, csidrivers | get, list | optional storage/CSI evidence | no Secret payloads |
| apiextensions.k8s.io | customresourcedefinitions | get, list | required for component/API evidence | no Secret payloads |
| discovery | server groups/resources/version | get | required for preflight and API availability | no Secret payloads |

Forbidden resources become limitations. Required evidence denial stops the
affected command or produces `INCONCLUSIVE`; optional denial produces `UNKNOWN`
findings with reduced confidence. Normal setup never asks for `cluster-admin`.

Secrets are explicitly excluded. ConfigMap contents are excluded from MVP
collection unless a later approved contract names specific non-sensitive keys.

## 2. External command allowlists

KUA invokes external processes directly without a shell, with bounded runtime and
output.

### Kubent

Approved MVP command shape:

```text
kubent --output json --helm3=false --target-version <minor-or-version> [kubeconfig/context flags]
```

Additional requirements:

- supported version is `0.7.3`;
- target-rule coverage must be verified before empty output can mean no API
  findings;
- malformed output, nonzero execution, missing binary, unsupported version, or
  missing rule coverage produces `INCONCLUSIVE`;
- stderr and diagnostics are redacted before logs or reports.

### Azure CLI

Approved AKS upgrade evidence command:

```text
az aks get-upgrades --resource-group <name> --name <cluster> [--subscription <id>] --output json
```

Approved read-only supporting operations are limited to version/account/context
inspection needed to determine whether the local CLI can run the command.

Prohibited operations include `az login`, browser or device-flow auth, `az
account set`, `az aks upgrade`, extension installation, token export, provider
mutation, and arbitrary shell/query text. `offline` mode must not start any
Azure CLI process.

## 3. Redaction matrix

| Data class | Examples | Local default | Redacted mode |
| --- | --- | --- | --- |
| Product metadata | KUA version, catalog version, Kubernetes version | show | show |
| Operational identifiers | cluster/context names, namespaces, workloads, nodes, node pools | show | stable aliases |
| Cloud identifiers | subscription, resource group, AKS name, region when sensitive | show if needed | stable aliases, region configurable |
| Registry identifiers | private registry hosts and image paths | show if needed | stable aliases |
| Event text | warning messages and object references | show sanitized | aliases and sensitive text redaction |
| Prohibited secrets | Secret data, tokens, certs, passwords, env values, kubeconfig contents | never collect | never collect |

Redaction preserves versions, counts, finding IDs, readiness, risk, destination,
stage order, and decision trace semantics.

## 4. Dependency policy

Every new module, tool, or release dependency requires user approval and a
recorded assessment before installation or commit. The assessment must include:

- purpose and package/tool name;
- version and checksum or lockfile impact;
- license compatibility with Apache-2.0 distribution;
- maintenance status and upstream trust;
- security/vulnerability review;
- rejected alternatives and reason;
- whether it runs locally, in CI, or in release packaging.

The approved planning baseline names Go, Cobra, Viper, `client-go`, schema
validation tooling, lint tooling, and kubent `0.7.3` as expected candidates, but
that baseline is not permission to install or add them before the relevant phase.

## 5. Threat responses

| Threat | Required behavior |
| --- | --- |
| RBAC denial | Record limitation and produce `UNKNOWN` or `INCONCLUSIVE` according to evidence class. |
| Secret exposure in input/output | Stop sharing artifacts, report incident, and wait for owner review. |
| Malformed catalog/provider/kubent JSON | Reject that evidence and avoid compatibility success claims. |
| Oversized process/API output | Stop affected adapter, record limitation, and continue only where safe. |
| Shell/argument injection attempt | Pass arguments as separate values; reject unsafe command shapes. |
| Output overwrite or symlink risk | Refuse unsafe writes unless a later approved `--force` contract covers the case. |
| HTML/Markdown injection | Escape renderer output and never execute user-provided markup. |
