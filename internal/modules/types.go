package modules

import (
	"embed"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
)

// ComponentType represents the type of module component.
type ComponentType string

const (
	ComponentTypeSubagent      ComponentType = "subagent"
	ComponentTypeHook          ComponentType = "hook"
	ComponentTypeSlashCommand  ComponentType = "slash_command"
	ComponentTypeMCP           ComponentType = "mcp"
)

// ComponentModule represents a single module definition (subagent, hook, slash command, or MCP).
type ComponentModule struct {
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"`
	DisplayName string            `yaml:"display_name,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	AssetPaths  []string          `yaml:"asset_paths,omitempty"`
	Defaults    map[string]any    `yaml:"defaults,omitempty"`
	Enabled     bool              `yaml:"enabled,omitempty"`
	Description string            // Extracted from markdown body
}

// ModuleDefinition represents a parsed module from a markdown file.
type ModuleDefinition struct {
	Path        string
	Frontmatter ComponentModule
	Body        string
}

// Validate checks if the module definition is valid.
func (m *ModuleDefinition) Validate() error {
	var errs []string

	// Validate required fields (Feature 008: only name and type are required)
	if m.Frontmatter.Name == "" {
		errs = append(errs, "name is required")
	}
	if m.Frontmatter.Type == "" {
		errs = append(errs, "type is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed for %s: %s", m.Path, strings.Join(errs, ", "))
	}

	return nil
}

// ModuleRegistry stores all loaded modules organized by type.
type ModuleRegistry struct {
	Subagents      map[string]*ComponentModule
	Hooks          map[string]*ComponentModule
	SlashCommands  map[string]*ComponentModule
	MCPs           map[string]*ComponentModule
}

// NewRegistry creates a new empty ModuleRegistry.
func NewRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		Subagents:     make(map[string]*ComponentModule),
		Hooks:         make(map[string]*ComponentModule),
		SlashCommands: make(map[string]*ComponentModule),
		MCPs:          make(map[string]*ComponentModule),
	}
}

// Get retrieves a module by type and name.
func (r *ModuleRegistry) Get(componentType ComponentType, name string) *ComponentModule {
	switch componentType {
	case ComponentTypeSubagent:
		return r.Subagents[name]
	case ComponentTypeHook:
		return r.Hooks[name]
	case ComponentTypeSlashCommand:
		return r.SlashCommands[name]
	case ComponentTypeMCP:
		return r.MCPs[name]
	default:
		return nil
	}
}

// List returns all modules of a given type, sorted by display name.
func (r *ModuleRegistry) List(componentType ComponentType) []*ComponentModule {
	var modules []*ComponentModule
	switch componentType {
	case ComponentTypeSubagent:
		for _, m := range r.Subagents {
			modules = append(modules, m)
		}
	case ComponentTypeHook:
		for _, m := range r.Hooks {
			modules = append(modules, m)
		}
	case ComponentTypeSlashCommand:
		for _, m := range r.SlashCommands {
			modules = append(modules, m)
		}
	case ComponentTypeMCP:
		for _, m := range r.MCPs {
			modules = append(modules, m)
		}
	}

	// Sort by display name (or name if display_name not set)
	slices.SortFunc(modules, func(a, b *ComponentModule) int {
		aName := a.DisplayName
		if aName == "" {
			aName = a.Name
		}
		bName := b.DisplayName
		if bName == "" {
			bName = b.Name
		}
		return strings.Compare(aName, bName)
	})

	return modules
}

// GetOptions returns huh.Option entries for form multi-select fields.
func (r *ModuleRegistry) GetOptions(componentType ComponentType) []huh.Option[string] {
	modules := r.List(componentType)
	options := make([]huh.Option[string], 0, len(modules))

	for _, m := range modules {
		displayName := m.DisplayName
		if displayName == "" {
			displayName = m.Name
		}

		// For subagents, append category badge if present
		if componentType == ComponentTypeSubagent && m.Category != "" {
			displayName = fmt.Sprintf("%s (%s)", displayName, m.Category)
		}

		options = append(options, huh.NewOption(displayName, m.Name))
	}

	return options
}

// Load populates the registry from an embedded filesystem.
// Returns a slice of errors encountered during loading (non-fatal).
func (r *ModuleRegistry) Load(fs embed.FS) []error {
	definitions, err := LoadFromMarkdown(fs)
	if err != nil {
		return []error{fmt.Errorf("failed to load modules: %w", err)}
	}

	var errs []error

	for _, def := range definitions {
		// Validate the module definition
		if err := def.Validate(); err != nil {
			errs = append(errs, err)
			continue
		}

		module := &def.Frontmatter
		module.Description = strings.TrimSpace(def.Body)

		// Validate the module has required assets
		if err := Validate(module, fs); err != nil {
			errs = append(errs, fmt.Errorf("validation failed for %s: %w", def.Path, err))
			continue
		}

		// Add to appropriate registry based on type
		switch ComponentType(module.Type) {
		case ComponentTypeSubagent:
			r.Subagents[module.Name] = module
		case ComponentTypeHook:
			r.Hooks[module.Name] = module
		case ComponentTypeSlashCommand:
			r.SlashCommands[module.Name] = module
		case ComponentTypeMCP:
			r.MCPs[module.Name] = module
		default:
			errs = append(errs, fmt.Errorf("unknown module type %q in %s", module.Type, def.Path))
		}
	}

	return errs
}
