---
asset_paths:
  - hooks/stop.sh
category: lifecycle
defaults:
    command: $CLAUDE_PROJECT_DIR/.claude/hooks/stop.sh
    hook_type: Stop
    timeout: 30
display_name: "\U0001F3C1 stop"
enabled: true
name: stop
type: hook
---

**Interruption handler hook.** Runs immediately when the user issues a stop command or cancels a response.

This hook handles user interruptions by:
- Logging stop events with timestamps to `.claude/logs/stop-events.log`
- Recording that an operation was stopped by user request
- Providing a hook point for interruption handling logic

Stop events occur when:
- User explicitly cancels a long-running operation
- User interrupts Claude mid-response
- User wants to abort the current task and redirect

Use cases for customization:
- Save partial work state for later resumption
- Clean up temporary files from interrupted operations
- Send notifications about stopped tasks
- Log context about what was interrupted
- Roll back incomplete multi-step operations
- Preserve draft content that wasn't finalized

The default implementation is lightweight - just logging the event - giving you a clean foundation to add recovery or cleanup logic specific to your workflow.