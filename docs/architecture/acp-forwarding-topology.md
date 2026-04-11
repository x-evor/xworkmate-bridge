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

## Three-Layer View

This view separates what the app sees, what the bridge owns, and what the
real upstream production targets are.

```mermaid
flowchart LR
    subgraph L1["APP 视角"]
        APP["xworkmate-app"]
        APPACP["ACP 能力发现<br/>acp.capabilities"]
        APPGW["Gateway 连接<br/>xworkmate.gateway.connect"]
        APP --> APPACP
        APP --> APPGW
    end

    subgraph L2["Bridge 视角"]
        BRIDGE["xworkmate-bridge<br/>唯一上游发现真源"]

        CAP["Bridge-owned ACP server list"]
        CAP1["codex"]
        CAP2["opencode"]
        CAP3["gemini"]

        GW["Bridge-owned gateway upstream"]
        GW1["remote mode -> openclaw"]

        BRIDGE --> CAP
        CAP --> CAP1
        CAP --> CAP2
        CAP --> CAP3

        BRIDGE --> GW
        GW --> GW1
    end

    subgraph L3["上游视角"]
        U1["https://acp-server.svc.plus/codex/acp/rpc"]
        U2["https://acp-server.svc.plus/opencode/acp/rpc"]
        U3["https://acp-server.svc.plus/gemini/acp/rpc"]
        U4["wss://openclaw.svc.plus<br/>reported as openclaw.svc.plus:443"]
    end

    APPACP --> BRIDGE
    APPGW --> BRIDGE

    CAP1 --> U1
    CAP2 --> U2
    CAP3 --> U3
    GW1 --> U4
```

Important distinction:

- `acp.capabilities.providerCatalog` currently advertises only the ACP
  single-agent providers: `codex`, `opencode`, and `gemini`
- `gateway` is not part of that provider catalog; it is exposed through the
  separate `xworkmate.gateway.*` bridge-owned runtime path
- for remote gateway mode, the bridge rewrites the upstream target to
  `wss://openclaw.svc.plus`

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
