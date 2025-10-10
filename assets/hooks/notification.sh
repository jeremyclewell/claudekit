#!/bin/bash
# Notification Hook
# Sends desktop notifications when Claude completes long-running operations

set -euo pipefail

# Hook metadata
# hook_type: Notification
# timeout: 30

# Get notification details from environment
TITLE="${CLAUDE_NOTIFICATION_TITLE:-Claude Code}"
MESSAGE="${CLAUDE_NOTIFICATION_MESSAGE:-Task completed}"
URGENCY="${CLAUDE_NOTIFICATION_URGENCY:-normal}"

# Detect OS and send notification
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS - use osascript for Notification Center
    osascript -e "display notification \"$MESSAGE\" with title \"$TITLE\""

elif command -v notify-send &> /dev/null; then
    # Linux - use notify-send (libnotify)
    notify-send -u "$URGENCY" "$TITLE" "$MESSAGE"

elif command -v terminal-notifier &> /dev/null; then
    # Alternative macOS notifier
    terminal-notifier -title "$TITLE" -message "$MESSAGE"

else
    # Fallback - just echo to stderr
    echo "[$TITLE] $MESSAGE" >&2
fi

exit 0
