# Composio API client

Package `api` talks to Composio REST API v3.1. It does not import Terraform types.

- Base URL default: `https://backend.composio.dev`
- Path prefix: `/api/v3.1` (pinned, not floating)
- Project calls send `x-api-key`
- Organization calls send `x-org-api-key`
- Retries apply to GET, PUT, PATCH, and DELETE. POST create is not retried.
- Secrets in error bodies are redacted by key name.
