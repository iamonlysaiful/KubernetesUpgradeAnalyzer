# ADR-0002: AKS-first provider scope

Status: Accepted  
Date: 2026-07-21

## Context

The product aims to be cloud agnostic, while the initial real validation cluster is AKS. Building every provider at once would dilute the MVP.

## Decision

Implement AKS provider analysis first. Keep core snapshots, findings, recommendation policy, and provider interfaces independent of Azure-specific clients and types.

## Consequences

MVP can validate a real use case sooner. EKS, GKE, OpenShift, and vanilla Kubernetes provider-specific recommendations remain later work, while generic inventory/health logic remains reusable.

