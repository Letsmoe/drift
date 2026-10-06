package innertube

import (
	"strings"

	"github.com/neoworks/otter-music/core/music"
)

// itemKinds maps the page a card opens to its kind; cards opening anything else
// (audiobooks, moods, profiles) are skipped.
var itemKinds = map[string]music.ItemKind{
	pageTypeAlbum:    music.ItemKindAlbum,
	pageTypePlaylist: music.ItemKindPlaylist,
	pageTypeArtist:   music.ItemKindArtist,
	pageTypePodcast:  music.ItemKindPodcast,
	pageTypeEpisode:  music.ItemKindEpisode,
}

func parseShelves(sections []section) []music.Shelf {
	shelves := []music.Shelf{}
	for _, section := range sections {
		if section.MusicCarouselShelfRenderer == nil {
			continue
		}
		shelf := parseShelf(*section.MusicCarouselShelfRenderer)
		if len(shelf.Items) > 0 {
			shelves = append(shelves, shelf)
		}
	}
	return shelves
}

func parseShelf(carousel carouselShelf) music.Shelf {
	shelf := music.Shelf{
		Title: joinText(carousel.Header.MusicCarouselShelfBasicHeaderRenderer.Title.Runs),
		Items: []music.Item{},
	}
	for _, content := range carousel.Contents {
		item, ok := parseShelfItem(content)
		if ok {
			shelf.Items = append(shelf.Items, item)
		}
	}
	return shelf
}

func parseShelfItem(content carouselItem) (music.Item, bool) {
	if content.MusicResponsiveListItemRenderer != nil {
		return parseSongRow(*content.MusicResponsiveListItemRenderer)
	}
	if content.MusicTwoRowItemRenderer != nil {
		return parseCard(*content.MusicTwoRowItemRenderer)
	}
	if content.MusicMultiRowListItemRenderer != nil {
		return parseEpisodeRowItem(*content.MusicMultiRowListItemRenderer)
	}
	return music.Item{}, false
}

// parseEpisodeRowItem reads an episode row whose secondTitle links its show.
func parseEpisodeRowItem(row episodeRow) (music.Item, bool) {
	showRuns := row.SecondTitle.Runs
	showName := joinText(showRuns)
	show := podcastShow{Title: showName, Author: music.Artist{Name: showName}}
	if len(showRuns) > 0 {
		show.ID = showRuns[0].browseID()
	}
	episode, ok := parseShowEpisode(row, show)
	if !ok {
		return music.Item{}, false
	}
	return episodeItem(episode.Track, show.Title), true
}

func parseSongRow(row listItem) (music.Item, bool) {
	track, ok := parseSong(row)
	if !ok {
		return music.Item{}, false
	}
	return songItem(track, joinText(row.column(1))), true
}

func parseCard(card twoRowItem) (music.Item, bool) {
	title := joinText(card.Title.Runs)
	subtitle := joinText(card.Subtitle.Runs)
	thumbnails := parseThumbnails(card.thumbnails())

	if card.NavigationEndpoint.isEpisode() {
		return parseEpisodeCard(card, title, subtitle), true
	}
	if videoID := card.videoID(); videoID != "" {
		track := music.Track{
			VideoID:    videoID,
			Title:      title,
			Artists:    parseLinkedArtists(card.Subtitle.Runs),
			Thumbnails: thumbnails,
			PlayCount:  parsePlayCount(card.Subtitle.Runs),
		}
		return songItem(track, subtitle), true
	}

	endpoint := card.NavigationEndpoint.BrowseEndpoint
	kind, known := itemKinds[endpoint.pageType()]
	if !known {
		return music.Item{}, false
	}
	return music.Item{
		Kind:       kind,
		ID:         endpoint.BrowseID,
		Title:      title,
		Subtitle:   subtitle,
		Thumbnails: thumbnails,
	}, true
}

func parseEpisodeCard(card twoRowItem, title, subtitle string) music.Item {
	videoID := card.videoID()
	if videoID == "" {
		videoID = strings.TrimPrefix(card.NavigationEndpoint.BrowseEndpoint.BrowseID, episodeBrowsePrefix)
	}
	track := parseEpisodeTrack(videoID, title, card.Subtitle.Runs, card.thumbnails())
	return episodeItem(track, subtitle)
}

// episodeBrowsePrefix precedes the video ID in an episode's browse ID.
const episodeBrowsePrefix = "MPED"

func songItem(track music.Track, subtitle string) music.Item {
	return music.Item{
		Kind:       music.ItemKindSong,
		ID:         track.VideoID,
		Title:      track.Title,
		Subtitle:   subtitle,
		Thumbnails: track.Thumbnails,
		Track:      &track,
	}
}

func episodeItem(track music.Track, subtitle string) music.Item {
	item := songItem(track, subtitle)
	item.Kind = music.ItemKindEpisode
	return item
}

// parseEpisodeTrack builds the playable track of an episode; the show is the
// podcast link among the subtitle runs, e.g. "Episode • Aug 30 • The Show".
func parseEpisodeTrack(videoID, title string, subtitle []run, thumbnails []thumbnail) music.Track {
	track := music.Track{
		VideoID:    videoID,
		Title:      title,
		Artists:    []music.Artist{},
		Thumbnails: parseThumbnails(thumbnails),
		PlayCount:  parsePlayCount(subtitle),
	}
	for _, textRun := range subtitle {
		if textRun.pageType() == pageTypePodcast {
			track.Artists = []music.Artist{{Name: textRun.Text}}
			track.Album = &music.Album{ID: textRun.browseID(), Name: textRun.Text}
			break
		}
	}
	return track
}

// parseLinkedArtists picks the artist links out of a card subtitle like "Song • Artist • 1.2M views".
func parseLinkedArtists(runs []run) []music.Artist {
	artists := []music.Artist{}
	for _, textRun := range runs {
		if textRun.pageType() == pageTypeArtist {
			artists = append(artists, music.Artist{ID: textRun.browseID(), Name: textRun.Text})
		}
	}
	return artists
}
