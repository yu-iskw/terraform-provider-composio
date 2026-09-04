# Research protocol

Step-by-step deep research against live Composio docs. Keep `SKILL.md` rules in force.

## 1. Frame

Write a one-paragraph brief:

- **Question** — what decision or implementation depends on the answer
- **Scope** — `provider` | `client` | `general`
- **Depth** — `overview` | `deep` | `diff`
- **Out of scope** — what you will not chase in this pass

## 2. Orient

1. Open [doc-map.md](doc-map.md) and select **one primary** API family page.
2. Always queue **auth** + **errors** if the question touches requests or failure handling.
3. For auth configs / toolkits, queue the matching **concept** docs under `docs/authentication/`.
4. If unsure which family applies, fetch https://docs.composio.dev/llms.txt and search for the noun (auth_config, toolkit, trigger, webhook).

Optional: run `scripts/list-reference-urls.sh` from the skill root to print current reference URLs.

## 3. Fetch

1. Fetch primary URL with `.md` suffix via WebFetch or `curl -fsSL`.
2. Strip or ignore the repeated "Composio SDK — Notes for AI Code Generators" footer when extracting REST facts; it is SDK guidance, not OpenAPI.
3. Follow in-page links to nested permission or scheme docs.
4. Do **not** paste entire pages into the final answer — extract structured facts.

### Fetch hygiene

- Prefer parallel fetches for primary + auth + errors.
- On 404, re-check llms.txt; Composio may rename pages.
- Prefer v3.1 paths. If only v3 content answers a historical question, label it **legacy**.

## 4. Extract (deep checklist)

For each endpoint relevant to the question, capture:

| Field              | Notes                                               |
| ------------------ | --------------------------------------------------- |
| Method + path      | Include `/api/v3.1` prefix as documented            |
| Auth header        | `x-api-key` vs `x-org-api-key` vs scoped key limits |
| Path params        | e.g. `{nanoid}`                                     |
| Query params       | filters, pagination, `toolkit_versions`             |
| Body fields        | required vs optional; managed vs custom variants    |
| Response           | id shape, nested objects, nullability               |
| Status / lifecycle | enable/disable vs delete; 404 on missing            |
| Side effects       | cascading deletes, connection impact                |
| Version params     | tools endpoints only — latest vs `00000000_00`      |

Also note:

- Auth schemes (`OAUTH2`, `API_KEY`, `BEARER_TOKEN`, `BASIC`) when researching auth configs
- Whether credentials are write-only / redacted on read
- List vs get field parity (Terraform often needs get-after-create)

## 5. Cross-check (diff / provider / client)

Compare extracted facts to:

1. `internal/composio/api/*.go` — method names, paths, JSON tags
2. `internal/composio/models/*.go` — domain structs
3. Matching `internal/provider/resource_*.go` or `data_source_*.go` if present
4. Generated docs under `docs/` only as a tertiary check (may lag)

Record mismatches as:

- **Docs → code gap** — client missing or wrong
- **Code → docs gap** — client has behavior docs omit (verify with another page or conservative assumption)
- **Provider policy** — API exists but must not become a Terraform resource (see [provider-scope.md](provider-scope.md))

## 6. Decide

Translate findings into actionable recommendations:

- Schema attributes (required / optional / computed / ForceNew / Sensitive)
- Client methods to add or change
- Import ID format
- Acceptance-test constraints (project key, toolkit slug, cleanup)
- Explicit non-goals

## 7. Report

Deliver:

1. Short verdict (2–5 sentences)
2. Structured endpoint / field notes
3. Citations (markdown links to `.md` URLs)
4. Open questions
5. Optional: fill [research-report-template.md](../assets/research-report-template.md)

Do not claim OpenAPI completeness if the page only lists endpoints without schemas — escalate to concept docs or SDK reference and mark confidence **medium/low**.
