# Init-provider checklist

Work top to bottom. Use `rg` (ripgrep) to confirm nothing from the template module path or old names remains before finishing.

## 1. Go module and imports

- [ ] [`go.mod`](../../../../go.mod): `module github.com/example/terraform-provider-template` → new path.
- [ ] All `import` strings under `internal/`, `main.go`, and [`.claude/skills/.../assets/*.go`](../../implement-terraform-provider-resource/assets/) that reference the old module path.
- [ ] Run `go mod tidy`.

## 2. Provider identity

- [ ] [`internal/provider/provider.go`](../../../../internal/provider/provider.go): `Metadata` `resp.TypeName` (default `template`).
- [ ] [`internal/provider/provider_test.go`](../../../../internal/provider/provider_test.go): `testAccProtoV6ProviderFactories` key must match the Terraform provider local name in examples; update `TestProviderMetadata` expected `TypeName` if asserted.
- [ ] [`internal/provider/resource_example_item_test.go`](../../../../internal/provider/resource_example_item_test.go) and [`data_source_example_item_test.go`](../../../../internal/provider/data_source_example_item_test.go): `ProviderTypeName` argument to `Metadata` if tests assert full type names.

## 3. Binary / registry

- [ ] [`main.go`](../../../../main.go): `ServeOpts.Address` (default `registry.terraform.io/example/template`).

## 4. Examples (Terraform)

- [ ] [`examples/provider/provider.tf`](../../../../examples/provider/provider.tf): `provider "template"` → new local name.
- [ ] [`examples/resources/template_example_item/resource.tf`](../../../../examples/resources/template_example_item/resource.tf): resource type prefix `template_` → `<new>_` (rename directory if you rename the example resource for clarity).
- [ ] [`examples/data-sources/template_example_item/data-source.tf`](../../../../examples/data-sources/template_example_item/data-source.tf): same for `data` blocks.

## 5. Documentation generation

- [ ] Run `go generate ./...` from repo root (requires `terraform` on PATH for fmt; tfplugindocs updates [`docs/`](../../../../docs/)).
- [ ] Scan `docs/**/*.md` for old `page_title` or registry strings if they still reference the template name.

## 6. Optional: rename YOUR_SERVICE client package

Skip if the fork keeps the placeholder until later.

- [ ] Rename directory `internal/your_service` → `internal/<product>api` (or similar).
- [ ] Set `package <matching>` in all `.go` files in that directory.
- [ ] Update imports in `internal/provider/*.go` and Claude boilerplate assets.
- [ ] Update [`internal/your_service/README.md`](../../../../internal/your_service/README.md) equivalent in the new path; update [`README.md`](../../../../README.md), [`AGENTS.md`](../../../../AGENTS.md), [`CONTRIBUTING.md`](../../../../CONTRIBUTING.md), [`.cursor/rules/*.mdc`](../../../../.cursor/rules/) if they mention `your_service`.

## 7. Final verification

- [ ] `gofmt` / `make format` on touched packages.
- [ ] `make test` (or at least `go test ./internal/...`).
- [ ] `go build -buildvcs=false ./`.
- [ ] If the repo vendors dependencies: `go mod vendor` and commit any intentional `vendor/` changes.
