# Kubernetes preflight contract

Status: Phase 2 contract artifact
Last updated: 2026-07-23

This contract covers P2-01: kubeconfig/context resolution and read-only
preflight design. It does not authorize live cluster execution by default.

## 1. Scope

P2-01 adds:

- kubeconfig loading through standard Kubernetes client-go rules;
- current-context or explicit `--context` selection;
- a normalized preflight result with selected context, server version, discovery
  status, RBAC checks, and limitations;
- fake-client and fixture-based tests;
- clear command behavior for `kua inventory` preflight-only output during this
  package.

P2-01 does not collect full inventory, read Secrets, invoke kubent, invoke Azure
CLI, call provider APIs, render final assessment reports, or access a live
cluster without separate user approval naming the context and operation.

## 2. Kubeconfig and context behavior

Resolution order:

1. explicit `--kubeconfig`;
2. normal client-go loading rules;
3. explicit `--context`;
4. kubeconfig current context.

Missing kubeconfig, missing context, invalid context, authentication failure,
and cluster reachability failure return exit code `5` for commands that require
Kubernetes access. Usage errors remain exit code `4`.

Before any approved live read, KUA must show or log the selected context and the
fact that it will perform read-only discovery/RBAC checks. It must not show
kubeconfig contents, tokens, certificates, or raw user credentials.

## 3. Read-only checks

Approved P2-01 live preflight checks, when separately authorized, are:

- Kubernetes server version discovery;
- API group/resource discovery;
- `SelfSubjectAccessReview` or equivalent permission checks for the approved
  get/list resources;
- no object listing beyond what is needed to verify discovery and permissions.

Secrets are excluded. `watch` is excluded. Mutation verbs are excluded.

## 4. RBAC evidence classes

P2-01 classifies permissions using the approved RBAC contract:

- required: server version, discovery, namespaces, nodes, pods, workloads, PVCs,
  CRDs;
- optional: ingresses, storage classes, CSIDrivers, events.

Denied required evidence makes the affected command fail safely or become
`INCONCLUSIVE` once analyzers exist. Denied optional evidence becomes an explicit
limitation and future `UNKNOWN` findings.

## 5. Test boundary

Automated tests use fake clients, synthetic kubeconfig data, and sanitized
fixtures only. Real kubeconfig files and live clusters require separate approval
and must not be committed.

## 6. Dependency approval

P2-01 requires Kubernetes client-go modules. The dependency assessment is
recorded in `docs/dependencies/client-go.md`; adding those modules to `go.mod`
is approved only for P2-01 preflight implementation.
