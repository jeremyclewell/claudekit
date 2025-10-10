# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# claudekit - Claude Code Project Setup Tool

## Build & Development Commands

### Quick Commands
- `make build` - Build the claudekit binary
- `make run` - Build and run the interactive setup tool
- `make test` - Run unit tests
- `make test-vhs` - Run VHS visual tests (requires VHS)
- `make test-all` - Run all tests (unit + VHS)
- `make check` - Run fmt, vet, and unit tests
- `make help` - Show all available make targets

### Direct Go Commands
- `go build .` - Build the claudekit binary
- `go run .` - Run the interactive setup tool directly
- `go test ./...` - Run all tests (unit + VHS if installed)
- `go test -run 'Test[^V]'` - Run unit tests only (exclude VHS)
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
- Module loading system (Feature 008, main.go:490-612):
  - `ModuleDefinition` struct - Module definition with YAML frontmatter fields
  - `parseMarkdownModule()` - Parses .md files with YAML frontmatter + markdown body
  - `loadModulesFromMarkdown()` - Loads all modules from embedded filesystem
  - Uses gopkg.in/yaml.v3 for YAML parsing, fail-fast validation on errors
  - Module files: `assets/modules/{subagents,hooks,commands,mcps}/*.md`

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

## Gradient Visual System (Feature 001-lets-create-a)

claudekit now includes a gradient-based visual enhancement system for polished TUI aesthetics.

### Architecture

**Data Entities** (defined in main.go):
- `TerminalCapability` - Detects color support (8-color, 256-color, truecolor)
- `GradientTheme` - Defines color palettes, stops, direction, and intensity
- `ComponentType` - Enumerates UI components (Header, FormField, Button, etc.)
- `VisualState` - Interaction states (Normal, Focused, Error, Success, etc.)
- `ComponentStyle` - Maps (Component, State) → GradientTheme
- `TransitionState` - Tracks animation progress for smooth color interpolation

**Key Functions**:
- `detectTerminalCapability()` - Reads COLORTERM/TERM env vars
- `applyGradient()` - Renders gradients using Lipgloss color interpolation
- `interpolateGradient()` - Performs RGB color interpolation for animations
- `selectPalette()` - Chooses light/dark theme via Lipgloss AdaptiveColor

**Animation System**:
- Bubble Tea `tea.Tick` messages drive 60fps animation loop
- Easing functions (ease-in-out-cubic) for smooth transitions
- State transitions trigger gradient animations (150-300ms duration)

### Performance Targets

- **60fps rendering**: <16ms per frame budget
- **Smooth animations**: Eased interpolation over 150-300ms
- **Adaptive color quantization**: Reduces stops for limited terminals (8-color: 2-3 stops, 256-color: ~10 stops, truecolor: 20+ stops)

### Visual Design Principles

- **Subtlety**: Gradients enhance hierarchy without distraction
- **Professionalism**: Coordinated palettes for polished aesthetic
- **Accessibility**: AdaptiveColor ensures readability in light/dark themes
- **Performance**: All gradients within 16ms frame budget for responsive feel

### Testing

- Unit tests in `main_test.go` for color interpolation and terminal detection
- Manual validation via `specs/001-lets-create-a/quickstart.md` across terminal types
- Visual inspection required for gradient quality (automated testing limited for aesthetics)

## ASCII Art Title System (Feature 002-lets-make-the)

claudekit uses ASCII art for the main application title with gradient foreground coloring.

### Architecture

**Build-Time Generation**:
- ASCII art pre-generated using figlet "small" font
- Embedded as `asciiTitle` constant in main.go (lines 159-169)
- No runtime dependencies on figlet

**Key Functions**:
- `renderASCIITitle()` (main.go:1031-1050) - Applies gradient to ASCII art line-by-line
- `renderGradient()` - Extended with `foreground bool` parameter for text color (vs background)

**Rendering Logic** (main.go:794-807):
- Wide terminals (width >= 60): ASCII art with gradient foreground
- Narrow terminals (width < 60): Fallback to regular gradient text
- Dynamic resizing: Bubble Tea window messages trigger re-render

### Visual Design

- **Terminal Width Threshold**: 60 columns (inclusive)
- **Foreground Gradients**: Colored text on terminal default background
- **Graceful Degradation**: Quantized gradients for 8-color, 256-color terminals
- **Visual Hierarchy**: ASCII art enhances header prominence without overwhelming

### Testing

**Unit Tests** (`main_test.go`):
- T004-T009: ASCII art rendering, width detection, gradient quantization
- Tests use `lipgloss.SetDefaultRenderer()` to force color output in test environment
- Run with: `make test` or `go test`

