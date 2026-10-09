package innertube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/letsmoe/drift/core/music"
)

func parseSuggestions(response suggestionsResponse) []string {
	suggestions := []string{}
	for _, section := range response.Contents {
		for _, entry := range section.SearchSuggestionsSectionRenderer.Contents {
			if entry.SearchSuggestionRenderer != nil {
				suggestions = append(suggestions, joinText(entry.SearchSuggestionRenderer.Suggestion.Runs))
			}
		}
	}
	return suggestions
}

func parseAlbumPage(browseID string, header pageHeader, shelf listShelf) music.AlbumPage {
	page := music.AlbumPage{
		ID:         browseID,
		PlaylistID: header.playlistID(),
		Title:      joinText(header.Title.Runs),
		Subtitle:   joinText(header.Subtitle.Runs),
		Artists:    parseArtists(splitSegments(header.StraplineTextOne.Runs)),
		Year:       parseYear(header.Subtitle.Runs),
		Thumbnails: parseThumbnails(header.Thumbnail.thumbnails()),
	}
	page.Tracks = parseSongs(rowsOf(shelf.Contents))
	for index := range page.Tracks {
		completeAlbumTrack(&page.Tracks[index], page)
	}
	return page
}

// completeAlbumTrack fills what album rows leave out: the album, its cover and,
// for rows without credited artists, the album artists.
func completeAlbumTrack(track *music.Track, page music.AlbumPage) {
	track.Album = &music.Album{ID: page.ID, Name: page.Title}
	if len(track.Thumbnails) == 0 {
		track.Thumbnails = page.Thumbnails
	}
	if len(track.Artists) == 0 {
		track.Artists = page.Artists
	}
}

// parseYear returns the four-digit run of a subtitle like "Album • 2010".
func parseYear(runs []run) string {
	for _, textRun := range runs {
		_, err := strconv.Atoi(textRun.Text)
		if err == nil && len(textRun.Text) == yearLength {
			return textRun.Text
		}
	}
	return ""
}

const yearLength = 4

func parsePlaylistPage(browseID string, header pageHeader, shelf listShelf) music.PlaylistPage {
	return music.PlaylistPage{
		ID:           browseID,
		Title:        joinText(header.Title.Runs),
		Author:       parsePlaylistAuthor(header),
		Description:  header.description(),
		Thumbnails:   parseThumbnails(header.Thumbnail.thumbnails()),
		Tracks:       parseSongs(rowsOf(shelf.Contents)),
		Continuation: continuationTokenOf(shelf.Contents),
	}
}

// parsePlaylistAuthor reads the facepile text, falling back to the strapline.
func parsePlaylistAuthor(header pageHeader) string {
	author := header.Facepile.AvatarStackViewModel.Text.Content
	if author != "" {
		return author
	}
	return joinText(header.StraplineTextOne.Runs)
}

func parseArtistPage(browseID string, header artistHeader, sections []section) music.ArtistPage {
	return music.ArtistPage{
		ID:          browseID,
		Name:        joinText(header.Title.Runs),
		Description: joinText(header.Description.Runs),
		Thumbnails:  parseThumbnails(header.Thumbnail.thumbnails()),
		TopSongs:    parseTopSongs(sections),
		Shelves:     parseShelves(sections),
	}
}

func parseTopSongs(sections []section) []music.Track {
	for _, section := range sections {
		if section.MusicShelfRenderer != nil {
			return parseSongs(rowsOf(section.MusicShelfRenderer.Contents))
		}
	}
	return []music.Track{}
}

