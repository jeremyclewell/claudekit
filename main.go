package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	huh "github.com/charmbracelet/huh"
)
//go:embed assets/*
var assets embed.FS

type Config struct {
	IsProjectLocal bool       // true = project-based, false = global/home directory
	ProjectName    string
	Languages      []string
	Subagents      []string
	Hooks          []string
	WantSlashCmd   bool
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

// Bubble Tea Model for the application
type model struct {
	form            *huh.Form
	config          *Config
	viewport        viewport.Model
	descViewport    viewport.Model
	ready           bool
	width           int
	height          int
	currentFocus    string
}

// Styles for the Uaud
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1).
			MarginLeft(2)

	formStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1)

	descStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#25A065")).
			Padding(1).
			MarginTop(1)
)

// Detailed descriptions for subagents
var subagentDescriptions = map[string]string{
	"code-reviewer": `🔍 CODE REVIEWER - Senior Code Review Specialist

EXPERTISE: 15+ years across multiple languages and architectures
MISSION: Ensure code quality, security, maintainability, and team knowledge transfer

REVIEW PROCESS:
• Context Gathering - Analyzes git diff and recent development context
• Critical Issues - Security vulnerabilities, memory leaks, breaking changes
• Warnings - Performance anti-patterns, code smells, missing error handling  
• Suggestions - Refactoring opportunities, better naming, documentation

LANGUAGE-SPECIFIC FOCUS:
• C++: RAII compliance, memory management, move semantics
• Go: Error handling, goroutine leaks, context usage
• TypeScript: Type safety, async/await patterns, bundle impact
• Python: PEP compliance, exception handling, type hints
• And 18+ other languages with specialized knowledge

OUTPUT FORMAT: Structured findings with severity levels, explanations, and fix suggestions`,

	"test-runner": `🧪 TEST RUNNER - Test Engineer & QA Specialist

CORE PRINCIPLE: Never weaken tests to make them pass; always fix the root cause
EXPERTISE: TDD, BDD, and modern testing practices across all major languages

TESTING STRATEGY:
1. Assessment Phase - Identifies changed files and affected tests
2. Quick Feedback Loop - Runs only relevant tests first for fast iteration
3. Full Test Suite - Comprehensive testing with coverage analysis
4. Failure Analysis - Diagnoses environment, test quality, and code issues

SUPPORTED FRAMEWORKS:
• JavaScript/TypeScript: Jest, Vitest, Mocha, Playwright, Cypress
• Python: pytest, unittest, hypothesis for property testing
• Go: Built-in testing, testify, table-driven tests
• Rust: Built-in framework, proptest, mockall
• Plus 18+ other languages with appropriate testing tools

TEST CATEGORIES:
• Unit Tests (First priority) - Individual functions with >90% coverage
• Integration Tests - Component interactions and database integration  
• End-to-End Tests - Full user workflows and cross-browser testing`,

	"bug-sleuth": `🕵️ BUG SLEUTH - Senior Debugging Specialist  

PHILOSOPHY: "The bug is always logical, never random" - 20+ years of debugging experience
METHODOLOGY: Systematic RIDDE approach for elusive bug hunting

THE RIDDE METHOD:
1. REPRODUCE - Make it happen consistently with minimal test cases
2. ISOLATE - Binary search to narrow the problem space
3. DIAGNOSE - Deep investigation with proper debugging tools
4. DEBUG - Strategic logging, memory analysis, timing issues
5. ELIMINATE - Fix root cause with minimal targeted changes

DEBUGGING TOOLS BY LANGUAGE:
• C++: GDB, Valgrind, AddressSanitizer, ThreadSanitizer
• JavaScript: Chrome DevTools, Node inspector, memory profiling
• Python: pdb, memory profiler, execution tracing
• Go: Delve debugger, race detection, execution tracing
• Plus comprehensive tool knowledge for 18+ other languages

BUG CATEGORIES:
• Logic Errors - Wrong output, trace data flow
• Race Conditions - Thread-safe operations, stress testing
• Memory Issues - Allocation tracking, leak detection
• Integration Failures - API errors, component isolation`,

	"security-auditor": `🔒 SECURITY AUDITOR - Cybersecurity & Code Auditor

EXPERTISE: Application security, penetration testing, secure code review
MISSION: Identify vulnerabilities before production using industry frameworks

OWASP TOP 10 FOCUSED REVIEW:
1. Injection Attacks - SQL, NoSQL, LDAP, OS Command injection prevention
2. Broken Authentication - Session management, password policies, MFA
3. Sensitive Data Exposure - Encryption at rest/transit, PII protection
4. XML External Entities - Safe XML parsing configurations
5. Broken Access Control - RBAC/ABAC, direct object reference protection
6. Security Misconfiguration - Debug settings, default credentials
7. Cross-Site Scripting - Input validation, output encoding, CSP
8. Insecure Deserialization - Safe serialization practices
9. Known Vulnerabilities - Dependency scanning and updates
10. Logging & Monitoring - Security event tracking and alerting

LANGUAGE-SPECIFIC SECURITY PATTERNS:
• C++: Buffer overflows, format string vulnerabilities, memory safety
• JavaScript: Prototype pollution, unsafe regex, code injection
• Python: Code injection, unsafe YAML/pickle loading, path traversal
• Plus security knowledge for 18+ other languages

COMPLIANCE CONSIDERATIONS: GDPR, PCI DSS, HIPAA, SOX, ISO 27001`,

	"perf-optimizer": `⚡ PERFORMANCE OPTIMIZER - Performance Engineering Specialist

PHILOSOPHY: "Measure first, optimize second, measure again"
METHODOLOGY: Data-driven MAPLE approach for systematic optimization

THE MAPLE METHOD:
1. MEASURE - Establish baseline performance with proper profiling
2. ANALYZE - Understand algorithmic complexity and bottlenecks  
3. PRIORITIZE - Apply 80/20 rule focusing on highest impact
4. LEVERAGE - Use proven optimization techniques and tools
5. EVALUATE - Verify improvements and prevent regressions

PROFILING TOOLS BY LANGUAGE:
• C++: GNU gprof, Valgrind, Intel VTune, perf
• JavaScript: Node --prof, Chrome DevTools, Clinic.js
• Python: cProfile, memory profiler, line profiler
• Go: pprof, race detection, benchmarking tools
• Plus comprehensive profiling knowledge for 18+ other languages

OPTIMIZATION CATEGORIES:
1. Algorithmic (Highest Impact) - O(n²) to O(n log n) improvements
2. I/O Optimization - Database queries, file operations, network calls
3. Memory Optimization - Cache strategies, allocation patterns
4. Concurrency - Parallelization, lock-free data structures

SUCCESS METRICS: User-visible improvements, maintainable solutions`,

	"docs-writer": `📚 DOCS WRITER - Technical Documentation Specialist

EXPERTISE: Comprehensive documentation, API docs, user guides
MISSION: Create clear, maintainable documentation that serves as project knowledge

DOCUMENTATION TYPES:
• API Documentation - OpenAPI specs, endpoint descriptions, examples
• User Guides - Step-by-step tutorials, getting started guides
• Technical Specifications - Architecture decisions, system design
• Code Documentation - Inline comments, module documentation
• Process Documentation - Development workflows, deployment guides

DOCUMENTATION STANDARDS:
• Clear Structure - Logical organization with consistent formatting
• Comprehensive Coverage - All public APIs and user-facing features
• Living Documentation - Automated updates from code annotations
• Accessibility - Screen reader compatible, multiple formats
• Searchable - Proper indexing and cross-referencing

TOOLS & FORMATS:
• Markdown - GitHub flavored, CommonMark specification
• OpenAPI/Swagger - RESTful API documentation
• JSDoc/TSDoc - JavaScript/TypeScript inline documentation  
• Sphinx/ReadTheDocs - Python documentation ecosystems
• Plus documentation tools for 18+ other languages

QUALITY CRITERIA:
• Accuracy - Up-to-date with current implementation
• Completeness - Covers all necessary information
• Clarity - Written for the target audience level
• Examples - Working code samples and use cases`,

	"release-manager": `🚀 RELEASE MANAGER - Release Engineering & DevOps Specialist

PHILOSOPHY: "Release early, release often, release safely"
METHODOLOGY: SHIP method for predictable, reliable software releases

THE SHIP METHOD:
1. SCAN - Assess release readiness, CI/CD status, quality gates
2. HARMONIZE - Coordinate dependencies, timing, rollback procedures  
3. INTEGRATE - Version bumps, changelogs, release artifacts
4. PUBLISH - Controlled deployment with monitoring and verification

MULTI-LANGUAGE BUILD SUPPORT:
• JavaScript/TypeScript: npm, webpack, rollup, vite
• Python: setuptools, poetry, wheel, conda
• Go: go build, cross-compilation, modules
• Rust: cargo, cross-platform targets
• Plus build systems for 18+ other languages

SEMANTIC VERSIONING GUIDELINES:
• MAJOR (x.0.0) - Breaking changes, API modifications
• MINOR (0.x.0) - New features, backward compatible  
• PATCH (0.0.x) - Bug fixes, security patches
• Pre-release - alpha, beta, rc identifiers

RELEASE ARTIFACTS:
• Automated Changelogs - Categorized by type (feat, fix, breaking)
• Release Notes - Technical and user-facing versions
• Migration Guides - Breaking change documentation
• Deployment Configurations - Environment-specific settings

QUALITY GATES: Test coverage, security scans, performance benchmarks`,

	"data-scientist": `📊 DATA SCIENTIST - Data Analysis & Insights Specialist

EXPERTISE: Statistical analysis, machine learning, data visualization
MISSION: Transform raw data into actionable insights and predictive models

DATA ANALYSIS CAPABILITIES:
• Exploratory Data Analysis - Statistical summaries, distribution analysis
• Data Cleaning - Missing value handling, outlier detection, normalization  
• Feature Engineering - Variable transformation, dimensionality reduction
• Statistical Testing - Hypothesis testing, A/B test analysis
• Time Series Analysis - Trend analysis, forecasting, seasonality

MACHINE LEARNING EXPERTISE:
• Supervised Learning - Classification, regression, ensemble methods
• Unsupervised Learning - Clustering, dimensionality reduction
• Model Selection - Cross-validation, hyperparameter tuning
• Model Evaluation - Performance metrics, validation strategies
• Model Deployment - Production ML pipelines, monitoring

SQL & DATABASE SKILLS:
• Complex Queries - JOINs, CTEs, window functions, subqueries
• Query Optimization - Index usage, execution plan analysis
• Data Warehousing - Star schema, ETL processes, data modeling
• Database Performance - Query tuning, partition strategies

VISUALIZATION TOOLS:
• Statistical Plots - Distributions, correlations, regression diagnostics
• Business Dashboards - KPI tracking, interactive visualizations  
• Exploratory Visualizations - Pattern discovery, anomaly detection

LANGUAGES: Python (pandas, scikit-learn), R, SQL, Julia statistical computing`,
}

