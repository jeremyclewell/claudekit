# claudekit

> Interactive TUI for creating complete Claude Code project configurations

![claudekit demo](assets/Claude%20Kit.gif)

A command-line tool that creates a complete Claude Code project setup through an interactive TUI form. Generate `.claude/` configurations, hooks, subagents, custom commands, and MCP server setups with a beautiful, gradient-powered terminal interface.

## What It Generates

claudekit creates a complete Claude Code project setup with:

- **CLAUDE.md** - Project documentation and build commands
- **.claude/settings.json** - Permissions, hooks, and environment config
- **.claude/agents/** - Specialized subagent definitions (code-reviewer, test-runner, bug-sleuth, etc.)
- **.claude/hooks/** - Shell/Python scripts for lifecycle events
- **.claude/commands/** - Custom slash commands for workflows
- **.mcp.json** - MCP server configurations (GitHub, Notion, Linear, etc.)

## Features

- 📝 **Interactive Forms** - Bubble Tea-powered configuration wizard
- 🔧 **Multi-Language Support** - Auto-detection for Go, TypeScript, Python
- 🤖 **Agent Library** - 8 pre-configured subagent templates
- 🪝 **Smart Hooks** - Automated linting, validation, and context injection
- 🔌 **MCP Integration** - 5 popular MCP server templates
- 📦 **Modular System** - Extensible via markdown module definitions

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/claudekit.git
cd claudekit

# Build the binary
make build

# Run the interactive setup
./claudekit
```

Or run directly without building:

```bash
make run
```

### Usage

1. Launch `./claudekit` in your terminal
2. Fill out the interactive form:
   - Project name and description
   - Primary programming language
   - Select subagents, hooks, commands, and MCP servers
3. Press Enter to generate your `.claude/` configuration
4. Start using Claude Code with your new setup!

### Development

```bash
# Run tests
make test

# Run all checks (fmt, vet, tests)
make check

# Run VHS visual tests (requires VHS)
make install-vhs
make test-vhs

# Clean build artifacts
make clean
```

## Testing

### Unit Tests

Comprehensive unit tests for gradient system and ASCII art rendering:

```bash
make test
# or
go test ./...
```

**Coverage:**
- Terminal capability detection
- Gradient theme validation
- Color interpolation
- ASCII art rendering (width detection, fallback, quantization)
- Gradient foreground vs background modes

### Visual Tests (VHS)

Automated screenshot generation for visual validation:

```bash
# Install VHS
make install-vhs

# Run visual tests
make test-vhs
```

**VHS generates screenshots for:**
- Wide terminal (80+ cols) ASCII art rendering
- Narrow terminal (<60 cols) fallback behavior
- Gradient foreground coloring
- 60-column boundary threshold
- 256-color terminal quantization
- 8-color terminal graceful degradation

Screenshots are saved to `specs/002-lets-make-the/vhs-tests/output/` and reviewed using the validation checklist.

### Manual Validation

Some scenarios require manual testing in a real terminal:
- Dynamic terminal resizing
- Subjective visual hierarchy assessment
- Multi-terminal emulator testing

See `specs/002-lets-make-the/quickstart.md` for detailed manual test procedures.

## Available Modules

### Subagents (8 total)
- **code-reviewer** - Code quality and security review
- **test-runner** - Test execution and failure fixing
- **bug-sleuth** - Root cause debugging and analysis
- **security-auditor** - Security vulnerability scanning
- **perf-optimizer** - Performance bottleneck identification
- **docs-writer** - Documentation generation
- **release-manager** - Release preparation and changelog generation
- **data-scientist** - Data analysis and SQL query assistance

### Hooks (8 total)
- **session-start** - Project context injection on session start
- **session-end** - Cleanup and summary generation
- **user-prompt-submit** - Pre-flight prompt validation
- **pre-tool-use** - Guard rails for sensitive operations
- **post-tool-use** - Post-execution validation and linting
- **pre-compact** - Context cleanup before compaction
- **stop** - Graceful shutdown handling
- **subagent-stop** - Subagent cleanup and reporting

### Custom Commands (10 total)
- `/add-feature` - Guided feature implementation workflow
- `/add-tests` - Test generation and coverage improvement
- `/debug-issue` - Structured debugging workflow
- `/fix-github-issue` - GitHub issue resolution workflow
- `/refactor-code` - Safe refactoring with validation
- `/optimize-performance` - Performance analysis and optimization
- `/security-audit` - Comprehensive security review
- `/generate-docs` - Documentation generation
- `/setup-ci` - CI/CD pipeline setup
- `/migrate-database` - Database migration workflow

### MCP Servers (5 total)
- **GitHub** - Repository, issues, PRs integration
- **Notion** - Wiki and documentation integration
- **Linear** - Issue tracking integration
- **Airtable** - Database and spreadsheet integration
- **Sentry** - Error monitoring integration

## Extending claudekit

### Adding New Modules

Create a markdown file in the appropriate `assets/modules/` subdirectory:

```markdown
---
name: my-custom-agent
description: Custom agent for specialized tasks
language: bash
tools: ["Read", "Write", "Bash"]
permissions: ["read:*", "write:**/*.md"]
---

# Agent Instructions

Your custom agent implementation here...
```

Modules are automatically loaded at runtime and validated against the schema.

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Run all checks**: `make check` (fmt, vet, tests)
2. **Add tests**: Unit tests for logic, VHS tests for UI changes
3. **Update docs**: Keep CLAUDE.md and README.md in sync
4. **Follow conventions**: Modular architecture, clean separation of concerns

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run `make check` and `make test-all`
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Excellent TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling and layout
- [Glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [Huh](https://github.com/charmbracelet/huh) - Interactive forms
- [VHS](https://github.com/charmbracelet/vhs) - Visual testing

