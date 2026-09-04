# Testing setup reference

Utilities used by provider tests in `internal/provider`.

## `isIntegrationTestMode()`

- Defined in `internal/provider/provider_test.go`.
- Returns true when `TF_ACC` is `1`.
- Skips acceptance-style tests during normal `make test` / `go test` runs.

## `testAccPreCheck(t)`

- Defined in `internal/provider/provider_test.go`.
- Skips unless `TF_ACC=1`. Extend this function when you add live-API acceptance tests to require environment variables (for example from `.env`).

## `testAccProtoV6ProviderFactories`

- Map key must match the provider local name in HCL: `"template"`.
- Uses `providerserver.NewProtocol6WithError(New("test")())`.

## Adding acceptance tests later

- Put Terraform fixtures under `internal/provider/acc_tests/...` if you adopt that layout.
- Build a `provider "template" { ... }` block from environment variables or literals in test code.
- Example resource address: `template_example_item.example` (see `examples/resources/template_example_item/resource.tf`).
