package gradient

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TerminalCapability represents terminal color support level.
type TerminalCapability int

const (
	Color8     TerminalCapability = iota // 8 ANSI colors
	Color256                              // 256-color palette
	Truecolor                             // 24-bit RGB
)

// Direction defines gradient orientation.
type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

// Theme defines a color palette for gradient rendering.
type Theme struct {
	Name       string
	StartColor lipgloss.AdaptiveColor
	EndColor   lipgloss.AdaptiveColor
	Stops      int
	Direction  Direction
	Intensity  float64
}

// ComponentType enumerates UI components for styling.
type ComponentType int

const (
	HeaderComponent ComponentType = iota
	FormFieldComponent
	ButtonComponent
	DividerComponent
	StatusComponent
	ErrorComponent
	SuccessComponent
	BackgroundComponent
)

// VisualState represents interaction states.
type VisualState int

const (
	NormalState VisualState = iota
	FocusedState
	ActiveState
	DisabledState
	ErrorState
	SuccessState
)

// ComponentStyle maps components/states to gradient themes.
type ComponentStyle struct {
	Component         ComponentType
	State             VisualState
	Theme             Theme
	AnimationDuration time.Duration
}

// EasingFunction defines animation easing curves.
type EasingFunction func(float64) float64

// TransitionState tracks gradient animation progress.
type TransitionState struct {
	Active     bool
	FromTheme  Theme
	ToTheme    Theme
	StartTime  time.Time
	Duration   time.Duration
	EasingFunc EasingFunction
}

// Progress returns current animation progress (0.0-1.0).
func (t *TransitionState) Progress() float64 {
	if !t.Active {
		return 1.0
	}
	elapsed := time.Since(t.StartTime)
	raw := float64(elapsed) / float64(t.Duration)
	if raw >= 1.0 {
		return 1.0
	}
	return t.EasingFunc(raw)
}

// Palettes holds color schemes for light/dark terminal themes.
type Palettes struct {
	// UI Component palettes
	Primary     lipgloss.AdaptiveColor
	Secondary   lipgloss.AdaptiveColor
	Accent      lipgloss.AdaptiveColor
	Success     lipgloss.AdaptiveColor
	Warning     lipgloss.AdaptiveColor
	Error       lipgloss.AdaptiveColor
	Muted       lipgloss.AdaptiveColor
	Background  lipgloss.AdaptiveColor

	// Markdown syntax highlighting palettes
	CodeBackground lipgloss.AdaptiveColor
	CodeForeground lipgloss.AdaptiveColor
	CommentColor   lipgloss.AdaptiveColor
	KeywordColor   lipgloss.AdaptiveColor
	StringColor    lipgloss.AdaptiveColor
	NumberColor    lipgloss.AdaptiveColor
	FunctionColor  lipgloss.AdaptiveColor
	OperatorColor  lipgloss.AdaptiveColor
}
