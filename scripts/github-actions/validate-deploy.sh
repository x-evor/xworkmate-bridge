#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-https://xworkmate-bridge.svc.plus}"
INGRESS_URL="${2:-https://acp-server.svc.plus}"

bridge_root="$(curl -fsS "${BASE_URL}/")"
test "${bridge_root}" = "xworkmate-bridge is running"

codex_root="$(curl -fsS "${INGRESS_URL}/codex")"
test "${codex_root}" = "xworkmate-bridge is running"

codex_rpc="$(
  curl -sS "${INGRESS_URL}/codex/acp/rpc" \
    -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","id":"cap-1","method":"acp.capabilities"}'
)"
opencode_rpc="$(
  curl -sS "${INGRESS_URL}/opencode/acp/rpc" \
    -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","id":"cap-1","method":"acp.capabilities"}'
)"
gemini_rpc="$(
  curl -sS "${INGRESS_URL}/gemini/acp/rpc" \
    -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","id":"cap-1","method":"acp.capabilities"}'
)"

grep -q '"missing bearer authorization"' <<<"${codex_rpc}"
grep -q '"missing bearer authorization"' <<<"${opencode_rpc}"
grep -q '"providers"' <<<"${gemini_rpc}"
