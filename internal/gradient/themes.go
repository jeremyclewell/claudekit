package gradient

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Palettes holds color schemes for light/dark terminal themes.
type paletteType struct {
	primary    lipgloss.AdaptiveColor
	secondary  lipgloss.AdaptiveColor
	accent     lipgloss.AdaptiveColor
	error      lipgloss.AdaptiveColor
	success    lipgloss.AdaptiveColor
	background lipgloss.AdaptiveColor

	// Markdown-specific theme colors
	markdownHeading  lipgloss.AdaptiveColor
	markdownCode     lipgloss.AdaptiveColor
	markdownEmphasis lipgloss.AdaptiveColor
	markdownLink     lipgloss.AdaptiveColor
}

// InitGradientPalettes initializes color palettes.
func InitGradientPalettes() paletteType {
	return paletteType{
		primary: lipgloss.AdaptiveColor{
			Light: "#6C5CE7", // Vibrant purple for light backgrounds
			Dark:  "#FF00FF", // Bright magenta for dark backgrounds
		},
		secondary: lipgloss.AdaptiveColor{
			Light: "#0984E3", // Deep blue for light
			Dark:  "#00FFFF", // Bright cyan for dark
		},
		accent: lipgloss.AdaptiveColor{
			Light: "#00B894", // Teal for light
			Dark:  "#55EFC4", // Bright teal for dark
		},
		error: lipgloss.AdaptiveColor{
			Light: "#D63031", // Deep red for light
			Dark:  "#FF7675", // Soft red for dark
		},
		success: lipgloss.AdaptiveColor{
			Light: "#00B894", // Green for light
			Dark:  "#55EFC4", // Bright green for dark
		},
		background: lipgloss.AdaptiveColor{
			Light: "#ECEFF1", // Light gray
			Dark:  "#263238", // Dark gray
		},
	}
}

// InitStyleMap populates component/state style mappings.
func InitStyleMap() map[ComponentType]map[VisualState]ComponentStyle {
	palettes := InitGradientPalettes()

	styleMap := make(map[ComponentType]map[VisualState]ComponentStyle)

	// Define default themes for each component
	components := []ComponentType{
		HeaderComponent, FormFieldComponent, ButtonComponent, DividerComponent,
		StatusComponent, ErrorComponent, SuccessComponent, BackgroundComponent,
	}

	states := []VisualState{
		NormalState, FocusedState, ActiveState, DisabledState, ErrorState, SuccessState,
	}

	for _, comp := range components {
		styleMap[comp] = make(map[VisualState]ComponentStyle)

		for _, state := range states {
			// Select theme based on component and state
			var theme Theme

			switch {
			case comp == ErrorComponent || state == ErrorState:
				theme = Theme{
					Name:       "error",
					StartColor: palettes.error,
					EndColor:   lipgloss.AdaptiveColor{Light: "#FF6B6B", Dark: "#FF7675"},
					Stops:      15,
					Direction:  Horizontal,
					Intensity:  0.9,
				}
			case comp == SuccessComponent || state == SuccessState:
				theme = Theme{
					Name:       "success",
					StartColor: palettes.success,
					EndColor:   lipgloss.AdaptiveColor{Light: "#00D2A0", Dark: "#7FFFD4"},
					Stops:      15,
					Direction:  Horizontal,
					Intensity:  0.85,
				}
			case state == FocusedState:
				theme = Theme{
					Name:       "focused",
					StartColor: palettes.primary,
					EndColor:   palettes.secondary,
					Stops:      20,
					Direction:  Horizontal,
					Intensity:  0.95,
				}
			default:
				theme = Theme{
					Name:       "normal",
					StartColor: palettes.primary,
					EndColor:   palettes.secondary,
					Stops:      15,
					Direction:  Horizontal,
					Intensity:  0.7,
				}
			}

			styleMap[comp][state] = ComponentStyle{
				Component:         comp,
				State:             state,
				Theme:             theme,
				AnimationDuration: 200 * time.Millisecond,
			}
		}
	}

	return styleMap
}

// GetPalettes returns the initialized gradient palettes for use in other packages.
func GetPalettes() paletteType {
	return InitGradientPalettes()
}
