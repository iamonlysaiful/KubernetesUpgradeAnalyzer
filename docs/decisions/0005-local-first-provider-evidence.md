# ADR-0005: Local-first provider evidence and catalog lifecycle

Status: Accepted  
Date: 2026-07-22

## Context

Operators already use local kubeconfig contexts and an authenticated Azure CLI to manage AKS clusters. Kubeconfig provides Kubernetes API connectivity but not Azure Resource Manager's cluster-specific upgrade offerings. Separately, live cluster metadata can reveal installed component versions but cannot prove their supported Kubernetes ranges.

## Decision

KUA is local-first rather than offline-by-default:

- Kubernetes analysis uses standard local kubeconfig resolution and the selected current/explicit context.
- Provider source defaults to `auto`. For detected AKS clusters, KUA invokes an installed, already authenticated Azure CLI using an allowlisted read-only `az aks get-upgrades` operation.
- `auto` falls back to a supplied JSON export and then to `UNKNOWN` provider availability without discarding independent Kubernetes findings.
- Explicit `azure`, `file`, `offline`, and `none` modes provide strict behavior. Offline mode makes no provider network call.
- KUA never initiates Azure login, changes subscriptions, or mutates provider resources.
- Compatibility knowledge is stored as reviewed YAML in the repository, validated and embedded into the binary. Runtime assessment never searches or scrapes the internet.
- Future catalog update commands may download only approved, signed catalog releases. Automation may propose catalog changes, but human review is required before publication.
- Missing, stale, ambiguous, conflicting, or unbounded component compatibility evidence never yields `PASS`.

## Consequences

Default AKS assessments can use current, cluster-specific upgrade availability with minimal user preparation. They are no longer guaranteed to be network-silent unless `offline` is selected. Azure authentication and identity-resolution failures become explicit limitations. Embedded catalogs retain deterministic component decisions, while catalog maintenance remains an ongoing governed responsibility.

## References

- <https://kubernetes.io/docs/concepts/overview/kubectl/>
- <https://learn.microsoft.com/azure/aks/upgrade-aks-control-plane?tabs=azure-cli>
- <https://learn.microsoft.com/cli/azure/aks#az-aks-get-upgrades>
