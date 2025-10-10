package config

// Config holds the user's configuration choices from the interactive form.
type Config struct {
	ProjectName    string
	ProjectDir     string
	Language       []string
	Subagents      []string
	Hooks          []string
	SlashCommands  []string
	MCPs           []string
	IncludeCLAUDE  bool
	IncludeAgents  bool
	IncludeHooks   bool
	IncludeExamples bool
}

// PersistenceConfig stores configuration state between sessions.
// This allows the form to remember previous selections when re-run in the same project.
type PersistenceConfig struct {
	ProjectName    string   `json:"project_name"`
	Language       []string `json:"language"`
	Subagents      []string `json:"subagents"`
	Hooks          []string `json:"hooks"`
	SlashCommands  []string `json:"slash_commands"`
	MCPs           []string `json:"mcps"`
	IncludeCLAUDE  bool     `json:"include_claude"`
	IncludeAgents  bool     `json:"include_agents"`
	IncludeHooks   bool     `json:"include_hooks"`
	IncludeExamples bool    `json:"include_examples"`
}
