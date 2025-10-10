---
asset_paths:
    - hooks/prewrite-guard.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/pre-tool-use.sh
    hook_type: PreToolUse
    timeout: 60
display_name: "\U0001F527 pre-tool-use"
enabled: true
name: pre-tool-use
type: hook
---

**Validation hook that runs before any tool use.** Guards against destructive operations and validates tool arguments before execution. Useful for preventing accidental writes to protected files or directories.