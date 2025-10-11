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

**Subagent completion tracking hook.** Triggered automatically when a specialized subagent finishes its work.

This hook tracks subagent activity by:
- Logging completion events to `.claude/logs/subagents.log`
- Recording subagent type, exit status, and timestamp
- Displaying success (✓) or failure (✗) status icons
- Providing environment variables: `CLAUDE_SUBAGENT_NAME`, `CLAUDE_SUBAGENT_TYPE`, `CLAUDE_SUBAGENT_EXIT_STATUS`

Subagents are specialized Claude instances that handle focused tasks like:
- `code-reviewer`: Code quality and security reviews
- `test-runner`: Test execution and debugging
- `security-auditor`: Vulnerability scanning
- `bug-sleuth`: Root cause analysis
- `perf-optimizer`: Performance profiling
- `docs-writer`: Documentation generation
- `release-manager`: Release preparation
- `data-scientist`: Data analysis

Use cases for customization:
- Parse test results and trigger notifications on failures
- Chain subagent workflows (code-reviewer → test-runner → release-manager)
- Collect metrics on subagent usage and success rates
- Trigger security alerts from security-auditor results
- Archive subagent reports and findings
- Send completion notifications for long-running agents

The default implementation logs all completions with status. Extend it to add workflow automation or metrics collection specific to your development process.