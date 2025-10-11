#!/usr/bin/env bash
# Session Start Hook
# Provides project context at the beginning of each Claude Code session

set -euo pipefail

# Hook metadata
# hook_type: SessionStart
# timeout: 30

echo "📋 Gathering project context..."
echo

# Git repository status
if command -v git >/dev/null && [ -d .git ]; then
    echo "=== Git Status ==="
    git status --short --branch 2>/dev/null || true
    echo

    echo "=== Recent Commits ==="
    git log --oneline --max-count=10 2>/dev/null || true
    echo

    # Uncommitted changes warning
    if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
        echo "⚠️  You have uncommitted changes"
        echo
    fi
fi

# GitHub issues (if gh CLI is available)
if command -v gh >/dev/null 2>&1; then
    echo "=== Open Issues ==="
    gh issue list --limit 5 2>/dev/null || true
    echo
fi

# Project stats
if [ -f go.mod ]; then
    echo "=== Go Project ==="
    go version 2>/dev/null || true
    echo "Module: $(grep '^module' go.mod | awk '{print $2}')"
    echo
elif [ -f package.json ]; then
    echo "=== Node.js Project ==="
    node --version 2>/dev/null || true
    npm --version 2>/dev/null || true
    echo
elif [ -f Cargo.toml ]; then
    echo "=== Rust Project ==="
    rustc --version 2>/dev/null || true
    echo
elif [ -f pyproject.toml ] || [ -f requirements.txt ]; then
    echo "=== Python Project ==="
    python --version 2>/dev/null || python3 --version 2>/dev/null || true
    echo
fi

echo "✅ Context gathered - ready to assist!"