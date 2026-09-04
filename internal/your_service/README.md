# `internal/your_service` (YOUR_SERVICE layer)

Placeholder package for your product API client (credentials, base URL, HTTP calls). Treat this as the **YOUR_SERVICE** layer: rename the directory and Go package when you adapt the template for a real API.

## HTTP client behavior

- **`Client.HTTP`**: standard `*http.Client` whose transport applies, in order: **token-bucket rate limit** (`requests_per_second`), **max concurrent in-flight requests** (`max_concurrent_requests`), then **Authorization: Bearer** using the provider `api_key`.
- **Defaults** (when provider omits optional attributes): `10` concurrent requests, `10` requests per second. See `DefaultMaxConcurrent` and `DefaultRPS` in [`client.go`](client.go).
- **Negative values** for optional limits are rejected; omit attributes to use defaults.

Provider attributes are defined in [`internal/provider/provider.go`](../provider/provider.go): `max_concurrent_requests`, `requests_per_second`.

When you initialize a provider from this template:

1. Rename this directory to something meaningful (for example `internal/acmeapi`).
2. Change the Go `package` line in `.go` files here to match (for example `package acmeapi`).
3. Update imports: replace `.../internal/your_service` with your new import path everywhere (for example under `internal/provider/`).
4. Run `go mod tidy` and, if you use vendoring, `go mod vendor`.
