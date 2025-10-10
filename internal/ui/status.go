package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"

	"claudekit/internal/util"
)

// renderStatus renders the status panel content based on form state.
func (m *Model) renderStatus() string {
	// If on the confirmation page, show configuration summary
	if m.Form.State == huh.StateCompleted || isOnConfirmationPage(m.Form) {
		return m.renderConfigurationSummary()
	}

	// Otherwise, show the current description
	return m.getCurrentDescription()
}

// isOnConfirmationPage checks if we're on the final confirmation page.
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

// renderConfigurationSummary renders a markdown summary of the user's configuration choices.
func (m *Model) renderConfigurationSummary() string {
	var status strings.Builder

	// Get config pointer from interface
	config, ok := m.Config.(*Config)
	if !ok {
		return "## Error: Invalid configuration type"
	}

	status.WriteString("## 📋 Configuration Summary\n\n")
	status.WriteString("\n\n-----\n\n")

	// Show configuration path based on project-local setting
	if config.IsProjectLocal {
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
	if len(config.Languages) > 0 {
		for _, lang := range config.Languages {
			status.WriteString(fmt.Sprintf("* %s\n", lang))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// Subagents
	status.WriteString("### 🤖 Subagents\n")
	if len(config.Subagents) > 0 {
		for _, agent := range config.Subagents {
			status.WriteString(fmt.Sprintf("* %s\n", util.CleanFormValue(agent)))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// Hooks
	status.WriteString("### 🪝 Hooks\n")
	if len(config.Hooks) > 0 {
		for _, hook := range config.Hooks {
			status.WriteString(fmt.Sprintf("* %s\n", util.CleanFormValue(hook)))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// Slash Commands
	status.WriteString("### 📟 Slash Commands\n")
	if len(config.SlashCommands) > 0 {
		for _, cmd := range config.SlashCommands {
			cleanCmd := util.CleanFormValue(cmd)
			status.WriteString(fmt.Sprintf("* /%s\n", cleanCmd))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}
	status.WriteString("\n")

	// MCP
	status.WriteString("### 🔌 MCP Integration\n")
	if len(config.MCPServers) > 0 {
		for _, server := range config.MCPServers {
			status.WriteString(fmt.Sprintf("* %s\n", util.CleanFormValue(server)))
		}
	} else {
		status.WriteString("* (none selected)\n")
	}

	return status.String()
}
