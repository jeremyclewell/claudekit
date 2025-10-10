package gradient

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ExtendColorPaletteForMarkdown extends palette with markdown colors.
func ExtendColorPaletteForMarkdown(palette *paletteType) {
	// Headings: blend primary and secondary (50/50) for purple-blue tone matching form headers
	headingLight := InterpolateColor(
		lipgloss.Color(palette.primary.Light),
		lipgloss.Color(palette.secondary.Light),
		0.5,
	)
	headingDark := InterpolateColor(
		lipgloss.Color(palette.primary.Dark),
		lipgloss.Color(palette.secondary.Dark),
		0.5,
	)
	palette.markdownHeading = lipgloss.AdaptiveColor{
		Light: string(headingLight),
		Dark:  string(headingDark),
	}

	// Code: will use glamour's default syntax highlighting (set to empty/nil in glamour config)
	// Using background color as placeholder to signal "use default"
	palette.markdownCode = palette.background

	// Emphasis (bold/italic): toned-down cyan for less jarring appearance
	// Reduce brightness by 20% in dark mode
	emphasisLight := palette.secondary.Light
	emphasisDark := AdjustSaturation(palette.secondary.Dark, 0.7) // Reduce saturation for softer look
	palette.markdownEmphasis = lipgloss.AdaptiveColor{
		Light: emphasisLight,
		Dark:  emphasisDark,
	}

	// Links: use secondary (cyan) same as toned-down emphasis
	palette.markdownLink = palette.markdownEmphasis
}

