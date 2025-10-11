---
asset_paths:
    - hooks/prompt-submit.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/user-prompt-submit.sh
    hook_type: UserPromptSubmit
    timeout: 10
display_name: "\U0001F4DD user-prompt-submit"
enabled: true
name: user-prompt-submit
type: hook
---

**Non-blocking prompt quality guidance hook.** Analyzes user prompts before submission and provides helpful tips.

This hook assists with prompt quality by:
- Detecting very short prompts (less than 10 characters) that might need more detail
- Checking if prompts contain action words or question words (what, how, why, please, help, fix, etc.)
- Suggesting improvements like "What is the desired outcome? Are there any constraints?"
- Injecting guidance messages into the conversation context (via stdout)
- **Never blocking** - always exits with code 0, providing suggestions only

Unlike validation hooks that block bad inputs, this hook takes a gentle approach by offering non-intrusive guidance that helps users craft better prompts without interrupting their workflow. All feedback is contextual and only appears when potentially helpful.