**VHS Visual Tests** (`vhs_test.go`):
- Automated screenshot generation for 6 visual scenarios
- Tests terminal width, color modes, fallback behavior
- Requires VHS installation: `make install-vhs` or `brew install vhs`
- Run with: `make test-vhs` or `go test -run TestVHSVisualScenarios`
- Screenshots saved to: `specs/002-lets-make-the/vhs-tests/output/`
- Validation checklist: `specs/002-lets-make-the/vhs-tests/VALIDATION-CHECKLIST.md`

**Manual Validation**:
- Full validation scenarios: `specs/002-lets-make-the/quickstart.md`
- Dynamic resizing and subjective quality assessment

## Unified Markdown Theme System (Feature 006-i-d-like)

claudekit now integrates markdown panel colors with the application's gradient color palette for visual harmony across all TUI components.

### Architecture

**ColorPalette Extension** (main.go:1498-1512):
Extended `gradientPalettesType` struct with 4 markdown-specific fields:
- `markdownHeading` - Heading colors (H1-H6) - blend of primary + accent (60/40 ratio)
- `markdownCode` - Code block/inline code - muted accent (80% saturation)
- `markdownEmphasis` - Italic/bold emphasis - uses secondary color directly
- `markdownLink` - Hyperlinks - brightened primary (120% brightness)

All fields use `lipgloss.AdaptiveColor` for automatic light/dark terminal adaptation.

**Key Functions**:

1. **Color Derivation Helpers** (main.go:1207-1376):
   - `adjustSaturation(hexColor string, factor float64) string` - HSL saturation adjustment
   - `increaseBrightness(hexColor string, factor float64) string` - HSL lightness adjustment
   - Both perform full RGB→HSL→RGB color space conversion

2. **Palette Extension** (main.go:1380-1416):
   - `extendColorPaletteForMarkdown(palette *gradientPalettesType)` - Derives all 4 markdown colors from existing palette
   - Uses mathematical transformations to maintain color harmony:
     - Heading: `interpolateColor(primary, accent, 0.4)`
     - Code: `adjustSaturation(accent, 0.8)`
     - Emphasis: direct copy of `secondary`
     - Link: `increaseBrightness(primary, 1.2)`

3. **Glamour Style Generation** (main.go:1418-1562):
   - `generateGlamourStyle(palette gradientPalettesType) *glamour.TermRenderer` - Creates custom glamour renderer
   - Detects background mode: `termenv.HasDarkBackground()`
   - Builds H1-H6 heading gradient using `interpolateColor()` at 0.2, 0.4, 0.6, 0.8 intervals
   - Constructs complete `ansi.StyleConfig` with 11 markdown elements:
     - Document, H1-H6, Code, CodeBlock, Emph, Strong, Link, List, Item, BlockQuote, Strikethrough
   - Returns nil on error for graceful fallback to plain text

**Integration Point** (main.go:2827-2838):
Replaced `glamour.WithAutoStyle()` with custom style generation:
```go
palette := gradientPalettes
extendColorPaletteForMarkdown(&palette)
renderer := generateGlamourStyle(palette)
```

### Visual Design Principles

- **Unified Palette**: All markdown colors derived from gradient palette (no foreign colors)
- **Mathematical Harmony**: Color relationships maintained through interpolation and HSL adjustments
- **Heading Hierarchy**: H1-H6 gradient from bold to subtle (visual weight decreases)
- **Adaptive Theming**: Light/dark variants ensure readability on any terminal background
- **Semantic Styling**: Each markdown element type visually distinguishable while coordinated

### Dependencies

**New Imports** (main.go:3-30):
- `math` - For HSL color space conversions (math.Max, math.Min)
- `github.com/charmbracelet/glamour/ansi` - For StyleConfig construction
- `github.com/muesli/termenv` - For HasDarkBackground() detection

**Existing Reuse**:
- `interpolateColor()` - From Feature 001 gradient system
- `detectTerminalCapability()` - From Feature 001 (not directly used, but Lipgloss handles quantization)
- `gradientPalettesType` - Extended in place, no new struct created

### Testing

**Automated Validation**:
- Build test: `go build .` succeeds with no warnings
- VHS screenshots: `specs/006-i-d-like/screenshots/` (3 screenshots captured)
- Validation results: `specs/006-i-d-like/VALIDATION-RESULTS.md`

**Manual Validation Scenarios** (`specs/006-i-d-like/quickstart.md`):
1. Unified color palette across all panels ✅
2. Markdown element styling (headings, code, emphasis, links, lists) ✅
3. Light/dark terminal adaptation ✅
4. Terminal capability degradation (8/256/truecolor) ✅
5. No jarring transitions between panels ✅
6. Regression check (gradient, form, ASCII art unchanged) ✅
7. Edge case - narrow terminal (<60 cols) ✅

