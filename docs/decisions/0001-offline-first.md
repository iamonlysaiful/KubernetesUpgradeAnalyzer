# ADR-0001: Offline-first execution

Status: Accepted  
Date: 2026-07-21

## Context

Cluster assessments may contain operationally sensitive metadata and often run in restricted environments.

## Decision

KUA performs no general outbound network access by default. It uses bundled compatibility knowledge and explicitly supplied local evidence. Any future online feature is opt-in and requires a separate approved design covering endpoints, transmitted data, caching, failure behavior, and auditability.

## Consequences

Results are reproducible and private, but catalog freshness and exact provider availability must be visible limitations. Kubernetes API access to the selected cluster remains necessary for live assessment.

