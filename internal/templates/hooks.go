package templates

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"jeremyclewell.com/claudekit/internal/util"
)

// PostWriteLintScript generates the post-write linting hook script based on selected languages
func PostWriteLintScript(langs []string, assetsFS embed.FS) string {
	tmplContent, err := assetsFS.ReadFile("assets/hooks/postwrite-lint.sh.tmpl")
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("postwrite-lint").Parse(string(tmplContent))
	if err != nil {
		panic(err)
	}

	data := struct {
		HasGo         bool
		HasTypeScript bool
		HasPython     bool
		HasRust       bool
		HasCpp        bool
		HasJava       bool
		HasCsharp     bool
		HasPhp        bool
		HasRuby       bool
		HasSwift      bool
		HasDart       bool
		HasShell      bool
		HasLua        bool
		HasElixir     bool
		HasHaskell    bool
		HasElm        bool
		HasJulia      bool
		HasSql        bool
	}{
		HasGo:         util.Includes(langs, "Go"),
		HasTypeScript: util.Includes(langs, "TypeScript"),
		HasPython:     util.Includes(langs, "Python"),
		HasRust:       util.Includes(langs, "Rust"),
		HasCpp:        util.Includes(langs, "C++"),
		HasJava:       util.Includes(langs, "Java") || util.Includes(langs, "Kotlin"),
		HasCsharp:     util.Includes(langs, "C#"),
		HasPhp:        util.Includes(langs, "PHP"),
		HasRuby:       util.Includes(langs, "Ruby"),
		HasSwift:      util.Includes(langs, "Swift"),
		HasDart:       util.Includes(langs, "Dart"),
		HasShell:      util.Includes(langs, "Shell"),
		HasLua:        util.Includes(langs, "Lua"),
		HasElixir:     util.Includes(langs, "Elixir"),
		HasHaskell:    util.Includes(langs, "Haskell"),
		HasElm:        util.Includes(langs, "Elm"),
		HasJulia:      util.Includes(langs, "Julia"),
		HasSql:        util.Includes(langs, "SQL"),
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		panic(err)
	}
	return b.String()
}

// GenerateHookScript creates a template hook script (bash or python) based on the hook name
func GenerateHookScript(hookName, description string) string {
	if strings.HasSuffix(hookName, ".py") || strings.Contains(hookName, "prompt") {
		// Generate Python script for Python-based hooks
		return fmt.Sprintf(`#!/usr/bin/env python3
"""
%s Hook - %s

This hook is called by Claude Code during specific events.
You can customize this script to add logging, validation, or other actions.

Environment variables available:
- CLAUDE_PROJECT_DIR: Current project directory
- CLAUDE_SESSION_ID: Current session identifier
- CLAUDE_USER_MESSAGE: User's message (for prompt hooks)
- CLAUDE_TOOL_NAME: Tool name (for tool hooks)
- CLAUDE_TOOL_ARGS: Tool arguments (for tool hooks)
"""

import os
import sys
from datetime import datetime

def main():
    print(f"[{datetime.now().isoformat()}] %s hook triggered")

    # Add your custom logic here
    # Example: Log to file, send notifications, validate inputs, etc.

    # Return 0 for success, non-zero for failure
    return 0

if __name__ == "__main__":
    sys.exit(main())
`, hookName, description, hookName)
	} else {
		// Generate bash script for shell-based hooks
		return fmt.Sprintf(`#!/usr/bin/env bash
# %s Hook - %s
#
# This hook is called by Claude Code during specific events.
# You can customize this script to add logging, validation, or other actions.
#
# Environment variables available:
# - CLAUDE_PROJECT_DIR: Current project directory
# - CLAUDE_SESSION_ID: Current session identifier
# - CLAUDE_USER_MESSAGE: User's message (for prompt hooks)
# - CLAUDE_TOOL_NAME: Tool name (for tool hooks)
# - CLAUDE_TOOL_ARGS: Tool arguments (for tool hooks)

echo "[$(date -Iseconds)] %s hook triggered"

# Add your custom logic here
# Examples:
# - Log events: echo "Event logged" >> "$CLAUDE_PROJECT_DIR/.claude/hooks.log"
# - Send notifications: curl -X POST ...
# - Validate inputs: [[ "$CLAUDE_TOOL_NAME" == "Write" ]] && echo "Validating write operation"

# Return 0 for success, non-zero for failure
exit 0
`, hookName, description, hookName)
	}
}

// PreWriteGuardScript returns the pre-write guard hook script from assets
func PreWriteGuardScript(assetsFS embed.FS) string {
	content, err := assetsFS.ReadFile("assets/hooks/prewrite-guard.sh")
	if err != nil {
		panic(err)
	}
	// Strip the shebang and set -euo since writeExecutable adds them
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(lines[0], "set -euo pipefail") {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}

// SessionStartScript returns the session start hook script from assets
func SessionStartScript(assetsFS embed.FS) string {
	content, err := assetsFS.ReadFile("assets/hooks/session-start-context.sh")
	if err != nil {
		panic(err)
	}
	// Strip the shebang and set -euo since writeExecutable adds them
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "#!") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(lines[0], "set -euo pipefail") {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}

// PromptLintPy returns the prompt linting Python script from assets
func PromptLintPy(assetsFS embed.FS) string {
	content, err := assetsFS.ReadFile("assets/hooks/prompt-lint.py")
	if err != nil {
		panic(err)
	}
	return string(content)
}
