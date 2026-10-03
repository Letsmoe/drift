package innertube

import "github.com/neoworks/otter-music/core/music"

// itemKinds maps the page a card opens to its kind; cards opening anything else
// (podcasts, audiobooks, moods) are skipped.
var itemKinds = map[string]music.ItemKind{
	pageTypeAlbum:    music.ItemKindAlbum,
	pageTypePlaylist: music.ItemKindPlaylist,
	pageTypeArtist:   music.ItemKindArtist,
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
	return music.Item{}, false
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

	if videoID := card.videoID(); videoID != "" {
		track := music.Track{
			VideoID:    videoID,
			Title:      title,
			Artists:    parseLinkedArtists(card.Subtitle.Runs),
			Thumbnails: thumbnails,
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
