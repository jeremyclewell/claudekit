---
asset_paths:
    - hooks/postwrite-lint.sh.tmpl
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/post-tool-use.sh
    hook_type: PostToolUse
    timeout: 120
display_name: ✅ post-tool-use
enabled: true
name: post-tool-use
type: hook
---

**Post-processing hook that runs after tool execution.** Runs language-specific linters and formatters after file writes. Automatically catches syntax errors and style issues immediately after code generation.