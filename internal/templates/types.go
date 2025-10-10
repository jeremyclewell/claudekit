package templates

// settings follows Anthropic's hooks schema.
type Settings struct {
	Permissions *struct {
		Allow []string `json:"allow,omitempty"`
		Ask   []string `json:"ask,omitempty"`
		Deny  []string `json:"deny,omitempty"`
	} `json:"permissions,omitempty"`
	Hooks map[string][]HookMatcher `json:"hooks,omitempty"`
	Env   map[string]string        `json:"env,omitempty"`
}

// HookCmd represents a single hook command.
type HookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// HookMatcher represents a hook matcher pattern.
type HookMatcher struct {
	Matcher string    `json:"matcher,omitempty"`
	Hooks   []HookCmd `json:"hooks"`
}
