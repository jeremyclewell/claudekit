package templates

import (
	"encoding/json"
)

// BuildMCPJSON generates the .mcp.json configuration for selected MCP servers
func BuildMCPJSON(selected []string) string {
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
