#!/usr/bin/env bash
set -euo pipefail

TARGET_HOST="${1:?target host is required}"
RUN_APPLY="${2:?run_apply flag is required}"
PLAYBOOK_DIR="${3:-playbooks}"
INTERNAL_SERVICE_TOKEN="${INTERNAL_SERVICE_TOKEN:-}"

cd "${PLAYBOOK_DIR}"

args=(
  ansible-playbook
  -i inventory.ini
  deploy_xworkmate_bridge_vhosts.yml
  -l "${TARGET_HOST}"
)

if [[ -n "${INTERNAL_SERVICE_TOKEN}" ]]; then
  args+=(--vault-password-file <(printf '%s' "${INTERNAL_SERVICE_TOKEN}"))
fi

if [[ "${RUN_APPLY}" != "true" ]]; then
  args+=(-C)
fi

"${args[@]}"
