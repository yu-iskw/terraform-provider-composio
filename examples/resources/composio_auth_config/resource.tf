data "composio_toolkit" "github" {
  slug = "github"
}

resource "composio_auth_config" "github" {
  toolkit_slug = data.composio_toolkit.github.slug

  managed_auth = {
    restrict_to_following_tools = [
      "GITHUB_CREATE_ISSUE",
      "GITHUB_GET_ISSUE",
    ]
  }

  enabled = true
}
