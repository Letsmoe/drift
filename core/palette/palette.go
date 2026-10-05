// Package palette picks the colors of a cover image, so the apps can tint their
// UI with the artwork of what is playing or shown.
package palette

import (
	"fmt"
	"image"
	"math"
)

// Palette holds the colors of one image as "#rrggbb"; both are empty when the
// image has no visible pixels.
type Palette struct {
	// Dominant is the most common color, ignoring near-black and near-white.
	Dominant string `json:"dominant"`
	// Vibrant is the strongest saturated color that covers a visible part of the
	// image; it falls back to Dominant for greyscale images.
	Vibrant string `json:"vibrant"`
}

const (
	// sampleGrid is how many pixels per side are sampled; covers are photos, so
	// a coarse grid finds the same colors as every pixel.
	sampleGrid = 64

	hueBins        = 24
	saturationBins = 4
	valueBins      = 4

	// vibrantMinimumShare keeps specks (a logo, a lens flare) from winning.
	vibrantMinimumShare = 0.01
	vibrantMinimumSat   = 0.3
	vibrantMinimumValue = 0.25

	nearBlackValue      = 0.12
	nearWhiteValue      = 0.92
	nearWhiteSaturation = 0.1
)

type rgb struct {
	red, green, blue float64
}

type hsv struct {
	hue, saturation, value float64
}

// bin accumulates the pixels of one hue/saturation/value cell.
type bin struct {
	count int
	sum   rgb
}

func (cell bin) average() rgb {
	count := float64(cell.count)
	return rgb{cell.sum.red / count, cell.sum.green / count, cell.sum.blue / count}
}

// Extract returns the dominant and vibrant colors of an image.
func Extract(img image.Image) Palette {
	bins, total := sample(img)
	if total == 0 {
		return Palette{}
	}
	dominant, found := pickDominant(bins)
	if !found {
		return Palette{}
	}
	vibrant, found := pickVibrant(bins, total)
	if !found {
		vibrant = dominant
	}
	return Palette{Dominant: hex(dominant), Vibrant: hex(vibrant)}
}

// sample buckets a grid of opaque pixels by hue, saturation and value.
func sample(img image.Image) (map[int]*bin, int) {
	bounds := img.Bounds()
	step := max(1, max(bounds.Dx(), bounds.Dy())/sampleGrid)
	bins := map[int]*bin{}
	total := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			color, opaque := pixel(img, x, y)
			if !opaque {
				continue
			}
			key := binKey(toHSV(color))
			cell, exists := bins[key]
			if !exists {
				cell = &bin{}
				bins[key] = cell
			}
			cell.count++
			cell.sum = rgb{cell.sum.red + color.red, cell.sum.green + color.green, cell.sum.blue + color.blue}
			total++
		}
	}
	return bins, total
}

func pixel(img image.Image, x, y int) (rgb, bool) {
	red, green, blue, alpha := img.At(x, y).RGBA()
	if alpha < 0x8000 {
		return rgb{}, false
	}
	// RGBA is alpha-premultiplied 16-bit; un-premultiply into [0, 1].
	scale := float64(alpha)
	return rgb{float64(red) / scale, float64(green) / scale, float64(blue) / scale}, true
}

func binKey(color hsv) int {
	hueIndex := min(hueBins-1, int(color.hue/360*hueBins))
	saturationIndex := min(saturationBins-1, int(color.saturation*saturationBins))
	valueIndex := min(valueBins-1, int(color.value*valueBins))
	return (hueIndex*saturationBins+saturationIndex)*valueBins + valueIndex
}

// pickDominant returns the most populated bin that is not near-black or
// near-white, or the most populated bin at all when the image is only those.
func pickDominant(bins map[int]*bin) (rgb, bool) {
	var best, fallback *bin
	for _, cell := range bins {
		if fallback == nil || cell.count > fallback.count {
			fallback = cell
		}
		if isExtreme(toHSV(cell.average())) {
			continue
		}
		if best == nil || cell.count > best.count {
			best = cell
		}
	}
	if best == nil {
		best = fallback
	}
	if best == nil {
		return rgb{}, false
	}
	return best.average(), true
}

// pickVibrant scores bins by saturation and brightness, weighted by the square
// root of their share so a large calm area doesn't drown a strong accent.
func pickVibrant(bins map[int]*bin, total int) (rgb, bool) {
	var best *bin
	bestScore := 0.0
	for _, cell := range bins {
		share := float64(cell.count) / float64(total)
		color := toHSV(cell.average())
		if share < vibrantMinimumShare || color.saturation < vibrantMinimumSat || color.value < vibrantMinimumValue {
			continue
		}
		score := math.Sqrt(share) * color.saturation * color.value
		if score > bestScore {
			best = cell
			bestScore = score
		}
	}
	if best == nil {
		return rgb{}, false
	}
	return best.average(), true
}

func isExtreme(color hsv) bool {
	if color.value < nearBlackValue {
		return true
	}
	return color.value > nearWhiteValue && color.saturation < nearWhiteSaturation
}

func toHSV(color rgb) hsv {
	highest := max(color.red, color.green, color.blue)
	lowest := min(color.red, color.green, color.blue)
	spread := highest - lowest
	result := hsv{value: highest}
	if highest > 0 {
		result.saturation = spread / highest
	}
	if spread == 0 {
		return result
	}
	switch highest {
	case color.red:
		result.hue = 60 * math.Mod((color.green-color.blue)/spread, 6)
	case color.green:
		result.hue = 60 * ((color.blue-color.red)/spread + 2)
	default:
		result.hue = 60 * ((color.red-color.green)/spread + 4)
	}
	if result.hue < 0 {
		result.hue += 360
	}
	return result
}

func hex(color rgb) string {
	return fmt.Sprintf("#%02x%02x%02x", channel(color.red), channel(color.green), channel(color.blue))
}

func channel(value float64) uint8 {
	return uint8(math.Round(math.Max(0, math.Min(1, value)) * 255))
}
