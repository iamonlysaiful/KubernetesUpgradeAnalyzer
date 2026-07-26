# Phase 9 staging approval record

Status: Draft record
Last updated: 2026-07-27

This record captures required explicit approval and execution boundaries before
any live AKS staging validation commands are run.

## 1. Approval checklist

- Owner approval date/time:
- Approved operator:
- Approved staging Kubernetes context:
- Approved Azure subscription alias:
- Approved resource group alias:
- Approved cluster alias:

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
| 1 |  |  |
| 2 |  |  |
| 3 |  |  |
| 4 |  |  |
| 5 |  |  |
| 6 |  |  |

## 3. Output handling and recovery

- Raw output local path (recoverable, not committed):
- Sanitized output local path:
- Backup/recovery location:
- Redaction reviewer:
- Approval to commit sanitized derivatives (yes/no):

## 4. Stop conditions

Stop immediately if any of the following occur:

- active context differs from approved context;
- command output contains secrets or disallowed identifiers;
- read scope exceeds approved RBAC/resource boundaries;
- command differs from approved allowlist.

## 5. Post-run validation summary

- Run started:
- Run completed:
- Commands executed exactly as approved (yes/no):
- Raw output retained locally (yes/no):
- Sanitized outputs reviewed (yes/no):
- Follow-up decisions required:

## 6. Sanitization checklist before any commit

- Remove cluster names, node names, namespace names, workload names, registry hosts.
- Remove subscription IDs, resource group names, cluster names.
- Remove tokens, kubeconfig content, authentication traces.
- Keep only approved sanitized derivatives and provenance metadata.
