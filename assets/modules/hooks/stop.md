---
asset_paths:
  - hooks/stop.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/stop.sh
    hook_type: Stop
    timeout: 30
display_name: "\U0001F3C1 stop"
enabled: true
name: stop
type: hook
---

**Cleanup hook triggered when conversation stops.** Runs when the user issues a stop command or cancels a response. Useful for cleaning up temporary files, saving state, or logging incomplete operations.