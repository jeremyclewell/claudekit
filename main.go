package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	huh "github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"

	"jeremyclewell.com/claudekit/internal/generation"
	"jeremyclewell.com/claudekit/internal/gradient"
)
//go:embed assets/* assets/modules/**/*
var assets embed.FS

// Version number
const Version = "0.0.1"

// Layout threshold constants for adaptive right panel display (Feature 007)
const (
	MIN_WIDTH_FOR_PANEL  = 140 // Minimum terminal columns for right panel
	MIN_HEIGHT_FOR_PANEL = 40  // Minimum terminal rows for right panel
	RESIZE_DEBOUNCE_MS   = 200 // Debounce delay in milliseconds
)

type Config struct {
	IsProjectLocal bool       // true = project-based, false = global/home directory
	ProjectName    string
	Languages      []string
	Subagents      []string
	Hooks          []string
	SlashCommands  []string
	MCPServers     []string
	ClaudeMDExtras string
	Confirmed      bool       // for final confirmation step
}

// PersistenceConfig stores previous choices for subsequent runs
type PersistenceConfig struct {
	LastUpdated    time.Time `json:"last_updated"`
	IsProjectLocal bool      `json:"is_project_local"`
	ProjectName    string    `json:"project_name"`
	Languages      []string  `json:"languages"`
	Subagents      []string  `json:"subagents"`
	Hooks          []string  `json:"hooks"`
	SlashCommands  []string  `json:"slash_commands"`
	MCPServers     []string  `json:"mcp_servers"`
	ClaudeMDExtras string    `json:"claude_md_extras"`
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

// Module Registry Types (Feature 004)

// ModuleComponentType represents the category of a component module
type ModuleComponentType string

const (
	TypeSubagent ModuleComponentType = "subagent"
	TypeHook     ModuleComponentType = "hook"
	TypeMCP      ModuleComponentType = "mcp"
	TypeCommand  ModuleComponentType = "command"
)

// ComponentModule represents a single modular component definition
type ComponentModule struct {
	// Required fields
	Name        string              `json:"name"`
	Type        ModuleComponentType `json:"type"`
	Description string              `json:"description"`
	AssetPaths  []string            `json:"asset_paths"`

	// Optional fields
	Category     string         `json:"category,omitempty"`
	DisplayName  string         `json:"display_name,omitempty"`
	Dependencies []string       `json:"dependencies,omitempty"`
	Defaults     map[string]any `json:"defaults,omitempty"`
	Enabled      bool           `json:"enabled,omitempty"`
}

// GetDescription implements generation.ComponentModule interface
func (m *ComponentModule) GetDescription() string {
	return m.Description
}

// GetCategory implements generation.ComponentModule interface
func (m *ComponentModule) GetCategory() string {
	return m.Category
}

// ModuleDefinition represents a module definition loaded from Markdown with YAML frontmatter
// (Feature 008: Module Loading System Migration)
type ModuleDefinition struct {
	// Required fields (from frontmatter)
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`

	// Optional fields (from frontmatter)
	DisplayName string                 `yaml:"display_name,omitempty"`
	Category    string                 `yaml:"category,omitempty"`
	AssetPaths  []string               `yaml:"asset_paths,omitempty"`
	Defaults    map[string]interface{} `yaml:"defaults,omitempty"`

	// Content field (from markdown body)
	Description string `yaml:"-"` // Not in YAML

	// Metadata field (derived)
	FilePath string `yaml:"-"` // Not in YAML
}

// Module loading errors (Feature 008)
var (
	ErrMissingName       = errors.New("missing required field: name")
	ErrMissingType       = errors.New("missing required field: type")
	ErrInvalidType       = errors.New("invalid module type")
	ErrMissingDelimiters = errors.New("missing frontmatter delimiters")
	ErrYAMLParse         = errors.New("YAML parse error")
)

// ModuleRegistry manages the collection of all component modules
type ModuleRegistry struct {
	modules map[ModuleComponentType]map[string]*ComponentModule
	loaded  bool
	errors  []error
}

// Load discovers and loads all modules from the embedded filesystem
func (r *ModuleRegistry) Load(fs embed.FS) []error {
	r.modules = make(map[ModuleComponentType]map[string]*ComponentModule)
	r.errors = []error{}

	// Try both paths: assets/modules (production) and testdata/modules (testing)
	basePaths := []string{"assets/modules", "testdata/modules"}
	var entries []os.DirEntry
	var err error
	var basePath string

	for _, path := range basePaths {
		entries, err = fs.ReadDir(path)
		if err == nil {
			basePath = path
			break
		}
	}

	if err != nil {
		r.errors = append(r.errors, fmt.Errorf("cannot read modules directory: %w", err))
		r.loaded = true
		return r.errors
	}

	// Iterate through type directories (subagents, hooks, mcps, commands)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // Skip files in root modules dir
		}

		typeName := entry.Name()
		var componentType ModuleComponentType

		switch typeName {
		case "subagents":
			componentType = TypeSubagent
		case "hooks":
			componentType = TypeHook
		case "mcps":
			componentType = TypeMCP
		case "commands":
			componentType = TypeCommand
		default:
			continue // Skip unknown directories
		}

		// Initialize map for this type
		if r.modules[componentType] == nil {
			r.modules[componentType] = make(map[string]*ComponentModule)
		}

		// Read JSON files in this type directory
		typeDir := basePath + "/" + typeName
		typeEntries, err := fs.ReadDir(typeDir)
		if err != nil {
			r.errors = append(r.errors, fmt.Errorf("cannot read %s directory: %w", typeDir, err))
			continue
		}

		for _, fileEntry := range typeEntries {
			if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".md") {
				continue // Skip directories and non-.md files
			}

			// Read and parse Markdown file with YAML frontmatter (Feature 008)
			filePath := typeDir + "/" + fileEntry.Name()
			data, err := fs.ReadFile(filePath)
			if err != nil {
				r.errors = append(r.errors, fmt.Errorf("cannot read %s: %w", filePath, err))
				continue
			}

			// Parse using new Markdown+YAML parser
			moduleDef, err := parseMarkdownModule(filePath, data)
			if err != nil {
				r.errors = append(r.errors, fmt.Errorf("cannot parse %s: %w", filePath, err))
				continue
			}

			// Convert ModuleDefinition to ComponentModule for compatibility
			module := ComponentModule{
				Name:        moduleDef.Name,
				Type:        ModuleComponentType(moduleDef.Type),
				Description: moduleDef.Description,
				DisplayName: moduleDef.DisplayName,
				Category:    moduleDef.Category,
				AssetPaths:  moduleDef.AssetPaths,
				Defaults:    moduleDef.Defaults,
				Enabled:     moduleDef.Enabled,
			}

			// Validate and apply defaults
			if err := validateModule(&module, fs); err != nil {
				r.errors = append(r.errors, fmt.Errorf("validation failed for %s: %w", filePath, err))
				// Continue loading with warnings
			}

			// Register module (last-loaded wins for duplicates)
			r.modules[componentType][module.Name] = &module
		}
	}

	r.loaded = true
	return r.errors
}

// Get retrieves a specific module by type and name
func (r *ModuleRegistry) Get(componentType ModuleComponentType, name string) *ComponentModule {
	if r == nil || r.modules == nil {
		return nil
	}
	if typeMap, ok := r.modules[componentType]; ok {
		return typeMap[name]
	}
	return nil
}

// List returns all modules of a given type, sorted by name
func (r *ModuleRegistry) List(componentType ModuleComponentType) []*ComponentModule {
	if r == nil || r.modules == nil {
		return []*ComponentModule{}
	}

	typeMap, ok := r.modules[componentType]
	if !ok {
		return []*ComponentModule{}
	}

	// Extract modules and sort by name
	modules := make([]*ComponentModule, 0, len(typeMap))
	for _, module := range typeMap {
		modules = append(modules, module)
	}

	// Sort by name for deterministic ordering
	slices.SortFunc(modules, func(a, b *ComponentModule) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})

	return modules
}

// GetOptions generates TUI form options for a component type
func (r *ModuleRegistry) GetOptions(componentType ModuleComponentType) []huh.Option[string] {
	modules := r.List(componentType)
	options := make([]huh.Option[string], 0, len(modules))

	for _, module := range modules {
		displayText := module.Name
		if module.DisplayName != "" {
			displayText = module.DisplayName
		}
		options = append(options, huh.NewOption(displayText, module.Name))
	}

	return options
}

// ============================================================================
// Feature 008: Module Loading from Markdown with YAML Frontmatter
// ============================================================================

// extractFrontmatter extracts YAML frontmatter and markdown body from content
// Returns frontmatter YAML, body content, and error if delimiters missing
func extractFrontmatter(content string) (frontmatter, body string, err error) {
	// Split on --- delimiters
	parts := strings.SplitN(content, "---", 3)

	// Need at least 3 parts: [empty/whitespace, frontmatter, body]
	if len(parts) < 3 {
		return "", "", ErrMissingDelimiters
	}

	// First part should be empty or whitespace only (before opening ---)
	if strings.TrimSpace(parts[0]) != "" {
		return "", "", fmt.Errorf("%w: opening delimiter not at start", ErrMissingDelimiters)
	}

	frontmatter = strings.TrimSpace(parts[1])
	body = strings.TrimSpace(parts[2])

	return frontmatter, body, nil
}

// Validate checks if ModuleDefinition has all required fields and valid values
func (m *ModuleDefinition) Validate() error {
	// Required field: name
	if m.Name == "" {
		return ErrMissingName
	}

	// Required field: type
	if m.Type == "" {
		return ErrMissingType
	}

	// Type must be valid enum (FR-008)
	validTypes := map[string]bool{
		"subagent": true,
		"hook":     true,
		"command":  true,
		"mcp":      true,
	}
	if !validTypes[m.Type] {
		return fmt.Errorf("%w: %s (must be subagent, hook, command, or mcp)", ErrInvalidType, m.Type)
	}

	// Note: Enabled is bool, zero value (false) is valid
	// Note: Optional fields can be empty/nil

	return nil
}

// parseMarkdownModule parses a single module file with YAML frontmatter
func parseMarkdownModule(path string, content []byte) (ModuleDefinition, error) {
	var module ModuleDefinition

	// Extract frontmatter and body
	frontmatterYAML, body, err := extractFrontmatter(string(content))
	if err != nil {
		return module, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Parse YAML frontmatter
	err = yaml.Unmarshal([]byte(frontmatterYAML), &module)
	if err != nil {
		return module, fmt.Errorf("failed to parse %s: %w: %v", path, ErrYAMLParse, err)
	}

	// Set description from markdown body
	module.Description = body

	// Set file path
	module.FilePath = path

	// Validate required fields
	err = module.Validate()
	if err != nil {
		return module, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	return module, nil
}

// loadModulesFromMarkdown loads all module files from embedded filesystem
func loadModulesFromMarkdown(fsys embed.FS) ([]ModuleDefinition, error) {
	var modules []ModuleDefinition

	// Walk the assets/modules directory
	err := fs.WalkDir(fsys, "assets/modules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-.md files, and README.md files
		if d.IsDir() || !strings.HasSuffix(path, ".md") || strings.HasSuffix(path, "README.md") {
			return nil
		}

		// Read file content
		content, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Parse module
		module, err := parseMarkdownModule(path, content)
		if err != nil {
			return err // Fail-fast on any parse error (FR-013)
		}

		modules = append(modules, module)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil
}

// ============================================================================
// Feature 005: Asset File Generation Types
// ============================================================================

// validateModule checks required fields and applies defaults
func validateModule(module *ComponentModule, fs embed.FS) error {
	var errs []string

	// Check required fields (Feature 008: only name and type are required)
	if module.Name == "" {
		errs = append(errs, "name is required")
	}
	if module.Type == "" {
		errs = append(errs, "type is required")
	}
	// Description and AssetPaths are optional (e.g., MCPs don't need asset_paths)

	// Validate asset paths exist (warning only, not fatal)
	for _, assetPath := range module.AssetPaths {
		// Try multiple base paths for assets
		found := false
		for _, base := range []string{"assets/", "testdata/", ""} {
			fullPath := base + assetPath
			if _, err := fs.ReadFile(fullPath); err == nil {
				found = true
				break
			}
		}
		if !found {
			// Log warning but don't fail - asset might be optional
			errs = append(errs, fmt.Sprintf("asset not found: %s", assetPath))
		}
	}

	// Apply defaults for optional fields
	if module.DisplayName == "" {
		module.DisplayName = module.Name
	}
	if module.Defaults == nil {
		module.Defaults = make(map[string]any)
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// ASCII Art Title (T011) - Pre-generated with figlet "small" font

// const asciiTitle = `▄▖▜      ▌    ▖▖▘▗ 
// ▌ ▐ ▀▌▌▌▛▌█▌  ▙▘▌▜▘
// ▙▖▐▖█▌▙▌▙▌▙▖  ▌▌▌▐▖`

// const asciiTitle = ` ▗▄▄▖▗▖    ▗▄▖ ▗▖ ▗▖▗▄▄▄ ▗▄▄▄▖    ▗▖ ▗▖▗▄▄▄▖▗▄▄▄▖
// ▐▌   ▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  █▐▌       ▐▌▗▞▘  █    █  
// ▐▌   ▐▌   ▐▛▀▜▌▐▌ ▐▌▐▌  █▐▛▀▀▘    ▐▛▚▖   █    █  
// ▝▚▄▄▖▐▙▄▄▖▐▌ ▐▌▝▚▄▞▘▐▙▄▄▀▐▙▄▄▖    ▐▌ ▐▌▗▄█▄▖  █  `

const asciiTitle = `┏━╸╻  ┏━┓╻ ╻╺┳┓┏━╸   ╻┏ ╻╺┳╸
┃  ┃  ┣━┫┃ ┃ ┃┃┣╸    ┣┻┓┃ ┃ 
┗━╸┗━╸╹ ╹┗━┛╺┻┛┗━╸   ╹ ╹╹ ╹ `

// Bubble Tea Model for the application
type model struct {
	form            *huh.Form
	config          *Config
	viewport        viewport.Model
	glamourRenderer *glamour.TermRenderer
	ready           bool
	width           int
	height          int
	currentFocus    string

	// Gradient system fields (T028)
	terminalCap  gradient.TerminalCapability
	currentTheme gradient.Theme
	transition   gradient.TransitionState
	styleMap     map[gradient.ComponentType]map[gradient.VisualState]gradient.ComponentStyle

	// Module registry (Feature 004)
	registry *ModuleRegistry

	// Adaptive right panel layout (Feature 007)
	showRightPanel  bool                // Computed: width >= 140 && height >= 40
	resizeDebouncer *time.Timer         // Active debounce timer (nil if none)
	pendingResize   *tea.WindowSizeMsg  // Cached resize message during debounce
}

// Styles for the Uaud
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Padding(1).
			MarginLeft(2)

	formStyle = lipgloss.NewStyle().
			Padding(1)

	descStyle = lipgloss.NewStyle().
			BorderTop(true).
			Border(lipgloss.RoundedBorder()).
			// BorderStyle(lipgloss.Border{
			// 	Top: "//",
			// }).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1).
			MarginTop(1)
)

// Descriptions for subagents, MCPs, hooks, and commands are now loaded from JSON modules (Feature 004)

// Detailed descriptions for programming languages
var languageDescriptions = map[string]string{
	"Go": "## 🐹 Go\nSimple, fast, concurrent. Master goroutines and channels for scalable microservices and cloud-native applications.\n\n### Key Features\n\n* Clean, readable syntax\n* Excellent standard library\n* Built-in concurrency primitives\n* Fast compilation and execution\n\n-------\n\n### Example\n\n```go\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n\n---\n\nPerfect for APIs, distributed systems, and microservices.",
	
	"TypeScript": "## 🟦 TypeScript\nJavaScript with static types. Build type-safe web applications with excellent IntelliSense and compile-time error catching.\n\n### Key Features\n* Static type checking\n* Excellent IDE support\n* Scales from small to enterprise projects\n* Works with React, Next.js, Node.js, Angular, Vue\n\n---\n\n### Example\n```typescript\nconsole.log(\"Hello, World!\");\n```\n\n---\n\nPerfect for full-stack web development and large-scale applications.",

	"Python": "## 🐍 Python\nReadable, versatile, powerful. Write clean code for data science, machine learning, web development, and automation.\n\n### Key Features\n* Clean, readable syntax\n* Rich ecosystem (Django, FastAPI, NumPy, pandas, PyTorch)\n* Excellent for data science and ML\n* Rapid prototyping and development\n\n---\n\n### Example\n```python\nprint(\"Hello, World!\")\n```\n\n---\n\nPerfect for scientific computing, web apps, and automation scripts.",

	"Java": "## ☕ Java\nEnterprise-grade reliability. Build scalable applications with Spring Boot, microservices architecture, and proven design patterns.\n\n### Key Features\n* Robust JVM performance\n* Extensive standard library\n* Strong typing and OOP\n* Cross-platform compatibility\n\n---\n\n### Example\n```java\npublic class HelloWorld {\n    public static void main(String[] args) {\n        System.out.println(\"Hello, World!\");\n    }\n}\n```\n\n---\n\nExcellent for large-scale enterprise systems and backend services.",

	"Rust": "## 🦀 Rust\nMemory-safe systems programming. Zero-cost abstractions with ownership model and fearless concurrency.\n\n### Key Features\n* Memory safety without garbage collection\n* Prevents data races at compile time\n* Zero-cost abstractions\n* Excellent performance\n\n---\n\n### Example\n```rust\nfn main() {\n    println!(\"Hello, World!\");\n}\n```\n\n---\n\nPerfect for operating systems, game engines, and performance-critical applications.",

	"C++": "## ⚡ C++\nHigh-performance systems programming. Modern C++20 features with RAII patterns and efficient low-level control.\n\n### Key Features\n* Maximum speed and control\n* Templates and metaprogramming\n* Smart pointers and RAII\n* Zero-overhead abstractions\n\n---\n\n### Example\n```cpp\n#include <iostream>\n\nint main() {\n    std::cout << \"Hello, World!\" << std::endl;\n    return 0;\n}\n```\n\n---\n\nPerfect for game engines, embedded systems, and performance-critical applications.",

	"C#": "## 💎 C#\nModern .NET development. Build cross-platform applications with LINQ, async/await, and excellent tooling.\n\n### Key Features\n* Rich type system and LINQ\n* Async/await for concurrency\n* Cross-platform with .NET Core\n* Desktop (WPF), Web (ASP.NET), Cloud (Azure)\n\n---\n\n### Example\n```csharp\nusing System;\n\nclass Program {\n    static void Main() {\n        Console.WriteLine(\"Hello, World!\");\n    }\n}\n```\n\n---\n\nPerfect for enterprise applications and Microsoft ecosystem integration.",
	
	"PHP": "## 🐘 PHP\nWeb development made easy. Modern frameworks like Laravel and Symfony for rapid application development.\n\n### Key Features\n* Easy database integration\n* Modern PHP 8+ features\n* Rich framework ecosystem (Laravel, Symfony)\n* Excellent for web applications\n\n---\n\n### Example\n```php\n<?php\necho \"Hello, World!\";\n?>\n```\n\n---\n\nPerfect for CMS, e-commerce, and dynamic web applications.",

	"Ruby": "## 💎 Ruby\nDeveloper happiness first. Elegant Rails development with convention over configuration and expressive syntax.\n\n### Key Features\n* Beautiful, readable syntax\n* Rails framework for rapid development\n* Rich gem ecosystem\n* Powerful metaprogramming\n\n---\n\n### Example\n```ruby\nputs \"Hello, World!\"\n```\n\n---\n\nPerfect for web applications, automation scripts, and developer-friendly APIs.",

	"Swift": "## 🍎 Swift\nApple's modern language. Build native iOS, macOS, and watchOS apps with SwiftUI and protocol-oriented programming.\n\n### Key Features\n* Optionals for null safety\n* SwiftUI for declarative UIs\n* Automatic memory management (ARC)\n* Protocol-oriented programming\n\n---\n\n### Example\n```swift\nprint(\"Hello, World!\")\n```\n\n---\n\nPerfect for iOS and macOS app development.",

	"Kotlin": "## 🎯 Kotlin\nConcise JVM language. Android development with coroutines, null safety, and 100% Java interoperability.\n\n### Key Features\n* Null safety built-in\n* Coroutines for async programming\n* 100% Java interoperability\n* Multiplatform support\n\n---\n\n### Example\n```kotlin\nfun main() {\n    println(\"Hello, World!\")\n}\n```\n\n---\n\nPerfect for Android apps and server-side development.",
	
	"Dart": "## 🎯 Dart\nFlutter's foundation. Build beautiful cross-platform apps for iOS, Android, web, and desktop from one codebase.\n\n### Key Features\n* Single codebase for all platforms\n* Hot reload for fast development\n* Rich widget library\n* Native performance\n\n---\n\n### Example\n```dart\nvoid main() {\n  print('Hello, World!');\n}\n```\n\n---\n\nPerfect for cross-platform mobile and web applications.",

	"Shell": "## 🐚 Shell/Bash\nSystem automation master. Write robust scripts for deployment, system administration, and file processing.\n\n### Key Features\n* Powerful text processing with pipes\n* System administration and automation\n* CI/CD pipeline integration\n* Universal Unix/Linux availability\n\n---\n\n### Example\n```bash\n#!/bin/bash\necho \"Hello, World!\"\n```\n\n---\n\nPerfect for automation, DevOps, and system administration.",

	"Lua": "## 🌙 Lua\nLightweight scripting. Embedded applications, game scripting, and configuration management with minimal footprint.\n\n### Key Features\n* Tiny footprint (~280KB)\n* Fast execution\n* Easy C integration\n* Simple, clean syntax\n\n---\n\n### Example\n```lua\nprint(\"Hello, World!\")\n```\n\n---\n\nPerfect for game scripting, embedded systems, and application extensions.",

	"Elixir": "## 💧 Elixir\nFault-tolerant concurrency. Actor model with millions of lightweight processes for distributed, real-time systems.\n\n### Key Features\n* Massive concurrency on Erlang VM\n* Built-in fault tolerance\n* Functional programming patterns\n* Excellent for real-time systems\n\n---\n\n### Example\n```elixir\nIO.puts \"Hello, World!\"\n```\n\n---\n\nPerfect for chat apps, IoT backends, and distributed systems.",
	
	"Haskell": "## λ Haskell\nPure functional programming. Type-driven development with mathematically elegant solutions and compile-time guarantees.\n\n### Key Features\n* Pure functional programming\n* Strong static typing\n* Lazy evaluation\n* Advanced type system\n\n---\n\n### Example\n```haskell\nmain = putStrLn \"Hello, World!\"\n```\n\n---\n\nPerfect for compilers, financial systems, and mathematically precise applications.",

	"Elm": "## 🌳 Elm\nDelightful web apps. No runtime exceptions with functional reactive programming and immutable data structures.\n\n### Key Features\n* No runtime exceptions\n* Excellent error messages\n* Time-travel debugging\n* Guaranteed refactoring safety\n\n---\n\n### Example\n```elm\nimport Html exposing (text)\n\nmain =\n  text \"Hello, World!\"\n```\n\n---\n\nPerfect for maintainable frontend web applications.",

	"Julia": "## 🔬 Julia\nScientific computing. High-performance numerical algorithms with Python-like syntax and C-like speed.\n\n### Key Features\n* Python-like syntax, C-like speed\n* Built-in parallel computing\n* Excellent for numerical analysis\n* Multiple dispatch system\n\n---\n\n### Example\n```julia\nprintln(\"Hello, World!\")\n```\n\n---\n\nPerfect for machine learning, scientific research, and computational mathematics.",

	"SQL": "## 🗄️ SQL\nData mastery. Write efficient queries, design normalized schemas, and optimize database performance.\n\n### Key Features\n* Declarative query language\n* Works with PostgreSQL, MySQL, SQLite\n* Essential for data analysis\n* Industry-standard for databases\n\n---\n\n### Example\n```sql\nSELECT 'Hello, World!' AS greeting;\n```\n\n---\n\nPerfect for data analysis, backend development, and business intelligence.",

	"Arduino": "## 🤖 Arduino\nHardware programming. Build IoT devices, sensor networks, and interactive electronic projects with C++ for microcontrollers.\n\n### Key Features\n* Easy hardware interfacing\n* Large sensor and module ecosystem\n* Digital and analog I/O\n* Serial communication protocols\n\n---\n\n### Example\n```cpp\nvoid setup() {\n  Serial.begin(9600);\n}\n\nvoid loop() {\n  Serial.println(\"Hello, World!\");\n  delay(1000);\n}\n```\n\n---\n\nPerfect for home automation, robotics, and IoT projects.",

	"Scheme": "## 🧠 Scheme\nMinimalist functional programming. Pure computational thinking and programming language fundamentals with elegant syntax.\n\n### Key Features\n* Minimal, elegant syntax\n* First-class functions\n* Powerful macro system\n* Educational and theoretical\n\n---\n\n### Example\n```scheme\n(display \"Hello, World!\")\n(newline)\n```\n\n---\n\nPerfect for learning computer science and programming language theory.",

	"Lisp": "## 🧠 Lisp\nSymbolic AI programming. Meta-programming with code-as-data philosophy and powerful macro systems.\n\n### Key Features\n* Homoiconic code-as-data\n* Powerful macro system\n* REPL-driven development\n* Excellent for symbolic computation\n\n---\n\n### Example\n```lisp\n(print \"Hello, World!\")\n```\n\n---\n\nPerfect for AI, symbolic computation, and domain-specific languages.",
}


// getPersistenceFilePath returns the path to the persistence file
func getPersistenceFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".claudekit.json"), nil
}

