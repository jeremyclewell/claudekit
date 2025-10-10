#!/bin/bash
# Subagent Stop Hook
# Triggered when a specialized subagent finishes its work

set -euo pipefail

# Hook metadata
# hook_type: SubagentStop
# timeout: 30

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
SUBAGENT_LOG="$PROJECT_DIR/.claude/subagent-history.log"

# Get subagent details from environment
SUBAGENT_NAME="${CLAUDE_SUBAGENT_NAME:-unknown}"
SUBAGENT_TYPE="${CLAUDE_SUBAGENT_TYPE:-unknown}"
EXIT_STATUS="${CLAUDE_SUBAGENT_EXIT_STATUS:-0}"

# Log subagent completion
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Subagent stopped: $SUBAGENT_NAME (type: $SUBAGENT_TYPE, status: $EXIT_STATUS)" >> "$SUBAGENT_LOG"

# Optional: Collect subagent metrics
METRICS_FILE="$PROJECT_DIR/.claude/metrics/subagents.json"
mkdir -p "$PROJECT_DIR/.claude/metrics"

# Update metrics file (append to JSON array)
if [ ! -f "$METRICS_FILE" ]; then
    echo "[]" > "$METRICS_FILE"
fi

# Add new entry (simplified - in production you'd use jq)
cat >> "$SUBAGENT_LOG" <<EOF
  Subagent: $SUBAGENT_NAME
  Type: $SUBAGENT_TYPE
  Exit Status: $EXIT_STATUS
  Timestamp: $TIMESTAMP
EOF

# Optional: Subagent-specific actions
case "$SUBAGENT_TYPE" in
    "code-reviewer")
        # Code review completed - could trigger notifications
        echo "  Code review completed" >> "$SUBAGENT_LOG"
        ;;

    "test-runner")
        # Tests completed - could parse results
        if [ "$EXIT_STATUS" = "0" ]; then
            echo "  Tests passed ✓" >> "$SUBAGENT_LOG"
        else
            echo "  Tests failed ✗" >> "$SUBAGENT_LOG"
        fi
        ;;

    "security-auditor")
        # Security scan completed - could trigger alerts
        echo "  Security audit completed" >> "$SUBAGENT_LOG"
        ;;

    "bug-sleuth")
        # Debugging session ended - could save findings
        echo "  Debug session completed" >> "$SUBAGENT_LOG"
        ;;

    "perf-optimizer")
        # Performance optimization done - could save benchmarks
        echo "  Performance optimization completed" >> "$SUBAGENT_LOG"
        ;;

    "docs-writer")
        # Documentation updated - could trigger doc rebuild
        echo "  Documentation update completed" >> "$SUBAGENT_LOG"
        ;;

    "release-manager")
        # Release prepared - could trigger deployment
        echo "  Release preparation completed" >> "$SUBAGENT_LOG"
        ;;

    "data-scientist")
        # Analysis completed - could save visualizations
        echo "  Data analysis completed" >> "$SUBAGENT_LOG"
        ;;

    *)
        echo "  Unknown subagent type: $SUBAGENT_TYPE" >> "$SUBAGENT_LOG"
        ;;
esac

# Optional: Chain subagent workflows
# If one subagent completes successfully, you could trigger another
# Example: code-reviewer -> test-runner -> release-manager

# Optional: Send completion notification
if [ "$EXIT_STATUS" = "0" ]; then
    # Success notification
    if command -v osascript &> /dev/null; then
        osascript -e "display notification \"$SUBAGENT_NAME completed successfully\" with title \"Claude Code - Subagent\"" 2>/dev/null || true
    fi
fi

# Success
exit 0
