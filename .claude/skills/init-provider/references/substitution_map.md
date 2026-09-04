# Substitution map (template defaults)

Replace left column with fork-specific values. Use repository-wide search (`rg "github.com/example/terraform-provider-template"`, `rg "registry.terraform.io/example/template"`, `rg '\"template\"' examples`) until clean.

| Concept                                          | Template value                                         | Typical fork value                                                   |
| ------------------------------------------------ | ------------------------------------------------------ | -------------------------------------------------------------------- |
| Go module                                        | `github.com/example/terraform-provider-template`       | `github.com/<org>/terraform-provider-<product>`                      |
| Import prefix                                    | same as module path                                    | same as new `go.mod` module line                                     |
| Registry address (`main.go` `ServeOpts.Address`) | `registry.terraform.io/example/template`               | `registry.terraform.io/<namespace>/<type>`                           |
| Provider short type (`Metadata.TypeName`)        | `template`                                             | short product slug (e.g. `acme`)                                     |
| Terraform provider local name                    | `template`                                             | must match `provider "..."` in examples and acc-test factory map key |
| Resource / data source prefix                    | `template_` (e.g. `template_example_item`)             | `<type>_` prefix matching provider short type                        |
| YOUR_SERVICE client import (default)             | `.../internal/your_service` with `package yourservice` | Optional second pass: `.../internal/<clientpkg>`                     |

## Grep starting points

```bash
rg 'github.com/example/terraform-provider-template'
rg 'registry.terraform.io/example/template'
rg 'ProviderTypeName: "template"' internal/
rg 'provider "template"' examples/
```

## Notes

- **Provider local name** and **`Metadata.TypeName`** should stay aligned with the first segment of managed resource types (`<type>_example_item`).
- After renaming the module, **Claude skill boilerplate** under `.claude/skills/implement-terraform-provider-resource/assets/` must use the new import path or copied examples will not compile when pasted into `internal/provider`.
