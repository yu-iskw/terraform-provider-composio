---
name: implement-terraform-provider-resource
description: Implement Terraform resources and data sources in `internal/provider` with testing aligned to this template.
---

# Implement Terraform provider resource

## Description

Implement a new Terraform resource or data source in `internal/provider` using `hashicorp/terraform-plugin-framework`, following patterns in this repository (for example `resource_example_item.go`, `data_source_example_item.go`, and the API client in `internal/your_service`).

## Input

The user should provide:

1. **Type**: Resource or data source.
2. **Terraform type name suffix** (e.g., `widget` → `template_widget` when `ProviderTypeName` is `template`).
3. **Schema**: Attributes (name, type, required / optional / computed).
4. **Client operations**: Methods on your API client (replace or extend `yourservice.Client` in `internal/your_service`).

## Workflow

### 1. Implementation

- **Files**: Add `internal/provider/resource_<name>.go` or `internal/provider/data_source_<name>.go`.
- **Model**: Go struct with `tfsdk` tags; use `types.String`, `types.Bool`, etc.
- **Interfaces**: Implement `resource.Resource` (+ `Configure`, `ImportState` as needed) or `datasource.DataSource` (+ `Configure` as needed).
- **Schema**: Define in `Schema`; use clear descriptions.
- **Configure**: Read `*yourservice.Client` (or your client type) from `req.ProviderData`, matching `internal/provider/provider.go` `Configure`.
- **CRUD / read**: Implement lifecycle methods; surface errors with `resp.Diagnostics`.

### 2. Unit tests

- Add `internal/provider/resource_<name>_test.go` or `data_source_<name>_test.go` for metadata, helpers, and behavior that does not need Terraform CLI.

### 3. Acceptance tests (optional)

- Use `isIntegrationTestMode()` and `testAccPreCheck(t)`.
- Use `testAccProtoV6ProviderFactories` with provider key `template`.
- Prefer small `.tf` fixtures under `internal/provider/acc_tests/` when you add them.
- Use `resource.Test` from `github.com/hashicorp/terraform-plugin-testing/helper/resource`.

## Example patterns

### Resource structure (same package as provider)

```go
package provider

import (
    "context"

    "github.com/example/terraform-provider-template/internal/your_service"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &widgetResource{}

type widgetResource struct {
    client *yourservice.Client
}

type widgetResourceModel struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}
```

### Acceptance test snippet

```go
func TestAccWidget_basic(t *testing.T) {
    if !isIntegrationTestMode() {
        t.Skip("acceptance tests require TF_ACC=1")
    }
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccWidgetConfigBasic,
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("template_widget.test", "name", "value"),
                ),
            },
        },
    })
}
```

## Reference

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Provider implementation rules](.cursor/rules/provider-implementation.mdc)
- [Project structure](.cursor/rules/project-structure.mdc)
- [Testing setup](references/testing_setup.md)

## Assets

- [Resource boilerplate](assets/resource_boilerplate.go)
- [Data source boilerplate](assets/data_source_boilerplate.go)
- [Acceptance test boilerplate](assets/test_boilerplate.go)
