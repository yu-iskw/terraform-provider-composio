---
name: lint-and-fix
description: Run `make lint` from the repository root, fix violations in recipe order with minimal changes, and re-run until lint passes.
---

# Lint and fix

## When to use

- CI or local feedback says lint failed, or you need a green `make lint` before a PR.
- Someone asks to "fix lint", "make lint pass", or resolve Trunk / security scanner / linter output.

## Command

From the **repository root** (where `GNUmakefile` lives):

```bash
make lint
```

Per `GNUmakefile`, `make lint` depends on targets **in this order**:

1. **`run-trunk-check`** — `trunk check --all -y` (linters and rules live in [`.trunk/trunk.yaml`](../../../.trunk/trunk.yaml)).
2. **`deadcode`** — `go run golang.org/x/tools/cmd/deadcode -test ./...`.
3. **`gosec`** — `gosec ./internal/...`.

The first failing step stops `make`; fix that step before expecting later steps to run.

## Fix loop

1. Run `make lint` and identify **which prerequisite failed first** from the log.
2. Apply the **smallest** change that addresses that failure (prefer real fixes over disabling linters).
3. Re-run `make lint` until exit code is zero.
4. Do not weaken [`.trunk/trunk.yaml`](../../../.trunk/trunk.yaml) unless the user explicitly asks to change project policy.

## Common failures

- **Trunk (`trunk check`)** — golangci-lint2, security scanners (e.g. grype, osv-scanner), tflint, markdown, YAML, etc. Use file paths and rule IDs in output; consult `.trunk/trunk.yaml` for enabled tools.
- **`deadcode`** — remove or use unreachable code; remember `-test` mode may keep symbols used only from tests.
- **`gosec`** — address security findings; only use suppressions or `#nosec` patterns if the repo already documents that approach.

## Boundaries

- **Formatting vs lint:** some issues are faster to clear with **`make format`** (`go fmt` + `trunk fmt`) than by hand; if output looks like pure formatting, run `make format` and re-run `make lint`.
- **Not tests:** use **test-and-fix** (`make test`) for unit test failures.
- **Not compile / generate:** `make lint` does **not** run `go generate` or `go build`. Failures there belong to **build-and-fix** (`make build`).

## Optional

- If `trunk: command not found`, install Trunk (see [`mise.toml`](../../../mise.toml) and [CONTRIBUTING](../../../CONTRIBUTING.md) for optional mise + `mise trust` / `mise install`).
