---
name: run-acceptance-tests
description: Guide execution of Terraform provider acceptance tests (TF_ACC) with correct environment setup and targeted runs.
---

# Run Acceptance Tests

## Purpose

Standardize running acceptance tests in this repository. Acceptance tests use `terraform-plugin-testing` with `TF_ACC=1` and may call a live API once you add real resources and wire credentials.

## Prerequisites

### 1. When acceptance tests exist

Before running acceptance tests, confirm how `testAccPreCheck` and test configs expect credentials. This template ships with fast unit tests only; when you add `TestAcc...` cases, document required environment variables (for example values from `.env.template` in the repository root: `TEMPLATE_ENDPOINT`, `TEMPLATE_API_KEY`).

### If credentials are required and missing

- Copy `.env.template` to `.env` when you introduce live-API tests: `cp .env.template .env`
- Fill in values before running `make testacc`.

## Workflow

### 1. Full acceptance suite

- **Command**: `make testacc`
- **Details**: Runs `go test ./internal/provider/...` with `TF_ACC=1`.
- **Warning**: Slow and may create or change real infrastructure once tests target a live API.

### 2. Targeted acceptance tests

- **Command**: `make testacc TESTARGS="-run <Pattern>"`
- **Example**: `make testacc TESTARGS="-run TestAccExampleItem"`

## Distinction from unit tests

- **Unit tests**: `make test`. Does not set `TF_ACC`.
- **Acceptance tests**: `make testacc`. Sets `TF_ACC=1`.

## Troubleshooting

- **Timeouts**: `make testacc TESTARGS="-timeout 120m"`.
- **Authentication**: Verify variables in `.env` match what `testAccPreCheck` and test configs expect.
- **Cleanup**: After failures against a real API, remove stray resources in the target system or via API if tests did not roll back.
