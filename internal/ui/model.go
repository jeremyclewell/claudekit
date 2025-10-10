package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"claudekit/internal/gradient"
	"claudekit/internal/modules"
)

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return m.Form.Init()
}

// getCurrentDescription returns the description for the currently focused field.
func (m *Model) getCurrentDescription() string {
	// Get current focus from form state
	if m.Form.State == huh.StateCompleted {
		return "✅ Configuration complete! Ready to generate your Claude Code setup."
	}

	// Get the currently focused field
	focusedField := m.Form.GetFocusedField()
	if focusedField == nil {
		return m.getDefaultDescription()
	}

	// Check field key to identify what type of selection we're in
	fieldKey := focusedField.GetKey()

	// Handle language selection
	if fieldKey == "languages" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				if desc, exists := LanguageDescriptions[hoveredItem]; exists {
					return desc
				}
			}
		}
		return "💻 Select programming languages used in your project. Claude will provide specialized assistance and optimized configurations for each language. Navigate with arrow keys to see how Claude can help."
	}

	// Get registry pointer from interface
	registry, ok := m.Registry.(*modules.ModuleRegistry)
	if !ok {
		return m.getDefaultDescription()
	}

	// Handle subagent selection (Feature 004: use registry)
	if fieldKey == "subagents" {
		if multiSelect, ok := focusedField.(*huh.MultiSelect[string]); ok {
			if hoveredItem, hasHovered := multiSelect.Hovered(); hasHovered {
				// Extract the subagent name (remove emoji prefix)
				subagentName := extractSubagentName(hoveredItem)
				if module := registry.Get(modules.ComponentTypeSubagent, subagentName); module != nil {
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
				if module := registry.Get(modules.ComponentTypeHook, hookName); module != nil {
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
				if module := registry.Get(modules.ComponentTypeSlashCommand, commandName); module != nil {
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
				if module := registry.Get(modules.ComponentTypeMCP, serverName); module != nil {
					return module.Description
				}
			}
		}
		return "🔌 Select external tool integrations to enhance Claude's capabilities via Model Context Protocol. Navigate with arrow keys to see detailed descriptions."
	}

	return m.getDefaultDescription()
}

// getDefaultDescription returns the default description when no field is focused.
func (m *Model) getDefaultDescription() string {
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

// renderMarkdown renders markdown content using the glamour renderer.
func (m *Model) renderMarkdown(content string) string {
	if m.GlamourRenderer == nil {
		return content // Fallback to plain text
	}

	rendered, err := m.GlamourRenderer.Render(content)
	if err != nil {
		return content // Fallback to plain text on error
	}

	return rendered
}

// startTransition initiates a gradient theme transition (T033).
func (m *Model) startTransition(to gradient.Theme, duration time.Duration) tea.Cmd {
	m.Transition = gradient.TransitionState{
		Active:     true,
		FromTheme:  m.CurrentTheme,
		ToTheme:    to,
		StartTime:  time.Now(),
		Duration:   duration,
		EasingFunc: gradient.EaseInOutCubic,
	}

	// Return initial tick command to start animation
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
