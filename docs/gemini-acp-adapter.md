# Gemini ACP Adapter

This document records the verified local behavior of `gemini --experimental-acp` and the recommended adapter design for integrating Gemini into `xworkmate-bridge` as a single-agent ACP backend.

## Goal

Keep the bridge semantics unchanged:

- `openclaw gateway` stays on gateway runtime forwarding
- `codex ACP` stays a single-agent ACP backend
- `opencode ACP` stays a single-agent ACP backend
- `gemini ACP` is added as another single-agent ACP backend

Gemini should therefore be integrated through an adapter layer, not through the gateway runtime path.

## Verified Local Findings

The local Gemini CLI binary supports an experimental ACP stdio mode:

```bash
gemini --experimental-acp
```

The bridge's current ACP RPC surface is not directly compatible with Gemini's ACP mode.

The following bridge-style methods were verified to be unsupported by Gemini ACP:

- `acp.capabilities`
- `session.start`
- `tools/list`

Example response:

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"\"Method not found\": acp.capabilities","data":{"method":"acp.capabilities"}}}
```

Gemini ACP does respond to `initialize`, and `protocolVersion` is required:

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":1}}' \
  | gemini --experimental-acp
```

Observed result:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": 1,
    "authMethods": [
      {
        "id": "oauth-personal",
        "name": "Log in with Google",
        "description": null
      },
      {
        "id": "gemini-api-key",
        "name": "Use Gemini API key",
        "description": "Requires setting the `GEMINI_API_KEY` environment variable"
      },
      {
        "id": "vertex-ai",
        "name": "Vertex AI",
        "description": null
      }
    ],
    "agentCapabilities": {
      "loadSession": false,
      "promptCapabilities": {
        "image": true,
        "audio": true,
        "embeddedContext": true
      },
      "mcpCapabilities": {
        "http": true,
        "sse": true
      }
    }
  }
}
```

## Conclusion

Gemini ACP is reachable over stdio JSON-RPC, but it does not expose the same RPC methods expected by `xworkmate-bridge` today.

This means Gemini should be connected behind a small ACP adapter that:

1. speaks the bridge-facing ACP surface already used by `xworkmate-bridge`
2. speaks Gemini's experimental ACP surface on stdio

## Recommended Topology

```text
xworkmate app / agent manager
  -> xworkmate-bridge
    -> external provider: gemini
      -> gemini ACP adapter
        -> stdio child process: gemini --experimental-acp
```

The adapter should be registered in bridge provider sync as an external ACP provider:

- `providerId: "gemini"`
- `label: "Gemini"`
- `endpoint: http://127.0.0.1:<port>/acp/rpc` or `ws://127.0.0.1:<port>/acp`
- `enabled: true`

`xworkmate-bridge` already supports this provider shape through `xworkmate.providers.sync`.

## Adapter Responsibilities

The Gemini ACP adapter should own all protocol translation.

### Bridge-facing side

Expose the same single-agent ACP methods the bridge already forwards:

- `acp.capabilities`
- `session.start`
- `session.message`
- `session.cancel`
- `session.close`

### Gemini-facing side

Launch one Gemini ACP stdio child process:

```bash
gemini --experimental-acp
```

Then initialize it first:

```json
{
  "jsonrpc": "2.0",
  "id": "init-1",
  "method": "initialize",
  "params": {
    "protocolVersion": 1,
    "clientInfo": {
      "name": "xworkmate-gemini-adapter",
      "version": "0.1.0"
    }
  }
}
```

The adapter should cache the initialization result and derive bridge-side capability information from it.

## Bridge-to-Gemini Mapping

The exact post-`initialize` Gemini method surface still needs to be enumerated. Until that discovery is finished, the adapter contract should be structured as follows.

### `acp.capabilities`

Return a bridge-compatible synthesized response. Suggested response:

```json
{
  "singleAgent": true,
  "multiAgent": false,
  "providers": ["gemini"],
  "capabilities": {
    "single_agent": true,
    "multi_agent": false,
    "providers": ["gemini"]
  },
  "upstream": {
    "protocolVersion": 1,
    "promptCapabilities": {
      "image": true,
      "audio": true,
      "embeddedContext": true
    },
    "mcpCapabilities": {
      "http": true,
      "sse": true
    },
    "authMethods": [
      "oauth-personal",
      "gemini-api-key",
      "vertex-ai"
    ]
  }
}
```

### `session.start`

Suggested adapter behavior:

1. Ensure Gemini stdio process is started
2. Ensure `initialize` has succeeded
3. Create adapter-local session state keyed by `sessionId`
4. Translate the bridge prompt and metadata into the Gemini request format
5. Return a bridge-compatible response envelope

If Gemini requires a different request method than bridge `session.start`, the translation should remain fully internal to the adapter.

### `session.message`

Suggested adapter behavior:

