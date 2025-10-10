#!/bin/bash
# Stop Hook
# Triggered when user issues stop command or cancels a response

set -euo pipefail

# Hook metadata
# hook_type: Stop
# timeout: 30

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
STOP_LOG="$PROJECT_DIR/.claude/stop-events.log"

# Log stop event
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Stop command received" >> "$STOP_LOG"

# Optional: Log context about what was interrupted
if [ -n "${CLAUDE_CURRENT_TASK:-}" ]; then
    echo "  Interrupted task: $CLAUDE_CURRENT_TASK" >> "$STOP_LOG"
fi

# Optional: Save incomplete state
# This is useful for resuming work later
STATE_FILE="$PROJECT_DIR/.claude/interrupted-state.json"
cat > "$STATE_FILE" <<EOF
{
  "timestamp": "$TIMESTAMP",
  "interrupted_at": "$(date '+%Y-%m-%dT%H:%M:%S%z')",
  "task": "${CLAUDE_CURRENT_TASK:-unknown}",
  "reason": "user_stop"
}
EOF

# Optional: Cleanup partial operations
# If Claude was in the middle of:
# - File writes: May want to save drafts
# - Long-running commands: May want to clean up processes
# - Multi-step operations: May want to checkpoint progress

# Check for temporary files that might need cleanup
TEMP_DIR="$PROJECT_DIR/.claude/temp"
if [ -d "$TEMP_DIR" ]; then
    # Move temp files to recovery location instead of deleting
    RECOVERY_DIR="$PROJECT_DIR/.claude/recovery/$(date '+%Y%m%d_%H%M%S')"
    mkdir -p "$RECOVERY_DIR"

    if [ "$(ls -A "$TEMP_DIR" 2>/dev/null)" ]; then
        mv "$TEMP_DIR"/* "$RECOVERY_DIR"/ 2>/dev/null || true
        echo "  Temporary files moved to: $RECOVERY_DIR" >> "$STOP_LOG"
    fi
fi

# Optional: Send notification about interruption
if command -v osascript &> /dev/null; then
    osascript -e 'display notification "Operation stopped by user" with title "Claude Code"' 2>/dev/null || true
elif command -v notify-send &> /dev/null; then
    notify-send "Claude Code" "Operation stopped by user" 2>/dev/null || true
fi

# Success
exit 0
