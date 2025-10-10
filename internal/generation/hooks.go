package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateHookAssetFile creates a hook script (shell or Python).
func GenerateHookAssetFile(desc AssetFileDescriptor, outputPath string) GenerationResult {
	// Determine language from file extension
	isPython := strings.HasSuffix(outputPath, ".py")
	isTemplate := strings.HasSuffix(outputPath, ".tmpl")

	var content string
	if isPython {
		content = fmt.Sprintf(`#!/usr/bin/env python3
"""
Hook: %s
Runs at specific lifecycle point in Claude Code
"""

import os
import sys

def main():
    # Environment variables from Claude Code:
    # - CLAUDE_TOOL_NAME = os.getenv('CLAUDE_TOOL_NAME')
    # - CLAUDE_TOOL_ARGS = os.getenv('CLAUDE_TOOL_ARGS')
    # - CLAUDE_PROJECT_DIR = os.getenv('CLAUDE_PROJECT_DIR')

    # TODO: Implement %s hook logic
    return 0

if __name__ == '__main__':
    sys.exit(main())
`, desc.Name, desc.Name)
	} else {
		// Shell script
		content = fmt.Sprintf(`#!/bin/bash
# Hook: %s
# Runs at specific lifecycle point in Claude Code

set -e

# Environment variables from Claude Code:
# - CLAUDE_TOOL_NAME
# - CLAUDE_TOOL_ARGS
# - CLAUDE_PROJECT_DIR

# TODO: Implement %s hook logic
exit 0
`, desc.Name, desc.Name)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return GenerationResult{
			FilePath: outputPath,
			Status:   StatusFailed,
			Error:    fmt.Errorf("failed to create directory: %w", err),
		}
	}

	// Write file with executable permissions (except for .tmpl files)
	perm := os.FileMode(0644)
	if !isTemplate {
		perm = 0755
	}

	if err := os.WriteFile(outputPath, []byte(content), perm); err != nil {
		return GenerationResult{
			FilePath: outputPath,
			Status:   StatusFailed,
			Error:    fmt.Errorf("failed to write file: %w", err),
		}
	}

	// Placeholders have TODO markers
	isPlaceholder := strings.Contains(content, "TODO")

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

// GeneratePlaceholderHook creates a placeholder hook script.
func GeneratePlaceholderHook(name string, outputPath string, language string) GenerationResult {
	desc := AssetFileDescriptor{Name: name, Type: AssetTypeHook}
	return GenerateHookAssetFile(desc, outputPath)
}
