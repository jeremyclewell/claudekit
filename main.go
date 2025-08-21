package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	huh "github.com/charmbracelet/huh"
)

type Config struct {
	ProjectDir     string
	ProjectName    string
	Languages      []string
	Subagents      []string
	Hooks          []string
	WantSlashCmd   bool
	WantMCP        bool
	MCPServers     []string
	ClaudeMDExtras string
}

// Hook structs follow Anthropic's hooks schema.
type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}
type hookMatcher struct {
	Matcher string    `json:"matcher,omitempty"`
	Hooks   []hookCmd `json:"hooks"`
}
type settings struct {
	Permissions *struct {
		Allow []string `json:"allow,omitempty"`
		Ask   []string `json:"ask,omitempty"`
		Deny  []string `json:"deny,omitempty"`
	} `json:"permissions,omitempty"`
	Hooks map[string][]hookMatcher `json:"hooks,omitempty"`
	Env   map[string]string        `json:"env,omitempty"`
}

func main() {
	cfg := Config{
		Languages:  []string{"Go"},
		Subagents:  []string{"code-reviewer", "test-runner", "bug-sleuth"},
		Hooks:      []string{"pre-write-guard", "post-write-lint", "session-start", "prompt-lint"},
		WantSlashCmd: true,
		WantMCP:      true,
		MCPServers:   []string{"notion", "linear", "sentry", "github"},
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Project directory").Placeholder(".").Value(&cfg.ProjectDir),
			huh.NewInput().Title("Project name").Placeholder("awesome-app").Value(&cfg.ProjectName),
			huh.NewMultiSelect[string]().
				Title("Primary languages (for defaults)").
				Options(huh.NewOptions("Go", "TypeScript", "Python", "Java", "Rust")...).
				Value(&cfg.Languages),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Subagents to add").
				Options(huh.NewOptions(
					"code-reviewer", "test-runner", "bug-sleuth", "security-auditor",
					"perf-optimizer", "docs-writer", "release-manager", "data-scientist")...).
				Value(&cfg.Subagents),
			huh.NewMultiSelect[string]().
				Title("Hooks to enable").
				Options(huh.NewOptions("pre-write-guard", "post-write-lint", "session-start", "prompt-lint")...).
				Value(&cfg.Hooks),
			huh.NewConfirm().Title("Add example slash command (/project:fix-github-issue)?").Value(&cfg.WantSlashCmd),
		),
		huh.NewGroup(
			huh.NewConfirm().Title("Configure MCP now?").Value(&cfg.WantMCP),
			huh.NewMultiSelect[string]().
				Title("MCP servers (project scope)").
				Options(huh.NewOptions("notion", "linear", "sentry", "github", "airtable")...).
				Value(&cfg.MCPServers),
		),
		huh.NewGroup(
			huh.NewText().Title("Extra content for CLAUDE.md (optional)").Value(&cfg.ClaudeMDExtras),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cancelled: %v\n", err)
		os.Exit(1)
	}

	if cfg.ProjectDir == "" {
		cfg.ProjectDir = "."
	}
	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\n✅ claudekit finished. Open Claude Code in this repo and start coding!")
}

func run(cfg Config) error {
	abs, err := filepath.Abs(cfg.ProjectDir)
	if err != nil {
		return err
	}
	// Create directories
	mustMkdir(filepath.Join(abs, ".claude"))
	mustMkdir(filepath.Join(abs, ".claude", "agents"))
	mustMkdir(filepath.Join(abs, ".claude", "hooks"))
	if cfg.WantSlashCmd {
		mustMkdir(filepath.Join(abs, ".claude", "commands"))
	}

	// Write CLAUDE.md
	if err := os.WriteFile(filepath.Join(abs, "CLAUDE.md"),
		[]byte(renderClaudeMD(cfg)), 0o644); err != nil {
		return err
	}

	// Write subagents
	for _, a := range cfg.Subagents {
		path := filepath.Join(abs, ".claude", "agents", a+".md")
		if err := os.WriteFile(path, []byte(renderAgent(a)), 0o644); err != nil {
			return err
		}
	}

	// Write hooks scripts
	if contains(cfg.Hooks, "pre-write-guard") {
		if err := writeExecutable(filepath.Join(abs, ".claude", "hooks", "prewrite-guard.sh"), preWriteGuardScript()); err != nil {
			return err
		}
	}
	if contains(cfg.Hooks, "post-write-lint") {
		if err := writeExecutable(filepath.Join(abs, ".claude", "hooks", "postwrite-lint.sh"), postWriteLintScript(cfg.Languages)); err != nil {
			return err
		}
	}
	if contains(cfg.Hooks, "session-start") {
		if err := writeExecutable(filepath.Join(abs, ".claude", "hooks", "session-start-context.sh"), sessionStartScript()); err != nil {
			return err
		}
	}
	if contains(cfg.Hooks, "prompt-lint") {
		if err := writeExecutable(filepath.Join(abs, ".claude", "hooks", "prompt-lint.py"), promptLintPy()); err != nil {
			return err
		}
	}

	// Write settings.json with hooks + permissions
	st := buildSettings(abs, cfg)
	buf, _ := json.MarshalIndent(st, "", "  ")
	if err := os.WriteFile(filepath.Join(abs, ".claude", "settings.json"), buf, 0o644); err != nil {
		return err
	}

	// Slash command example
	if cfg.WantSlashCmd {
		if err := os.WriteFile(
			filepath.Join(abs, ".claude", "commands", "fix-github-issue.md"),
			[]byte(sampleSlashCommand()), 0o644); err != nil {
			return err
		}
	}

	// MCP project config
	if cfg.WantMCP {
		mcp := buildMCPJSON(cfg.MCPServers)
		if err := os.WriteFile(filepath.Join(abs, ".mcp.json"), []byte(mcp), 0o644); err != nil {
			return err
		}
	}

	// Gentle reminder if claude CLI is missing
	if _, err := exec.LookPath("claude"); err != nil {
		fmt.Println("\nℹ️  Claude Code CLI not found on PATH. Install with:")
		fmt.Println("   curl -fsSL https://claude.ai/install.sh | bash   # macOS/Linux/WSL")
	}

	return nil
}

