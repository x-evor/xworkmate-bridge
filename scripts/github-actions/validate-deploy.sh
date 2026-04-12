#!/usr/bin/env bash
set -euo pipefail

IMAGE_REF="${1:?image_ref is required}"

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

image_ref="$(printf '%s' "${IMAGE_REF}" | tr -d '\n' | xargs)"
if [[ -z "${image_ref}" ]]; then
  echo "image_ref is required" >&2
  exit 1
fi

image_no_digest="${image_ref%@*}"
tag="${image_no_digest##*:}"
if [[ "${image_no_digest}" == "${tag}" ]]; then
  tag=""
fi

commit=""
version="${tag}"

if [[ "${tag}" =~ ^[0-9a-f]{40}$ ]]; then
  commit="${tag}"
fi

BASE_URL="$(normalize_url "${BRIDGE_SERVER_URL:-${2:-https://xworkmate-bridge.svc.plus}}")"
OPENCLAW_HTTP_PROBE_URL="$(websocket_probe_url "${OPENCLAW_URL:-${3:-wss://openclaw.svc.plus}}")"
CODEX_RPC_URL="$(normalize_url "${CODEX_RPC_URL:-${4:-https://acp-server.svc.plus/codex/acp/rpc}}")"
OPENCODE_RPC_URL="$(normalize_url "${OPENCODE_RPC_URL:-${5:-https://acp-server.svc.plus/opencode/acp/rpc}}")"
GEMINI_RPC_URL="$(normalize_url "${GEMINI_RPC_URL:-${6:-https://acp-server.svc.plus/gemini/acp/rpc}}")"
AUTH_TOKEN="${BRIDGE_AUTH_TOKEN:-${INTERNAL_SERVICE_TOKEN:-${7:-}}}"

curl_common=(
  --silent
  --show-error
  --fail
  --location
  --max-time 20
)

auth_headers=()
if [[ -n "${AUTH_TOKEN}" ]]; then
  auth_headers+=(-H "Authorization: Bearer ${AUTH_TOKEN}")
fi

probe_jsonrpc_capabilities() {
  local endpoint="$1"
  local response
  local headers=(
    -H 'Content-Type: application/json'
  )

  headers+=("${auth_headers[@]}")

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
      "${auth_headers[@]}" \
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

ping_json="$(
  curl \
    "${curl_common[@]}" \
    "${BASE_URL}/api/ping"
)"

PING_JSON="${ping_json}" python3 - "${image_ref}" "${tag}" "${commit}" "${version}" <<'PY'
import json
import os
import sys

image_ref, tag, commit, version = sys.argv[1:5]
payload = json.loads(os.environ["PING_JSON"])

if payload.get("status") != "ok":
    raise SystemExit("ping status not ok")

if payload.get("image") != image_ref:
    raise SystemExit(f"expected image {image_ref!r}, got {payload.get('image')!r}")

if tag and payload.get("tag") != tag:
    raise SystemExit(f"expected tag {tag!r}, got {payload.get('tag')!r}")

if commit and payload.get("commit") != commit:
    raise SystemExit(f"expected commit {commit!r}, got {payload.get('commit')!r}")

if version and payload.get("version") != version:
    raise SystemExit(f"expected version {version!r}, got {payload.get('version')!r}")
PY

bridge_root="$(curl "${curl_common[@]}" "${auth_headers[@]}" "${BASE_URL}/")"
grep -qi 'xworkmate-bridge' <<<"${bridge_root}"

probe_safe_http_endpoint "${OPENCLAW_HTTP_PROBE_URL}"
probe_jsonrpc_capabilities "${CODEX_RPC_URL}"
probe_jsonrpc_capabilities "${OPENCODE_RPC_URL}"
probe_jsonrpc_capabilities "${GEMINI_RPC_URL}"
