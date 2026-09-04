# AGENTS.md

Shared guidance for agents working in this repository.

## Overview

This is a template Terraform provider written in Go with the HashiCorp Terraform Plugin Framework. The provider is intentionally generic and includes one example resource and one example data source.

## mise (optional)

If you use [mise](https://mise.jdx.dev/), run `mise trust` in the repo root on first use if mise refuses to load the config, then `mise install`. [`mise.toml`](mise.toml) installs **only the Trunk CLI** (pinned to match `.trunk/trunk.yaml` `cli.version`). **Go is not installed via mise**; use your normal Go install and the `go` / `toolchain` lines in `go.mod`. When you upgrade Trunk, bump `cli.version` in `.trunk/trunk.yaml` and the `trunk` version in `mise.toml` together.

## Key Commands

- Unit tests: `make test`
- Build: `go build -v ./`
- Generate docs: `go generate ./...`
- Format: `make format`
- Lint: `make lint`

## Project Structure

- `main.go`: provider server entry point
- `internal/provider`: provider schema, configuration, resources, data sources, embedded docs, and tests
- `internal/your_service`: HTTP API client (`Client.HTTP`, YOUR_SERVICE placeholder) with optional rate and concurrency limits (rename when you fork; see `internal/your_service/README.md`)
- `examples`: Terraform examples used by docs generation
- `docs`: provider documentation
- `tools`: Go tool dependencies
- `mise.toml`: optional Trunk CLI version for mise users (no Go via mise)

## Development Notes

- Keep the template small and easy to adapt.
- Prefer real, deterministic tests over mocks.
- Do not add product-specific API behavior unless adapting the template for a concrete provider.
- Keep provider examples generic until the repository is renamed for a real provider.
- Git hooks for this repo are **Trunk-only** (`make setup-dev` runs `trunk git-hooks sync`); there is no separate pre-commit install step.