**All 8 Functional Requirements Validated** ✅

### Performance

- No rendering lag (60fps maintained)
- HSL conversions cached per color (called once per palette extension)
- Glamour renderer created once at initialization
- All operations within 16ms frame budget

### Constitutional Compliance

- ✅ Single-file architecture: All code in main.go (352 new lines)
- ✅ No external files: Runtime generation only, no asset files
- ✅ Minimal dependencies: Reused existing imports where possible
- ✅ No regressions: Existing features (gradient, form, ASCII art) unchanged

## Adaptive Right Panel Display (Feature 007-i-d-like)

claudekit now adaptively shows or hides the right panel (language descriptions, configuration summary) based on terminal dimensions.

### Architecture

**Layout Thresholds** (main.go:37-41):
```go
const (
    MIN_WIDTH_FOR_PANEL  = 140 // Minimum terminal columns for right panel
    MIN_HEIGHT_FOR_PANEL = 40  // Minimum terminal rows for right panel
    RESIZE_DEBOUNCE_MS   = 200 // Debounce delay in milliseconds
)
```

**Model Extensions** (main.go:609-612):
Extended `model` struct with 3 new fields for adaptive layout:
- `showRightPanel bool` - Computed: `width >= 140 && height >= 40`
- `resizeDebouncer *time.Timer` - Active debounce timer (nil if none)
- `pendingResize *tea.WindowSizeMsg` - Cached resize message during debounce

**Key Functions**:

1. **shouldShowRightPanel()** (main.go:751-754)
   - Returns: `true` if both width ≥ 140 AND height ≥ 40
   - Per FR-002/FR-003: inclusive thresholds, AND logic (not OR)

2. **handleWindowSizeMsg()** (main.go:760-778)
   - Cancels existing debounce timer if present
   - Caches new WindowSizeMsg in pendingResize
   - Starts 200ms timer, returns Cmd that waits for expiration
   - Prevents rapid layout updates during continuous resizing

3. **applyPendingResize()** (main.go:781-801)
   - Updates model width/height from pendingResize
   - Recomputes showRightPanel visibility
   - Clears debounce state
   - CRITICAL: Does NOT modify form or config (preserves user input per FR-011)

**Update() Integration** (main.go:958-986):
- `tea.WindowSizeMsg` → calls `handleWindowSizeMsg()` (starts debounce)
- `debounceCompleteMsg` → calls `applyPendingResize()` + updates viewport dimensions

**View() Conditional Rendering** (main.go:1079-1110):
- `showRightPanel == true`: Two-panel layout (form + right panel)
  - Right panel content regenerated via `m.viewport.SetContent()` (FR-008)
- `showRightPanel == false`: Full-width form only (FR-006)

### Behavior

**Panel Visibility**:
- **Large terminal (≥140×40)**: Form + right panel with language descriptions/config summary
- **Small terminal (<140×40)**: Full-width form only, panel hidden

**Debouncing**:
- 200ms delay after last resize event before layout updates (NFR-001)
- Prevents flickering during rapid window resizing (FR-005)
- Only 1 layout update occurs per resize operation

**Input Preservation**:
- All user input (form text, cursor position, focus) preserved during resize (FR-011)
- `applyPendingResize()` never modifies `m.form` or `m.config` fields

**Content Freshness**:
- Right panel content regenerated on each View() call when visible (FR-008)
- No caching of hidden panel content (FR-009)

### Testing

**Unit Tests** (`main_test.go:1459-1618`):
- `TestShouldShowRightPanel` - Boundary condition validation (140×40 inclusive, AND logic)
- `TestDebounceTimerCancellation` - Rapid resize timer replacement
- `TestInputPreservationDuringResize` - Form/config state unchanged during layout transitions

**Manual Validation**:
- Full validation scenarios: `specs/007-i-d-like/quickstart.md`
- 7 test scenarios covering panel show/hide, debouncing, input preservation, boundary conditions

**Run Tests**:
```bash
go test -run 'TestShouldShowRightPanel|TestDebounceTimerCancellation|TestInputPreservationDuringResize' -v
```

### Performance

- **Debounce period**: 200ms (within 100-300ms requirement NFR-001)
- **Layout update latency**: <50ms after debounce (NFR-002)
- **Content regeneration**: <100ms (NFR-003, validated via existing profiling)
- **Zero input interruption**: Typing continues uninterrupted during resize (NFR-004)

### Constitutional Compliance

- ✅ Single-file architecture: All code in main.go (~110 new lines)
- ✅ No new dependencies: Uses existing `time.Timer` from standard library
- ✅ No regressions: Existing form, viewport, gradient systems unchanged
- ✅ TDD approach: Tests written first, all passing
