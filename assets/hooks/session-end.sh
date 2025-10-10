#!/bin/bash
# Session End Hook
# Runs when Claude Code session terminates

set -euo pipefail

# Hook metadata
# hook_type: SessionEnd
# timeout: 30

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
SESSION_LOG="$PROJECT_DIR/.claude/session-history.log"

# Log session end
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Session ended" >> "$SESSION_LOG"

# Optional: Generate session summary
if [ -n "${CLAUDE_SESSION_START:-}" ]; then
    DURATION=$(($(date +%s) - CLAUDE_SESSION_START))
    HOURS=$((DURATION / 3600))
    MINUTES=$(((DURATION % 3600) / 60))
    echo "  Duration: ${HOURS}h ${MINUTES}m" >> "$SESSION_LOG"
fi

# Optional: Archive conversation logs
# This is a good place to:
# - Save conversation transcripts
# - Generate summary of changes made
# - Update project documentation
# - Clean up temporary files

# Optional: Run final validation checks
# Example: Check for uncommitted changes
if command -v git &> /dev/null && [ -d "$PROJECT_DIR/.git" ]; then
    if ! git diff --quiet 2>/dev/null; then
        echo "  Warning: Uncommitted changes detected" >> "$SESSION_LOG"
    fi
fi

# Optional: Save session metrics
if [ -n "${CLAUDE_METRICS_FILE:-}" ]; then
    cat > "${CLAUDE_METRICS_FILE}" <<EOF
{
  "session_end": "$TIMESTAMP",
  "duration_seconds": ${DURATION:-0},
  "project": "$(basename "$PROJECT_DIR")"
}
EOF
fi

# Optional: Cleanup
# Remove temporary files, cache, etc.
# Be careful not to remove important working files

# Success
exit 0
