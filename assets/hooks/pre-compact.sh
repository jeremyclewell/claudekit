#!/bin/bash
# Pre-Compact Hook
# Runs before Claude compacts conversation history to save context space

set -euo pipefail

# Hook metadata
# hook_type: PreCompact
# timeout: 60

# This hook allows you to save important conversation state before compaction
# Example use cases:
# - Save critical conversation excerpts to a file
# - Mark important messages that shouldn't be summarized
# - Checkpoint current task state
# - Archive conversation for later review

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-.}"
COMPACT_LOG="$PROJECT_DIR/.claude/compact-history.log"

# Log compaction event with timestamp
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] Pre-compact hook triggered" >> "$COMPACT_LOG"

# Optional: Save conversation snapshot
if [ -n "${CLAUDE_CONVERSATION_ID:-}" ]; then
    echo "  Conversation ID: $CLAUDE_CONVERSATION_ID" >> "$COMPACT_LOG"
fi

# Optional: Check for important markers in conversation
# You could scan for TODO markers, decisions, or action items
# and preserve them before compaction

# Example: Create checkpoint file
CHECKPOINT_FILE="$PROJECT_DIR/.claude/checkpoints/$(date '+%Y%m%d_%H%M%S').txt"
mkdir -p "$PROJECT_DIR/.claude/checkpoints"

cat > "$CHECKPOINT_FILE" <<EOF
Conversation checkpoint before compaction
Timestamp: $TIMESTAMP
Project: $(basename "$PROJECT_DIR")

# Add any important state you want to preserve
# This could include:
# - Current task status
# - Important decisions made
# - Action items
# - Code locations being worked on
EOF

echo "  Checkpoint saved: $CHECKPOINT_FILE" >> "$COMPACT_LOG"

# Success
exit 0
