# ADR-0003: External kubent adapter for MVP

Status: Accepted  
Date: 2026-07-21

## Context

Deprecated and removed API detection is essential, but building a trustworthy native analyzer would expand the initial scope.

## Decision

MVP invokes a separately installed kubent binary using a controlled, version-aware adapter. KUA normalizes kubent results into its own finding contract. A native analyzer is planned for Phase 2 behind the same interface.

## Consequences

Delivery is faster, but users must install a compatible kubent version and KUA must handle missing, incompatible, failed, or malformed tool output explicitly. External execution receives no shell interpolation and remains subject to offline/privacy controls.

