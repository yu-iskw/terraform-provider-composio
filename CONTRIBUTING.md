# Contributing

This repository is the Terraform provider for Composio.

## Prerequisites

- Go, using the version declared in `go.mod` (install Go yourself; it is not provided via mise here)
- Terraform 1.11 or later (write-only credentials). See `.terraform-version`
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

```shell
export TF_ACC=1
export COMPOSIO_API_KEY=...
make testacc
```

## Repository Structure

- `main.go`: provider server entry point
- `internal/provider`: Terraform wiring, resources, data sources, embedded docs, tests
- `internal/composio/api`: HTTP client for `/api/v3.1`
- `internal/composio/models`: domain models
- `examples`: Terraform examples used by documentation generation
- `docs`: generated provider documentation
- `tools`: Go tool dependency tracking
- `mise.toml`: optional Trunk CLI pin for mise users
