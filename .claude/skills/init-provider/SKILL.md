---
name: init-provider
description: Bootstrap a new Terraform provider from this template by renaming the Go module, registry address, provider type, examples, and optional YOUR_SERVICE client package.
---

# Initialize provider from template

## When to use

Use this skill when someone forks or copies this repository to create `terraform-provider-<product>` and wants a consistent rename pass (module path, `main.go` registry address, Terraform `provider` local name, examples, tests) so `make test` and `go build` stay green.

## Inputs to collect first

1. **Go module path** (for example `github.com/acme/terraform-provider-acme`).
2. **Registry provider address** for [`main.go`](../../../main.go) `providerserver.ServeOpts.Address`: `registry.terraform.io/<namespace>/<type>` (must match where the provider will be published).
3. **Terraform provider local name** — short name used in `provider "..."` blocks and in [`internal/provider/provider_test.go`](../../../internal/provider/provider_test.go) `testAccProtoV6ProviderFactories` (this template uses `template`).
4. (Optional) Whether to **keep** [`internal/your_service`](../../../internal/your_service) as the YOUR_SERVICE placeholder or rename directory + `package` to a product-specific client.

Do not edit plan files under `.cursor/plans/`.

## Workflow

1. **Replace module path** — Set `module` in [`go.mod`](../../../go.mod). Replace every import and string occurrence of `github.com/example/terraform-provider-template` with the new module path (see [substitution_map.md](references/substitution_map.md)).
2. **Provider type name** — In [`internal/provider/provider.go`](../../../internal/provider/provider.go), set `resp.TypeName` in `Metadata` to the new short name. Update unit tests that assert the old name.
3. **Serve address** — In [`main.go`](../../../main.go), update `ServeOpts.Address` to the new registry source string.
4. **Examples** — Update [`examples/provider/provider.tf`](../../../examples/provider/provider.tf) and resource/data source examples under [`examples/`](../../../examples/) so `provider "<name>"` and `resource "<name>_..."` / `data "<name>_..."` match the new type prefix.
5. **Acceptance-test wiring** — If present, align `testAccProtoV6ProviderFactories` map keys and any `TestAcc` HCL with the new provider local name (see [checklist.md](references/checklist.md)).
6. **Optional client rename** — Rename `internal/your_service` and `package yourservice` only if the fork wants a product-specific import path; then fix imports under `internal/provider/` and skill boilerplate under [`.claude/skills/implement-terraform-provider-resource/assets/`](../implement-terraform-provider-resource/assets/).
7. **Verify** — `go mod tidy`, `make test`, `go build -buildvcs=false ./`, then `go generate ./...` and review `docs/` for intentional updates.

After mechanical edits, optionally run the **verify-and-fix** skill for format and lint.

## References

- [Detailed checklist](references/checklist.md)
- [Substitution map](references/substitution_map.md)
