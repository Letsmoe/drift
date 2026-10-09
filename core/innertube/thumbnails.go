package innertube

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/letsmoe/drift/core/music"
)

// largeVariantWidths are the widths added to resizable thumbnail lists, enough for the
// Now playing column and the full-screen player on 2x screens.
var largeVariantWidths = []int{544, 1080}

// resizableSizePattern matches the size options of googleusercontent image URLs,
// e.g. "=w120-h120-l90-rj"; the server renders any size asked for.
var resizableSizePattern = regexp.MustCompile(`=w(\d+)-h(\d+)(-[^/?#]*)?$`)

// withLargeVariants appends larger sizes of the largest thumbnail when its URL can be
// resized and the list has nothing that large.
func withLargeVariants(thumbnails []music.Thumbnail) []music.Thumbnail {
	if len(thumbnails) == 0 {
		return thumbnails
	}
	largest := slices.MaxFunc(thumbnails, func(first, second music.Thumbnail) int {
		return first.Width - second.Width
	})
	match := resizableSizePattern.FindStringSubmatchIndex(largest.URL)
	if match == nil || largest.Width <= 0 || largest.Height <= 0 {
		return thumbnails
	}
	options := ""
	if match[6] >= 0 {
		options = largest.URL[match[6]:match[7]]
	}
	base := largest.URL[:match[0]]
	for _, width := range largeVariantWidths {
		if width <= largest.Width {
			continue
		}
		height := width * largest.Height / largest.Width
		thumbnails = append(thumbnails, music.Thumbnail{
			URL:    fmt.Sprintf("%s=w%d-h%d%s", base, width, height, options),
			Width:  width,
			Height: height,
		})
	}
	return thumbnails
}
