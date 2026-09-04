resource "composio_auth_config" "test" {
  toolkit_slug = "github"
  name         = "__NAME__"

  managed_auth = {
    restrict_to_following_tools = [
      "GITHUB_CREATE_ISSUE",
    ]
  }

  enabled = true
}
