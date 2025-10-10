package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	huh "github.com/charmbracelet/huh"

	"claudekit/internal/gradient"
)

// Layout threshold constants for adaptive right panel display.
const (
	MinWidthForPanel  = 140 // Minimum terminal columns for right panel
	MinHeightForPanel = 40  // Minimum terminal rows for right panel
	ResizeDebounceMS  = 200 // Debounce delay in milliseconds
)

// Config holds the user's configuration choices from the interactive form.
type Config struct {
	IsProjectLocal bool     // true = project-based, false = global/home directory
	ProjectName    string
	Languages      []string
	Subagents      []string
	Hooks          []string
	SlashCommands  []string
	MCPServers     []string
	ClaudeMDExtras string
	Confirmed      bool // for final confirmation step
}

// Model represents the TUI application state.
type Model struct {
	Form            *huh.Form
	Config          interface{} // Will be *Config
	Viewport        viewport.Model
	GlamourRenderer *glamour.TermRenderer
	Ready           bool
	Width           int
	Height          int
	CurrentFocus    string

	// Gradient system fields
	TerminalCap  gradient.TerminalCapability
	CurrentTheme gradient.Theme
	Transition   gradient.TransitionState
	StyleMap     map[gradient.ComponentType]map[gradient.VisualState]gradient.ComponentStyle

	// Module registry
	Registry interface{} // Will be *modules.ModuleRegistry

	// Adaptive right panel layout
	ShowRightPanel  bool               // Computed: width >= 140 && height >= 40
	ResizeDebouncer *time.Timer        // Active debounce timer (nil if none)
	PendingResize   *tea.WindowSizeMsg // Cached resize message during debounce
}

// DebounceCompleteMsg signals that resize debounce period has elapsed.
type DebounceCompleteMsg struct{}

// TickMsg is our custom message for gradient animations.
type TickMsg time.Time

// LanguageDescriptions maps language names to their markdown descriptions.
var LanguageDescriptions = map[string]string{
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
}
