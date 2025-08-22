#!/usr/bin/env bash
set -euo pipefail

# Read stdin JSON into variable
payload="$(cat)"
# Very basic path extraction (avoid jq dependency)
filePath="$(printf "%s" "$payload" | sed -n 's/.*"file_path":"$begin:math:text$[^"]*$end:math:text$".*/\1/p')"
if [[ -z "${filePath}" ]]; then
  # Not all tools send file_path; allow
  exit 0
fi

# Disallow edits to sensitive paths
case "$filePath" in
  */.env|*/.env.*|*/secrets/*|*/config/production/*)
    echo "Blocking edit to sensitive path: $filePath" 1>&2
    exit 2 # per Anthropic docs: exit 2 = blocking error
    ;;
esac

exit 0