package neoworks

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/letsmoe/drift/core/music"
)

// syncedThumbnailWidth is the width of the one cover URL a synced item carries.
const syncedThumbnailWidth = 544

// rebuiltThumbnailWidths are the sizes YouTube Music hands Drift for a cover.
var rebuiltThumbnailWidths = []int{60, 120, 226, 544}

// sizeSuffixPattern matches the size options of googleusercontent image URLs,
// e.g. "=w544-h544-l90-rj"; the image server renders any size asked for.
var sizeSuffixPattern = regexp.MustCompile(`=w(\d+)-h(\d+)(-[^/?#]*)?$`)

// syncedThumbnailURL is the cover URL a synced item carries: the largest
// thumbnail, resized to syncedThumbnailWidth when its URL carries a size. Every
// device derives the same URL from the same cover, so a synced copy never
// differs from the local one only by size.
func syncedThumbnailURL(thumbnails []music.Thumbnail) string {
	if len(thumbnails) == 0 {
		return ""
	}
	largest := slices.MaxFunc(thumbnails, func(first, second music.Thumbnail) int {
		return first.Width - second.Width
	})
	base, aspect, options, resizable := splitSizedURL(largest.URL)
	if !resizable {
		return largest.URL
	}
	return sizedURL(base, syncedThumbnailWidth, aspect, options)
}

// rebuildThumbnails turns a synced cover URL back into Drift's usual size list;
// a URL without a size stays a single entry of unknown size.
func rebuildThumbnails(url string) []music.Thumbnail {
	if url == "" {
		return []music.Thumbnail{}
	}
	base, aspect, options, resizable := splitSizedURL(url)
	if !resizable {
		return []music.Thumbnail{{URL: url}}
	}
	thumbnails := make([]music.Thumbnail, 0, len(rebuiltThumbnailWidths))
	for _, width := range rebuiltThumbnailWidths {
		height := scaledHeight(width, aspect)
		thumbnails = append(thumbnails, music.Thumbnail{URL: sizedURL(base, width, aspect, options), Width: width, Height: height})
	}
	return thumbnails
}

// splitSizedURL splits a resizable URL into what precedes its size options, the
// height-to-width ratio and the options after the size.
func splitSizedURL(url string) (base string, aspect float64, options string, resizable bool) {
	match := sizeSuffixPattern.FindStringSubmatch(url)
	if match == nil {
		return "", 0, "", false
	}
	width, _ := strconv.Atoi(match[1])
	height, _ := strconv.Atoi(match[2])
	if width <= 0 || height <= 0 {
		return "", 0, "", false
	}
	suffixLength := len(match[0])
	return url[:len(url)-suffixLength], float64(height) / float64(width), match[3], true
}

func scaledHeight(width int, aspect float64) int {
	return int(float64(width)*aspect + 0.5)
}

func sizedURL(base string, width int, aspect float64, options string) string {
	return fmt.Sprintf("%s=w%d-h%d%s", base, width, scaledHeight(width, aspect), options)
}