// GenerateGlamourStyle creates custom glamour renderer from palette.
func GenerateGlamourStyle(palette paletteType) *glamour.TermRenderer {
	// Detect background mode for adaptive colors
	isDark := termenv.HasDarkBackground()

	// Helper to select appropriate color variant
	selectColor := func(adaptive lipgloss.AdaptiveColor) string {
		if isDark {
			return adaptive.Dark
		}
		return adaptive.Light
	}

	// Helper to create string pointer for ansi config
	strPtr := func(s string) *string { return &s }
	boolPtr := func(b bool) *bool { return &b }

	// Select colors based on background
	headingColor := selectColor(palette.markdownHeading)    // Magenta for headers
	emphasisColor := selectColor(palette.markdownEmphasis)  // Cyan for bold/italic
	linkColor := selectColor(palette.markdownLink)          // Cyan for links

	// Plain text color: use white for dark bg, black for light bg
	var textColor string
	if isDark {
		textColor = "#FFFFFF" // White for dark backgrounds
	} else {
		textColor = "#000000" // Black for light backgrounds
	}

	// Construct ansi.StyleConfig
	config := ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: strPtr(textColor), // Plain text: white or black
			},
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: strPtr(headingColor), // Magenta
				Bold:  boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(headingColor), // Magenta
				Bold:   boolPtr(true),
				Prefix: "# ",
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(headingColor), // Magenta (all headers same color)
				Bold:   boolPtr(true),
				Prefix: "## ",
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(headingColor), // Magenta
				Bold:   boolPtr(true),
				Prefix: "### ",
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(headingColor), // Magenta
				Bold:   boolPtr(true),
				Prefix: "#### ",
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(headingColor), // Magenta
				Bold:   boolPtr(true),
				Prefix: "##### ",
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(headingColor), // Magenta
				Bold:   boolPtr(true),
				Prefix: "###### ",
			},
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				// No color specified - glamour will use default syntax highlighting
				Prefix: "`",
				Suffix: "`",
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					// No background color - transparent
				},
			},
			Chroma: &ansi.Chroma{
				Text:                ansi.StylePrimitive{Color: strPtr(textColor)},
				Error:               ansi.StylePrimitive{Color: strPtr("#FF0000")},
				Comment:             ansi.StylePrimitive{Color: strPtr("#6B7280")},
				CommentPreproc:      ansi.StylePrimitive{Color: strPtr("#9CA3AF")},
				Keyword:             ansi.StylePrimitive{Color: strPtr("#A78BFA")},  // Purple for keywords
				KeywordReserved:     ansi.StylePrimitive{Color: strPtr("#A78BFA")},
				KeywordNamespace:    ansi.StylePrimitive{Color: strPtr("#60A5FA")},  // Blue
				KeywordType:         ansi.StylePrimitive{Color: strPtr("#34D399")},  // Green
				Operator:            ansi.StylePrimitive{Color: strPtr("#F59E0B")},  // Orange
				Punctuation:         ansi.StylePrimitive{Color: strPtr(textColor)},
				Name:                ansi.StylePrimitive{Color: strPtr(textColor)},
				NameBuiltin:         ansi.StylePrimitive{Color: strPtr("#60A5FA")},  // Blue for builtins
				NameTag:             ansi.StylePrimitive{Color: strPtr("#A78BFA")},
				NameAttribute:       ansi.StylePrimitive{Color: strPtr("#34D399")},
				NameClass:           ansi.StylePrimitive{Color: strPtr("#FBBF24")},  // Yellow
				NameConstant:        ansi.StylePrimitive{Color: strPtr("#EC4899")},  // Pink
				NameDecorator:       ansi.StylePrimitive{Color: strPtr("#F59E0B")},
				NameException:       ansi.StylePrimitive{Color: strPtr("#EF4444")},  // Red
				NameFunction:        ansi.StylePrimitive{Color: strPtr("#60A5FA")},  // Blue
				NameOther:           ansi.StylePrimitive{Color: strPtr(textColor)},
				Literal:             ansi.StylePrimitive{Color: strPtr("#34D399")},
				LiteralNumber:       ansi.StylePrimitive{Color: strPtr("#F97316")},  // Orange
				LiteralDate:         ansi.StylePrimitive{Color: strPtr("#34D399")},
				LiteralString:       ansi.StylePrimitive{Color: strPtr("#10B981")},  // Green
				LiteralStringEscape: ansi.StylePrimitive{Color: strPtr("#F59E0B")},
				GenericDeleted:      ansi.StylePrimitive{Color: strPtr("#EF4444")},
				GenericEmph:         ansi.StylePrimitive{Color: strPtr(emphasisColor)},
				GenericInserted:     ansi.StylePrimitive{Color: strPtr("#10B981")},
				GenericStrong:       ansi.StylePrimitive{Color: strPtr(emphasisColor)},
				GenericSubheading:   ansi.StylePrimitive{Color: strPtr(headingColor)},
				Background:          ansi.StylePrimitive{},  // Transparent
			},
		},
		Emph: ansi.StylePrimitive{
			Color:  strPtr(emphasisColor), // Cyan for italic
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Color: strPtr(emphasisColor), // Cyan for bold
			Bold:  boolPtr(true),
		},
		Link: ansi.StylePrimitive{
			Color:     strPtr(linkColor), // Cyan for links
			Underline: boolPtr(true),
		},
		List: ansi.StyleList{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: strPtr(textColor), // Plain text color for lists
				},
			},
			LevelIndent: 2,
		},
		Enumeration: ansi.StylePrimitive{
			Color:       strPtr(textColor), // Plain text color for numbered lists
			BlockPrefix: "",
		},
		Item: ansi.StylePrimitive{
			Color:       strPtr(textColor), // Plain text color for list items
			BlockPrefix: "• ",
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr(emphasisColor), // Cyan for blockquotes
				Italic: boolPtr(true),
			},
		},
		Strikethrough: ansi.StylePrimitive{
			Color:      strPtr(textColor), // Plain text with strikethrough
			CrossedOut: boolPtr(true),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  strPtr(emphasisColor), // Cyan horizontal rules
			Format: "─────────────────────────────────────────────────────────",
		},
	}

	// Create renderer with custom style and syntax highlighting
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(config),
		glamour.WithWordWrap(60),
	)
	if err != nil {
		return nil // Fallback to nil, caller will handle
	}

	return renderer
}