// loadPersistenceConfig loads previous choices from the persistence file
func loadPersistenceConfig() (*PersistenceConfig, error) {
	filePath, err := getPersistenceFilePath()
	if err != nil {
		return nil, err
	}
	
	// If file doesn't exist, return empty config
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &PersistenceConfig{}, nil
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var config PersistenceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// savePersistenceConfig saves current choices to the persistence file
func savePersistenceConfig(config Config) error {
	filePath, err := getPersistenceFilePath()
	if err != nil {
		return err
	}
	
	persistConfig := PersistenceConfig{
		LastUpdated:    time.Now(),
		IsProjectLocal: config.IsProjectLocal,
		ProjectName:    config.ProjectName,
		Languages:      config.Languages,
		Subagents:      config.Subagents,
		Hooks:          config.Hooks,
		SlashCommands:  config.SlashCommands,
		MCPServers:     config.MCPServers,
		ClaudeMDExtras: config.ClaudeMDExtras,
	}
	
	data, err := json.MarshalIndent(persistConfig, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filePath, data, 0644)
}

// shouldShowRightPanel returns true if terminal dimensions meet thresholds for right panel display.
// Per FR-002/FR-003: Requires BOTH width >= 140 AND height >= 40 (inclusive thresholds).
func shouldShowRightPanel(width, height int) bool {
	return width >= MIN_WIDTH_FOR_PANEL && height >= MIN_HEIGHT_FOR_PANEL
}

// debounceCompleteMsg signals that resize debounce period has elapsed
type debounceCompleteMsg struct{}

// handleWindowSizeMsg processes terminal resize events with debouncing.
// Cancels any existing timer, caches the new dimensions, and starts 200ms countdown.
func handleWindowSizeMsg(m model, msg tea.WindowSizeMsg) (model, tea.Cmd) {
	// Cancel existing debounce timer if present
	if m.resizeDebouncer != nil {
		m.resizeDebouncer.Stop()
	}

	// Cache resize message
	m.pendingResize = &msg

	// Start new debounce timer
	m.resizeDebouncer = time.NewTimer(RESIZE_DEBOUNCE_MS * time.Millisecond)

	// Return Cmd that waits for timer to expire
	return m, func() tea.Msg {
		<-m.resizeDebouncer.C
		return debounceCompleteMsg{}
	}
}

// applyPendingResize updates model dimensions and recomputes panel visibility.
// CRITICAL: MUST NOT modify m.form or m.config (FR-011: preserve all user input).
func applyPendingResize(m model) (model, tea.Cmd) {
	if m.pendingResize == nil {
		return m, nil // No pending resize, nothing to do
	}

	// Update dimensions
	m.width = m.pendingResize.Width
	m.height = m.pendingResize.Height

	// Recompute panel visibility (FR-002, FR-003)
	m.showRightPanel = shouldShowRightPanel(m.width, m.height)

	// Clear debounce state
	m.pendingResize = nil
	m.resizeDebouncer = nil

	// IMPORTANT: Do NOT modify m.form or m.config - preserves input per FR-011

	return m, nil
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m *model) getCurrentDescription() string {
	// Get current focus from form state
	if m.form.State == huh.StateCompleted {
		return "✅ Configuration complete! Ready to generate your Claude Code setup."
	}
	
	// Get the currently focused field
	focusedField := m.form.GetFocusedField()
	if focusedField == nil {
		return m.getDefaultDescription()
	}
	
	// Check field key to identify what type of selection we're in
	fieldKey := focusedField.GetKey()
	
	// Handle language selection
	if fieldKey == "languages" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				if desc, exists := languageDescriptions[hoveredItem]; exists {
					return desc
				}
			}
		}
		return "💻 Select programming languages used in your project. Claude will provide specialized assistance and optimized configurations for each language. Navigate with arrow keys to see how Claude can help."
	}
	
	// Handle subagent selection (Feature 004: use registry)
	if fieldKey == "subagents" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				// Extract the subagent name (remove emoji prefix)
				subagentName := extractSubagentName(hoveredItem)
				if module := m.registry.Get(TypeSubagent, subagentName); module != nil {
					return module.Description
				}
			}
		}
		return "🤖 Select specialized AI assistants for your development workflow. Navigate with arrow keys to see detailed descriptions."
	}
	
	// Handle hook selection (Feature 004: use registry)
	if fieldKey == "hooks" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				// Extract the hook name (remove emoji prefix)
				hookName := extractSubagentName(hoveredItem)
				if module := m.registry.Get(TypeHook, hookName); module != nil {
					return module.Description
				}
			}
		}
		return "🪝 Select automation hooks to enhance your development workflow. These scripts run at specific points to provide safety, quality control, and context. Navigate with arrow keys to see detailed descriptions."
	}
	
	// Handle slash command selection (Feature 004: use registry)
	if fieldKey == "slash-commands" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				// Extract the command name (remove emoji prefix)
				commandName := extractSubagentName(hoveredItem)
				if module := m.registry.Get(TypeCommand, commandName); module != nil {
					return module.Description
				}
			}
		}
		return "⚡ Select custom slash commands for common development tasks. These powerful shortcuts automate complex workflows and boost productivity. Navigate with arrow keys to see detailed descriptions."
	}
	
	// Handle MCP server selection (Feature 004: use registry)
	if fieldKey == "mcp-servers" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				// Extract the MCP server name (remove emoji prefix)
				serverName := extractSubagentName(hoveredItem)
				if module := m.registry.Get(TypeMCP, serverName); module != nil {
					return module.Description
				}
			}
		}
		return "🔌 Select external tool integrations to enhance Claude's capabilities via Model Context Protocol. Navigate with arrow keys to see detailed descriptions."
	}
	
	return m.getDefaultDescription()
}

