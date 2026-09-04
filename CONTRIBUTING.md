# Contributing

This repository is the Terraform provider for Composio.

## Prerequisites

- Go, using the version declared in `go.mod` (install Go yourself; it is not provided via mise here)
- Terraform 1.15 or later. See `.terraform-version`
- GNU Make

Optional: [mise](https://mise.jdx.dev/) — run `mise trust` in the repo root if needed, then `mise install` to install the **Trunk CLI** only from [`mise.toml`](mise.toml). Go stays outside mise. When upgrading Trunk, update both `.trunk/trunk.yaml` `cli.version` and `mise.toml` `trunk = "..."`.

Quality checks use **Trunk** (`trunk check` / `trunk fmt`) plus **`gosec`** and **`deadcode`** via `GNUmakefile` targets (`make lint`, `make build`). Git hooks: run `make setup-dev` to clear conflicting `core.hooksPath` and run `trunk git-hooks sync` (no pre-commit framework).

## Local Development

```shell
make test
go build -v ./
go generate ./...
make format
```

Acceptance tests:

Use a **dedicated** Composio project API key. Current Acc coverage exercises `composio_toolkit` (read), managed `composio_auth_config` (create/update/import/destroy), and optionally `composio_custom_toolkit` when `COMPOSIO_ACC_CUSTOM_MCP_URL` is set to a public HTTPS MCP endpoint. Do not point Acc at a production project.

```shell
export TF_ACC=1
export COMPOSIO_API_KEY=...
# optional: export COMPOSIO_ENDPOINT=https://backend.composio.dev
# optional: export COMPOSIO_ACC_CUSTOM_MCP_URL=https://mcp.example.com/mcp
make testacc
```

`make testacc` sets `TF_ACC_TERRAFORM_VERSION` from `.terraform-version` (1.15+) so Acc does not pick a tfenv global default when the plugin-testing helper runs Terraform from a temp directory.

Optional `.env` (Make does **not** auto-load it):

```shell
cp .env.template .env
# set COMPOSIO_API_KEY
set -a && source .env && set +a
make testacc
```

Targeted runs:

```shell
make testacc TESTARGS='-run TestAccToolkit'
make testacc TESTARGS='-run TestAccAuthConfig'
```

CI: `.github/workflows/acceptance.yml` is `workflow_dispatch` and injects `COMPOSIO_API_KEY` from the `composio-acceptance` environment.

## Repository Structure

- `main.go`: provider server entry point
- `internal/provider`: Terraform wiring, resources, data sources, embedded docs, tests
- `internal/composio/api`: HTTP client for `/api/v3.1`
- `internal/composio/models`: domain models
- `examples`: Terraform examples used by documentation generation
- `docs`: generated provider documentation
- `tools`: Go tool dependency tracking
- `mise.toml`: optional Trunk CLI pin for mise users
