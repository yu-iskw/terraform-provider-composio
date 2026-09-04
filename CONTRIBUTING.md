# Contributing

This repository is a starting point for custom Terraform providers built with the Terraform Plugin Framework.

## Prerequisites

- Go, using the version declared in `go.mod` (install Go yourself; it is not provided via mise here)
- Terraform, using the version declared in `.terraform-version`
- GNU Make

Optional: [mise](https://mise.jdx.dev/) — run `mise trust` in the repo root if needed, then `mise install` to install the **Trunk CLI** only from [`mise.toml`](mise.toml). Go stays outside mise. When upgrading Trunk, update both `.trunk/trunk.yaml` `cli.version` and `mise.toml` `trunk = "..."`.

Quality checks use **Trunk** (`trunk check` / `trunk fmt`) plus **`gosec`** and **`deadcode`** via `GNUmakefile` targets (`make lint`, `make build`). Git hooks: run `make setup-dev` to clear conflicting `core.hooksPath` and run `trunk git-hooks sync` (no pre-commit framework).

## Local Development

Run the common checks:

```shell
make test
go build -v ./
go generate ./...
```

Run formatting before sending changes:

```shell
make format
```

## Repository Structure

- `main.go`: provider server entry point
- `internal/provider`: provider implementation, resources, data sources, docs embedded by tests, and unit tests
- `internal/your_service`: HTTP API client (`Client.HTTP`, YOUR_SERVICE placeholder) with optional rate and concurrency limits (rename when you fork; see `internal/your_service/README.md`)
- `examples`: Terraform examples used by documentation generation
- `docs`: generated or hand-maintained provider documentation
- `tools`: Go tool dependency tracking for code generation and analysis
- `mise.toml`: optional Trunk CLI pin for mise users

## Adapting The Template

1. Rename the module path in `go.mod`.
2. Update the provider address in `main.go`.
3. Rename the provider type in `internal/provider/provider.go`.
4. Rename `internal/your_service` to a meaningful import path and package name for your API client; update imports under `internal/provider/` (see `internal/your_service/README.md`).
5. Replace or extend `internal/your_service` (see `Client.HTTP` and optional provider rate or concurrency attributes) and the example resource/data source files with real API-backed implementations.
6. Update examples and generated docs.
7. Run tests, build, and documentation generation.

Acceptance tests should be added once the provider has a real external API and deterministic test environment.
