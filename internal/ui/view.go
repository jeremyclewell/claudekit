package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"claudekit/internal/gradient"
)

// Styles for the UI components.
var (
	formStyle = lipgloss.NewStyle().
		Padding(1)

	statusStyle = lipgloss.NewStyle().
		Padding(1).
		MarginLeft(2)
)

// asciiTitle is the ASCII art logo for the application.
const asciiTitle = `┏━╸╻  ┏━┓╻ ╻╺┳┓┏━╸   ╻┏ ╻╺┳╸
┃  ┃  ┣━┫┃ ┃ ┃┃┣╸    ┣┻┓┃ ┃
┗━╸┗━╸╹ ╹┗━┛╺┻┛┗━╸   ╹ ╹╹ ╹ `

// Version is the application version.
const Version = "0.0.1"

// View renders the TUI application view.
func (m Model) View() string {
	if !m.Ready {
		return "Initializing..."
	}

	// Account for border + padding
	// Border adds 2 chars left/right (1 for border char, 1 for automatic border spacing)
	// Padding adds 2 chars left/right (via Padding(1, 2))
	// Total per side: 2 + 2 = 4, so 8 total width
	const borderPadding = 10 // Extra space for border + padding on left/right
	const borderHeight = 4   // Border (2 lines: top + bottom) + Padding (2 lines: top + bottom via Padding(1, 2))

	innerWidth := m.Width - borderPadding
	innerHeight := m.Height - borderHeight

	if innerWidth < 20 {
		innerWidth = 20
	}
	if innerHeight < 10 {
		innerHeight = 10
	}

	// Calculate dimensions with fixed percentages for stability
	formWidth := int(float64(innerWidth) * 0.6)
	statusWidth := innerWidth - formWidth - 6

	// Reserve space for title (3 lines ASCII + 1 line gradient border + 1 line spacing)
	titleHeight := 5
	availableHeight := innerHeight - titleHeight
	if availableHeight < 20 {
		availableHeight = 20
		titleHeight = innerHeight - 20 // Reduce title space if needed
	}

	formHeight := availableHeight
	statusHeight := availableHeight // Status panel should match form height

	// Title with gradient (T035)
	// T015: Width-based conditional rendering for ASCII art title
	headerTheme := m.StyleMap[gradient.HeaderComponent][gradient.NormalState].Theme
	var title string

	// Version string with subtle styling
	versionText := "v" + Version
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Faint(true)
	version := versionStyle.Render(versionText)

	if innerWidth >= 60 {
		// Wide terminal: render ASCII art with gradient foreground + version
		gradientASCII := gradient.RenderASCIITitle(asciiTitle, headerTheme, m.TerminalCap)

		// Split ASCII art into lines and add version to the first line
		asciiLines := strings.Split(gradientASCII, "\n")
		if len(asciiLines) > 0 {
			// Calculate padding to right-align version
			firstLineWidth := lipgloss.Width(asciiLines[0])
			padding := innerWidth - firstLineWidth - lipgloss.Width(version)
			if padding < 0 {
				padding = 0
			}
			asciiLines[0] = asciiLines[0] + strings.Repeat(" ", padding) + version
		}
		title = strings.Join(asciiLines, "\n")
	} else {
		// Narrow terminal: fallback to regular gradient text with version
		titleText := "🛠️  ClaudeKit"
		gradientTitle := gradient.RenderGradient(titleText, headerTheme, m.TerminalCap, true)

		// Add version on same line with padding
		titleWidth := lipgloss.Width(gradientTitle)
		padding := innerWidth - titleWidth - lipgloss.Width(version)
		if padding < 0 {
			padding = 0
		}
		title = gradientTitle + strings.Repeat(" ", padding) + version
	}

	// Create gradient top border with "/" characters
	borderWidth := innerWidth
	borderText := strings.Repeat("/", borderWidth)
	gradientBorder := gradient.RenderGradient(borderText, headerTheme, m.TerminalCap, true)

	// Feature 007: Adaptive right panel based on terminal size
	var content string

	if m.ShowRightPanel {
		// Update viewport height to match available content height
		m.Viewport.Height = statusHeight
		m.Viewport.Width = statusWidth

		// Large terminal: show form + right panel
		formContent := m.Form.View()
		leftContent := formStyle.
			Width(formWidth).
			Height(formHeight).
			Render(formContent)

		// Regenerate right panel content (FR-008: always fresh)
		m.Viewport.SetContent(m.renderMarkdown(m.renderStatus()))

		// Status panel (right side, fixed height to match form)
		statusPanel := statusStyle.
			Width(statusWidth).
			Height(statusHeight). // Use consistent height
			Render(m.Viewport.View())

		// Main content (left content + status)
		// Ensure exact height by padding if necessary
		leftContent = ensureExactHeight(leftContent, formHeight)
		statusPanel = ensureExactHeight(statusPanel, statusHeight)

		content = lipgloss.JoinHorizontal(lipgloss.Top, leftContent, statusPanel)
	} else {
		// Small terminal: full-width form only (FR-006)
		formContent := m.Form.View()
		leftContent := formStyle.
			Width(innerWidth - 4). // Full width minus padding
			Height(formHeight).
			Render(formContent)

		// Ensure exact height
		leftContent = ensureExactHeight(leftContent, formHeight)
		content = leftContent
	}

	// Combine title, border, and content
	app := lipgloss.JoinVertical(lipgloss.Left, title, gradientBorder, content)

	// Add border around entire application with gradient start color and padding
	borderColor := headerTheme.StartColor
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2) // 1 line top/bottom, 2 chars left/right padding

	// Render the app with border
	rendered := borderStyle.Render(app)

	// Enforce exact terminal dimensions to prevent height overflow
	// Truncate content to fit within terminal bounds
	lines := strings.Split(rendered, "\n")
	if len(lines) > m.Height {
		lines = lines[:m.Height]
	}

	// Ensure each line doesn't exceed width
	for i, line := range lines {
		if lipgloss.Width(line) > m.Width {
			// Truncate line to fit width (accounting for ANSI codes)
			lines[i] = truncateLine(line, m.Width)
		}
	}

	return strings.Join(lines, "\n")
}
