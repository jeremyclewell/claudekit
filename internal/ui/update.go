package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/huh"

	"claudekit/internal/gradient"
)

// handleWindowSizeMsg processes terminal resize events with debouncing.
// Cancels any existing timer, caches the new dimensions, and starts 200ms countdown.
func handleWindowSizeMsg(m Model, msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	// Cancel existing debounce timer if present
	if m.ResizeDebouncer != nil {
		m.ResizeDebouncer.Stop()
	}

	// Cache resize message
	m.PendingResize = &msg

	// Start new debounce timer
	m.ResizeDebouncer = time.NewTimer(ResizeDebounceMS * time.Millisecond)

	// Return Cmd that waits for timer to expire
	return m, func() tea.Msg {
		<-m.ResizeDebouncer.C
		return DebounceCompleteMsg{}
	}
}

// applyPendingResize updates model dimensions and recomputes panel visibility.
// CRITICAL: MUST NOT modify m.Form or m.Config (FR-011: preserve all user input).
func applyPendingResize(m Model) (Model, tea.Cmd) {
	if m.PendingResize == nil {
		return m, nil // No pending resize, nothing to do
	}

	// Update dimensions
	m.Width = m.PendingResize.Width
	m.Height = m.PendingResize.Height

	// Recompute panel visibility (FR-002, FR-003)
	m.ShowRightPanel = shouldShowRightPanel(m.Width, m.Height)

	// Clear debounce state
	m.PendingResize = nil
	m.ResizeDebouncer = nil

	// IMPORTANT: Do NOT modify m.Form or m.Config - preserves input per FR-011

	return m, nil
}

// Update is the main Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Feature 007: Debounced resize handling
		return handleWindowSizeMsg(m, msg)

	case DebounceCompleteMsg:
		// Feature 007: Apply pending resize after debounce period
		m, cmd := applyPendingResize(m)

		// After applying resize, update viewport dimensions if panel is visible
		if m.ShowRightPanel {
			// Calculate layout dimensions with fixed percentages for stability
			formWidth := int(float64(m.Width) * 0.6)        // 60% width for left side
			statusWidth := m.Width - formWidth - 6          // Remaining width for right side

			// Calculate height consistently with View() function
			const borderPadding = 10
			const borderHeight = 4
			innerHeight := m.Height - borderHeight
			if innerHeight < 10 {
				innerHeight = 10
			}
			titleHeight := 5
			availableHeight := innerHeight - titleHeight
			if availableHeight < 20 {
				availableHeight = 20
			}
			statusHeight := availableHeight

			if !m.Ready {
				m.Viewport = viewport.New(statusWidth, statusHeight)
				m.Ready = true
			} else {
				m.Viewport.Width = statusWidth
				m.Viewport.Height = statusHeight
			}
		}

		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	// T032: Handle gradient animation ticks
	case TickMsg:
		if m.Transition.Active {
			progress := m.Transition.Progress()
			if progress >= 1.0 {
				// Transition complete
				m.Transition.Active = false
				m.CurrentTheme = m.Transition.ToTheme
			} else {
				// Continue animating
				m.CurrentTheme = gradient.InterpolateGradient(
					m.Transition.FromTheme,
					m.Transition.ToTheme,
					progress,
				)
				// Schedule next tick for smooth animation
				return m, tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
					return TickMsg(t)
				})
			}
		}
	}

	// Update form
	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
	}

	// Handle viewport scrolling for status panel
	var viewportCmd tea.Cmd
	m.Viewport, viewportCmd = m.Viewport.Update(msg)
	cmd = tea.Batch(cmd, viewportCmd)

	// Update viewport content with current status/descriptions
	m.Viewport.SetContent(m.renderMarkdown(m.renderStatus()))

	// Check if form is complete
	if m.Form.State == huh.StateCompleted {
		return m, tea.Quit
	}

	return m, cmd
}
