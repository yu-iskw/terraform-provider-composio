# Composio API reference doc map

Stable navigation for research. Always prefer **v3.1** (paths without `/reference/v3/`). Confirm URLs against https://docs.composio.dev/llms.txt if a link 404s.

## Canonical bases

| Item              | Value                                                               |
| ----------------- | ------------------------------------------------------------------- |
| Human overview    | https://docs.composio.dev/reference                                 |
| Agent overview    | https://docs.composio.dev/reference.md                              |
| Agent index       | https://docs.composio.dev/llms.txt                                  |
| Full dump (large) | https://docs.composio.dev/llms-full.txt                             |
| REST base URL     | `https://backend.composio.dev/api/v3.1`                             |
| Legacy REST base  | `https://backend.composio.dev/api/v3` (supported; not for new work) |

Fetch pattern: for any docs path, append `.md` (e.g. `/reference/api-reference/toolkits` → `.../toolkits.md`).

## Cross-cutting (read first for any deep dive)

| Topic                       | URL                                                                                           |
| --------------------------- | --------------------------------------------------------------------------------------------- |
| Authenticating              | https://docs.composio.dev/reference/authenticating-to-composio.md                             |
| Project API key permissions | https://docs.composio.dev/reference/authenticating-to-composio/project-api-key-permissions.md |
| Errors                      | https://docs.composio.dev/reference/errors.md                                                 |
| Rate limits                 | https://docs.composio.dev/reference/rate-limits.md                                            |
| Glossary                    | https://docs.composio.dev/reference/glossary.md                                               |

## REST API families (v3.1)

| Family                  | URL                                                                          | Typical use in this provider                                            |
| ----------------------- | ---------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| Auth configs            | https://docs.composio.dev/reference/api-reference/auth-configs.md            | Resource CRUD (`composio_auth_config`)                                  |
| Toolkits                | https://docs.composio.dev/reference/api-reference/toolkits.md                | Data source + Custom MCP upsert/sync/delete (`composio_custom_toolkit`) |
| Tools                   | https://docs.composio.dev/reference/api-reference/tools.md                   | Rarely as resources; may inform schemas                                 |
| Connected accounts      | https://docs.composio.dev/reference/api-reference/connected-accounts.md      | Out of provider resource scope (user OAuth)                             |
| Triggers                | https://docs.composio.dev/reference/api-reference/triggers.md                | Possible future control-plane; not runtime execute                      |
| Tool Router             | https://docs.composio.dev/reference/api-reference/tool-router.md             | Sessions — out of provider resource scope                               |
| Projects                | https://docs.composio.dev/reference/api-reference/projects.md                | Org/project management                                                  |
| Organization            | https://docs.composio.dev/reference/api-reference/organization.md            | Org-scoped ops (`x-org-api-key`)                                        |
| Organization management | https://docs.composio.dev/reference/api-reference/organization-management.md | Org admin                                                               |
| API keys                | https://docs.composio.dev/reference/api-reference/api-keys.md                | Key lifecycle                                                           |
| MCP                     | https://docs.composio.dev/reference/api-reference/mcp.md                     | Legacy servers — deprecated; do not wrap                                |
| Files                   | https://docs.composio.dev/reference/api-reference/files.md                   | Usually out of scope                                                    |
| Logs                    | https://docs.composio.dev/reference/api-reference/logs.md                    | Observability                                                           |
| Webhook endpoints       | https://docs.composio.dev/reference/api-reference/webhook-endpoints.md       | Possible future resources                                               |
| Webhook events          | https://docs.composio.dev/reference/api-reference/webhook-events.md          | Events catalog                                                          |
| Webhook subscriptions   | https://docs.composio.dev/reference/api-reference/webhook-subscriptions.md   | Possible future resources                                               |

## Concept docs that clarify auth configs / toolkits

Useful when the OpenAPI-generated API page lacks narrative:

| Topic                       | URL                                                                        |
| --------------------------- | -------------------------------------------------------------------------- |
| Authentication overview     | https://docs.composio.dev/docs/authentication.md                           |
| Programmatic auth configs   | https://docs.composio.dev/docs/authentication/programmatic-auth-configs.md |
| Custom vs managed app       | https://docs.composio.dev/docs/authentication/custom-app-vs-managed-app.md |
| Controlling scopes          | https://docs.composio.dev/docs/authentication/controlling-scopes.md        |
| Custom MCP (remote servers) | https://docs.composio.dev/docs/extending-sessions/custom-mcp.md            |
| How Composio works          | https://docs.composio.dev/docs/how-composio-works.md                       |

## Legacy v3 (compare only)

Prefix: `https://docs.composio.dev/reference/v3/` — same family names under `api-reference/`. Do not use for new client paths.

## SDK reference (secondary)

Use SDKs only to clarify intent when REST pages are incomplete. Prefer REST for Go client design.

- TypeScript: https://docs.composio.dev/reference/sdk-reference/typescript.md
- Python: https://docs.composio.dev/reference/sdk-reference/python.md
- Auth configs (TS): https://docs.composio.dev/reference/sdk-reference/typescript/auth-configs.md
- Auth configs (Python): https://docs.composio.dev/reference/sdk-reference/python/auth-configs.md

## Local code anchors

| Area                | Path                                    |
| ------------------- | --------------------------------------- |
| Client README       | `internal/composio/README.md`           |
| HTTP client         | `internal/composio/api/client.go`       |
| Auth configs client | `internal/composio/api/auth_configs.go` |
| Toolkits client     | `internal/composio/api/toolkits.go`     |
| Models              | `internal/composio/models/`             |
| Provider resources  | `internal/provider/`                    |
