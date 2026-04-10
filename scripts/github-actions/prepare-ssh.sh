#!/usr/bin/env bash
set -euo pipefail

TARGET_HOST="${1:?target host is required}"
SSH_KNOWN_HOSTS_PAYLOAD="${2:-}"

test -n "${SINGLE_NODE_VPS_SSH_PRIVATE_KEY:-}"

mkdir -p "${HOME}/.ssh"
chmod 700 "${HOME}/.ssh"

python3 .github/scripts/normalize-private-key.py normalize > "${HOME}/.ssh/id_rsa"
chmod 600 "${HOME}/.ssh/id_rsa"
ssh-keygen -y -f "${HOME}/.ssh/id_rsa" >/dev/null

touch "${HOME}/.ssh/known_hosts"
chmod 600 "${HOME}/.ssh/known_hosts"

if [[ -n "${SSH_KNOWN_HOSTS_PAYLOAD}" ]]; then
  printf '%s\n' "${SSH_KNOWN_HOSTS_PAYLOAD}" >> "${HOME}/.ssh/known_hosts"
fi

ssh-keyscan -H "${TARGET_HOST}" >> "${HOME}/.ssh/known_hosts" 2>/dev/null || true
