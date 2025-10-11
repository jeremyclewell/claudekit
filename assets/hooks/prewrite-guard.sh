#!/usr/bin/env bash
# PreWrite Guard Hook
# Blocks edits to sensitive files and paths

set -euo pipefail

# Hook metadata
# hook_type: PreWrite
# timeout: 10

# Read stdin JSON into variable
payload="$(cat)"

# Extract file_path from JSON (avoid jq dependency for portability)
# Matches: "file_path": "some/path/file.txt"
filePath="$(printf "%s" "$payload" | grep -o '"file_path":"[^"]*"' | cut -d'"' -f4)"

# If no file_path found, allow the operation
if [[ -z "${filePath}" ]]; then
  exit 0
fi

# Disallow edits to sensitive paths
case "$filePath" in
  .env|*/.env|.env.*|*/.env.*|*/secrets/*|*/config/production/*|*/.git/*|*.key|*.pem)
    echo "❌ Blocked edit to sensitive path: $filePath" >&2
    exit 2 # exit 2 = blocking error (shown to user)
    ;;
esac

exit 0