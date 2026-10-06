package palette

import "testing"

func TestAccentIsReadableOnDarkPanels(t *testing.T) {
	colors := map[string]rgb{
		"dark blue": {0.05, 0.1, 0.45},
		"dark red":  {0.4, 0.05, 0.05},
		"black":     {0, 0, 0},
		"yellow":    {0.95, 0.85, 0.1},
	}
	for name, vibrant := range colors {
		accent := accentFor(vibrant)
		if ratio := contrastRatio(luminance(accent), accentBackgroundLuminance); ratio < accentMinimumContrast {
			t.Errorf("%s: accent %s has contrast %.2f", name, hex(accent), ratio)
		}
	}
}

func TestAccentKeepsBrightColorsSaturated(t *testing.T) {
	accent := accentFor(rgb{0.95, 0.85, 0.1})

	if want := hex(mixWithWhite(rgb{0.95, 0.85, 0.1}, accentMinimumWhite)); hex(accent) != want {
		t.Errorf("accent = %s, want the minimum lift %s", hex(accent), want)
	}
}
