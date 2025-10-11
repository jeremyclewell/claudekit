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

**Pre-write validation hook that blocks edits to sensitive files.** Runs before Claude writes or edits any file.

This hook acts as a safety guard by:
- Parsing the file path from the write/edit tool request
- Blocking writes to sensitive files and directories including:
  - Environment files: `.env`, `.env.*` (root or any subdirectory)
  - Secret directories: `*/secrets/*`
  - Production configs: `*/config/production/*`
  - Git internals: `*/.git/*`
  - Private keys: `*.key`, `*.pem`
- Displaying a clear error message when blocking (exit code 2)
- Allowing all other file operations to proceed normally (exit code 0)

This prevents accidental commits of secrets or credentials and protects critical configuration from unintended modifications. Customize the case patterns to match your project's security requirements.