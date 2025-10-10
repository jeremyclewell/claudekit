package templates

import (
	"embed"
	"fmt"
	"strings"

	"jeremyclewell.com/claudekit/internal/modules"
)

// SampleSlashCommand returns the sample slash command template from assets
func SampleSlashCommand(assetsFS embed.FS) string {
	content, err := assetsFS.ReadFile("assets/templates/fix-github-issue.md")
	if err != nil {
		panic(err)
	}
	return string(content)
}

// GenerateSlashCommand creates a slash command markdown file based on the command name
func GenerateSlashCommand(cmdName string, registry *modules.ModuleRegistry) string {
	// Generate custom slash command content based on the command name (Feature 004: use registry)
	module := registry.Get(modules.ComponentTypeSlashCommand, cmdName)
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
