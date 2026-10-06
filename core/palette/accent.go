package palette

import "math"

const (
	// accentMinimumContrast is the WCAG ratio for normal text.
	accentMinimumContrast = 4.5
	// accentBackgroundLuminance is the relative luminance of #3f3f46, the lightest
	// dark panel (hovered rows) the accent is drawn on.
	accentBackgroundLuminance = 0.0513
	// accentMinimumWhite always lifts the accent a little, so it reads as a highlight.
	accentMinimumWhite = 0.2
	accentWhiteStep    = 0.05
)

// accentFor returns vibrant mixed with the least white that makes it readable on
// the dark panels.
func accentFor(vibrant rgb) rgb {
	for white := accentMinimumWhite; white < 1; white += accentWhiteStep {
		accent := mixWithWhite(vibrant, white)
		if contrastRatio(luminance(accent), accentBackgroundLuminance) >= accentMinimumContrast {
			return accent
		}
	}
	return rgb{1, 1, 1}
}

func mixWithWhite(color rgb, white float64) rgb {
	return rgb{
		red:   color.red + (1-color.red)*white,
		green: color.green + (1-color.green)*white,
		blue:  color.blue + (1-color.blue)*white,
	}
}

// luminance is the WCAG relative luminance of an sRGB color.
func luminance(color rgb) float64 {
	return 0.2126*linear(color.red) + 0.7152*linear(color.green) + 0.0722*linear(color.blue)
}

func linear(channel float64) float64 {
	if channel <= 0.04045 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}

func contrastRatio(first, second float64) float64 {
	lighter := math.Max(first, second)
	darker := math.Min(first, second)
	return (lighter + 0.05) / (darker + 0.05)
}
