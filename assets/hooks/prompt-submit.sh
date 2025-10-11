#!/usr/bin/env bash
# Prompt Submit Hook
# Provides guidance for better prompts without being overly restrictive

set -euo pipefail

# Hook metadata
# hook_type: PromptSubmit
# timeout: 10

# Read stdin JSON
payload="$(cat)"

# Extract prompt text from JSON (avoid jq dependency)
prompt="$(printf "%s" "$payload" | grep -o '"prompt":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "")"

# If no prompt found, allow the operation
if [[ -z "${prompt}" ]]; then
  exit 0
fi

# Provide helpful guidance via stdout (injected into context)
# This is non-blocking - just adds suggestions to the conversation

# Check for very short prompts (might need more detail)
if [[ "${#prompt}" -lt 10 ]]; then
  echo "💡 Tip: More detailed prompts help me provide better assistance."
  exit 0
fi

# Check if prompt seems well-formed (contains action words or questions)
if [[ ! "$prompt" =~ (what|how|why|when|where|can|should|could|would|please|help|fix|add|update|create|refactor|implement) ]]; then
  # Prompt might benefit from more context
  echo "💡 Consider adding: What is the desired outcome? Are there any constraints?"
fi

# Always exit 0 - this is guidance only, not blocking
exit 0
