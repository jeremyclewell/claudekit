#!/usr/bin/env bash
# Pre-Compact Hook
# Runs before Claude compacts conversation history to save context

set -euo pipefail

# Hook metadata
# hook_type: PreCompact
# timeout: 30

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
LOG_FILE="$PROJECT_DIR/.claude/logs/compact.log"

# Create log directory if needed
mkdir -p "$PROJECT_DIR/.claude/logs"

# Log compaction event
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Conversation compaction initiated" >> "$LOG_FILE"

# Note: Customize this hook to save important state before compaction
# Examples:
# - Save TODO list or action items to a file
# - Archive current task state
# - Mark important decisions or context

exit 0
