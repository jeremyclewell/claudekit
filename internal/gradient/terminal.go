package gradient

import (
	"os"
	"strings"
)

// detectTerminalCapability detects the terminal's color support level.
func DetectTerminalCapability() TerminalCapability {
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return Truecolor
	}

	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") {
		return Color256
	}

	return Color8 // Conservative fallback
}

// QuantizeStops reduces gradient stops for limited terminals.
func QuantizeStops(capability TerminalCapability, desiredStops int) int {
	switch capability {
	case Color8:
		return 3 // Minimal gradient with 8 colors
	case Color256:
		return 10 // Moderate gradient
	case Truecolor:
		return desiredStops // Full fidelity
	default:
		return 3
	}
}
