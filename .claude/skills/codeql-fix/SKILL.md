---
name: codeql-fix
description: Run CodeQL security/quality analysis and fix findings. Use when the user asks to run CodeQL, security scan, static analysis, or fix CodeQL findings.
compatibility: Requires [CodeQL CLI](https://github.com/github/codeql-cli-binaries/releases) on PATH (e.g. brew install codeql). Go toolchain matching [`go.mod`](../../../go.mod) / `toolchain`. Run `go mod download` (or rely on the traced `go build` in [`dev/codeql.sh`](../../../dev/codeql.sh)) before or during database creation. Optional GitHub workflow [`.github/workflows/codeql.yml`](../../../.github/workflows/codeql.yml) is not shipped with this template; local flow matches [`dev/codeql.sh`](../../../dev/codeql.sh).
---

# CodeQL Fix

Use when the user asks to run CodeQL or static analysis, or to fix CodeQL findings (see frontmatter `description`).

## Preconditions

- [CodeQL CLI](https://github.com/github/codeql-cli-binaries/releases) on `PATH` (e.g. `brew install codeql`).
- **Go** installed and compatible with [`go.mod`](../../../go.mod) (including the `toolchain` directive when present).

## Run analysis (repository root)

All commands below assume `cd "$(git rev-parse --show-toplevel)"`.

Do not commit CodeQL databases or SARIF outputs (large, machine-specific). They belong in [`.gitignore`](../../../.gitignore) (for example `.codeql_db/`, `codeql-results.sarif`).

### 1. Preferred: `dev/codeql.sh`

```bash
./dev/codeql.sh
```

This runs **`go mod download`**, creates **`.codeql_db`** with **`go`** using a traced **`go build ./...`**, analyzes with **`codeql/go-queries:codeql-suites/go-security-and-quality.qls`**, writes **`codeql-results.sarif`**, and passes **`--download`** to resolve query packs.

**Autobuild alternative:** omit `--command` and use the Go autobuilder, or set **`CODEQL_EXTRACTOR_GO_BUILD_TRACING=on`** per [CodeQL Go docs](https://docs.github.com/en/code-security/codeql-cli/getting-started-with-the-codeql-cli/preparing-your-code-for-codeql-analysis#go). This template’s script uses an explicit **`go build ./...`** for predictable extraction.

### 2. Manual CLI (equivalent to the script)

After **`go mod download`** (recommended):

```bash
codeql database create .codeql_db --language=go --source-root . --overwrite \
  --command='go build ./...'
```

Analyze and emit SARIF:

```bash
codeql database analyze .codeql_db \
  "codeql/go-queries:codeql-suites/go-security-and-quality.qls" \
  --format=sarif-latest \
  --output=codeql-results.sarif \
  --download
```

- For a narrower suite closer to default GitHub code scanning, use `codeql/go-queries:codeql-suites/go-code-scanning.qls` instead.
- If packs are missing and you are not using `--download`, run `codeql pack download codeql/go-queries` once.

View SARIF in the VS Code SARIF extension (or upload where your org uses code scanning).

### 3. Optional: code scanning config (`paths-ignore`)

Use the renderer when you want `paths-ignore` for large or generated trees, hand-edited query blocks, or parity with GitHub code scanning YAML.

```bash
REPO="$(git rev-parse --show-toplevel)"
"$REPO/.claude/skills/codeql-fix/scripts/render-code-scanning-config.sh" "$REPO" /tmp/codeql-config.yml
codeql database create .codeql_db --language=go --source-root . \
  --command='go build ./...' \
  --codescanning-config=/tmp/codeql-config.yml --overwrite
```

Then run `codeql database analyze` as in section 2. See [references/code-scanning-config.md](references/code-scanning-config.md).

## Fixer loop

If the relevant SARIF has an empty `runs[].results` array, there are **no CodeQL alerts to fix** for that suite; stop unless the user wants a broader suite or diagnostic queries.

When SARIF findings remain:

1. **Identify:** Read the SARIF or CLI output for reported findings.
2. **Fix:** Apply the minimum necessary edit to resolve each finding.
3. **Verify:** From the repository root, run **`make test`**, then **`make lint`** (see [AGENTS.md](../../../AGENTS.md)).
4. **Re-scan:** Run `./dev/codeql.sh` or repeat the manual create + analyze steps until clean or up to 3 iterations to avoid unbounded loops.

## Optional: code scanning config details

See [references/code-scanning-config.md](references/code-scanning-config.md) and the official [code scanning configuration](https://aka.ms/code-scanning-docs/config-file) reference.
