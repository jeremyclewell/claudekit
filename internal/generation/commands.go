package generation

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateSlashCommandAssetFile creates a slash command markdown template.
func GenerateSlashCommandAssetFile(desc AssetFileDescriptor, outputPath string) GenerationResult {
	description := "Custom slash command"
	if desc.Module != nil && desc.Module.GetDescription() != "" {
		description = desc.Module.GetDescription()
	}

	content := fmt.Sprintf(`---
name: %s
description: %s
---

# %s Command

## Purpose
Define the purpose and when to use this command.

## Workflow
1. Step 1: Describe the first action
2. Step 2: Describe the second action
3. Step 3: Describe the final action

## Context Integration
Explain how this command uses project context and available tools.

## Example Usage
Provide example scenarios and expected outcomes.
`, desc.Name, description, desc.Name)

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return GenerationResult{
			FilePath: outputPath,
			Status:   StatusFailed,
			Error:    fmt.Errorf("failed to create directory: %w", err),
		}
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return GenerationResult{
			FilePath: outputPath,
			Status:   StatusFailed,
			Error:    fmt.Errorf("failed to write file: %w", err),
		}
	}

	// Mark as placeholder if using default description
	isPlaceholder := description == "Custom slash command"
	status := StatusSuccess
	if isPlaceholder {
		status = StatusPlaceholderGenerated
	}

	return GenerationResult{
		FilePath:      outputPath,
		Status:        status,
		BytesWritten:  len(content),
		IsPlaceholder: isPlaceholder,
	}
}

// GeneratePlaceholderSlashCommand creates a placeholder slash command.
func GeneratePlaceholderSlashCommand(name string, outputPath string) GenerationResult {
	desc := AssetFileDescriptor{Name: name, Type: AssetTypeSlashCommand}
	return GenerateSlashCommandAssetFile(desc, outputPath)
}
