package modules

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFromMarkdown loads all module definitions from markdown files in the embedded filesystem.
func LoadFromMarkdown(fsys embed.FS) ([]ModuleDefinition, error) {
	var modules []ModuleDefinition

	// Walk the assets/modules directory
	err := fs.WalkDir(fsys, "assets/modules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Read the file
		content, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Parse the module
		module, err := ParseMarkdown(path, content)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		modules = append(modules, module)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil
}

// ParseMarkdown parses a markdown file with YAML frontmatter into a ModuleDefinition.
func ParseMarkdown(path string, content []byte) (ModuleDefinition, error) {
	frontmatter, body, err := extractFrontmatter(string(content))
	if err != nil {
		return ModuleDefinition{}, fmt.Errorf("failed to extract frontmatter from %s: %w", path, err)
	}

	var module ComponentModule
	if err := yaml.Unmarshal([]byte(frontmatter), &module); err != nil {
		return ModuleDefinition{}, fmt.Errorf("failed to parse YAML frontmatter in %s: %w", path, err)
	}

	return ModuleDefinition{
		Path:        path,
		Frontmatter: module,
		Body:        body,
	}, nil
}

// extractFrontmatter separates YAML frontmatter from markdown body.
// Returns frontmatter (YAML between --- delimiters) and body (everything after).
func extractFrontmatter(content string) (frontmatter, body string, err error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || lines[0] != "---" {
		return "", "", fmt.Errorf("missing frontmatter opening delimiter")
	}

	// Find closing delimiter
	closeIdx := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closeIdx = i
			break
		}
	}

	if closeIdx == -1 {
		return "", "", fmt.Errorf("missing frontmatter closing delimiter")
	}

	frontmatter = strings.Join(lines[1:closeIdx], "\n")
	body = strings.Join(lines[closeIdx+1:], "\n")

	return frontmatter, body, nil
}
