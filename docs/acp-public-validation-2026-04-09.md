# ACP Public Validation - 2026-04-09

This document records the post-deployment public validation for `xworkmate-bridge.svc.plus` and the unified ACP ingress at `acp-server.svc.plus`.

It is intended as an app-integration reference so future clients use the verified public endpoints and expected JSON-RPC methods.

## Verified Public Endpoints

### Bridge root

- URL: `https://xworkmate-bridge.svc.plus/`
- Result: `200 OK`
- Body: `xworkmate-bridge is running`

### ACP public ingress

The public ACP JSON-RPC endpoint is the `.../acp/rpc` path.

Do not send JSON-RPC requests to `.../acp` for HTTP clients.

Verified public HTTP JSON-RPC endpoints:

- Codex: `https://acp-server.svc.plus/codex/acp/rpc`
- OpenCode: `https://acp-server.svc.plus/opencode/acp/rpc`
- Gemini: `https://acp-server.svc.plus/gemini/acp/rpc`

The `.../acp` path remains reserved for WebSocket ACP.

## Auth Contract

All verified public ACP HTTP requests used:

- header: `Authorization: Bearer <INTERNAL_SERVICE_TOKEN>`
- header: `Content-Type: application/json`

Missing bearer auth returns a JSON-RPC error envelope with code `-32001`.

## Public Validation Results

### Codex

Verified `acp.capabilities` over the public ingress:

```json
{
  "method": "acp.capabilities",
  "result": {
    "providers": ["codex", "gemini", "opencode"],
    "singleAgent": true,
    "multiAgent": true
  }
}
```

Verified `session.start` reached the Codex execution layer, but the task failed upstream.

Observed upstream failure summary:

- repeated `wss://api.openai.com/v1/responses` `500 Internal Server Error`
- final `https://api.openai.com/v1/responses` `401 Unauthorized`
- message: `Missing bearer or basic authentication in header`

### OpenCode

Verified `acp.capabilities` over the public ingress:

```json
{
  "method": "acp.capabilities",
  "result": {
    "providers": ["opencode"],
    "singleAgent": true,
    "multiAgent": true
  }
}
```

Verified `session.start` end to end with prompt `Reply with exactly pong`.

Observed result:

```json
{
  "success": true,
  "provider": "opencode",
  "output": "pong"
}
```

### Gemini

Verified `acp.capabilities` over the public ingress:

```json
{
  "method": "acp.capabilities",
  "result": {
    "providers": ["gemini"],
    "singleAgent": true,
    "multiAgent": false
  }
}
```

Before the compatibility layer landed, the upstream Gemini ACP returned:

```json
{
  "success": false,
  "error": "\"Method not found\": session.start"
}
```

The adapter has now been updated so `session.start` and `session.message` default to adapter-local prompt compatibility instead of forwarding unsupported upstream methods.

## App Integration Notes

### Recommended request shape

Use JSON-RPC `POST` requests against `.../acp/rpc`.

For capability discovery:

```json
{
  "jsonrpc": "2.0",
  "id": "cap-1",
  "method": "acp.capabilities"
}
```

For single-agent task execution:

```json
{
  "jsonrpc": "2.0",
  "id": "task-1",
  "method": "session.start",
  "params": {
    "sessionId": "session-1",
    "threadId": "thread-1",
    "taskPrompt": "Reply with exactly pong",
    "workingDirectory": "/tmp",
    "routing": {
      "routingMode": "explicit",
      "explicitExecutionTarget": "singleAgent",
      "explicitProviderId": "opencode"
    }
  }
}
```

### Provider-specific notes

- `opencode` is the currently verified public task path.
- `gemini` now depends on the adapter compatibility layer, not an upstream Gemini ACP conversation method.
- `codex` public routing is healthy, but task execution requires upstream OpenAI auth to be present in the runtime environment.

## Codex Runtime Root Cause

Remote inspection on `jp-xhttp-contabo.svc.plus` showed:

- `codex-app-server` only had `HOME=/root TERM=xterm-256color NODE_NO_WARNINGS=1`
- `/root/.codex` existed, but no auth/config JSON files were present
- `codex --version` was `0.117.0`

That means the deployed Codex runtime was able to start, but it did not have a usable OpenAI auth source.

For future deployments, the systemd units should provide:

- `CODEX_HOME`
- `OPENAI_API_KEY` when API-key auth is used
- optional custom base URL variables when running against a non-default upstream
