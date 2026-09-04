# Security Policy

## Supported Versions

Document supported provider release lines here as versions are published.

## Credentials and state

Provider API keys are sensitive. Prefer environment variables in CI. `Sensitive` hides values in the Terraform UI. It does not make Terraform state secret-free. Encrypt and restrict access to the state backend.

Do not log `x-api-key`, `x-org-api-key`, OAuth client secrets, or webhook signing secrets.

## Reporting A Vulnerability

Use the repository's private vulnerability reporting channel or security contact. Include:

- A clear description of the issue
- Steps to reproduce
- Potential impact
- Suggested remediation, if known

Do not include secrets, credentials, or production data in reports.
