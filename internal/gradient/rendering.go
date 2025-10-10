package gradient

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ApplyGradient creates a Lipgloss style with gradient.
func ApplyGradient(theme Theme, capability TerminalCapability) lipgloss.Style {
	stops := QuantizeStops(capability, theme.Stops)

	// Create base style with first color
	// Note: Full gradient rendering happens in RenderGradient()
	// This creates a style reference for the theme
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.StartColor.Dark))

	// Adjust intensity (not fully implemented for brevity, would affect alpha/saturation)
	_ = stops

	return style
}

// RenderGradient renders text with gradient colors applied.
func RenderGradient(text string, theme Theme, capability TerminalCapability, foreground bool) string {
	if text == "" {
		return ""
	}

	stops := QuantizeStops(capability, theme.Stops)
	if stops < 2 {
		stops = 2
	}

	// Split text into segments
	runes := []rune(text)
	segmentSize := len(runes) / stops
	if segmentSize < 1 {
		segmentSize = 1
	}

	var result strings.Builder

	for i := 0; i < len(runes); i += segmentSize {
		end := i + segmentSize
		if end > len(runes) {
			end = len(runes)
		}

		segment := string(runes[i:end])
		progress := float64(i) / float64(len(runes))

		// Interpolate color for this segment
		color := InterpolateColor(
			lipgloss.Color(theme.StartColor.Dark),
			lipgloss.Color(theme.EndColor.Dark),
			progress,
		)

		// Apply color and render
		var styled string
		if foreground {
			styled = lipgloss.NewStyle().Foreground(color).Render(segment)
		} else {
			styled = lipgloss.NewStyle().Background(color).Render(segment)
		}
		result.WriteString(styled)
	}

	return result.String()
}

// RenderASCIITitle applies gradient to ASCII art line-by-line.
func RenderASCIITitle(asciiArt string, theme Theme, capability TerminalCapability) string {
	lines := strings.Split(asciiArt, "\n")
	var result strings.Builder

	for _, line := range lines {
		// Apply horizontal gradient to each line independently with foreground coloring
		gradientLine := RenderGradient(line, theme, capability, true)
		result.WriteString(gradientLine)
		result.WriteString("\n")
	}

	// Remove trailing newline
	output := result.String()
	if len(output) > 0 && output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}

	return output
}