func mustMkdir(p string) {
	_ = os.MkdirAll(p, 0o755)
}
func writeExecutable(path string, content string) error {
	if strings.HasSuffix(path, ".py") {
		return os.WriteFile(path, []byte(content), 0o755)
	}
	return os.WriteFile(path, []byte("#!/usr/bin/env bash\nset -euo pipefail\n"+content+"\n"), 0o755)
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func buildSettings(projectDir string, cfg Config) settings {
	s := settings{
		Permissions: &struct {
			Allow []string `json:"allow,omitempty"`
			Ask   []string `json:"ask,omitempty"`
			Deny  []string `json:"deny,omitempty"`
		}{
			Allow: []string{"Read", "LS", "Grep", "Glob"},
			Ask:   []string{"Bash(git *:*)", "WebFetch"},
			Deny:  []string{"Read(./.env)", "Read(./.env.*)", "Read(./secrets/**)"},
		},
		Env: map[string]string{
			"CLAUDE_CODE_MAX_OUTPUT_TOKENS": "8192",
			"MCP_TOOL_TIMEOUT":              "180000",
		},
		Hooks: map[string][]hookMatcher{},
	}

	// PreToolUse: guard write/edit, matchers are case-sensitive per docs.
	if contains(cfg.Hooks, "pre-write-guard") {
		s.Hooks["PreToolUse"] = append(s.Hooks["PreToolUse"],
			hookMatcher{
				Matcher: "Write|Edit|MultiEdit",
				Hooks: []hookCmd{{
					Type:    "command",
					Command: "$CLAUDE_PROJECT_DIR/.claude/hooks/prewrite-guard.sh",
					Timeout: 60,
				}},
			},
		)
	}

	// PostToolUse: run lints/tests after writes/edits
	if contains(cfg.Hooks, "post-write-lint") {
		s.Hooks["PostToolUse"] = append(s.Hooks["PostToolUse"],
			hookMatcher{
				Matcher: "Write|Edit|MultiEdit",
				Hooks: []hookCmd{{
					Type:    "command",
					Command: "$CLAUDE_PROJECT_DIR/.claude/hooks/postwrite-lint.sh",
					Timeout: 120,
				}},
			},
		)
	}

	// SessionStart
	if contains(cfg.Hooks, "session-start") {
		s.Hooks["SessionStart"] = append(s.Hooks["SessionStart"],
			hookMatcher{
				Hooks: []hookCmd{{
					Type:    "command",
					Command: "$CLAUDE_PROJECT_DIR/.claude/hooks/session-start-context.sh",
					Timeout: 30,
				}},
			},
		)
	}

	// UserPromptSubmit (prompt linter)
	if contains(cfg.Hooks, "prompt-lint") {
		s.Hooks["UserPromptSubmit"] = append(s.Hooks["UserPromptSubmit"],
			hookMatcher{
				Hooks: []hookCmd{{
					Type:    "command",
					Command: "$CLAUDE_PROJECT_DIR/.claude/hooks/prompt-lint.py",
					Timeout: 10,
				}},
			},
		)
	}

	return s
}

func renderClaudeMD(cfg Config) string {
	now := time.Now().Format("2006-01-02")
	var b bytes.Buffer
	fmt.Fprintf(&b, "# %s — Engineering Ground Rules\n\n", or(cfg.ProjectName, "Your Project"))
	b.WriteString("## Build & Test Commands\n")
	if includes(cfg.Languages, "Go") {
		b.WriteString("- `go build ./...`\n- `go test ./... -run . -v`\n- `golangci-lint run`\n")
	}
	if includes(cfg.Languages, "TypeScript") {
		b.WriteString("- `npm run build` / `pnpm build`\n- `npm run test -w` or `vitest`\n- `eslint . && prettier -c .`\n")
	}
	if includes(cfg.Languages, "Python") {
		b.WriteString("- `pytest -q`\n- `ruff check . && ruff format --check .`\n- `mypy`\n")
	}
	b.WriteString("\n## Code Style\n- Prefer small, pure functions\n- Comprehensive unit tests before large changes\n- Security & privacy by default\n\n")
	b.WriteString("## Workflow\n- Plan → Implement → Verify → Review → Merge\n- Use subagents proactively for review, tests, and debugging\n\n")
	b.WriteString("## Important Files to Know\n- @README\n- @.github/workflows (CI)\n\n")
	b.WriteString("## Claude Usage\n- Think first, then code; iterate with tests.\n- Prefer targeted file edits; do not modify secrets or prod configs.\n")
	b.WriteString("\n")
	if strings.TrimSpace(cfg.ClaudeMDExtras) != "" {
		b.WriteString("## Project‑Specific Notes\n")
		b.WriteString(cfg.ClaudeMDExtras + "\n\n")
	}
	b.WriteString(fmt.Sprintf("> Initialized by claudekit on %s\n", now))
	return b.String()
}

func renderAgent(name string) string {
	type agent struct {
		Front string
		Body  string
	}
	lib := map[string]agent{
		"code-reviewer": {
			Front: `---
name: code-reviewer
description: Expert code review specialist. Proactively reviews code for quality, security, and maintainability. Use immediately after writing or modifying code.
tools: Read, Grep, Glob, Bash
---
`,
			Body: `You are a senior code reviewer focused on clarity, correctness, security, and performance.
1) Run ` + "`git diff`" + ` to see recent changes.
2) Review only modified files unless instructed otherwise.
Output: Critical issues, Warnings, Suggestions, each with specific fixes.`,
		},
		"test-runner": {
			Front: `---
name: test-runner
description: Proactively run tests and fix failures. Use after code changes.
tools: Bash, Read, Edit
---
`,
			Body: `You write/maintain tests, run them, and iterate until green. Do not weaken tests; fix root causes.`,
		},
		"bug-sleuth": {
			Front: `---
name: bug-sleuth
description: Debug specialist for errors and unexpected behavior. Use proactively upon failures.
tools: Read, Edit, Bash, Grep, Glob
---
`,
			Body: `Do root-cause analysis. Reproduce, isolate, implement minimal fix, and verify.`,
		},
		"security-auditor": {
			Front: `---
name: security-auditor
description: Audit code for security issues. Use proactively on PRs or sensitive changes.
tools: Read, Grep, Glob
---
`,
			Body: `Check for secrets, injection, unsafe deserialization, authz gaps, and insecure defaults.`,
		},
		"perf-optimizer": {
			Front: `---
name: perf-optimizer
description: Identify hotspots and propose pragmatic optimizations. Use explicitly when perf degrades.
tools: Read, Grep, Glob, Bash
---
`,
			Body: `Focus on big-O and IO/alloc hotspots. Propose measurable changes and micro-bench references.`,
		},
		"docs-writer": {
			Front: `---
name: docs-writer
description: Produce/update docs, READMEs, and ADRs with clarity and examples.
tools: Read, Write, Edit
---
`,
			Body: `Ensure concise, accurate documentation with runnable snippets and task-oriented sections.`,
		},
		"release-manager": {
			Front: `---
name: release-manager
description: Prepare changelogs, version bumps, and release notes.
tools: Read, Write, Bash
---
`,
			Body: `Draft semver changes, summarize user-facing changes, and verify CI is green.`,
		},
		"data-scientist": {
			Front: `---
name: data-scientist
description: Data analysis expert. Use for SQL/BigQuery/data insights.
tools: Bash, Read, Write
---
`,
			Body: `Write efficient queries; summarize results with clear findings and next steps.`,
		},
	}
	if a, ok := lib[name]; ok {
		return a.Front + "\n" + a.Body + "\n"
	}
	return `---\nname: ` + name + `\ndescription: Custom subagent\n---\nProvide a focused role and steps.`
}

func postWriteLintScript(langs []string) string {
	var lines []string
	lines = append(lines, `echo "PostWrite: running linters/tests if available..."`)
	if includes(langs, "Go") {
		lines = append(lines, `command -v golangci-lint >/dev/null && golangci-lint run || true`)
		lines = append(lines, `go test ./... -run . -count=1 -v || true`)
	}
	if includes(langs, "TypeScript") {
		lines = append(lines, `[ -f package.json ] && (npm run -s lint || npx eslint . || true) || true`)
		lines = append(lines, `[ -f package.json ] && (npm test -w || npx vitest run || true) || true`)
	}
	if includes(langs, "Python") {
		lines = append(lines, `command -v ruff >/dev/null && (ruff check . || true); command -v pytest >/dev/null && (pytest -q || true)`)
	}
	return strings.Join(lines, "\n")
}

func preWriteGuardScript() string {
	return `
# Read stdin JSON into variable
payload="$(cat)"
# Very basic path extraction (avoid jq dependency)
filePath="$(printf "%s" "$payload" | sed -n 's/.*"file_path":"$begin:math:text$[^"]*$end:math:text$".*/\1/p')"
if [[ -z "${filePath}" ]]; then
  # Not all tools send file_path; allow
  exit 0
fi

# Disallow edits to sensitive paths
case "$filePath" in
  */.env|*/.env.*|*/secrets/*|*/config/production/*)
    echo "Blocking edit to sensitive path: $filePath" 1>&2
    exit 2 # per Anthropic docs: exit 2 = blocking error
    ;;
esac

exit 0
`
}

func sessionStartScript() string {
	return `
echo "Gathering project context..."
echo "Recent commits:"; git log --oneline -n 20 || true
echo
echo "Open issues (if gh installed):"; command -v gh >/dev/null && gh issue list --limit 10 || true
`
}

func promptLintPy() string {
	return `#!/usr/bin/env python3
import sys, json
data = sys.stdin.read()
try:
    obj = json.loads(data)
    prompt = obj.get("prompt","")
except Exception:
    prompt = ""
# Simple checks: ask user to add acceptance criteria
if len(prompt) > 0 and "acceptance" not in prompt.lower():
    sys.stderr.write("Please include acceptance criteria in your prompt (exit=2 to block).\n")
    sys.exit(2)  # block & show to user
# Otherwise, inject brief guidance to context via stdout
sys.stdout.write("Consider: objective, constraints, acceptance criteria, and test plan.")
sys.exit(0)
`
}

func sampleSlashCommand() string {
	return `Please analyze and fix the GitHub issue: $ARGUMENTS.

Follow these steps:
1. Use "gh issue view" to get details.
2. Identify affected files and tests.
3. Implement changes, keep commits small.
4. Run tests and linters.
5. Create a PR with a clear description.
`
}

func buildMCPJSON(selected []string) string {
	// Project-scoped .mcp.json using type/http or stdio servers; env expansion supported by Claude Code.
	// See docs for exact schema and variable expansion semantics.
	type server struct {
		Type    string            `json:"type,omitempty"`
		URL     string            `json:"url,omitempty"`
		Command string            `json:"command,omitempty"`
		Args    []string          `json:"args,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
		Headers map[string]string `json:"headers,omitempty"`
	}
	m := map[string]server{}
	for _, name := range selected {
		switch name {
		case "notion":
			m["notion"] = server{Type: "http", URL: "https://mcp.notion.com/mcp",
				Headers: map[string]string{"Authorization": "Bearer ${NOTION_TOKEN}"}} // env expansion supported
		case "linear":
			m["linear"] = server{Type: "sse", URL: "https://mcp.linear.app/sse",
				Headers: map[string]string{"Authorization": "Bearer ${LINEAR_TOKEN}"}}
		case "sentry":
			m["sentry"] = server{Type: "http", URL: "https://mcp.sentry.dev/mcp"}
		case "github":
			// Example stdio: npx server (official server names may vary; adjust to your org's choice)
			m["github"] = server{Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"},
				Env: map[string]string{"GITHUB_TOKEN": "${GITHUB_TOKEN}"}}
		case "airtable":
			// Cli-installed server (JS community)
			m["airtable"] = server{Command: "npx", Args: []string{"-y", "airtable-mcp-server"},
				Env: map[string]string{"AIRTABLE_API_KEY": "${AIRTABLE_API_KEY}"}}
		}
	}
	root := struct {
		MCPServers map[string]server `json:"mcpServers"`
	}{MCPServers: m}
	out, _ := json.MarshalIndent(root, "", "  ")
	return string(out)
}

func includes(ss []string, s string) bool {
	for _, x := range ss {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}
func or(a, b string) string {
	if strings.TrimSpace(a) == "" {
		return b
	}
	return a
}
