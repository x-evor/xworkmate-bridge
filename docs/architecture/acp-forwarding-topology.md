# ACP Forwarding Topology

This document describes how `xworkmate-bridge.svc.plus` forwards requests to the public ACP and gateway endpoints.

## Topology

```mermaid
flowchart TD
  U[Client / App] --> B[xworkmate-bridge.svc.plus]

  B -->|HTTP POST /acp/rpc| ACPRPC[ACP HTTP RPC handler]
  B -->|WebSocket /acp| ACPWS[ACP WebSocket handler]

  ACPRPC --> R{Method?}
  ACPWS --> R

  R -->|acp.capabilities| CAP[Return available provider list]
  R -->|session.start / session.message| ENQ[Resolve routing and enqueue turn]
  R -->|session.cancel / session.close| LIFE[Session lifecycle control]
  R -->|xworkmate.providers.sync| SYNC[Sync external provider catalog]
  R -->|xworkmate.gateway.*| GWAPI[Gateway control methods]
  R -->|xworkmate.dispatch.resolve| DISPATCH[Dispatch resolution]
  R -->|xworkmate.routing.resolve| ROUTE[Routing resolution]

  ENQ --> D{Resolved execution target}
  D -->|gateway / openclaw| GW[gatewayruntime.Manager]
  D -->|singleAgent + codex| C[codex provider]
  D -->|singleAgent + opencode| O[opencode provider]
  D -->|singleAgent + gemini| G[gemini provider]

  GW --> OCLAW[wss://openclaw.svc.plus]
  C --> CODR[https://acp-server.svc.plus/codex/acp/rpc]
  O --> OPR[https://acp-server.svc.plus/opencode/acp/rpc]
  G --> GMR[https://acp-server.svc.plus/gemini/acp/rpc]

  SYNC --> CAT[providerCatalog]
  CAT --> C
  CAT --> O
  CAT --> G
```

## Request Flow

The bridge accepts ACP JSON-RPC over `POST /acp/rpc` and ACP WebSocket traffic over `/acp`.

For `session.start` and `session.message`, the server resolves routing metadata, selects either the gateway runtime or a single-agent provider, and then forwards the turn to the resolved endpoint.

For the public single-agent ACP providers, `http` and `https` endpoints are forwarded as JSON-RPC `POST .../acp/rpc` requests, while `ws` and `wss` endpoints are forwarded as WebSocket ACP sessions on `/acp`.

## Current Public Endpoints

- `wss://openclaw.svc.plus`
- `https://acp-server.svc.plus/codex/acp/rpc`
- `https://acp-server.svc.plus/opencode/acp/rpc`
- `https://acp-server.svc.plus/gemini/acp/rpc`
