# AGENTS

## Deployment Naming Rules

### Caddy config file naming

Use the following filename format for generated Caddy config files:

```text
<server-name>-<release_id>-<hostname>-<domain>.caddy
```

Example:

```text
console-6ebcdd6-jp-xhttp-contabo-console-svc-plus.caddy
```

Notes:

- `server-name` is the service or site identifier, such as `console`
- `release_id` is the normalized release identifier; use the release tag when available, otherwise use a short git commit id
- `hostname` is the target host name
- `domain` should be encoded in a filesystem-friendly form; replace `.` with `-`
