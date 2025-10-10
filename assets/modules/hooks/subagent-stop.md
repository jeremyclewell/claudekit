---
asset_paths:
  - hooks/subagent-stop.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/subagent-stop.sh
    hook_type: SubagentStop
    timeout: 30
display_name: "\U0001F916 subagent-stop"
enabled: true
name: subagent-stop
type: hook
---

**Subagent completion hook.** Triggered when a specialized subagent finishes its work. Useful for collecting metrics, validating subagent outputs, or chaining subagent workflows.