# Testing setup reference

Utilities used by provider acceptance tests in `internal/provider`.

## `isIntegrationTestMode()`

- Defined in `internal/provider/provider_test.go`.
- Returns true when `TF_ACC` is `1`.
- Skips acceptance-style tests during normal `make test` / `go test` runs.

## `testAccPreCheck(t)`

- Defined in `internal/provider/provider_test.go`.
- Skips unless `TF_ACC=1`.
- Fatals if `COMPOSIO_API_KEY` is unset when Acc mode is on.

## `testAccProtoV6ProviderFactories`

- Map key must match the provider local name in HCL: `"composio"`.
- Uses `providerserver.NewProtocol6WithError(New("test")())`.

## Acc helpers (`acc_helpers.go`)

- `getProviderConfig()` reads `acc_tests/provider.tf` (`provider "composio" {}`). Credentials come from the environment via provider Configure.
- `ReadAccTestResource(pathParts)` reads `.tf` fixtures under `internal/provider/acc_tests/`.
- `substituteAccTestName` replaces `__NAME__` placeholders for unique resource names.

## Adding acceptance tests

- Put Terraform fixtures under `internal/provider/acc_tests/{resources|data_sources}/...`.
- Prefer managed_auth Acc cases; avoid write-only `custom_auth.credentials` in CI.
- Example addresses: `composio_toolkit` / `composio_auth_config` (see `examples/`).
