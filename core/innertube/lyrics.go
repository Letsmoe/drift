package innertube

import (
	"context"
	"strconv"
	"strings"

	"github.com/letsmoe/drift/core/music"
)

// lyricsBrowsePrefix starts the browse ID of a track's lyrics page.
const lyricsBrowsePrefix = "MPLYt"

const lyricsSourcePrefix = "Source: "

type nextRequest struct {
	Context requestContext `json:"context"`
	VideoID string         `json:"videoId"`
}

// Lyrics returns YouTube Music's lyrics for videoID, synced when YouTube has
// timestamps for them. The bool is false when the track has no lyrics, which is
// not an error.
func (client *Client) Lyrics(ctx context.Context, videoID string) (music.Lyrics, bool, error) {
	browseID, err := client.lyricsBrowseID(ctx, videoID)
	if err != nil {
		return music.Lyrics{}, false, err
	}
	if browseID == "" {
		return music.Lyrics{}, false, nil
	}

	// The timed page is best effort: any failure falls back to the plain page.
	var timedResponse timedLyricsResponse
	if err := client.browseAs(ctx, AndroidMusic, browseRequest{BrowseID: browseID}, &timedResponse); err == nil {
		if lyrics, found := parseTimedLyrics(timedResponse); found {
			return lyrics, true, nil
		}
	}

	var response lyricsResponse
	if err := client.browse(ctx, browseRequest{BrowseID: browseID}, &response); err != nil {
		return music.Lyrics{}, false, err
	}
	lyrics, found := parseLyrics(response)
	return lyrics, found, nil
}

// lyricsBrowseID asks the watch-next tabs of videoID for the lyrics page; it is
// empty when the track has none.
func (client *Client) lyricsBrowseID(ctx context.Context, videoID string) (string, error) {
	call, err := client.visitorCall(ctx, WebRemix, "next", func(context requestContext) any {
		return nextRequest{Context: context, VideoID: videoID}
	})
	if err != nil {
		return "", err
	}

	var response watchTabsResponse
	if err := client.post(ctx, call, &response); err != nil {
		return "", err
	}
	return findLyricsBrowseID(response), nil
}

func findLyricsBrowseID(response watchTabsResponse) string {
	tabs := response.Contents.SingleColumnMusicWatchNextResultsRenderer.TabbedRenderer.WatchNextTabbedResultsRenderer.Tabs
	for _, tab := range tabs {
		browseID := tab.TabRenderer.Endpoint.BrowseEndpoint.BrowseID
		if strings.HasPrefix(browseID, lyricsBrowsePrefix) {
			return browseID
		}
	}
	return ""
}

// parseLyrics reads the description shelf; a "Lyrics not available" message or
// an empty text means no lyrics.
func parseLyrics(response lyricsResponse) (music.Lyrics, bool) {
	for _, section := range response.Contents.SectionListRenderer.Contents {
		shelf := section.MusicDescriptionShelfRenderer
		if shelf == nil {
			continue
		}
		text := strings.TrimSpace(joinText(shelf.Description.Runs))
		if text == "" {
			return music.Lyrics{}, false
		}
		source := strings.TrimPrefix(strings.TrimSpace(joinText(shelf.Footer.Runs)), lyricsSourcePrefix)
		return music.Lyrics{Lines: splitLyricLines(text), Source: source}, true
	}
	return music.Lyrics{}, false
}

// parseTimedLyrics reads the ANDROID_MUSIC lyrics model. Lines without a cue
// range make the lyrics plain.
func parseTimedLyrics(response timedLyricsResponse) (music.Lyrics, bool) {
	model := response.Contents.ElementRenderer.NewElement.Type.ComponentType.Model.TimedLyricsModel
	if model == nil || len(model.LyricsData.TimedLyricsData) == 0 {
		return music.Lyrics{}, false
	}
	data := model.LyricsData
	lyrics := music.Lyrics{
		Source: strings.TrimPrefix(strings.TrimSpace(data.SourceMessage), lyricsSourcePrefix),
		Synced: true,
	}
	for _, timedLine := range data.TimedLyricsData {
		line := music.LyricLine{Text: strings.TrimSpace(timedLine.LyricLine)}
		startMs, hasStart := cueStartMs(timedLine.CueRange)
		if !hasStart {
			lyrics.Synced = false
		}
		line.StartMs = startMs
		lyrics.Lines = append(lyrics.Lines, line)
	}
	if !lyrics.Synced {
		for index := range lyrics.Lines {
			lyrics.Lines[index].StartMs = 0
		}
	}
	return lyrics, true
}

func cueStartMs(cue *lyricCueRange) (int64, bool) {
	if cue == nil {
		return 0, false
	}
	startMs, err := strconv.ParseInt(cue.StartTimeMilliseconds, 10, 64)
	if err != nil {
		return 0, false
	}
	return startMs, true
}

func splitLyricLines(text string) []music.LyricLine {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	rawLines := strings.Split(text, "\n")
	lines := make([]music.LyricLine, 0, len(rawLines))
	for _, rawLine := range rawLines {
		lines = append(lines, music.LyricLine{Text: strings.TrimSpace(rawLine)})
	}
	return lines
}
