---
name: testacc-and-fix
description: Run `make testacc` with acceptance-test safeguards, fix failures, and re-run until acceptance tests pass or skips are understood.
---

# Testacc and fix

## When to use

- You are working on `TestAcc...` tests under `internal/provider/` and need them green with `TF_ACC=1`.
- Someone asks to "fix acceptance tests" or "debug `make testacc`".

## Safety and boundaries

- **Acceptance tests may create, modify, or destroy real infrastructure** once wired to a live API. Do **not** run `make testacc` against unknown production accounts or shared environments.
- If required credentials, endpoints, or cloud projects are **missing or ambiguous**, **stop and ask the user** before running or re-running tests.
- This template may have **few or no** real `TestAcc` cases yet; output may show skips—confirm that is expected before "fixing" nonexistent tests.

## Command

From the **repository root**:

```bash
make testacc
```

This sets **`TF_ACC=1`** and runs:

`go test ./internal/provider/... -v $(TESTARGS) -timeout 120m`

So failures are **not** caused by forgetting `TF_ACC` when using this Makefile target (unlike ad-hoc `go test` without the env).

Targeted run (optional):

```bash
make testacc TESTARGS='-run ^TestAccYourPattern$'
```

## Fix loop

1. Confirm environment and blast radius with the user if tests touch live APIs.
2. Run `make testacc` (or with `TESTARGS`) and capture the first failing test and error.
3. Fix **acceptance test code**, provider wiring, or test fixtures with minimal changes; re-run.
4. Repeat until green or until remaining skips are explained by `testAccPreCheck` / missing credentials.

## Common outcomes

- **All skipped** — often `testAccPreCheck` skips when prerequisites are not met; see environment guidance in [run-acceptance-tests](../run-acceptance-tests/SKILL.md).
- **Terraform CLI / provider config errors** — fix HCL fixtures or provider factory in tests.
- **API 401/403/404** — credentials or endpoint; do not hardcode secrets into the repo.

## Further reading

- [run-acceptance-tests](../run-acceptance-tests/SKILL.md) — prerequisites, `TESTARGS`, `.env` patterns, and troubleshooting.

## Optional

- For unit tests only, use **test-and-fix** (`make test`). For broader lint/format, use `make format` / `make lint`.