func (m *model) getDefaultDescription() string {
	return `## 📋 Claude Code Project Setup

Welcome to the interactive **Claude Code** project configuration tool! This wizard will help you set up a comprehensive development environment either _globally_, or on a _per project basis_.

### 🔍 NAVIGATION:
* Use **tab** & **shift-tab** to move between form fields
* Use **arrow** keys to navigate between options
* Use **space** to select/deselect items in multi-select lists
* Use **enter** to proceed/confirm to the next field

### 📚 WHAT YOU'RE CONFIGURING:
* Project basics (directory, name, languages)
* AI subagents for specialized development tasks
* Automation hooks for workflow enhancement
* External tool integrations via MCP

Choose the options that best fit your development workflow and project needs. Your choices will persist and you may _use this tool again to make changes_.`
}

func extractSubagentName(displayName string) string {
	// Remove emoji and space prefix (e.g., "🔍 code-reviewer" -> "code-reviewer")
	parts := strings.SplitN(displayName, " ", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return displayName
}

func (m *model) renderMarkdown(content string) string {
	if m.glamourRenderer == nil {
		return content // Fallback to plain text
	}
	
	rendered, err := m.glamourRenderer.Render(content)
	if err != nil {
		return content // Fallback to plain text on error
	}
	
	return rendered
}

// tickMsg is our custom message for gradient animations
type tickMsg time.Time

// startTransition initiates a gradient theme transition (T033)
func (m *model) startTransition(to gradient.Theme, duration time.Duration) tea.Cmd {
	m.transition = gradient.TransitionState{
		Active:     true,
		FromTheme:  m.currentTheme,
		ToTheme:    to,
		StartTime:  time.Now(),
		Duration:   duration,
		EasingFunc: gradient.EaseInOutCubic,
	}

	// Return initial tick command to start animation
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Feature 007: Debounced resize handling
		return handleWindowSizeMsg(m, msg)

	case debounceCompleteMsg:
		// Feature 007: Apply pending resize after debounce period
		m, cmd := applyPendingResize(m)

		// After applying resize, update viewport dimensions if panel is visible
		if m.showRightPanel {
			// Calculate layout dimensions with fixed percentages for stability
			formWidth := int(float64(m.width) * 0.6)        // 60% width for left side
			statusWidth := m.width - formWidth - 6          // Remaining width for right side

			// Calculate height consistently with View() function
			const borderPadding = 10
			const borderHeight = 4
			innerHeight := m.height - borderHeight
			if innerHeight < 10 {
				innerHeight = 10
			}
			titleHeight := 5
			availableHeight := innerHeight - titleHeight
			if availableHeight < 20 {
				availableHeight = 20
			}
			statusHeight := availableHeight

			if !m.ready {
				m.viewport = viewport.New(statusWidth, statusHeight)
				m.ready = true
			} else {
				m.viewport.Width = statusWidth
				m.viewport.Height = statusHeight
			}
		}

		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	// T032: Handle gradient animation ticks
	case tickMsg:
		if m.transition.Active {
			progress := m.transition.Progress()
			if progress >= 1.0 {
				// Transition complete
				m.transition.Active = false
				m.currentTheme = m.transition.ToTheme
			} else {
				// Continue animating
				m.currentTheme = gradient.InterpolateGradient(
					m.transition.FromTheme,
					m.transition.ToTheme,
					progress,
				)
				// Schedule next tick for smooth animation
				return m, tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
					return tickMsg(t)
				})
			}
		}
	}

	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	
	// Handle viewport scrolling for status panel
	var viewportCmd tea.Cmd
	m.viewport, viewportCmd = m.viewport.Update(msg)
	cmd = tea.Batch(cmd, viewportCmd)

	// Update viewport content with current status/descriptions
	m.viewport.SetContent(m.renderMarkdown(m.renderStatus()))

	// Check if form is complete
	if m.form.State == huh.StateCompleted {
		return m, tea.Quit
	}

	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Account for border + padding
	// Border adds 2 chars left/right (1 for border char, 1 for automatic border spacing)
	// Padding adds 2 chars left/right (via Padding(1, 2))
	// Total per side: 2 + 2 = 4, so 8 total width
	const borderPadding = 10  // Extra space for border + padding on left/right
	const borderHeight = 4    // Border (2 lines: top + bottom) + Padding (2 lines: top + bottom via Padding(1, 2))

	innerWidth := m.width - borderPadding
	innerHeight := m.height - borderHeight

	if innerWidth < 20 {
		innerWidth = 20
	}
	if innerHeight < 10 {
		innerHeight = 10
	}

	// Calculate dimensions with fixed percentages for stability
	formWidth := int(float64(innerWidth) * 0.6)
	statusWidth := innerWidth - formWidth - 6

	// Reserve space for title (3 lines ASCII + 1 line gradient border + 1 line spacing)
	titleHeight := 5
	availableHeight := innerHeight - titleHeight
	if availableHeight < 20 {
		availableHeight = 20
		titleHeight = innerHeight - 20  // Reduce title space if needed
	}

	formHeight := availableHeight
	statusHeight := availableHeight // Status panel should match form height

	// Title with gradient (T035)
	// T015: Width-based conditional rendering for ASCII art title
	headerTheme := m.styleMap[gradient.HeaderComponent][gradient.NormalState].Theme
	var title string

	// Version string with subtle styling
	versionText := "v" + Version
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Faint(true)
	version := versionStyle.Render(versionText)

	if innerWidth >= 60 {
		// Wide terminal: render ASCII art with gradient foreground + version
		gradientASCII := gradient.RenderASCIITitle(asciiTitle, headerTheme, m.terminalCap)

		// Split ASCII art into lines and add version to the first line
		asciiLines := strings.Split(gradientASCII, "\n")
		if len(asciiLines) > 0 {
			// Calculate padding to right-align version
			firstLineWidth := lipgloss.Width(asciiLines[0])
			padding := innerWidth - firstLineWidth - lipgloss.Width(version)
			if padding < 0 {
				padding = 0
			}
			asciiLines[0] = asciiLines[0] + strings.Repeat(" ", padding) + version
		}
		title = strings.Join(asciiLines, "\n")
	} else {
		// Narrow terminal: fallback to regular gradient text with version
		titleText := "🛠️  ClaudeKit"
		gradientTitle := gradient.RenderGradient(titleText, headerTheme, m.terminalCap, true)

		// Add version on same line with padding
		titleWidth := lipgloss.Width(gradientTitle)
		padding := innerWidth - titleWidth - lipgloss.Width(version)
		if padding < 0 {
			padding = 0
		}
		title = gradientTitle + strings.Repeat(" ", padding) + version
	}

	// Create gradient top border with "/" characters
	borderWidth := innerWidth
	borderText := strings.Repeat("/", borderWidth)
	gradientBorder := gradient.RenderGradient(borderText, headerTheme, m.terminalCap, true)

	// Feature 007: Adaptive right panel based on terminal size
	var content string

	if m.showRightPanel {
		// Update viewport height to match available content height
		m.viewport.Height = statusHeight
		m.viewport.Width = statusWidth

		// Large terminal: show form + right panel
		formContent := m.form.View()
		leftContent := formStyle.
			Width(formWidth).
			Height(formHeight).
			Render(formContent)

		// Regenerate right panel content (FR-008: always fresh)
		m.viewport.SetContent(m.renderMarkdown(m.renderStatus()))

		// Status panel (right side, fixed height to match form)
		statusPanel := statusStyle.
			Width(statusWidth).
			Height(statusHeight). // Use consistent height
			Render(m.viewport.View())

		// Main content (left content + status)
		// Ensure exact height by padding if necessary
		leftContent = ensureExactHeight(leftContent, formHeight)
		statusPanel = ensureExactHeight(statusPanel, statusHeight)

		content = lipgloss.JoinHorizontal(lipgloss.Top, leftContent, statusPanel)
	} else {
		// Small terminal: full-width form only (FR-006)
		formContent := m.form.View()
		leftContent := formStyle.
			Width(innerWidth - 4). // Full width minus padding
			Height(formHeight).
			Render(formContent)

		// Ensure exact height
		leftContent = ensureExactHeight(leftContent, formHeight)
		content = leftContent
	}

	// Combine title, border, and content
	app := lipgloss.JoinVertical(lipgloss.Left, title, gradientBorder, content)

	// Add border around entire application with gradient start color and padding
	borderColor := headerTheme.StartColor
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2) // 1 line top/bottom, 2 chars left/right padding

	// Render the app with border
	rendered := borderStyle.Render(app)

	// Enforce exact terminal dimensions to prevent height overflow
	// Truncate content to fit within terminal bounds
	lines := strings.Split(rendered, "\n")
	if len(lines) > m.height {
		lines = lines[:m.height]
	}

	// Ensure each line doesn't exceed width
	for i, line := range lines {
		if lipgloss.Width(line) > m.width {
			// Truncate line to fit width (accounting for ANSI codes)
			lines[i] = truncateLine(line, m.width)
		}
	}

	return strings.Join(lines, "\n")
}

