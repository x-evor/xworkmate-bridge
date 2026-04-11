# ACP Forwarding Topology

This document describes the bridge-only production forwarding model for `xworkmate-bridge.svc.plus`.

## Topology

```mermaid
flowchart TD
  U["xworkmate-app"] --> B["https://xworkmate-bridge.svc.plus"]

  B -->|POST /acp/rpc| RPC["ACP RPC handler"]
  B -->|WS /acp| WS["ACP WebSocket handler"]

  RPC --> R{"method"}
  WS --> R

  R -->|acp.capabilities| CAP["built-in provider catalog"]
  R -->|xworkmate.routing.resolve| ROUTE["bridge-owned routing resolve"]
  R -->|session.start / session.message| RUN["bridge-owned execution"]
  R -->|xworkmate.gateway.*| GWAPI["gateway runtime proxy"]
  R -->|session.cancel / session.close| LIFE["session lifecycle"]

  RUN --> ACP1["codex -> https://acp-server.svc.plus/codex/acp/rpc"]
  RUN --> ACP2["opencode -> https://acp-server.svc.plus/opencode/acp/rpc"]
  RUN --> ACP3["gemini -> https://acp-server.svc.plus/gemini/acp/rpc"]

  GWAPI --> GW["wss://openclaw.svc.plus"]
```

## Production Truth

Bridge owns the production map:

- `codex` -> `https://acp-server.svc.plus/codex/acp/rpc`
- `opencode` -> `https://acp-server.svc.plus/opencode/acp/rpc`
- `gemini` -> `https://acp-server.svc.plus/gemini/acp/rpc`
- gateway -> `wss://openclaw.svc.plus`

Upstream auth is bridge-internal:

- `Authorization: Bearer $INTERNAL_SERVICE_TOKEN`

## Invariants

- app-facing cloud entry is only `https://xworkmate-bridge.svc.plus`
- `acp.capabilities` returns the built-in production catalog
- no production `xworkmate.providers.sync`
- no app direct call to `acp-server.svc.plus/*`
- no app direct call to `openclaw.svc.plus`
- remote gateway runtime status is reported as `openclaw.svc.plus:443`, but the app still talks only to the bridge
