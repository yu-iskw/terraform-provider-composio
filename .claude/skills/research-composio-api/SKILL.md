---
name: research-composio-api
description: Deep-research Composio REST API v3.1 reference docs (auth configs, toolkits, tools, connected accounts, triggers, auth headers, errors, rate limits). Use when designing Terraform resources/data sources, extending internal/composio/api, verifying endpoint schemas, or answering Composio API questions. Triggers include "research Composio API", "Composio reference", "auth_configs endpoint", "v3.1 OpenAPI".
compatibility: Requires network access to docs.composio.dev (WebFetch or curl). Prefer .md doc URLs.
metadata:
  author: terraform-provider-composio
  version: "1.0"
  docs_entry: https://docs.composio.dev/reference
  api_base: https://backend.composio.dev/api/v3.1
---

# Research Composio API reference

## Purpose

Produce grounded, citation-backed findings from the live Composio API reference so provider and client changes match current v3.1 behavior. Do not rely on memory or stale snapshots for request/response shapes.

## When to use

Trigger phrases and situations:

- "Research the Composio auth configs API"
- "What does PATCH /auth_configs return?"
- "Map toolkit list fields for a data source"
- Before implementing or changing code under `internal/composio/api` or a new Terraform resource
- When docs and client code disagree

## Inputs to collect first

1. **Research question** — endpoint family, resource, or decision to answer.
2. **Scope** — `provider` (Terraform/control-plane only), `client` (full REST client), or `general` (any API surface). Default: `provider`.
3. **Depth** — `overview` (endpoint list + auth), `deep` (fields, status codes, edge cases), or `diff` (docs vs `internal/composio`). Default: `deep`.
4. **Output** — chat summary, or a filled [research report](assets/research-report-template.md). Default: chat summary plus citations.

## Hard rules

1. Prefer **v3.1** docs under `/reference/` (not `/reference/v3/`). Base path: `https://backend.composio.dev/api/v3.1`.
2. Fetch **live** pages. Append `.md` for agent-readable content (e.g. `https://docs.composio.dev/reference/api-reference/auth-configs.md`).
3. Start from [llms.txt](https://docs.composio.dev/llms.txt) or [reference overview](https://docs.composio.dev/reference.md) when the target page is unknown.
4. Use **current terminology** only (auth config, toolkit, tool, connected account, user_id). See [terminology.md](references/terminology.md).
5. Never invent fields, status codes, or auth headers. If a page is thin or auto-generated without schemas, say so and fetch related concept docs.
6. For this repository, respect provider boundaries in [provider-scope.md](references/provider-scope.md): do not recommend wrapping sessions, tool execute, or user OAuth as Terraform resources.

## Workflow

Follow [research-protocol.md](references/research-protocol.md). Checklist:

- [ ] **Frame** — Restate the question, scope, and success criteria.
- [ ] **Orient** — Open [doc-map.md](references/doc-map.md); pick primary + related URLs.
- [ ] **Fetch** — Load primary `.md` pages; then related auth/errors/rate-limit pages as needed.
- [ ] **Extract** — Methods, paths, headers, path/query/body params, response shape, idempotency, enable/disable vs delete.
- [ ] **Cross-check** — For `provider`/`client`/`diff` depth, compare with `internal/composio/api` and `internal/composio/models` (and existing resources if relevant).
- [ ] **Decide** — Implications for Terraform schema, client methods, tests, or docs.
- [ ] **Report** — Cite URLs; flag unknowns and follow-ups. Use [research-report-template.md](assets/research-report-template.md) when the user wants a durable artifact.

## Quick entry points

| Need                     | Start here                                                        |
| ------------------------ | ----------------------------------------------------------------- |
| API overview             | https://docs.composio.dev/reference.md                            |
| Auth headers / key types | https://docs.composio.dev/reference/authenticating-to-composio.md |
| Auth configs CRUD        | https://docs.composio.dev/reference/api-reference/auth-configs.md |
| Toolkits                 | https://docs.composio.dev/reference/api-reference/toolkits.md     |
| Errors                   | https://docs.composio.dev/reference/errors.md                     |
| Rate limits              | https://docs.composio.dev/reference/rate-limits.md                |
| Full index               | https://docs.composio.dev/llms.txt                                |

Optional helper to list current reference URLs from llms.txt:

```bash
.claude/skills/research-composio-api/scripts/list-reference-urls.sh
```

## Success criteria

- Findings cite live docs URLs (prefer `.md`).
- Endpoint paths use `/api/v3.1` (or clearly note legacy v3).
- Terminology matches v3 (no "integration" / "entity ID" in recommendations).
- Provider-scoped research explicitly separates control-plane vs runtime surfaces.
- Gaps are labeled (missing schema, ambiguous field, docs/client mismatch).

## References

- [Doc map](references/doc-map.md)
- [Research protocol](references/research-protocol.md)
- [Provider scope](references/provider-scope.md)
- [Terminology](references/terminology.md)
- [Report template](assets/research-report-template.md)

## Optional / Related

After research, implementing a Terraform resource may use the repo's resource implementation skill if available — not required to complete this skill.