// truncateLine truncates a line to the specified width, preserving ANSI codes
func truncateLine(line string, width int) string {
	// Use lipgloss's Truncate which handles ANSI codes properly
	return lipgloss.NewStyle().Width(width).Render(line)
}

// ensureExactHeight pads or truncates content to be exactly the specified height
func ensureExactHeight(content string, targetHeight int) string {
	lines := strings.Split(content, "\n")
	currentHeight := len(lines)

	if currentHeight == targetHeight {
		return content
	}

	if currentHeight > targetHeight {
		// Truncate to target height
		return strings.Join(lines[:targetHeight], "\n")
	}

	// Pad with empty lines to reach target height
	padding := make([]string, targetHeight-currentHeight)
	for i := range padding {
		padding[i] = ""
	}
	return content + "\n" + strings.Join(padding, "\n")
}

func (m *model) renderStatus() string {
	// If on the confirmation page, show configuration summary
	if m.form.State == huh.StateCompleted || isOnConfirmationPage(m.form) {
		return m.renderConfigurationSummary()
	}
	
	// Otherwise, show the current description
	return m.getCurrentDescription()
}

// isOnConfirmationPage checks if we're on the final confirmation page
func isOnConfirmationPage(form *huh.Form) bool {
	// Check if the form has a focused field with confirmation-related text
	focusedField := form.GetFocusedField()
	if focusedField != nil {
		// Check if the field title contains "Generate Claude Code configuration"
		// This is a simple way to detect the confirmation page
		if confirm, ok := focusedField.(*huh.Confirm); ok {
			// We can check some property that would indicate it's our confirmation field
			_ = confirm // Use the confirm variable to avoid unused variable error
			// For now, let's assume any confirm field on the last page is our confirmation
			return true
		}
	}
	return false
}

