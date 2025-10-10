package gradient

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

// InterpolateColor performs RGB interpolation between two colors.
func InterpolateColor(start, end lipgloss.Color, progress float64) lipgloss.Color {
	// Parse hex colors
	startStr := string(start)
	endStr := string(end)

	// Extract RGB components
	var sr, sg, sb, er, eg, eb int
	fmt.Sscanf(startStr, "#%02x%02x%02x", &sr, &sg, &sb)
	fmt.Sscanf(endStr, "#%02x%02x%02x", &er, &eg, &eb)

	// Interpolate
	r := int(float64(sr) + float64(er-sr)*progress)
	g := int(float64(sg) + float64(eg-sg)*progress)
	b := int(float64(sb) + float64(eb-sb)*progress)

	// Return as hex color
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

// AdjustSaturation adjusts the saturation of a hex color.
func AdjustSaturation(hexColor string, factor float64) string {
	// Parse hex color
	var r, g, b int
	fmt.Sscanf(hexColor, "#%02x%02x%02x", &r, &g, &b)

	// Convert RGB to HSL
	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l := (max + min) / 2.0

	var h, s float64
	if max == min {
		h, s = 0, 0 // achromatic
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h /= 6
	}

	// Adjust saturation
	s = s * factor
	if s > 1.0 {
		s = 1.0
	}
	if s < 0.0 {
		s = 0.0
	}

	// Convert HSL back to RGB
	var r2, g2, b2 float64
	if s == 0 {
		r2, g2, b2 = l, l, l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r2 = hue2rgb(p, q, h+1.0/3.0)
		g2 = hue2rgb(p, q, h)
		b2 = hue2rgb(p, q, h-1.0/3.0)
	}

	return fmt.Sprintf("#%02X%02X%02X", int(r2*255), int(g2*255), int(b2*255))
}

// IncreaseBrightness increases the brightness of a hex color.
func IncreaseBrightness(hexColor string, factor float64) string {
	// Parse hex color
	var r, g, b int
	fmt.Sscanf(hexColor, "#%02x%02x%02x", &r, &g, &b)

	// Convert RGB to HSL
	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	l := (max + min) / 2.0

	var h, s float64
	if max == min {
		h, s = 0, 0 // achromatic
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h /= 6
	}

	// Increase lightness
	l = l * factor
	if l > 1.0 {
		l = 1.0
	}
	if l < 0.0 {
		l = 0.0
	}

	// Convert HSL back to RGB
	var r2, g2, b2 float64
	if s == 0 {
		r2, g2, b2 = l, l, l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r2 = hue2rgb(p, q, h+1.0/3.0)
		g2 = hue2rgb(p, q, h)
		b2 = hue2rgb(p, q, h-1.0/3.0)
	}

	return fmt.Sprintf("#%02X%02X%02X", int(r2*255), int(g2*255), int(b2*255))
}
