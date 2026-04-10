#!/usr/bin/env bash
set -euo pipefail

TARGET_HOST="${1:?target host is required}"
RUN_APPLY="${2:?run_apply flag is required}"
PLAYBOOK_DIR="${3:-playbooks}"

cd "${PLAYBOOK_DIR}"

temp_config="$(mktemp)"
trap 'rm -f "${temp_config}"' EXIT

awk '
  BEGIN { skip = 0 }
  /^[[:space:]]*vault_password_file[[:space:]]*=/ { skip = 1; next }
  { print }
' ansible.cfg > "${temp_config}"

args=(
  ansible-playbook
  -i inventory.ini
  deploy_xworkmate_bridge_vhosts.yml
  -l "${TARGET_HOST}"
)

if [[ "${RUN_APPLY}" != "true" ]]; then
  args+=(-C)
fi

ANSIBLE_CONFIG="${temp_config}" \
ANSIBLE_VAULT_PASSWORD_FILE="" \
"${args[@]}"
