---
asset_paths:
  - hooks/session-end.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/session-end.sh
    hook_type: SessionEnd
    timeout: 30
display_name: "\U0001F44B session-end"
enabled: true
name: session-end
type: hook
---

**Session cleanup and summary hook.** Runs when the Claude Code session terminates. Perfect for generating session summaries, archiving conversation logs, or running final validation checks.