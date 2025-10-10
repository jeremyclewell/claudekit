---
asset_paths:
    - hooks/prompt-lint.py
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/user-prompt-submit.py
    hook_type: UserPromptSubmit
    timeout: 30
display_name: "\U0001F4DD user-prompt-submit"
enabled: true
name: user-prompt-submit
type: hook
---

**Prompt validation and enrichment hook.** Validates user prompts before submission, checks for clarity issues, and can automatically append context. Helps prevent ambiguous requests and improves response quality.