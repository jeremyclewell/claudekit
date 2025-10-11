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

**Context compaction preparation hook.** Runs before Claude compacts the conversation history to save context space.

This hook logs compaction events by:
- Creating `.claude/logs/compact.log` to track when compactions occur
- Recording timestamps for each compaction event
- Providing a foundation for custom state-saving logic

Compaction happens when the conversation becomes too long and Claude needs to summarize older messages to free up context space. This hook gives you an opportunity to:
- Save TODO lists or action items before they're summarized
- Archive current task state for later review
- Mark important decisions or context that should be preserved
- Log conversation checkpoints

The default implementation simply logs the event. Customize this hook to add your own state preservation logic based on your workflow needs.