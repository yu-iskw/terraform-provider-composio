---
name: run-acceptance-tests
description: Guide execution of Terraform provider acceptance tests (TF_ACC) with correct environment setup and targeted runs.
---

# Run Acceptance Tests

## Purpose

Standardize running acceptance tests in this repository. Acceptance tests use `terraform-plugin-testing` with `TF_ACC=1` and call the live Composio API.

## Prerequisites

### 1. Credentials

Acceptance tests require a **dedicated** Composio project API key (do not use a production project).

- `COMPOSIO_API_KEY` (required): project key (`x-api-key`)
- `COMPOSIO_ENDPOINT` (optional): API origin; defaults to `https://backend.composio.dev`
- `COMPOSIO_ORG_API_KEY` is **not** required for current Acc coverage (toolkit + managed auth config)

Optional local `.env` (not loaded by Make):

```shell
cp .env.template .env
# edit COMPOSIO_API_KEY
set -a && source .env && set +a
make testacc
```

Or export variables directly:

```shell
export TF_ACC=1
export COMPOSIO_API_KEY=...
make testacc
```

### If credentials are missing

`testAccPreCheck` fails the suite when `TF_ACC=1` and `COMPOSIO_API_KEY` is empty.

## Workflow

### 1. Full acceptance suite

- **Command**: `make testacc`
- **Details**: Runs `go test ./internal/provider/...` with `TF_ACC=1`.
- **Warning**: Creates and destroys real auth configs in the project tied to the API key.

### 2. Targeted acceptance tests

- **Command**: `make testacc TESTARGS="-run <Pattern>"`
- **Example (data source)**: `make testacc TESTARGS="-run TestAccToolkit"`
- **Example (resource)**: `make testacc TESTARGS="-run TestAccAuthConfig"`

## Distinction from unit tests

- **Unit tests**: `make test`. Does not set `TF_ACC`.
- **Acceptance tests**: `make testacc`. Sets `TF_ACC=1`.

## Troubleshooting

- **Timeouts**: `make testacc TESTARGS="-timeout 120m"`.
- **Authentication**: Confirm `COMPOSIO_API_KEY` is a valid project key for a non-production project.
- **Cleanup**: After failures, delete orphaned auth configs in the Composio project if destroy did not complete.
- **Terraform version**: Acc requires Terraform CLI 1.15 or later (`tfversion.SkipBelow(1.15.0)`). `make testacc` sets `TF_ACC_TERRAFORM_VERSION` from `.terraform-version` so tests do not skip when a tfenv shim resolves an older global default from the plugin-testing temp directory. Override with `make testacc TF_ACC_TERRAFORM_VERSION=1.15.0` or `TF_ACC_TERRAFORM_PATH=/path/to/terraform` if needed.
