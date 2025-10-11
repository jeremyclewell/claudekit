#!/usr/bin/env bash
# Subagent Stop Hook
# Triggered when a specialized subagent finishes its work

set -euo pipefail

# Hook metadata
# hook_type: SubagentStop
# timeout: 30

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
LOG_FILE="$PROJECT_DIR/.claude/logs/subagents.log"

# Create log directory if needed
mkdir -p "$PROJECT_DIR/.claude/logs"

# Get subagent details from environment (if available)
SUBAGENT_NAME="${CLAUDE_SUBAGENT_NAME:-unknown}"
SUBAGENT_TYPE="${CLAUDE_SUBAGENT_TYPE:-unknown}"
EXIT_STATUS="${CLAUDE_SUBAGENT_EXIT_STATUS:-0}"

# Log subagent completion
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
STATUS_ICON="✓"
if [ "$EXIT_STATUS" != "0" ]; then
    STATUS_ICON="✗"
fi

echo "[$TIMESTAMP] $STATUS_ICON $SUBAGENT_TYPE completed (status: $EXIT_STATUS)" >> "$LOG_FILE"

# Note: Customize this hook for subagent-specific actions
# Examples:
# - Parse test results (test-runner)
# - Trigger security alerts (security-auditor)
# - Chain workflows (code-reviewer -> test-runner)
# - Send notifications on completion

exit 0
