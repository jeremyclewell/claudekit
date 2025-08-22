# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# claudekit - Claude Code Project Setup Tool

## Build & Development Commands

- `go build .` - Build the claudekit binary
- `go run .` - Run the interactive setup tool directly
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up dependencies

## Architecture Overview

claudekit is a command-line tool that creates a complete Claude Code project setup through an interactive TUI form. The tool generates:

- **CLAUDE.md** - Project documentation for Claude Code
- **.claude/settings.json** - Permissions, hooks, and environment configuration  
- **.claude/agents/*** - Specialized subagent definitions (code-reviewer, test-runner, etc.)
- **.claude/hooks/*** - Shell/Python scripts for lifecycle hooks
- **.claude/commands/*** - Custom slash commands
- **.mcp.json** - MCP server configurations

### Core Components

- `Config` struct (main.go:16-26) - Configuration collected from user input
- `settings` struct (main.go:38-46) - Claude Code settings schema with hooks and permissions  
- `renderClaudeMD()` (main.go:283) - Generates project-specific CLAUDE.md content
- `buildSettings()` (main.go:208) - Creates settings.json with hooks configuration
- Hook generators:
  - `preWriteGuardScript()` - Blocks edits to sensitive paths
  - `postWriteLintScript()` - Runs language-specific lints after writes
  - `sessionStartScript()` - Provides project context on session start
  - `promptLintPy()` - Python script for prompt validation

### Generated Agent Types

The tool supports these predefined agent types with specific tools and behaviors:
- `code-reviewer` - Code quality and security review
- `test-runner` - Test execution and failure fixing
- `bug-sleuth` - Root cause debugging
- `security-auditor` - Security vulnerability scanning
- `perf-optimizer` - Performance bottleneck identification
- `docs-writer` - Documentation generation
- `release-manager` - Release preparation
- `data-scientist` - Data analysis and SQL queries

### Language Support

The tool detects and configures for:
- **Go**: golangci-lint, go test, go build
- **TypeScript**: eslint, prettier, vitest/npm test
- **Python**: ruff, mypy, pytest

## Key Files

- `main.go` - Single-file application with all functionality
- `go.mod` - Dependencies (primarily Charm/Bubble Tea for TUI)