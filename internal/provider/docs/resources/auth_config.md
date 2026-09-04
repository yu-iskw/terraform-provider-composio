Manages a Composio auth config.

Requires a project API key (`api_key` / `COMPOSIO_API_KEY`). Import uses the auth config id.

`Sensitive` hides values in the Terraform UI. It does not make state secret-free. `custom_auth.credentials` is write-only and is not stored in state. A credentials-only edit does not produce a plan.

Changing `toolkit_slug` forces replacement. Switching between `managed_auth` and `custom_auth` also forces replacement.
