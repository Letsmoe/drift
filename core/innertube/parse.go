package innertube

import (
	"slices"
	"strconv"
	"strings"

	"github.com/neoworks/otter-music/core/music"
)

// segmentSeparator splits the subtitle column into artists • album • duration.
const segmentSeparator = " • "

func parseSongs(items []listItem) []music.Track {
	tracks := []music.Track{}
	for _, item := range items {
		track, ok := parseSong(item)
		if ok {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

func parseSong(item listItem) (music.Track, bool) {
	videoID := item.PlaylistItemData.VideoID
	title := joinText(item.column(0))
	if videoID == "" || title == "" {
		return music.Track{}, false
	}

	segments := splitSegments(item.column(1))
	return music.Track{
		VideoID: videoID,
		Title:   title,
		Artists: parseArtists(segments),
		// Search puts the album in the subtitle column, the home feed in a column of its own.
		Album:           parseAlbum(slices.Concat(item.column(1), item.column(2))),
		DurationSeconds: parseDuration(segments),
		Thumbnails:      parseThumbnails(item.thumbnails()),
	}, true
}

func splitSegments(runs []run) [][]run {
	segments := [][]run{{}}
	for _, textRun := range runs {
		if textRun.Text == segmentSeparator {
			segments = append(segments, []run{})
			continue
		}
		last := len(segments) - 1
		segments[last] = append(segments[last], textRun)
	}
	return segments
}

// parseArtists reads the first segment, e.g. "Daft Punk", ", ", "Pharrell Williams", " & ", "Nile Rodgers".
func parseArtists(segments [][]run) []music.Artist {
	artists := []music.Artist{}
	for _, textRun := range segments[0] {
		if isArtistJoiner(textRun.Text) {
			continue
		}
		artists = append(artists, music.Artist{ID: textRun.browseID(), Name: textRun.Text})
	}
	return artists
}

func isArtistJoiner(text string) bool {
	trimmed := strings.TrimSpace(text)
	return trimmed == "" || trimmed == "," || trimmed == "&"
}

func parseAlbum(runs []run) *music.Album {
	for _, textRun := range runs {
		if textRun.pageType() == pageTypeAlbum {
			return &music.Album{ID: textRun.browseID(), Name: textRun.Text}
		}
	}
	return nil
}

// parseDuration reads the last segment as "m:ss" or "h:mm:ss"; 0 means unknown.
func parseDuration(segments [][]run) int {
	lastSegment := segments[len(segments)-1]
	if len(lastSegment) != 1 {
		return 0
	}

	totalSeconds := 0
	for _, part := range strings.Split(lastSegment[0].Text, ":") {
		value, err := strconv.Atoi(part)
		if err != nil {
			return 0
		}
		totalSeconds = totalSeconds*60 + value
	}
	return totalSeconds
}

func parseThumbnails(thumbnails []thumbnail) []music.Thumbnail {
	parsed := make([]music.Thumbnail, 0, len(thumbnails))
	for _, thumbnail := range thumbnails {
		parsed = append(parsed, music.Thumbnail{
			URL:    thumbnail.URL,
			Width:  thumbnail.Width,
			Height: thumbnail.Height,
		})
	}
	return parsed
}

func joinText(runs []run) string {
	var builder strings.Builder
	for _, textRun := range runs {
		builder.WriteString(textRun.Text)
	}
	return builder.String()
}
