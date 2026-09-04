provider "template" {
  endpoint = "https://api.example.com"
  api_key  = "example-api-key"

  # Optional: tune HTTP client (defaults: 10 concurrent, 10 RPS)
  # max_concurrent_requests = 5
  # requests_per_second     = 20.5
}
