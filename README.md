# claudekit

A command-line tool that creates a complete Claude Code project setup through an interactive TUI form.

## Features

- 🎨 **Gradient Visual System** - Polished TUI with smooth color gradients
- 🖼️ **ASCII Art Titles** - Dynamic width-based rendering with gradient foreground
- 🎯 **Interactive Configuration** - Bubble Tea-powered forms
- 🔧 **Multi-Language Support** - Go, TypeScript, Python detection
- 🤖 **Agent Templates** - Pre-configured subagent definitions
- 🪝 **Lifecycle Hooks** - Automated linting and validation

## Quick Start

### Build and Run

```bash
make build
./claudekit
```

Or run directly:

```bash
make run
```

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

## Architecture

### Single-File Design

All application code is in `main.go` (Constitution Principle I):
- TUI model and update logic (Bubble Tea)
- Gradient rendering system
- ASCII art title with width-based conditional rendering
- Form configuration and file generation

### Visual System

**Gradient Engine:**
- Terminal capability detection (8-color, 256-color, truecolor)
- RGB color interpolation with easing functions
- Adaptive color quantization
- Foreground and background gradient modes

**ASCII Art Title:**
- Pre-generated figlet "small" font (embedded constant)
- Dynamic rendering based on terminal width (>= 60 cols)
- Fallback to regular gradient text for narrow terminals
- Gradient applied line-by-line for smooth effect

### Testing Architecture

**3-Layer Testing Strategy:**
1. **Unit Tests** (`main_test.go`) - Logic validation
2. **VHS Tests** (`vhs_test.go`) - Automated visual screenshots
3. **Manual Tests** (`quickstart.md`) - Human validation

This approach balances automation with the need for subjective visual quality assessment.

## Project Structure

```
claudekit/
├── main.go              # Single-file application
├── main_test.go         # Unit tests
├── vhs_test.go          # VHS visual test integration
├── Makefile             # Build and test automation
├── CLAUDE.md            # Claude Code guidance
├── README.md            # This file
└── specs/
    ├── 001-lets-create-a/    # Gradient system feature
    └── 002-lets-make-the/    # ASCII art title feature
        ├── spec.md
        ├── plan.md
        ├── tasks.md
        ├── quickstart.md
        └── vhs-tests/        # VHS test scripts
            ├── *.tape files
            ├── run-all-tests.sh
            ├── VALIDATION-CHECKLIST.md
            └── QUICKSTART.md
```

## Contributing

### Before Submitting PR

1. Run all checks: `make check`
2. Run visual tests: `make test-vhs` (if VHS installed)
3. Review screenshots against validation checklist
4. Test manually in different terminal emulators
5. Update documentation if adding features

### Development Workflow

1. **Unit Tests First** (TDD) - Write failing tests before implementation
2. **Implement Feature** - Maintain single-file architecture
3. **VHS Tests** - Add visual test scenarios if UI changes
4. **Manual Validation** - Test in real terminal
5. **Documentation** - Update CLAUDE.md and README.md

## License

See LICENSE file for details.

## Learn More

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [VHS](https://github.com/charmbracelet/vhs) - Visual testing tool
- [Claude Code](https://claude.com/claude-code) - AI-powered coding assistant
