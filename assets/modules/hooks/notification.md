---
asset_paths:
  - hooks/notification.sh
category: integration
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/notification.sh
    hook_type: Notification
    timeout: 30
display_name: "\U0001F514 notification"
enabled: true
name: notification
type: hook
---

**Desktop notification system for long-running operations.** Sends OS-level notifications when Claude completes tasks, allowing you to work on other things while waiting. Integrates with macOS Notification Center and Linux notify-send.