// Detailed descriptions for MCP servers  
var mcpDescriptions = map[string]string{
	"notion": `📝 NOTION MCP SERVER - Comprehensive Notion Integration

CAPABILITIES:
• Page Management - Create, read, update, and delete Notion pages
• Database Operations - Query databases, add records, update properties
• Content Manipulation - Rich text formatting, blocks, embeds
• Search Functionality - Full-text search across workspaces
• Template Management - Create and use page templates

USE CASES:
• Documentation - Automatically update project documentation in Notion
• Task Management - Create and track issues directly from Claude Code
• Knowledge Base - Build and maintain technical knowledge repositories  
• Meeting Notes - Generate and organize development meeting minutes
• Project Planning - Create project roadmaps and requirement documents

INTEGRATION EXAMPLES:
• "Create a new project spec in Notion based on this codebase"
• "Update the API documentation page with these new endpoints"  
• "Search our knowledge base for information about this error"
• "Create a task in our sprint planning database"
• "Generate a project status report from recent commits"

CONFIGURATION:
• Requires Notion API token with appropriate workspace permissions
• Supports both personal and team Notion workspaces
• Configurable database and page access scopes`,

	"linear": `📋 LINEAR MCP SERVER - Advanced Issue & Project Management

CAPABILITIES:  
• Issue Management - Create, update, assign, and track Linear issues
• Project Operations - Manage projects, milestones, and roadmaps
• Team Coordination - Handle team assignments and workflows
• Label & Status Management - Organize issues with labels and custom statuses
• Comment & Activity Tracking - Add comments and track issue activity

USE CASES:
• Bug Reporting - Automatically create Linear issues from code analysis
• Feature Planning - Convert code TODOs into tracked Linear issues
• Sprint Planning - Organize and prioritize development work
• Code Review Integration - Link pull requests to Linear issues
• Progress Tracking - Monitor development velocity and completion rates

INTEGRATION EXAMPLES:
• "Create a Linear issue for this performance bottleneck I found"
• "Update the status of issue LIN-123 to 'In Review'"
• "Show me all open issues assigned to the backend team"  
• "Create a project milestone for the v2.0 release"
• "Convert these TODO comments into Linear issues"

WORKFLOW AUTOMATION:
• Automatic issue creation from code comments
• Status updates based on git branch/PR status  
• Integration with CI/CD pipelines for deployment tracking
• Custom workflows for different issue types`,

	"sentry": `🐛 SENTRY MCP SERVER - Error & Performance Monitoring

CAPABILITIES:
• Error Tracking - Monitor application errors and exceptions in real-time
• Performance Monitoring - Track application performance metrics and bottlenecks
• Release Management - Associate errors with specific releases and deployments  
• Alert Configuration - Set up custom alerts for error thresholds and patterns
• Issue Management - Group, assign, and resolve error issues

USE CASES:
• Error Analysis - Deep dive into production errors with full context
• Performance Debugging - Identify slow queries and performance regressions
• Release Health - Monitor error rates after deployments  
• Proactive Monitoring - Get notified before users report issues
• Root Cause Analysis - Trace errors through distributed systems

INTEGRATION EXAMPLES:
• "Show me all new errors introduced in the latest release"
• "Analyze the performance impact of the recent database changes"
• "Create alerts for any errors affecting more than 1% of users"
• "What's the current error rate for the payment processing service?"
• "Show me the stack trace for the most frequent error this week"

MONITORING CAPABILITIES:
• Real-time error tracking across multiple environments
• Performance profiling for web applications and APIs
• Custom metric tracking for business-specific KPIs
• Integration with popular frameworks and languages
• Distributed tracing for microservices architectures`,

	"github": `🐙 GITHUB MCP SERVER - Advanced Repository Management

CAPABILITIES:
• Repository Management - Create, clone, and manage GitHub repositories
• Issue & PR Operations - Advanced issue tracking and pull request management
• Code Analysis - Repository statistics, contributor analysis, code quality metrics
• Workflow Automation - GitHub Actions integration and workflow management
• Team Management - Collaborator management, permissions, team coordination

USE CASES:
• Code Review Automation - Streamline pull request reviews and approvals
• Issue Triage - Intelligent issue labeling and assignment
• Repository Analytics - Track development metrics and team productivity
• Release Automation - Automated release creation and change log generation
• Security Monitoring - Vulnerability scanning and security alerts

INTEGRATION EXAMPLES:
• "Create a new repository for this microservice with standard templates"
• "Analyze the commit history to identify the most active contributors"
• "Set up automated PR checks for code quality and test coverage"
• "Create a release with change log from recent commits"
• "Show me all open security vulnerabilities across our repositories"

ADVANCED FEATURES:
• Multi-repository operations and batch processing
• Custom GitHub App integration for enhanced permissions
• Webhook configuration for real-time event processing
• Integration with GitHub Packages for dependency management
• Support for GitHub Enterprise and advanced security features`,

	"airtable": `📊 AIRTABLE MCP SERVER - Collaborative Database Management

CAPABILITIES:
• Base Management - Create and manage Airtable bases with custom schemas
• Record Operations - Add, update, delete, and query records across tables
• View Management - Work with different views, filters, and sorting options
• Formula Integration - Use and create custom formulas for data processing
• Attachment Handling - Manage file attachments and media within records

USE CASES:
• Project Tracking - Maintain project databases with custom fields and views
• Bug Tracking - Alternative to traditional issue trackers with custom workflows
• Documentation - Maintain structured documentation with rich metadata
• Team Coordination - Track team member availability, skills, and assignments
• Data Analysis - Store and analyze structured data from various sources

INTEGRATION EXAMPLES:
• "Add this performance benchmark data to our metrics tracking base"
• "Create a new project record with the current sprint information"
• "Update the bug tracking table with issues found during code review"
• "Query the team skills base to find experts in React and TypeScript"
• "Generate a report from our feature request tracking table"

COLLABORATION FEATURES:
• Multi-user collaboration with real-time updates
• Rich field types including attachments, links, and formulas
• Custom views for different team roles and responsibilities
• API-driven automation for data synchronization
• Integration with popular tools and services through Airtable's ecosystem`,
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m *model) getCurrentDescription() string {
	// Get current focus from form state
	if m.form.State == huh.StateCompleted {
		return "✅ Configuration complete! Ready to generate your Claude Code setup."
	}
	
	// For now, provide general descriptions based on what might be selected
	// This could be enhanced later when we can better detect current focus
	
	return `📋 Claude Code Project Setup

Welcome to the interactive Claude Code project configuration tool! This wizard will help you set up a comprehensive development environment with AI-powered assistants and external tool integrations.

🔍 NAVIGATION:
• Use arrow keys to navigate between options
• Use space to select/deselect items in multi-select lists
• Use tab to move between form fields
• Use enter to proceed to the next page

📚 WHAT YOU'RE CONFIGURING:
• Project basics (directory, name, languages)
• AI subagents for specialized development tasks
• Automation hooks for workflow enhancement
• External tool integrations via MCP

Each selection will enhance your Claude Code experience with expert knowledge, proven methodologies, and powerful integrations. Choose the options that best fit your development workflow and project needs.`
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Calculate layout dimensions with proper spacing accounting
		formWidth := int(float64(msg.Width) * 0.6)        // 60% width for left side
		statusWidth := msg.Width - formWidth - 6          // Remaining width for right side
		
		descHeight := int(float64(msg.Height) * 0.3) - 50  // 30% of available height minus large buffer
		statusHeight := msg.Height - 50                // Use height minus large buffer
		
		if descHeight < 3 {
			descHeight = 3
		}
		if statusHeight < 10 {
			statusHeight = 10
		}
		
		if !m.ready {
			m.viewport = viewport.New(statusWidth, statusHeight)          // Status: full available height
			m.descViewport = viewport.New(formWidth-4, descHeight)        // Description: 30% height, account for borders
			m.ready = true
		} else {
			m.viewport.Width = statusWidth
			m.viewport.Height = statusHeight
			m.descViewport.Width = formWidth - 4
			m.descViewport.Height = descHeight
		}
		
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Update form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Update viewport content with current config status
	m.viewport.SetContent(m.renderStatus())
	
	// Update description viewport with current focus information
	m.descViewport.SetContent(m.getCurrentDescription())

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

	// Calculate dimensions with proper overflow prevention
	formWidth := int(float64(m.width) * 0.6)
	statusWidth := m.width - formWidth - 6
	
	descHeight := int(float64(m.height) * 0.3) - 50    // 30% for description minus large buffer
	formHeight := m.height - descHeight - 50 - 1       // Remaining height minus buffers
	
	if descHeight < 3 {
		descHeight = 3  // Ensure minimum description height
	}
	if formHeight < 3 {
		formHeight = 3  // Ensure minimum form height
	}

	// Title
	title := titleStyle.Render("🛠️  Claude Code Project Setup")
	
	// Left side panels
	formPanel := formStyle.
		Width(formWidth-2).  // Account for border
		Height(formHeight).
		Render(m.form.View())
	
	descPanel := descStyle.
		Width(formWidth-2).  // Account for border
		Height(descHeight).
		Render(m.descViewport.View())
	
	// Left side content (form above description)
	leftContent := lipgloss.JoinVertical(lipgloss.Left, formPanel, descPanel)
	
	// Status panel (right side, available height)
	statusPanel := statusStyle.
		Width(statusWidth-2).    // Account for border
		Height(m.height - 50).  // Use height minus buffer
		Render(m.viewport.View())

	// Main content (left content + status)
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, statusPanel)
	
	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

