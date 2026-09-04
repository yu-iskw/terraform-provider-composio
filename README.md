# Terraform Provider for Composio

Manage durable Composio control-plane configuration with Terraform. This provider is for platform configuration, not for user OAuth, sessions, or tool execution.

Source address: `yu-iskw/composio`.

## Example Usage

```hcl
terraform {
  required_providers {
    composio = {
      source = "yu-iskw/composio"
    }
  }
}

provider "composio" {
  # Falls back to COMPOSIO_API_KEY / COMPOSIO_ORG_API_KEY
}

data "composio_toolkit" "github" {
  slug = "github"
}

resource "composio_auth_config" "github" {
  toolkit_slug = data.composio_toolkit.github.slug

  managed_auth = {
    restrict_to_following_tools = [
      "GITHUB_CREATE_ISSUE",
      "GITHUB_GET_ISSUE",
    ]
  }
}
```

## Provider configuration

| Attribute | Env | Purpose |
| --- | --- | --- |
| `api_key` | `COMPOSIO_API_KEY` | Project key (`x-api-key`) |
| `org_api_key` | `COMPOSIO_ORG_API_KEY` | Organization key (`x-org-api-key`) |
| `endpoint` | `COMPOSIO_ENDPOINT` | API origin. Default `https://backend.composio.dev` |
| `max_concurrent_requests` | | In-flight HTTP cap. Default 8 |
| `request_timeout` | | Go duration. Default `30s` |

At least one credential is required. Project-scoped resources need `api_key`. The client always calls `/api/v3.1` under the origin. That prefix does not float.

## Current surface

- Resource `composio_auth_config` (import: auth config id)
- Data source `composio_toolkit`

`Sensitive` hides values in the Terraform UI. It does not make state secret-free. `custom_auth.credentials` is write-only (Terraform 1.11+) and is not stored in state. A credentials-only edit does not produce a plan.

## Development

```shell
make test
go build -v ./
go generate ./...
```

Acceptance tests need `TF_ACC=1` and `COMPOSIO_API_KEY`. They are skipped in default `make test`.
