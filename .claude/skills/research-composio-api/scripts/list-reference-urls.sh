#!/usr/bin/env bash
# List Composio docs URLs under /reference from the live llms.txt index.
# Usage: list-reference-urls.sh [v3.1|v3|all]
set -euo pipefail

FILTER="${1:-v3.1}"
INDEX_URL="${COMPOSIO_LLMS_TXT:-https://docs.composio.dev/llms.txt}"

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

if ! curl -fsSL "${INDEX_URL}" -o "${tmp}"; then
	echo "error: failed to fetch ${INDEX_URL}" >&2
	exit 1
fi

case "${FILTER}" in
v3.1)
	# Current API reference: /reference/... but not /reference/v3/
	grep -E 'https://docs\.composio\.dev/reference' "${tmp}" |
		grep -v '/reference/v3' |
		grep -v 'sdk-reference' |
		sort -u
	;;
v3)
	grep -E 'https://docs\.composio\.dev/reference/v3' "${tmp}" | sort -u
	;;
all)
	grep -E 'https://docs\.composio\.dev/reference' "${tmp}" | sort -u
	;;
*)
	echo "usage: $0 [v3.1|v3|all]" >&2
	exit 2
	;;
esac
