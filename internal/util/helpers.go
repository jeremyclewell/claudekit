package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// MustMkdir creates a directory or panics.
func MustMkdir(p string) {
	if err := os.MkdirAll(p, 0755); err != nil {
		panic(fmt.Sprintf("failed to create directory %s: %v", p, err))
	}
}

// WriteExecutable writes content to path with executable permissions.
func WriteExecutable(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// Contains checks if a string slice contains a string.
func Contains(ss []string, s string) bool {
	for _, item := range ss {
		if item == s {
			return true
		}
	}
	return false
}

// Includes checks if a string slice contains a string (case-insensitive).
func Includes(ss []string, s string) bool {
	for _, item := range ss {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

// Or returns a if non-empty, otherwise b.
func Or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// CommandExists checks if a command exists in PATH.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// CleanFormValue removes emoji and space prefix (e.g., "🔍 code-reviewer" -> "code-reviewer")
func CleanFormValue(value string) string {
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return value
}

// CleanFormValues cleans multiple form values.
func CleanFormValues(values []string) []string {
	cleaned := make([]string, len(values))
	for i, v := range values {
		cleaned[i] = CleanFormValue(v)
	}
	return cleaned
}
