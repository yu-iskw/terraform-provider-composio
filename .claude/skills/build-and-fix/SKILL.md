---
name: build-and-fix
description: Run `make build` from the repository root, fix failures across generate/tidy/security/deadcode/compile stages, and re-run until the build succeeds.
---

# Build and fix

## When to use

- `go build` or release prep fails, or CI reports a broken build.
- Someone asks to "fix the build", "make build pass", or resolve errors from `make build`.

## Command

From the **repository root**:

```bash
make build
```

Per `GNUmakefile`, `make build` runs **in order**:

1. **`gen-docs`** — `go generate ./...` (may invoke Terraform fmt and tfplugindocs; Terraform should be on `PATH`).
2. **`go-tidy`** — `go mod tidy`.
3. **`gosec`** — `gosec ./internal/...`.
4. **`deadcode`** — `go run golang.org/x/tools/cmd/deadcode -test ./...`.
5. **`go build -v ./`** — compile the provider module.

A failure at **any** step stops the recipe; fix that stage before expecting a successful compile.

## Fix loop

1. Run `make build` and read output **from the top of the failure** (which target failed first).
2. Fix the underlying issue with the **minimal** diff (security findings may need code changes or documented false positives—prefer fixing real issues).
3. Re-run `make build` until it completes with exit code zero.
4. Do not paper over failures by removing `gosec` / `deadcode` from the Makefile in this workflow unless the user explicitly asks to change project policy.

## Common failures

- **`go generate` / tfplugindocs** — missing `terraform`, wrong working directory, or template/docs errors under `examples/` / `internal/provider`.
- **`go mod tidy`** — unused requires or inconsistent `go.sum`; run tidy and commit intentional `go.sum` changes.
- **`gosec`** — address findings or narrow with approved patterns only if the project already allows them.
- **`deadcode`** — remove or use dead symbols; tests may keep symbols reachable—check `-test` mode behavior.
- **`go build`** — type errors, missing imports, build tags.

## Boundaries

- This skill is **not** a substitute for `make test` or `make testacc`.
- For broad style and linter fixes beyond `gosec`, use `make format` / `make lint` after the build is green if needed.
