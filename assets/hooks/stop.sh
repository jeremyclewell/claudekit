#!/usr/bin/env bash
# Stop Hook
# Triggered when user issues stop command or cancels a response

set -euo pipefail

# Hook metadata
# hook_type: Stop
# timeout: 10

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
LOG_FILE="$PROJECT_DIR/.claude/logs/stop-events.log"

# Create log directory if needed
mkdir -p "$PROJECT_DIR/.claude/logs"

# Log stop event
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Operation stopped by user" >> "$LOG_FILE"

# Note: Customize this hook to handle interrupted operations
# Examples:
# - Save partial work state
# - Clean up temporary files
# - Send notifications

exit 0
