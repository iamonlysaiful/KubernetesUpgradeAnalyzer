# ADR-0004: Separate destination from upgrade path

Status: Accepted  
Date: 2026-07-21

## Context

An AKS staging assessment may identify Kubernetes `1.33.12` as the desired destination from `1.30.0`. Upstream Kubernetes and AKS guidance prohibit skipping minor versions during supported upgrades.

## Decision

KUA reports both:

1. the recommended final destination; and
2. the ordered provider-valid upgrade stages required to reach it.

Thus `1.33.12` may be the destination for a `1.30.x` cluster, while the plan includes `1.31`, `1.32`, and `1.33.12` stages with exact patches governed by AKS availability evidence.

## Consequences

The recommendation matches long-term intent without implying an unsupported one-step operation. Compatibility and API checks must evaluate every stage, not only the destination.

## References

- <https://kubernetes.io/releases/version-skew-policy/>
- <https://learn.microsoft.com/azure/aks/upgrade-aks-control-plane>
- <https://learn.microsoft.com/azure/aks/supported-kubernetes-versions>

