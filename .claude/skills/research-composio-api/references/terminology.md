# Composio terminology (v3)

Always use **current** terms in research output, code suggestions, and Terraform docs. Translate user or legacy wording before searching docs.

| Old (v1/v2)                     | Current (v3)                 | In code / APIs                         |
| ------------------------------- | ---------------------------- | -------------------------------------- |
| entity ID                       | user ID                      | `user_id`                              |
| actions                         | tools                        | e.g. `GITHUB_CREATE_ISSUE`             |
| apps / appType                  | toolkits                     | e.g. `github`                          |
| integration / integration ID    | auth config / auth config ID | `auth_config_id`, auth config `nanoid` |
| connection                      | connected account            | `connected_accounts`                   |
| ComposioToolSet / OpenAIToolSet | `Composio` + provider        | SDK only                               |
| toolset                         | provider (SDK)               | e.g. OpenAI provider package           |

## Auth config vocabulary

| Term               | Meaning                                                                  |
| ------------------ | ------------------------------------------------------------------------ |
| Auth config        | Project-level blueprint for toolkit authentication                       |
| Managed auth       | Composio-managed OAuth app / credentials                                 |
| Custom auth        | Customer-supplied OAuth client or secrets                                |
| Auth scheme        | `OAUTH2`, `API_KEY`, `BEARER_TOKEN`, `BASIC`                             |
| Connected account  | Per-user tokens created when a user authenticates against an auth config |
| Enabled / disabled | Lifecycle without delete; disabled configs cannot start new connections  |

## API version vocabulary

| Term                    | Meaning                                                    |
| ----------------------- | ---------------------------------------------------------- |
| v3.1                    | Current REST API; default for new work                     |
| v3                      | Previous REST API; supported; frozen tool-version defaults |
| `latest` / omit version | On v3.1 tool endpoints, selects latest toolkit version     |
| `00000000_00`           | Pinned toolkit version (v3 default when version omitted)   |

## Search tips

- User says "integration" → search **auth config**
- User says "app" → search **toolkit**
- User says "action" → search **tool**
- User says "entity" → search **user_id** / connected accounts (usually out of Terraform scope)
