package formatting

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// NormalizeWhitespace removes trailing spaces and normalizes line endings.
func NormalizeWhitespace(content []byte) []byte {
	lines := bytes.Split(content, []byte("\n"))

	for i, line := range lines {
		// Remove trailing whitespace from each line
		lines[i] = bytes.TrimRight(line, " \t\r")
	}

	// Join with LF line endings
	result := bytes.Join(lines, []byte("\n"))

	// Ensure file ends with single newline
	result = bytes.TrimRight(result, "\n")
	result = append(result, '\n')

	return result
}

// AtomicWriteFile writes content to a file atomically using temp file + rename.
func AtomicWriteFile(path string, content []byte) error {
	// Create temp file in same directory as target
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".mdformat-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure temp file is removed on error
	defer func() {
		if tmpFile != nil {
			os.Remove(tmpPath)
		}
	}()

	// Write content
	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Close before rename
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Success - don't delete temp file
	tmpFile = nil
	return nil
}
