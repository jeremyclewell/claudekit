#!/usr/bin/env bash
set -euo pipefail

echo "Gathering project context..."
echo "Recent commits:"; git log --oneline -n 20 || true
echo
echo "Open issues (if gh installed):"; command -v gh >/dev/null && gh issue list --limit 10 || true