func (m *model) renderStatus() string {
	var status strings.Builder
	
	status.WriteString("📋 Current Configuration\n\n")
	
	// Project Setup
	status.WriteString("📁 Project Setup:\n")
	if m.config.IsProjectLocal {
		status.WriteString("  Scope: Project-specific (current directory)\n")
	} else {
		status.WriteString("  Scope: Global (home directory)\n")
	}
	if m.config.ProjectName != "" {
		status.WriteString(fmt.Sprintf("  Name: %s\n", m.config.ProjectName))
	}
	if len(m.config.Languages) > 0 {
		status.WriteString(fmt.Sprintf("  Languages: %s\n", strings.Join(m.config.Languages, ", ")))
	}
	status.WriteString("\n")
	
	// Subagents
	status.WriteString("🤖 Subagents:\n")
	if len(m.config.Subagents) > 0 {
		for _, agent := range m.config.Subagents {
			status.WriteString(fmt.Sprintf("  ✓ %s\n", cleanFormValue(agent)))
		}
	} else {
		status.WriteString("  (none selected)\n")
	}
	status.WriteString("\n")
	
	// Hooks
	status.WriteString("🪝 Hooks:\n")
	if len(m.config.Hooks) > 0 {
		for _, hook := range m.config.Hooks {
			status.WriteString(fmt.Sprintf("  ✓ %s\n", cleanFormValue(hook)))
		}
	} else {
		status.WriteString("  (none selected)\n")
	}
	if m.config.WantSlashCmd {
		status.WriteString("  ✓ Example slash command\n")
	}
	status.WriteString("\n")
	
	// MCP
	status.WriteString("🔌 MCP Integration:\n")
	if len(m.config.MCPServers) > 0 {
		for _, server := range m.config.MCPServers {
			status.WriteString(fmt.Sprintf("  ✓ %s\n", cleanFormValue(server)))
		}
	} else {
		status.WriteString("  (none selected)\n")
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

func main() {
	cfg := Config{
		IsProjectLocal: true,  // Default to project-specific
		Languages:      []string{"Go"},
		Subagents:      []string{"code-reviewer", "test-runner", "bug-sleuth"},
		Hooks:          []string{"pre-write-guard", "post-write-lint", "session-start", "prompt-lint"},
		WantSlashCmd:   true,
		MCPServers:     []string{"notion", "linear", "sentry", "github"},
	}

	form := huh.NewForm(
		// Page 1: Project Setup
		huh.NewGroup(
			huh.NewNote().Title("📁 Project Setup").Description("Configure your project basics and language support"),
			huh.NewConfirm().
				Title("Project-specific configuration?").
				Description("Yes = Configure for this project only\nNo = Global configuration in your home directory").
				Value(&cfg.IsProjectLocal),
			huh.NewInput().
				Title("Project name").
				Description("Used in generated documentation and configurations").
				Placeholder("awesome-app").
				Value(&cfg.ProjectName),
			huh.NewMultiSelect[string]().
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
				Title("Select subagents to include").
				Description("Choose the AI specialists you want available for your project").
				Options(huh.NewOptions(
					"🔍 code-reviewer", "🧪 test-runner", "🕵️ bug-sleuth", "🔒 security-auditor",
					"⚡ perf-optimizer", "📚 docs-writer", "🚀 release-manager", "📊 data-scientist")...).
				Value(&cfg.Subagents),
		),
		
		// Page 3: Hook Configuration
		huh.NewGroup(
			huh.NewNote().Title("🪝 Hook Setup").Description("Configure automation and lifecycle scripts"),
			huh.NewMultiSelect[string]().
				Title("Select hooks to enable").
				Description("Automation scripts that run at specific points in your workflow").
				Options(huh.NewOptions("🛡️ pre-write-guard", "🔧 post-write-lint", "🚀 session-start", "✏️ prompt-lint")...).
				Value(&cfg.Hooks),
			huh.NewConfirm().
				Title("Add example slash command?").
				Description("Creates /project:fix-github-issue as a template for custom commands").
				Value(&cfg.WantSlashCmd),
			huh.NewText().
				Title("Extra CLAUDE.md content (optional)").
				Description("Project-specific instructions to include in CLAUDE.md").
				Value(&cfg.ClaudeMDExtras),
		),
		
		// Page 4: MCP Configuration
		huh.NewGroup(
			huh.NewNote().Title("🔌 MCP Integration").Description("Connect to external tools and services via Model Context Protocol"),
			huh.NewMultiSelect[string]().
				Title("Select MCP servers to include").
				Description("Choose external tool integrations to enhance Claude's capabilities (optional)").
				Options(huh.NewOptions("📝 notion", "📋 linear", "🐛 sentry", "🐙 github", "📊 airtable")...).
				Value(&cfg.MCPServers),
		),
	)

	// Create Bubble Tea model with form
	m := model{
		form:   form,
		config: &cfg,
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
	if err := run(cfg); err != nil {
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

func run(cfg Config) error {
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
