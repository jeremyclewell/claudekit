package generation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ValidateJSONFile validates a JSON file can be parsed (used for hooks/MCPs).
func ValidateJSONFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// ValidateAgentMarkdown validates an agent markdown file has required frontmatter.
func ValidateAgentMarkdown(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("missing YAML frontmatter")
	}

	// Find end of frontmatter
	endIdx := strings.Index(content[4:], "\n---\n")
	if endIdx == -1 {
		return fmt.Errorf("invalid YAML frontmatter (missing closing ---)")
	}

	frontmatter := content[4 : 4+endIdx]

	// Check for required fields
	requiredFields := []string{"name:", "description:", "tools:"}
	for _, field := range requiredFields {
		if !strings.Contains(frontmatter, field) {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// ValidateShebang checks if a script has a valid shebang.
func ValidateShebang(path string, expected string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "#!") {
		return fmt.Errorf("missing shebang")
	}

	if lines[0] != expected {
		return fmt.Errorf("expected shebang %q, got %q", expected, lines[0])
	}

	return nil
}

// ValidateYAMLFrontmatter checks if a markdown file has valid YAML frontmatter.
func ValidateYAMLFrontmatter(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("missing YAML frontmatter")
	}

	// Find end of frontmatter
	endIdx := strings.Index(content[4:], "\n---\n")
	if endIdx == -1 {
		return fmt.Errorf("invalid YAML frontmatter (missing closing ---)")
	}

	return nil
}
