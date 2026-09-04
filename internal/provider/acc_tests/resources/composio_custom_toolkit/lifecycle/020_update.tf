resource "composio_custom_toolkit" "test" {
  slug    = "__SLUG__"
  name    = "__NAME__"
  app_url = "__MCP_URL__"

  auth_scheme = {
    mode = "NO_AUTH"
  }
}
