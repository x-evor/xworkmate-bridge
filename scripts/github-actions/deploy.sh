#!/usr/bin/env bash
set -euo pipefail

TARGET_HOST="${1:?target host is required}"
RUN_APPLY="${2:?run_apply flag is required}"
PLAYBOOK_DIR="${3:-playbooks}"
XWORKMATE_BRIDGE_ARTIFACT_PATH="${XWORKMATE_BRIDGE_ARTIFACT_PATH:?artifact path is required}"

cd "${PLAYBOOK_DIR}"

args=(
  ansible-playbook
  -i inventory.ini
  deploy_xworkmate_bridge_vhosts.yml
  -l "${TARGET_HOST}"
  -e "xworkmate_bridge_artifact_path=${XWORKMATE_BRIDGE_ARTIFACT_PATH}"
)

if [[ "${RUN_APPLY}" != "true" ]]; then
  args+=(-C)
fi

ANSIBLE_CONFIG="${PWD}/ansible.cfg" \
"${args[@]}"
