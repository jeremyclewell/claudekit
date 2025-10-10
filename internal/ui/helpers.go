package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// truncateLine truncates a line to the specified width, preserving ANSI codes.
func truncateLine(line string, width int) string {
	// Use lipgloss's Truncate which handles ANSI codes properly
	return lipgloss.NewStyle().Width(width).Render(line)
}

// ensureExactHeight pads or truncates content to be exactly the specified height.
func ensureExactHeight(content string, targetHeight int) string {
	lines := strings.Split(content, "\n")
	currentHeight := len(lines)

	if currentHeight == targetHeight {
		return content
	}

	if currentHeight > targetHeight {
		// Truncate to target height
		return strings.Join(lines[:targetHeight], "\n")
	}

	// Pad with empty lines to reach target height
	padding := make([]string, targetHeight-currentHeight)
	for i := range padding {
		padding[i] = ""
	}
	return content + "\n" + strings.Join(padding, "\n")
}

// extractSubagentName removes emoji and space prefix (e.g., "🔍 code-reviewer" -> "code-reviewer").
func extractSubagentName(displayName string) string {
	parts := strings.SplitN(displayName, " ", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return displayName
}

// shouldShowRightPanel determines if the right panel should be shown based on terminal size.
// Per FR-002/FR-003: Requires BOTH width >= 140 AND height >= 40 (inclusive thresholds).
func shouldShowRightPanel(width, height int) bool {
	return width >= MinWidthForPanel && height >= MinHeightForPanel
}
