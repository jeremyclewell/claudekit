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

**Session cleanup and finalization hook.** Runs automatically when the Claude Code session terminates.

This hook performs end-of-session tasks by:
- Logging session end timestamp to `.claude/logs/sessions.log`
- Checking for uncommitted git changes and warning if any remain
- Providing a clean exit point for session-level cleanup

The hook serves as a reminder and audit trail for:
- Sessions that end with uncommitted work
- Session duration and timing patterns
- Custom cleanup logic you might want to add

Use cases for customization:
- Generate session summaries or reports
- Archive conversation transcripts
- Trigger backups or snapshots
- Send notifications about session completion
- Clean up temporary files or state
- Update project documentation based on session changes

The default implementation is minimal - just logging and git status checking - giving you a foundation to build on.