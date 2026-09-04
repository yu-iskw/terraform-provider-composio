Manages a Composio auth config.

Requires a project API key (`api_key` / `COMPOSIO_API_KEY`). Import uses the auth config id.

`Sensitive` hides values in the Terraform UI. It does not make state secret-free. `custom_auth.credentials` is write-only and is not stored in state.

Changing `toolkit_slug` forces replacement.
