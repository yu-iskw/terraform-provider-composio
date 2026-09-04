# Terraform Provider Template

This repository is a template for building custom Terraform providers with the HashiCorp Terraform Plugin Framework.

It includes:

- A generic provider configuration with `endpoint` and sensitive `api_key` attributes
- Optional `max_concurrent_requests` and `requests_per_second` to cap HTTP concurrency and average rate (defaults: 10 / 10)
- One example resource, `template_example_item`
- One example data source, `template_example_item`
- Documentation, examples, tests, build, lint, and release scaffolding

Replace the placeholder client in `internal/your_service` (YOUR_SERVICE layer), and the example resource and data source, with code for your product API.

## Example Usage

```hcl
terraform {
  required_providers {
    template = {
      source = "example/template"
    }
  }
}

provider "template" {
  endpoint = "https://api.example.com"
  api_key  = var.api_key

  # max_concurrent_requests = 5
  # requests_per_second     = 20
}

resource "template_example_item" "example" {
  name        = "example"
  description = "Created by the provider template"
}

data "template_example_item" "example" {
  name = template_example_item.example.name
}
```

## Development

Optional: with [mise](https://mise.jdx.dev/), run `mise trust` if prompted, then `mise install` to get the pinned Trunk CLI from `mise.toml` (Go is not installed via mise).

```shell
make test
go build -v ./
go generate ./...
```

Use `internal/provider` for Terraform wiring and `internal/your_service` for your API client (rename when you fork; see that directory’s `README.md`).
