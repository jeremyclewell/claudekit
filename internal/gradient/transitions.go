package gradient

import (
	"fmt"
)

// InterpolateGradient interpolates between two gradient themes.
func InterpolateGradient(from, to Theme, progress float64) Theme {
	return Theme{
		Name:       fmt.Sprintf("interpolated-%f", progress),
		StartColor: from.StartColor, // AdaptiveColor interpolation complex, keep from
		EndColor:   from.EndColor,
		Stops:      int(float64(from.Stops) + float64(to.Stops-from.Stops)*progress),
		Direction:  from.Direction,
		Intensity:  from.Intensity + (to.Intensity-from.Intensity)*progress,
	}
}

// EaseInOutCubic applies cubic easing for smooth animations.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - ((-2*t+2)*(-2*t+2)*(-2*t+2))/2
}
