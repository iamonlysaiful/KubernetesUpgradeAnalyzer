#!/usr/bin/env bash
set -euo pipefail

# Guarded helper for Phase 9 staging validation preparation.
# This script validates required inputs and prints exact commands to run.
# It does not execute live kubectl/az commands.

approved_context="${APPROVED_CONTEXT:-}"
approved_kubeconfig="${APPROVED_KUBECONFIG:-}"
approved_subscription="${APPROVED_SUBSCRIPTION:-}"
approved_resource_group="${APPROVED_RESOURCE_GROUP:-}"
approved_cluster_name="${APPROVED_CLUSTER_NAME:-}"
raw_output_dir="${RAW_OUTPUT_DIR:-./local-output/raw}"
rc_version="${RC_VERSION:-0.1.0-rc.manual}"

missing=0
for v in \
  approved_context approved_kubeconfig approved_subscription \
  approved_resource_group approved_cluster_name
 do
  if [[ -z "${!v}" ]]; then
    echo "missing required variable: ${v}" >&2
    missing=1
  fi
 done

if [[ "$missing" -ne 0 ]]; then
  echo "set required environment variables and re-run" >&2
  exit 1
fi

mkdir -p "$raw_output_dir"

cat <<EOF
Phase 9 approved command sequence (copy/paste after explicit approval):

1) kubectl config current-context
2) ./bin/kua inventory --format json --context "${approved_context}" --kubeconfig "${approved_kubeconfig}" > "${raw_output_dir}/inventory-core.raw.json"
3) az aks get-upgrades --subscription "${approved_subscription}" --resource-group "${approved_resource_group}" --name "${approved_cluster_name}" -o json > "${raw_output_dir}/aks-upgrades.raw.json"
4) ./scripts/release-candidate.sh "${rc_version}"
5) scripts/ci-local.sh
6) go test -race ./internal/recommendation ./internal/report

Reminder:
- Do not commit raw outputs.
- Sanitize derived files before any commit.
- Record results in docs/validation/phase-9-approval-record.md and docs/validation/phase-9-release-record.md.
EOF
