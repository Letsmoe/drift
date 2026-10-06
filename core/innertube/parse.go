package innertube

import (
	"regexp"
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
	videoID := item.videoID()
	title := joinText(item.column(0))
	if videoID == "" || title == "" {
		return music.Track{}, false
	}

	segments := splitSegments(item.subtitleRuns())
	return music.Track{
		VideoID: videoID,
		Title:   title,
		Artists: parseArtists(segments),
		// Search puts the album in the subtitle column, the home feed and artist pages in columns of their own.
		Album:           parseAlbum(slices.Concat(item.subtitleRuns(), item.column(2), item.column(3))),
		DurationSeconds: parseRowDuration(item, segments),
		Thumbnails:      parseThumbnails(item.thumbnails()),
		PlayCount:       parsePlayCount(item.allColumnRuns()),
	}, true
}

// playCountPattern matches YouTube's count segments: "6.2M plays", "922K views", "1 view".
var playCountPattern = regexp.MustCompile(`^[\d.,]+[KMB]? (plays?|views?)$`)

// parsePlayCount returns the first play or view count among runs, as shown.
func parsePlayCount(runs []run) string {
	for _, textRun := range runs {
		text := strings.TrimSpace(textRun.Text)
		if playCountPattern.MatchString(text) {
			return text
		}
	}
	return ""
}

// typeLabels are the leading "Song • " style segments unfiltered search puts before the details.
var typeLabels = map[string]bool{
	"Song": true, "Video": true, "Episode": true, "Album": true, "Single": true,
	"EP": true, "Playlist": true, "Artist": true, "Podcast": true, "Profile": true,
}

// dropTypeLabel removes a leading type label and its separator from subtitle runs.
func dropTypeLabel(runs []run) []run {
	if len(runs) < 2 || runs[0].NavigationEndpoint != nil {
		return runs
	}
	if !typeLabels[runs[0].Text] || runs[1].Text != segmentSeparator {
		return runs
	}
	return runs[2:]
}

// subtitleRuns is the second column without the type label.
func (item listItem) subtitleRuns() []run {
	return dropTypeLabel(item.column(1))
}

// parseRowDuration reads the duration from the subtitle, else from the fixed
// column album and playlist rows keep it in.
func parseRowDuration(item listItem, segments [][]run) int {
	duration := parseDuration(segments)
	if duration > 0 {
		return duration
	}
	return parseDuration([][]run{item.fixedColumn(0)})
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
	// Rows without credited artists start with the duration instead.
	if len(segments) == 1 && parseDuration(segments) > 0 {
		return artists
	}
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
