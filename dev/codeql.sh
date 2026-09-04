#!/usr/bin/env bash
# Local CodeQL: Go database + security-and-quality suite -> codeql-results.sarif
# Requires: codeql on PATH, Go toolchain matching go.mod.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z ${ROOT} ]]; then
	echo "Run from inside a git checkout." >&2
	exit 1
fi
cd "${ROOT}"

if ! command -v codeql >/dev/null 2>&1; then
	echo "codeql not found on PATH (e.g. brew install codeql)." >&2
	exit 1
fi

go mod download

codeql database create .codeql_db \
	--language=go \
	--source-root . \
	--overwrite \
	--command='go build ./...'

codeql database analyze .codeql_db \
	"codeql/go-queries:codeql-suites/go-security-and-quality.qls" \
	--format=sarif-latest \
	--output=codeql-results.sarif \
	--download

echo "Wrote ${ROOT}/codeql-results.sarif" >&2
