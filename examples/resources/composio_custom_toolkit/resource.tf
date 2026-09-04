resource "composio_custom_toolkit" "example" {
  slug    = "ACME"
  name    = "Acme"
  app_url = "https://mcp.example.com/mcp"

  auth_scheme = {
    mode = "NO_AUTH"
  }
}