func parseRadio(response radioResponse) []music.Track {
	tracks := []music.Track{}
	tabs := response.Contents.SingleColumnMusicWatchNextResultsRenderer.TabbedRenderer.WatchNextTabbedResultsRenderer.Tabs
	if len(tabs) == 0 {
		return tracks
	}
	for _, entry := range tabs[0].TabRenderer.Content.MusicQueueRenderer.Content.PlaylistPanelRenderer.Contents {
		if entry.PlaylistPanelVideoRenderer == nil {
			continue
		}
		track, ok := parsePanelVideo(*entry.PlaylistPanelVideoRenderer)
		if ok {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

func parsePanelVideo(video panelVideo) (music.Track, bool) {
	title := joinText(video.Title.Runs)
	if video.VideoID == "" || title == "" {
		return music.Track{}, false
	}
	byline := video.LongBylineText.Runs
	return music.Track{
		VideoID: video.VideoID,
		Title:   title,
		Artists: parseArtists(splitSegments(byline)),
		Album:   parseAlbum(byline),
		// The byline ends with the release year, so only lengthText holds the duration.
		DurationSeconds: parseDuration([][]run{video.LengthText.Runs}),
		Thumbnails:      parseThumbnails(video.Thumbnail.Thumbnails),
		PlayCount:       parsePlayCount(byline),
	}, true
}

// podcastShow is what episodes inherit from their show.
type podcastShow struct {
	ID     string       `json:"id"`
	Title  string       `json:"title"`
	Author music.Artist `json:"author"`
}

// podcastToken is the continuation handed to callers. Episode rows of a
// continuation response do not name their show, so the token carries it along.
type podcastToken struct {
	Token string      `json:"token"`
	Show  podcastShow `json:"show"`
}

// next wraps the following YouTube token; no token means the last page.
func (token podcastToken) next(youtubeToken string) string {
	if youtubeToken == "" {
		return ""
	}
	token.Token = youtubeToken
	encoded, err := json.Marshal(token)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func decodePodcastToken(token string) (podcastToken, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return podcastToken{}, fmt.Errorf("innertube podcast continuation: invalid token: %w", err)
	}
	var parsed podcastToken
	if err := json.Unmarshal(decoded, &parsed); err != nil {
		return podcastToken{}, fmt.Errorf("innertube podcast continuation: invalid token: %w", err)
	}
	return parsed, nil
}

func parsePodcastPage(browseID string, header pageHeader, shelf listShelf) music.PodcastPage {
	title := joinText(header.Title.Runs)
	author := parseShowAuthor(header)
	show := podcastShow{ID: browseID, Title: title, Author: author}

	nextToken := podcastToken{Show: show}.next(shelf.Continuations.token())
	return music.PodcastPage{
		ID:           browseID,
		Title:        title,
		Author:       author,
		Description:  header.description(),
		Thumbnails:   parseThumbnails(header.Thumbnail.thumbnails()),
		Episodes:     parseShowEpisodes(shelf.Contents, show),
		Continuation: nextToken,
	}
}

func parseShowAuthor(header pageHeader) music.Artist {
	runs := header.StraplineTextOne.Runs
	if len(runs) == 0 {
		return music.Artist{}
	}
	return music.Artist{ID: runs[0].browseID(), Name: joinText(runs)}
}

func parseShowEpisodes(entries []shelfItem, show podcastShow) []music.Episode {
	episodes := []music.Episode{}
	for _, entry := range entries {
		if entry.MusicMultiRowListItemRenderer == nil {
			continue
		}
		episode, ok := parseShowEpisode(*entry.MusicMultiRowListItemRenderer, show)
		if ok {
			episodes = append(episodes, episode)
		}
	}
	return episodes
}

func parseShowEpisode(row episodeRow, show podcastShow) (music.Episode, bool) {
	videoID := ""
	if row.OnTap.WatchEndpoint != nil {
		videoID = row.OnTap.WatchEndpoint.VideoID
	}
	title := joinText(row.Title.Runs)
	if videoID == "" || title == "" {
		return music.Episode{}, false
	}

	duration := row.PlaybackProgress.MusicPlaybackProgressRenderer.DurationText.Runs
	return music.Episode{
		Track: music.Track{
			VideoID:         videoID,
			Title:           title,
			Artists:         []music.Artist{show.Author},
			Album:           &music.Album{ID: show.ID, Name: show.Title},
			DurationSeconds: parseSpokenDuration(joinText(duration)),
			Thumbnails:      parseThumbnails(row.Thumbnail.thumbnails()),
			PlayCount:       parsePlayCount(row.Subtitle.Runs),
		},
		Description: joinText(row.Description.Runs),
		Published:   lastSegmentText(row.Subtitle.Runs),
	}, true
}

// lastSegmentText returns the text after the last separator, e.g. the date of "1 view • Aug 30".
func lastSegmentText(runs []run) string {
	segments := splitSegments(runs)
	return joinText(segments[len(segments)-1])
}

var spokenDurationUnits = map[string]int{"hr": 3600, "min": 60, "sec": 1}

// parseSpokenDuration reads durations like " • 1 hr 5 min"; 0 means unknown.
func parseSpokenDuration(text string) int {
	fields := strings.Fields(strings.ReplaceAll(text, segmentSeparator, " "))
	totalSeconds := 0
	for index := 0; index+1 < len(fields); index += 2 {
		amount, err := strconv.Atoi(fields[index])
		if err != nil {
			return 0
		}
		totalSeconds += amount * spokenDurationUnits[fields[index+1]]
	}
	return totalSeconds
}
