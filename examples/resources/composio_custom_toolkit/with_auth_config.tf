# API key Custom MCP + auth config for session account matching.
resource "composio_custom_toolkit" "acme" {
  slug    = "ACME"
  name    = "Acme"
  app_url = "https://mcp.example.com/mcp"

  auth_scheme = {
    mode = "API_KEY"
    headers = {
      Authorization = "Bearer {{generic_api_key}}"
    }
  }
}

resource "composio_auth_config" "acme" {
  toolkit_slug            = composio_custom_toolkit.acme.id
  enabled_for_tool_router = true

  custom_auth = {
    auth_scheme = "API_KEY"
  }
}
