---
asset_paths:
    - hooks/session-start-context.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/session-start.sh
    hook_type: SessionStart
    timeout: 30
display_name: "\U0001F680 session-start"
enabled: true
name: session-start
type: hook
---

**Session initialization and context-loading hook.** Runs when a new Claude Code session begins. Provides project context, loads environment state, and can display helpful information about the current project.