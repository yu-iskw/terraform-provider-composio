---
name: test-and-fix
description: Run `make test` from the repository root, fix failures with minimal changes, and re-run until tests pass.
---

# Test and fix

## When to use

- Unit tests are failing or you need a green `make test` before a PR.
- Someone asks to "fix tests", "make test pass", or "debug `make test` failures".

## Command

From the **repository root** (where `GNUmakefile` lives):

```bash
make test
```

This runs `unset TF_ACC && cd internal && go test -count=1 -v ./...` (unit tests only; acceptance tests are not selected by this target).

## Fix loop

1. Run `make test` and capture the **first** failing package and test name.
2. Apply the **smallest** change that addresses that failure (prefer fixing code or test expectations over disabling tests).
3. Re-run `make test`. Repeat until exit code is zero.
4. Avoid unrelated refactors, new dependencies, or broad formatting-only edits in the same pass.

## Targeting a single test (optional)

`make test` does not pass `TESTARGS` through. To narrow scope while debugging:

```bash
cd internal && go test -count=1 -v -run '^TestName$' ./path/to/package
```

Restore confidence with a full `make test` when done.

## Common failures

- **Import / compile errors** — fix `internal/provider` or `internal/your_service` (or other packages under `internal/`).
- **Assertion / golden failures** — adjust implementation or test; keep behavior intentional.
- **Flaky timing** — prefer deterministic tests; avoid sleep-only fixes unless already the project pattern.

## Boundaries

- Do not set `TF_ACC=1` for this skill; use `testacc-and-fix` for acceptance tests.
- For formatting and linters, use `make format` / `make lint` separately; this skill is tests only.
