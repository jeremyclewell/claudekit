package templates

import (
	"embed"

	"jeremyclewell.com/claudekit/internal/config"
	"jeremyclewell.com/claudekit/internal/modules"
	"jeremyclewell.com/claudekit/internal/util"
)

// BuildSettings creates the settings.json structure with permissions, hooks, and environment
func BuildSettings(projectDir string, cfg config.Config, registry *modules.ModuleRegistry, assetsFS embed.FS) Settings {
	s := Settings{
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
		Hooks: map[string][]HookMatcher{},
	}

	// Add all selected hooks using registry (Feature 004)
	for _, hookDisplay := range cfg.Hooks {
		hookName := util.CleanFormValue(hookDisplay)

		// Get hook module from registry
		hookModule := registry.Get(modules.ComponentTypeHook, hookName)
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
			HookMatcher{
				Hooks: []HookCmd{{
					Type:    "command",
					Command: command,
					Timeout: int(timeout),
				}},
			},
		)
	}

	return s
}
