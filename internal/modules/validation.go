package modules

import (
	"embed"
	"fmt"
	"strings"
)

// Validate checks that a module's referenced asset files exist in the embedded filesystem.
func Validate(module *ComponentModule, fs embed.FS) error {
	var errs []string

	// Check required fields (Feature 008: only name and type are required)
	if module.Name == "" {
		errs = append(errs, "name is required")
	}
	if module.Type == "" {
		errs = append(errs, "type is required")
	}

	// Description and AssetPaths are optional (e.g., MCPs don't need asset_paths)

	// Validate asset paths exist (if provided)
	for _, assetPath := range module.AssetPaths {
		fullPath := "assets/" + assetPath
		if _, err := fs.ReadFile(fullPath); err != nil {
			errs = append(errs, fmt.Sprintf("asset file %q not found", assetPath))
		}
	}

	// Type-specific validation
	switch ComponentType(module.Type) {
	case ComponentTypeSubagent:
		// Subagents should have a category
		if module.Category == "" {
			errs = append(errs, "subagents should have a category")
		}
		// Subagents should have asset_paths (template files)
		if len(module.AssetPaths) == 0 {
			errs = append(errs, "subagents should have at least one asset_path")
		}
	case ComponentTypeHook:
		// Hooks should have asset_paths (script templates)
		if len(module.AssetPaths) == 0 {
			errs = append(errs, "hooks should have at least one asset_path")
		}
	case ComponentTypeSlashCommand:
		// Slash commands should have asset_paths (script templates)
		if len(module.AssetPaths) == 0 {
			errs = append(errs, "slash commands should have at least one asset_path")
		}
	case ComponentTypeMCP:
		// MCPs don't need asset_paths (they're configured via .mcp.json)
		// No additional validation needed
	default:
		errs = append(errs, fmt.Sprintf("unknown module type: %q", module.Type))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}
