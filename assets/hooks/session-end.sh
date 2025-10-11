#!/usr/bin/env bash
# Session End Hook
# Runs when Claude Code session terminates

set -euo pipefail

# Hook metadata
# hook_type: SessionEnd
# timeout: 30

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
LOG_FILE="$PROJECT_DIR/.claude/logs/sessions.log"

# Create log directory if needed
mkdir -p "$PROJECT_DIR/.claude/logs"

# Log session end
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Session ended" >> "$LOG_FILE"

# Check for uncommitted changes
if command -v git >/dev/null 2>&1 && [ -d "$PROJECT_DIR/.git" ]; then
    if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
        echo "  ⚠️  Uncommitted changes remain" >> "$LOG_FILE"
    fi
fi

exit 0