func (m *model) renderConfigurationSummary() string {
	var status strings.Builder
	
	status.WriteString("## 📋 Configuration Summary\n\n")
	status.WriteString("\n\n-----\n\n")
	
	// Show configuration path based on project-local setting
	if m.config.IsProjectLocal {
		currentDir, err := os.Getwd()
		if err != nil {
			currentDir = "<current directory>"
		}
		status.WriteString("### 📁 Configuration Path:\n")
		status.WriteString(fmt.Sprintf("  %s/.claude/\n\n", currentDir))
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "<home directory>"
		}
		status.WriteString("### 🏠 Configuration Path:\n")
		status.WriteString(fmt.Sprintf("  %s/.claude/\n\n", homeDir))
	}
	
	// Language Setup
	status.WriteString("### 💻 Languages\n")
	if len(m.config.Languages) > 0 {
		for _, lang := range m.config.Languages {
			status.WriteString(fmt.Sprintf("* %s\n", lang))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// Subagents
	status.WriteString("### 🤖 Subagents\n")
	if len(m.config.Subagents) > 0 {
		for _, agent := range m.config.Subagents {
			status.WriteString(fmt.Sprintf("* %s\n", cleanFormValue(agent)))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// Hooks
	status.WriteString("### 🪝 Hooks\n")
	if len(m.config.Hooks) > 0 {
		for _, hook := range m.config.Hooks {
			status.WriteString(fmt.Sprintf("* %s\n", cleanFormValue(hook)))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// Slash Commands
	status.WriteString("### 📟 Slash Commands\n")
	if len(m.config.SlashCommands) > 0 {
		for _, cmd := range m.config.SlashCommands {
			cleanCmd := cleanFormValue(cmd)
			status.WriteString(fmt.Sprintf("* /%s\n", cleanCmd))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// MCP
	status.WriteString("### 🔌 MCP Integration\n")
	if len(m.config.MCPServers) > 0 {
		for _, server := range m.config.MCPServers {
			status.WriteString(fmt.Sprintf("* %s\n", cleanFormValue(server)))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	
	return status.String()
}


// Helper function to clean emoji prefixes from form selections
func cleanFormValue(value string) string {
	// Remove emoji and space prefix (e.g., "🔍 code-reviewer" -> "code-reviewer")
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return value
}

func cleanFormValues(values []string) []string {
	cleaned := make([]string, len(values))
	for i, v := range values {
		cleaned[i] = cleanFormValue(v)
	}
	return cleaned
}

// Gradient System Functions (T021-T027)

// detectTerminalCapability detects terminal color support from env vars (T021)

// interpolateColor performs RGB interpolation between two colors (T022)

// interpolateGradient interpolates between two gradient themes (T023)

// adjustSaturation adjusts the saturation of a hex color (Feature 006: T007)

// increaseBrightness increases the brightness of a hex color (Feature 006: T008)

// extendColorPaletteForMarkdown extends palette with markdown colors (Feature 006: T011)

// generateGlamourStyle creates custom glamour renderer from palette (Feature 006: T012)

// gradient.EaseInOutCubic applies cubic easing for smooth animations (T024)

// quantizeStops reduces gradient stops for limited terminals (T025)

// applyGradient creates a Lipgloss style with gradient (T026)

// renderASCIITitle applies gradient to ASCII art line-by-line (T014)

// renderGradient renders text with gradient colors applied (T027)

// validateGradientTheme validates theme constraints (helper for tests)
func validateGradientTheme(theme gradient.Theme) error {
	if theme.Stops < 2 {
		return fmt.Errorf("stops must be >= 2, got %d", theme.Stops)
	}
	if theme.Intensity < 0.0 || theme.Intensity > 1.0 {
		return fmt.Errorf("intensity must be in [0.0, 1.0], got %f", theme.Intensity)
	}
	return nil
}

// Gradient palette definitions (T030)
type gradientPalettesType struct {
	primary    lipgloss.AdaptiveColor
	secondary  lipgloss.AdaptiveColor
	accent     lipgloss.AdaptiveColor
	error      lipgloss.AdaptiveColor
	success    lipgloss.AdaptiveColor
	background lipgloss.AdaptiveColor

	// Markdown-specific theme colors (Feature 006: T010)
	markdownHeading  lipgloss.AdaptiveColor // Heading color (H1-H6)
	markdownCode     lipgloss.AdaptiveColor // Code block/inline code
	markdownEmphasis lipgloss.AdaptiveColor // Italic/bold emphasis
	markdownLink     lipgloss.AdaptiveColor // Hyperlinks
}

var gradientPalettes = gradient.InitGradientPalettes()

// initGradientPalettes initializes color palettes (T030)

// initStyleMap populates component/state style mappings (T031)

// ============================================================================
// Feature 005: Asset File Generation Functions
// ============================================================================

// generateAllAssets generates all asset files from the module registry
func generateAllAssets(registry *ModuleRegistry) error {
	// Get current directory (repository root)
	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	assetsDir := filepath.Join(repoRoot, "assets")

	// Collect all asset paths from all modules
	var descriptors []generation.AssetFileDescriptor

	// Process subagents
	subagents := registry.List(TypeSubagent)
	for _, module := range subagents {
		for _, assetPath := range module.AssetPaths {
			desc := generation.AssetFileDescriptor{
				Name:           module.Name,
				Type:           generation.AssetTypeSubagent,
				Path:           assetPath,
				SourceTemplate: assetPath,
				Module:         module,
			}
			descriptors = append(descriptors, desc)
		}
	}

	// Process hooks
	hooks := registry.List(TypeHook)
	for _, module := range hooks {
		for _, assetPath := range module.AssetPaths {
			desc := generation.AssetFileDescriptor{
				Name:           module.Name,
				Type:           generation.AssetTypeHook,
				Path:           assetPath,
				SourceTemplate: assetPath,
				Module:         module,
			}
			descriptors = append(descriptors, desc)
		}
	}

	// Process slash commands
	commands := registry.List(TypeCommand)
	for _, module := range commands {
		for _, assetPath := range module.AssetPaths {
			desc := generation.AssetFileDescriptor{
				Name:           module.Name,
				Type:           generation.AssetTypeSlashCommand,
				Path:           assetPath,
				SourceTemplate: assetPath,
				Module:         module,
			}
			descriptors = append(descriptors, desc)
		}
	}

	// Check for existing files
	warning := generation.CheckExistingFiles(descriptors, assetsDir)
	if len(warning.ExistingFiles) > 0 {
		fmt.Printf("\n⚠️  WARNING: The following files will be overwritten:\n")
		for _, file := range warning.ExistingFiles {
			fmt.Printf("  - %s\n", file)
		}
		fmt.Printf("\nAny customizations will be lost.\n\n")
		fmt.Printf("Continue? (y/n): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("\nℹ️  Generation cancelled. No files were modified.")
			return nil
		}
	}

	// Generate all files
	fmt.Printf("\nGenerating asset files...\n")
	report := generation.GenerateAssetFiles(descriptors, assetsDir)

	// Display results
	for _, result := range report.Results {
		status := "✅"
		if result.Status == generation.StatusFailed {
			status = "❌"
		} else if result.Status == generation.StatusPlaceholderGenerated {
			status = "ℹ️ "
		}
		relPath, _ := filepath.Rel(repoRoot, result.FilePath)
		if result.Error != nil {
			fmt.Printf("%s %s - %v\n", status, relPath, result.Error)
		} else {
			fmt.Printf("%s %s\n", status, relPath)
		}
	}

	// Summary
	fmt.Printf("\n")
	if report.Failed > 0 {
		fmt.Printf("⚠️  %d files failed to generate. %d files succeeded.\n", report.Failed, report.Successful+report.PlaceholdersGenerated)

		if report.ShouldPromptRetry() {
			fmt.Printf("\nRetry failed files? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response == "y" || response == "Y" {
				fmt.Printf("\nRetrying failed files...\n")
				retryReport := generation.RetryFailedGeneration(&report, assetsDir)

				for _, result := range retryReport.Results {
					status := "✅"
					if result.Status == generation.StatusFailed {
						status = "❌"
					}
					relPath, _ := filepath.Rel(repoRoot, result.FilePath)
					if result.Error != nil {
						fmt.Printf("%s %s - %v\n", status, relPath, result.Error)
					} else {
						fmt.Printf("%s %s\n", status, relPath)
					}
				}

				if retryReport.Failed == 0 {
					fmt.Printf("\n✅ All files generated successfully.\n")
				} else {
					fmt.Printf("\n⚠️  %d files still failed.\n", retryReport.Failed)
				}
			}
		}
	} else {
		fmt.Printf("✅ Generated %d asset files successfully.\n", report.Successful+report.PlaceholdersGenerated)
		if report.PlaceholdersGenerated > 0 {
			fmt.Printf("ℹ️  %d placeholder files were generated. Please review and customize files marked with TODO.\n", report.PlaceholdersGenerated)
		}
	}

	return nil
}

func main() {
	// Initialize module registry (Feature 004)
	registry := &ModuleRegistry{}
	registryErrs := registry.Load(assets)
	if len(registryErrs) > 0 {
		fmt.Fprintf(os.Stderr, "warning: module registry errors: %d issues\n", len(registryErrs))
		for _, regErr := range registryErrs {
			fmt.Fprintf(os.Stderr, "  - %v\n", regErr)
		}
	}

	// Feature 005: Check for --generate-assets flag
	if len(os.Args) > 1 && os.Args[1] == "--generate-assets" {
		if err := generateAllAssets(registry); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Get current directory name for project name default
	currentDir, err := os.Getwd()
	dirName := "awesome-app" // default fallback
	if err == nil {
		baseName := filepath.Base(currentDir)
		if baseName != "." && baseName != "/" && baseName != "" {
			dirName = baseName
		}
	}

	// Load previous choices from persistence file
	persistedConfig, err := loadPersistenceConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load previous choices: %v\n", err)
		persistedConfig = &PersistenceConfig{}
	}

	// Initialize config with defaults, then override with persisted values
	cfg := Config{
		IsProjectLocal: true,  // Default to project-specific
		ProjectName:    dirName, // Set directory name as default
		Languages:      []string{"Go"},
		Subagents:      []string{"code-reviewer", "test-runner", "bug-sleuth"},
		Hooks:          []string{"session-start", "pre-tool-use", "post-tool-use"},
		SlashCommands:  []string{"example", "fix-github-issue"},
		MCPServers:     []string{"notion", "linear", "sentry", "github"},
	}
	
	// Override with persisted choices if they exist
	if len(persistedConfig.Languages) > 0 {
		cfg.Languages = persistedConfig.Languages
	}
	if len(persistedConfig.Subagents) > 0 {
		cfg.Subagents = persistedConfig.Subagents
	}
	if len(persistedConfig.Hooks) > 0 {
		cfg.Hooks = persistedConfig.Hooks
	}
	if len(persistedConfig.SlashCommands) > 0 {
		cfg.SlashCommands = persistedConfig.SlashCommands
	}
	if len(persistedConfig.MCPServers) > 0 {
		cfg.MCPServers = persistedConfig.MCPServers
	}
	if persistedConfig.ClaudeMDExtras != "" {
		cfg.ClaudeMDExtras = persistedConfig.ClaudeMDExtras
	}
	// Always use persisted boolean and project name if available
	if persistedConfig.ProjectName != "" {
		cfg.IsProjectLocal = persistedConfig.IsProjectLocal
		// Only override project name if it's not the current directory default
		if persistedConfig.ProjectName != dirName {
			cfg.ProjectName = persistedConfig.ProjectName
		}
	}

	form := huh.NewForm(
		// Page 1: Project Setup
		huh.NewGroup(
			huh.NewNote().Title("📁 Project Setup").Description("Configure your project basics and language support"),
			huh.NewInput().
				Title("Project name").
				Description("Used in generated documentation and configurations").
				Value(&cfg.ProjectName),
			huh.NewConfirm().
				Title("Project-specific configuration?").
				Description("Yes = Configure for this project only\nNo = Global configuration in your home directory").
				Value(&cfg.IsProjectLocal),
			huh.NewMultiSelect[string]().
				Key("languages").
				Title("Primary languages").
				Description("Select all languages used in your project for optimized defaults").
				Options(huh.NewOptions(
					"Go", "TypeScript", "Python", "Java", "Rust", "C++", "C#", 
					"PHP", "Ruby", "Swift", "Kotlin", "Dart", "Shell", "Lua",
					"Elixir", "Haskell", "Elm", "Julia", "SQL", "Arduino", 
					"Scheme", "Lisp")...).
				Height(8).
				Value(&cfg.Languages),
		),
		
		// Page 2: Subagent Selection
		huh.NewGroup(
			huh.NewNote().Title("🤖 Subagent Configuration").Description("Choose specialized AI assistants for your development workflow"),
			huh.NewMultiSelect[string]().
				Key("subagents").
				Title("Select subagents to include").
				Description("Choose the AI specialists you want available for your project").
				Options(registry.GetOptions(TypeSubagent)...).
				Value(&cfg.Subagents),
		),
		
		// Page 3: Hook Configuration
		huh.NewGroup(
			huh.NewNote().Title("🪝 Hook Setup").Description("Configure automation and lifecycle scripts"),
			huh.NewMultiSelect[string]().
				Key("hooks").
				Title("Select hooks to enable").
				Description("Automation scripts that run at specific points in your workflow").
				Options(registry.GetOptions(TypeHook)...).
				Value(&cfg.Hooks),
		),
		
		// Page 4: Slash Commands
		huh.NewGroup(
			huh.NewNote().Title("⚡ Custom Commands").Description("Add powerful slash commands for common development tasks"),
			huh.NewMultiSelect[string]().
				Key("slash-commands").
				Title("Select custom slash commands").
				Description("Choose useful commands for common development tasks").
				Options(registry.GetOptions(TypeCommand)...).
				Value(&cfg.SlashCommands),
		),
		
		// Page 5: MCP Configuration
		huh.NewGroup(
			huh.NewNote().Title("🔌 MCP Integration").Description("Connect to external tools and services via Model Context Protocol"),
			huh.NewMultiSelect[string]().
				Key("mcp-servers").
				Title("Select MCP servers to include").
				Description("Choose external tool integrations to enhance Claude's capabilities (optional)").
				Options(registry.GetOptions(TypeMCP)...).
				Value(&cfg.MCPServers),
		),
		
		// Page 6: Final Configuration  
		huh.NewGroup(
			huh.NewNote().Title("📝 Final Setup").Description("Add custom instructions and complete your configuration"),
			huh.NewText().
				Title("Extra CLAUDE.md content (optional)").
				Description("Project-specific instructions to include in CLAUDE.md").
				Value(&cfg.ClaudeMDExtras),
		),
		
		// Page 7: Confirmation
		huh.NewGroup(
			huh.NewNote().Title("✅ Confirmation").Description("Review your configuration and confirm to generate Claude Code setup"),
			huh.NewConfirm().
				Title("Generate Claude Code configuration?").
				Description("This will create/update the Claude Code configuration files with your selections.\nReview the configuration summary in the right panel.").
				Affirmative("Yes, generate configuration").
				Negative("No, go back to make changes").
				Value(&cfg.Confirmed),
		),
	)

	// Create Bubble Tea model with form (T029: initialize gradient system)
	termCap := gradient.DetectTerminalCapability()
	styleMap := gradient.InitStyleMap()
	primaryTheme := styleMap[gradient.HeaderComponent][gradient.NormalState].Theme

	// Extend color palette with markdown colors (Feature 006: T013)
	palette := gradientPalettes
	gradient.ExtendColorPaletteForMarkdown(&palette)

	// Create custom glamour renderer from palette (Feature 006: T013)
	renderer := gradient.GenerateGlamourStyle(palette)
	// renderer is nil-checked by existing code (will fallback to plain text)

	m := model{
		form:            form,
		config:          &cfg,
		glamourRenderer: renderer,

		// Gradient system initialization
		terminalCap:  termCap,
		currentTheme: primaryTheme,
		transition: gradient.TransitionState{
			Active:     false,
			EasingFunc: gradient.EaseInOutCubic,
		},
		styleMap: styleMap,

		// Module registry (Feature 004)
		registry: registry,

		// Adaptive right panel layout (Feature 007)
		// showRightPanel will be computed on first WindowSizeMsg
		showRightPanel:  true, // Default to showing panel (will be adjusted on first resize)
		resizeDebouncer: nil,
		pendingResize:   nil,
	}

	// Run the Bubble Tea application
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running application: %v\n", err)
		os.Exit(1)
	}

	// Check if user cancelled
	if finalModel, ok := finalModel.(model); ok {
		if finalModel.form.State != huh.StateCompleted {
			fmt.Fprintf(os.Stderr, "cancelled\n")
			os.Exit(1)
		}
	}

	// Clean up emoji prefixes from form selections
	cfg.Subagents = cleanFormValues(cfg.Subagents)
	cfg.Hooks = cleanFormValues(cfg.Hooks)
	cfg.MCPServers = cleanFormValues(cfg.MCPServers)
	
	// Save current choices for future runs
	if err := savePersistenceConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save choices for future runs: %v\n", err)
		// Continue execution - this is not a fatal error
	}
	
	// Clean up deselected items before generating new configuration
	var targetDir string
	if cfg.IsProjectLocal {
		targetDir, _ = os.Getwd()
	} else {
		homeDir, _ := os.UserHomeDir()
		targetDir = filepath.Join(homeDir, ".claude")
	}
	if targetDir != "" {
		if err := cleanupDeselectedItems(cfg, persistedConfig, targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clean up deselected items: %v\n", err)
		}
	}
	
	if err := run(cfg, registry); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if cfg.IsProjectLocal {
		fmt.Println("\n✅ claudekit finished. Project-specific Claude Code configuration created!")
		fmt.Println("   Open Claude Code in this directory and start coding!")
	} else {
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".claude")
		fmt.Printf("\n✅ claudekit finished. Global Claude Code configuration created!\n")
		fmt.Printf("   Configuration saved to: %s\n", configPath)
		fmt.Println("   This configuration will apply to all your Claude Code sessions.")
	}
}

// cleanupDeselectedItems removes files for items that were previously selected but now deselected
func cleanupDeselectedItems(cfg Config, persistedConfig *PersistenceConfig, targetDir string) error {
	claudeDir := filepath.Join(targetDir, ".claude")
	
	// Clean up deselected subagents
	for _, oldAgent := range persistedConfig.Subagents {
		if !slices.Contains(cfg.Subagents, oldAgent) {
			agentFile := filepath.Join(claudeDir, "agents", oldAgent+".md")
			if _, err := os.Stat(agentFile); err == nil {
				if err := os.Remove(agentFile); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to remove deselected agent %s: %v\n", oldAgent, err)
				}
			}
		}
	}
	
	// Clean up deselected hooks
	for _, oldHook := range persistedConfig.Hooks {
		if !slices.Contains(cfg.Hooks, oldHook) {
			hookFile := filepath.Join(claudeDir, "hooks", oldHook+".sh")
			if _, err := os.Stat(hookFile); err == nil {
				if err := os.Remove(hookFile); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to remove deselected hook %s: %v\n", oldHook, err)
				}
			}
		}
	}
	
	// Clean up deselected slash commands
	for _, oldCmd := range persistedConfig.SlashCommands {
		if !slices.Contains(cfg.SlashCommands, oldCmd) {
			// Remove both .md and .py files (legacy .py support)
			for _, ext := range []string{".md", ".py"} {
				cmdFile := filepath.Join(claudeDir, "commands", oldCmd+ext)
				if _, err := os.Stat(cmdFile); err == nil {
					if err := os.Remove(cmdFile); err != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to remove deselected command %s: %v\n", oldCmd, err)
					}
				}
			}
		}
	}
	
	return nil
}

func run(cfg Config, registry *ModuleRegistry) error {
	var targetDir string
	var err error
	
	if cfg.IsProjectLocal {
		// Project-specific: use current directory
		targetDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else {
		// Global: use home directory with .claude subdirectory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		targetDir = filepath.Join(homeDir, ".claude")
	}
	
	abs, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	// Create directories
	mustMkdir(filepath.Join(abs, ".claude"))
	mustMkdir(filepath.Join(abs, ".claude", "agents"))
	mustMkdir(filepath.Join(abs, ".claude", "hooks"))
	if len(cfg.SlashCommands) > 0 {
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

	// Write selected hook scripts
	for _, hookDisplay := range cfg.Hooks {
		hookName := cleanFormValue(hookDisplay)
		var content string
		var filename string
		
		switch hookName {
		case "pre-tool-use":
			content = generateHookScript(hookName, "Runs before Claude executes any tool")
			filename = "pre-tool-use.sh"
		case "post-tool-use":
			content = generateHookScript(hookName, "Runs after successful tool execution")
			filename = "post-tool-use.sh"
		case "notification":
			content = generateHookScript(hookName, "Runs when Claude needs permission or when prompts idle")
			filename = "notification.sh"
		case "user-prompt-submit":
			content = generateHookScript(hookName, "Runs when users submit prompts, before Claude processes them")
			filename = "user-prompt-submit.py"
		case "stop":
			content = generateHookScript(hookName, "Runs when Claude finishes responding")
			filename = "stop.sh"
		case "subagent-stop":
			content = generateHookScript(hookName, "Runs when Claude Code subagents finish responding")
			filename = "subagent-stop.sh"
		case "session-end":
			content = generateHookScript(hookName, "Runs when Claude Code sessions terminate")
			filename = "session-end.sh"
		case "pre-compact":
			content = generateHookScript(hookName, "Runs before context compaction operations")
			filename = "pre-compact.sh"
		case "session-start":
			content = sessionStartScript() // Use existing script
			filename = "session-start.sh"
		default:
			continue
		}
		
		if err := writeExecutable(filepath.Join(abs, ".claude", "hooks", filename), content); err != nil {
			return err
		}
	}

	// Write settings.json with hooks + permissions
	st := buildSettings(abs, cfg, registry)
	buf, _ := json.MarshalIndent(st, "", "  ")
	if err := os.WriteFile(filepath.Join(abs, ".claude", "settings.json"), buf, 0o644); err != nil {
		return err
	}

	// Create selected slash commands
	for _, cmdDisplay := range cfg.SlashCommands {
		cmdName := cleanFormValue(cmdDisplay)
		var content string
		if cmdName == "example" {
			content = sampleSlashCommand()
		} else {
			content = generateSlashCommand(cmdName, registry)
		}
		
		if err := os.WriteFile(
			filepath.Join(abs, ".claude", "commands", cmdName+".md"),
			[]byte(content), 0o644); err != nil {
			return err
		}
	}

	// MCP project config
	if len(cfg.MCPServers) > 0 {
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

func buildSettings(projectDir string, cfg Config, registry *ModuleRegistry) settings {
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

	// Add all selected hooks using registry (Feature 004)
	for _, hookDisplay := range cfg.Hooks {
		hookName := cleanFormValue(hookDisplay)

		// Get hook module from registry
		hookModule := registry.Get(TypeHook, hookName)
		if hookModule == nil {
			continue // Skip unknown hooks
		}

		// Extract defaults from module
		hookType, _ := hookModule.Defaults["hook_type"].(string)
		command, _ := hookModule.Defaults["command"].(string)
		timeout, _ := hookModule.Defaults["timeout"].(float64) // JSON numbers are float64

		if hookType == "" || command == "" {
			continue // Skip malformed hook modules
		}

		s.Hooks[hookType] = append(s.Hooks[hookType],
			hookMatcher{
				Hooks: []hookCmd{{
					Type:    "command",
					Command: command,
					Timeout: int(timeout),
				}},
			},
		)
	}

	return s
}

func renderClaudeMD(cfg Config) string {
	tmplContent, err := assets.ReadFile("assets/templates/CLAUDE.md.tmpl")
	if err != nil {
		panic(err)
	}
	
	tmpl, err := template.New("claude").Funcs(template.FuncMap{
		"or": or,
	}).Parse(string(tmplContent))
	if err != nil {
		panic(err)
	}
	
	data := struct {
		Config
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
		Date          string
	}{
		Config:        cfg,
		HasGo:         includes(cfg.Languages, "Go"),
		HasTypeScript: includes(cfg.Languages, "TypeScript"),
		HasPython:     includes(cfg.Languages, "Python"),
		HasRust:       includes(cfg.Languages, "Rust"),
		HasCpp:        includes(cfg.Languages, "C++"),
		HasJava:       includes(cfg.Languages, "Java") || includes(cfg.Languages, "Kotlin"),
		HasCsharp:     includes(cfg.Languages, "C#"),
		HasPhp:        includes(cfg.Languages, "PHP"),
		HasRuby:       includes(cfg.Languages, "Ruby"),
		HasSwift:      includes(cfg.Languages, "Swift"),
		HasDart:       includes(cfg.Languages, "Dart"),
		HasShell:      includes(cfg.Languages, "Shell"),
		HasLua:        includes(cfg.Languages, "Lua"),
		HasElixir:     includes(cfg.Languages, "Elixir"),
		HasHaskell:    includes(cfg.Languages, "Haskell"),
		HasElm:        includes(cfg.Languages, "Elm"),
		HasJulia:      includes(cfg.Languages, "Julia"),
		HasSql:        includes(cfg.Languages, "SQL"),
		Date:          time.Now().Format("2006-01-02"),
	}
	
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		panic(err)
	}
	return b.String()
}

func renderAgent(name string) string {
	content, err := assets.ReadFile("assets/agents/" + name + ".md")
	if err != nil {
		return `---
name: ` + name + `
description: Custom subagent
---
Provide a focused role and steps.`
	}
	return string(content)
}

func postWriteLintScript(langs []string) string {
	tmplContent, err := assets.ReadFile("assets/hooks/postwrite-lint.sh.tmpl")
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
		HasGo:         includes(langs, "Go"),
		HasTypeScript: includes(langs, "TypeScript"),
		HasPython:     includes(langs, "Python"),
		HasRust:       includes(langs, "Rust"),
		HasCpp:        includes(langs, "C++"),
		HasJava:       includes(langs, "Java") || includes(langs, "Kotlin"),
		HasCsharp:     includes(langs, "C#"),
		HasPhp:        includes(langs, "PHP"),
		HasRuby:       includes(langs, "Ruby"),
		HasSwift:      includes(langs, "Swift"),
		HasDart:       includes(langs, "Dart"),
		HasShell:      includes(langs, "Shell"),
		HasLua:        includes(langs, "Lua"),
		HasElixir:     includes(langs, "Elixir"),
		HasHaskell:    includes(langs, "Haskell"),
		HasElm:        includes(langs, "Elm"),
		HasJulia:      includes(langs, "Julia"),
		HasSql:        includes(langs, "SQL"),
	}
	
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		panic(err)
	}
	return b.String()
}

func generateHookScript(hookName, description string) string {
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

func preWriteGuardScript() string {
	content, err := assets.ReadFile("assets/hooks/prewrite-guard.sh")
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

func sessionStartScript() string {
	content, err := assets.ReadFile("assets/hooks/session-start-context.sh")
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

func promptLintPy() string {
	content, err := assets.ReadFile("assets/hooks/prompt-lint.py")
	if err != nil {
		panic(err)
	}
	return string(content)
}

func sampleSlashCommand() string {
	content, err := assets.ReadFile("assets/templates/fix-github-issue.md")
	if err != nil {
		panic(err)
	}
	return string(content)
}

func generateSlashCommand(cmdName string, registry *ModuleRegistry) string {
	// Generate custom slash command content based on the command name (Feature 004: use registry)
	module := registry.Get(TypeCommand, cmdName)
	if module == nil {
		return fmt.Sprintf(`---
name: %s
description: Custom command
---

# %s Command

Add your custom command implementation here.
`, cmdName, strings.Title(strings.ReplaceAll(cmdName, "-", " ")))
	}

	desc := module.Description

	// Extract command name from description (between ** markers)
	titleStart := strings.Index(desc, "**")
	titleEnd := strings.Index(desc[titleStart+2:], "**")
	var title string
	if titleStart != -1 && titleEnd != -1 {
		title = desc[titleStart+2 : titleStart+2+titleEnd]
	} else {
		title = "/" + cmdName
	}

	// Extract description after the title
	descStart := strings.Index(desc, " - ")
	var description string
	if descStart != -1 {
		description = strings.TrimSpace(desc[descStart+3:])
	} else {
		description = "Custom development command"
	}

	return fmt.Sprintf(`---
name: %s
description: %s
---

# %s

%s

## Usage

Use this command to automate complex development tasks. The command will:

1. Analyze the current project context
2. Execute the requested operation
3. Provide detailed feedback and results
4. Ensure code quality and best practices

Add specific implementation details and parameters as needed.
`, cmdName, description, title, description)
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