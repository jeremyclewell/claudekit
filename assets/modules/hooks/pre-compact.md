---
asset_paths:
  - hooks/pre-compact.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/pre-compact.sh
    hook_type: PreCompact
    timeout: 60
display_name: "\U0001F4E6 pre-compact"
enabled: true
name: pre-compact
type: hook
---

**Context compaction preparation hook.** Runs before Claude compacts the conversation history to save context space. Allows saving important conversation state or marking critical messages before they're summarized.