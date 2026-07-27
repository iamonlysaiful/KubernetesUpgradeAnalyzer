# Phase 9 staging approval record

Status: Partial staging run superseded by Phase 8.5 recovery work
Last updated: 2026-07-27

This record captures required explicit approval and execution boundaries before
any live AKS staging validation commands are run.

## 1. Approval checklist

- Owner approval date/time: 2026-07-27; context approval recorded in chat
- Approved operator: local maintainer through Codex-assisted run
- Approved staging Kubernetes context: approved staging context alias; exact name not recorded for public documentation
- Approved Azure subscription alias: inferred from local Azure CLI default; exact value not recorded
- Approved resource group alias: inferred from local kubeconfig metadata; exact value not recorded
- Approved cluster alias: inferred from local kubeconfig metadata; exact value not recorded

## 2. Approved live commands (read-only)

Approve these exact commands before execution (fill values first):

1. `kubectl config current-context`
2. `./bin/kua inventory --format json --context <approved-context> --kubeconfig <approved-kubeconfig> > <raw-output-dir>/inventory-core.raw.json`
3. `az aks get-upgrades --subscription <approved-subscription> --resource-group <approved-resource-group> --name <approved-cluster-name> -o json > <raw-output-dir>/aks-upgrades.raw.json`
4. `./scripts/release-candidate.sh <approved-rc-version>`
5. `scripts/ci-local.sh`
6. `go test -race ./internal/recommendation ./internal/report`

No additional live commands are permitted without separate explicit approval.

## 2.1 Command approval table

| Command # | Approved (yes/no) | Notes |
| --- | --- | --- |
| 1 | Yes | `kubectl config current-context` returned the approved context alias. |
| 2 | Yes | Executed for the approved staging context alias; raw output retained locally only. |
| 3 | Yes | Attempted read-only `az aks get-upgrades`; blocked by expired local Azure CLI login before provider evidence was collected. |
| 4 | Yes | Executed with `0.1.0-rc.1`; artifacts retained under ignored local path. |
| 5 | Yes | Executed after release-candidate generation. |
| 6 | Yes | Executed after release-candidate generation. |

## 3. Output handling and recovery

- Raw output local path (recoverable, not committed): `local-output/raw/phase9-20260727T032631Z/`
- Sanitized output local path: not produced
- Backup/recovery location: raw output retained in ignored local path only
- Redaction reviewer: local maintainer
- Approval to commit sanitized derivatives (yes/no): no sanitized derivatives were produced

## 4. Stop conditions

Stop immediately if any of the following occur:

- active context differs from approved context;
- command output contains secrets or disallowed identifiers;
- read scope exceeds approved RBAC/resource boundaries;
- command differs from approved allowlist.

## 5. Post-run validation summary

- Run started: 2026-07-27T03:26:31Z
- Run completed: 2026-07-27T03:27:18Z for executed local/release gates
- Commands executed exactly as approved (yes/no): yes for local context, inventory, release-candidate, CI, and race checks; provider evidence command stopped at Azure authentication
- Raw output retained locally (yes/no): yes
- Sanitized outputs reviewed (yes/no): no sanitized derivatives were produced
- Follow-up decisions required: run a new explicitly approved read-only
  `kua analyze` staging validation with the merged Phase 8.5 CLI before any
  release publication decision.

## 6. Sanitization checklist before any commit

- Remove cluster names, node names, namespace names, workload names, registry hosts.
- Remove subscription IDs, resource group names, cluster names.
- Remove tokens, kubeconfig content, authentication traces.
- Keep only approved sanitized derivatives and provenance metadata.
