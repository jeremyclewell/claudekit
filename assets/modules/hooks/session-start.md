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

**Session initialization and context-loading hook.** Runs automatically when a new Claude Code session begins.

This hook provides comprehensive project context by:
- Displaying git repository status (branch, uncommitted changes)
- Showing the last 10 git commits for recent activity
- Warning about uncommitted changes that might need attention
- Listing open GitHub issues (if `gh` CLI is available)
- Detecting project type (Go/Node.js/Rust/Python) and showing relevant version info
- Providing a clean, organized overview to help Claude understand your project state

The context gathering helps Claude provide more informed assistance by understanding what you've been working on recently and what the current state of your repository is.