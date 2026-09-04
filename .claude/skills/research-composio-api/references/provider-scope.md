# Provider scope (terraform-provider-composio)

This repository manages **durable Composio control-plane objects**. It does not model user OAuth sessions, Tool Router sessions, or tool execution.

Source of truth for product boundaries: repo `AGENTS.md` and `internal/composio/README.md`.

## In scope for Terraform resources / data sources

Prefer APIs that are project-scoped, long-lived, and CRUD-friendly:

| Surface                           | Why                                                                            |
| --------------------------------- | ------------------------------------------------------------------------------ |
| Auth configs                      | Blueprint for how a toolkit authenticates; already `composio_auth_config`      |
| Custom toolkits (Custom MCP)      | Register remote MCP servers as `CUSTOM_*` toolkits (`composio_custom_toolkit`) |
| Toolkits (read)                   | Catalog / metadata for data sources                                            |
| Projects / org (careful)          | Only if clearly admin control-plane and safe to automate                       |
| Webhook endpoints / subscriptions | Possible future resources if durable and non-user-session                      |
| Triggers (config)                 | Possible if configuration is durable; not event delivery runtime               |

## Out of scope as Terraform resources

Do **not** recommend wrapping these as resources even if the REST API exists:

| Surface                          | Reason                           |
| -------------------------------- | -------------------------------- |
| Tool Router / sessions           | Runtime agent sessions           |
| Tools execute / search-as-action | Runtime execution                |
| Connected accounts               | Per-user OAuth connections       |
| Legacy `/mcp/servers` API        | Deprecated; replaced by sessions |
| MCP session URLs / headers       | Runtime transport                |
| Files / workbench / remote bash  | Runtime sandbox                  |

Researching these APIs is still valid for **client completeness**, threat modeling, or explaining why they stay out of Terraform — label findings **runtime / out of provider scope**.

## Auth expectations in this codebase

From `internal/composio/README.md`:

- Default origin: `https://backend.composio.dev`
- Path prefix: `/api/v3.1` (pinned)
- Project calls: `x-api-key`
- Organization calls: `x-org-api-key`
- Retries: GET, PUT, PATCH, DELETE — not POST create
- Error bodies: redact secrets by key name

When researching auth configs, assume **project API key** unless docs say otherwise.

## Research modes mapped to scope

| Mode       | Include runtime APIs?    | Compare to `internal/composio`? |
| ---------- | ------------------------ | ------------------------------- |
| `provider` | Only to exclude / warn   | Yes, for candidate resources    |
| `client`   | Yes                      | Yes                             |
| `general`  | Yes                      | Optional                        |
| `diff`     | As needed for the target | Required                        |