1. Reuse adapter-local session state
2. Translate `taskPrompt`, attachments, and working directory metadata as supported
3. Stream or collect Gemini output
4. Repackage into the bridge's current single-agent result shape

### `session.cancel` and `session.close`

Because Gemini reported `loadSession: false`, the first adapter version should assume sessions are adapter-local, not durable upstream sessions.

Recommended behavior:

- `session.cancel`: cancel the current in-flight Gemini request or kill and restart the child process if fine-grained cancellation is unavailable
- `session.close`: drop adapter-local state and optionally recycle the child process

## Startup Configuration

## Environment

At minimum, support these startup modes:

### API key mode

```bash
export GEMINI_API_KEY=your_api_key
gemini --experimental-acp
```

### OAuth mode

Use the Gemini CLI's own authenticated environment, then run:

```bash
gemini --experimental-acp
```

### Vertex AI mode

Run Gemini CLI in an environment already configured for Vertex AI auth, then launch:

```bash
gemini --experimental-acp
```

## Recommended adapter process contract

Example adapter startup:

```bash
./build/bin/xworkmate-go-core gemini-acp-adapter \
  --listen 127.0.0.1:8791 \
  --gemini-bin /opt/homebrew/bin/gemini \
  --gemini-args="--experimental-acp"
```

Recommended adapter environment:

```bash
XWORKMATE_GEMINI_BIN=/opt/homebrew/bin/gemini
XWORKMATE_GEMINI_ARGS=--experimental-acp
XWORKMATE_GEMINI_INIT_PROTOCOL_VERSION=1
```

The implemented adapter exposes:

- `POST /acp/rpc`
- `GET /acp` WebSocket ACP

Supported adapter methods:

- `acp.capabilities`
- `session.start`
- `session.message`
- `session.cancel`
- `session.close`
- `gemini.initialize`
- `gemini.raw`

`session.start` and `session.message` now use a compatibility layer by default:

- the adapter still initializes Gemini ACP first so `acp.capabilities` remains grounded in the real upstream ACP surface
- if `GEMINI_ADAPTER_UPSTREAM_METHOD` is unset, session traffic runs through adapter-local prompt mode
- the adapter keeps session-local history keyed by `sessionId`
- `session.start` resets adapter-local history for that session
- `session.message` replays prior user turns plus the new turn as one prompt to the Gemini CLI
- the adapter returns a bridge-compatible single-agent payload with `output`, `provider`, `mode`, and `upstreamMethod: "prompt"`
- `session.close` drops adapter-local state

This default exists because the verified Gemini ACP upstream did not expose bridge-compatible `session.start` / `session.message` methods during testing.

If Gemini ACP later gains a compatible conversation method, you can override the forwarded method name with:

```bash
export GEMINI_ADAPTER_UPSTREAM_METHOD=your-discovered-gemini-method
```

## Bridge Provider Sync Example

Once the adapter is running, register it as a normal external provider:

```json
{
  "jsonrpc": "2.0",
  "id": "providers-sync-1",
  "method": "xworkmate.providers.sync",
  "params": {
    "providers": [
      {
        "providerId": "gemini",
        "label": "Gemini",
        "endpoint": "http://127.0.0.1:8791/acp/rpc",
        "enabled": true
      }
    ]
  }
}
```

Example local startup and sync flow:

```bash
./build/bin/xworkmate-go-core gemini-acp-adapter \
  --listen 127.0.0.1:8791 \
  --gemini-bin /opt/homebrew/bin/gemini \
  --gemini-args="--experimental-acp"
```

Then sync this provider into the bridge:

```json
{
  "jsonrpc": "2.0",
  "id": "providers-sync-gemini",
  "method": "xworkmate.providers.sync",
  "params": {
    "providers": [
      {
        "providerId": "gemini",
        "label": "Gemini",
        "endpoint": "http://127.0.0.1:8791",
        "enabled": true
      }
    ]
  }
}
```

## Recommended First Implementation Scope

The first adapter milestone should stay intentionally small:

1. Start `gemini --experimental-acp`
2. Send `initialize`
3. Expose bridge-compatible `acp.capabilities`
4. Translate one synchronous prompt path for `session.start` and `session.message`
5. Return bridge-compatible `output`, `provider`, `mode`, and `turnId`

Do not block the first version on:

- durable upstream session restore
- multimodal attachment parity
- complex streaming semantics
- full cancellation fidelity

## Next Validation Tasks

Before implementing the adapter, enumerate Gemini's post-`initialize` callable methods and request schema.

Recommended next probes:

- inspect whether Gemini sends notifications after `initialize`
- test likely conversation methods after `initialize`
- determine whether the protocol requires explicit auth selection or out-of-band auth only
- verify whether one process supports multiple sequential requests safely
- verify whether one process supports concurrent requests or must be serialized
