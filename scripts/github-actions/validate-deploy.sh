#!/usr/bin/env bash
set -euo pipefail

normalize_url() {
  local value="$1"
  if [[ "${value}" =~ ^https:([^/].*)$ ]]; then
    printf 'https://%s\n' "${BASH_REMATCH[1]}"
    return
  fi
  if [[ "${value}" =~ ^http:([^/].*)$ ]]; then
    printf 'http://%s\n' "${BASH_REMATCH[1]}"
    return
  fi
  printf '%s\n' "${value}"
}

websocket_probe_url() {
  local value="$1"
  if [[ "${value}" =~ ^wss://(.*)$ ]]; then
    printf 'https://%s\n' "${BASH_REMATCH[1]}"
    return
  fi
  if [[ "${value}" =~ ^ws://(.*)$ ]]; then
    printf 'http://%s\n' "${BASH_REMATCH[1]}"
    return
  fi
  printf '%s\n' "${value}"
}

BASE_URL="$(normalize_url "${BRIDGE_SERVER_URL:-${1:-https://xworkmate-bridge.svc.plus}}")"
INGRESS_URL="$(normalize_url "${ACP_INGRESS_URL:-${2:-https://acp-server.svc.plus}}")"
AUTH_TOKEN="${BRIDGE_AUTH_TOKEN:-${INTERNAL_SERVICE_TOKEN:-${3:-}}}"
OPENCLAW_HTTP_PROBE_URL="$(websocket_probe_url "${OPENCLAW_URL:-${4:-wss://openclaw.svc.plus}}")"

curl_common=(
  --silent
  --show-error
  --fail
  --location
  --max-time 20
)

probe_jsonrpc_capabilities() {
  local endpoint="$1"
  local response
  local headers=(
    -H 'Content-Type: application/json'
  )

  if [[ -n "${AUTH_TOKEN}" ]]; then
    headers+=(-H "Authorization: Bearer ${AUTH_TOKEN}")
  fi

  response="$(
    curl "${curl_common[@]}" \
      "${headers[@]}" \
      --data '{"jsonrpc":"2.0","id":"cap-1","method":"acp.capabilities"}' \
      "${endpoint}"
  )"

  grep -q '"jsonrpc":"2.0"' <<<"${response}"
  grep -Eq '"result"|"providers"' <<<"${response}"
}

probe_safe_http_endpoint() {
  local endpoint="$1"
  local status
  status="$(
    curl \
      --silent \
      --show-error \
      --output /dev/null \
      --write-out '%{http_code}' \
      --location \
      --max-time 20 \
      "${endpoint}"
  )"

  case "${status}" in
    2*|3*|401|403|404|405|426)
      return 0
      ;;
    *)
      printf 'Unexpected HTTP status %s for %s\n' "${status}" "${endpoint}" >&2
      return 1
      ;;
  esac
}

bridge_root="$(curl "${curl_common[@]}" "${BASE_URL}/")"
grep -qi 'xworkmate-bridge' <<<"${bridge_root}"

probe_safe_http_endpoint "${OPENCLAW_HTTP_PROBE_URL}"
probe_jsonrpc_capabilities "${INGRESS_URL}/codex/acp/rpc"
probe_jsonrpc_capabilities "${INGRESS_URL}/opencode/acp/rpc"
probe_jsonrpc_capabilities "${INGRESS_URL}/gemini/acp/rpc"
