package innertube

import (
	"strings"

	"github.com/letsmoe/drift-music/core/music"
)

const (
	topResultTitle = "Top result"
	videosTitle    = "Videos"
)

// groupTitles name the shelves unfiltered search results are grouped into when
// YouTube sends them as one flat list instead of titled shelves.
var groupTitles = map[music.ItemKind]string{
	music.ItemKindSong:     "Songs",
	music.ItemKindAlbum:    "Albums",
	music.ItemKindArtist:   "Artists",
	music.ItemKindPlaylist: "Playlists",
	music.ItemKindPodcast:  "Podcasts",
	music.ItemKindEpisode:  "Episodes",
}

// searchShelves collects shelves in order of appearance and merges flat rows into one shelf per kind.
type searchShelves struct {
	shelves     []music.Shelf
	groupByName map[string]int
}

func parseSearch(sections []section) music.SearchResults {
	collected := searchShelves{shelves: []music.Shelf{}, groupByName: map[string]int{}}
	for _, section := range sections {
		collected.addSection(section)
	}
	return music.SearchResults{Shelves: collected.shelves}
}

func (collected *searchShelves) addSection(section section) {
	switch {
	case section.MusicCardShelfRenderer != nil:
		collected.addShelf(parseTopCard(*section.MusicCardShelfRenderer))
	case section.MusicShelfRenderer != nil:
		collected.addShelf(parseSearchShelf(*section.MusicShelfRenderer))
	case section.ItemSectionRenderer != nil:
		collected.addRows(rowsOf(section.ItemSectionRenderer.Contents))
	}
}

func (collected *searchShelves) addShelf(shelf music.Shelf) {
	if len(shelf.Items) > 0 {
		collected.shelves = append(collected.shelves, shelf)
	}
}

func (collected *searchShelves) addRows(rows []listItem) {
	for _, row := range rows {
		item, ok := parseSearchRow(row)
		if ok {
			collected.addToGroup(rowGroupTitle(row, item), item)
		}
	}
}

func (collected *searchShelves) addToGroup(title string, item music.Item) {
	index, exists := collected.groupByName[title]
	if !exists {
		index = len(collected.shelves)
		collected.groupByName[title] = index
		collected.shelves = append(collected.shelves, music.Shelf{Title: title, Items: []music.Item{}})
	}
	collected.shelves[index].Items = append(collected.shelves[index].Items, item)
}

func rowGroupTitle(row listItem, item music.Item) string {
	if item.Kind == music.ItemKindSong && isVideoRow(row) {
		return videosTitle
	}
	return groupTitles[item.Kind]
}

func isVideoRow(row listItem) bool {
	subtitle := row.column(1)
	return len(subtitle) > 0 && subtitle[0].Text == "Video"
}

func parseSearchShelf(shelf listShelf) music.Shelf {
	title := joinText(shelf.Title.Runs)
	if title == "" {
		title = topResultTitle
	}
	return music.Shelf{Title: title, Items: parseSearchRows(rowsOf(shelf.Contents))}
}

func parseSearchRows(rows []listItem) []music.Item {
	items := []music.Item{}
	for _, row := range rows {
		item, ok := parseSearchRow(row)
		if ok {
			items = append(items, item)
		}
	}
	return items
}

func parseSearchRow(row listItem) (music.Item, bool) {
	kind, known := rowKind(row)
	if !known {
		return music.Item{}, false
	}
	switch kind {
	case music.ItemKindSong:
		return parseSongRow(row)
	case music.ItemKindEpisode:
		return parseEpisodeRow(row), true
	default:
		return parseBrowseRow(row, kind), true
	}
}

// rowKind decides by the page the row opens, else by the video it plays.
func rowKind(row listItem) (music.ItemKind, bool) {
	kind, known := itemKinds[row.browseEndpoint().pageType()]
	if known {
		return kind, true
	}
	if row.videoID() != "" {
		return music.ItemKindSong, true
	}
	return "", false
}

func parseBrowseRow(row listItem, kind music.ItemKind) music.Item {
	return music.Item{
		Kind:       kind,
		ID:         row.browseEndpoint().BrowseID,
		Title:      joinText(row.column(0)),
		Subtitle:   viewsAsPlays(joinText(row.column(1))),
		Thumbnails: parseThumbnails(row.thumbnails()),
	}
}

func parseEpisodeRow(row listItem) music.Item {
	videoID := row.videoID()
	if videoID == "" {
		videoID = strings.TrimPrefix(row.browseEndpoint().BrowseID, episodeBrowsePrefix)
	}
	subtitle := row.subtitleRuns()
	track := parseEpisodeTrack(videoID, joinText(row.column(0)), subtitle, row.thumbnails())
	return episodeItem(track, joinText(subtitle))
}

// parseTopCard turns the top result card into a shelf of the card's own item
// followed by the songs listed under it.
func parseTopCard(card cardShelf) music.Shelf {
	shelf := music.Shelf{Title: topResultTitle, Items: []music.Item{}}
	cardItem, ok := parseCardItem(card)
	if !ok {
		return shelf
	}
	shelf.Items = append(shelf.Items, cardItem)

	for _, item := range parseSearchRows(rowsOf(card.Contents)) {
		shelf.Items = append(shelf.Items, withCardArtist(item, cardItem))
	}
	return shelf
}

// withCardArtist credits the card's artist on listed songs, which omit their artist.
func withCardArtist(item, cardItem music.Item) music.Item {
	if cardItem.Kind != music.ItemKindArtist || item.Track == nil || len(item.Track.Artists) > 0 {
		return item
	}
	item.Track.Artists = []music.Artist{{ID: cardItem.ID, Name: cardItem.Title}}
	return item
}

func parseCardItem(card cardShelf) (music.Item, bool) {
	if len(card.Title.Runs) == 0 || card.Title.Runs[0].NavigationEndpoint == nil {
		return music.Item{}, false
	}
	navigation := *card.Title.Runs[0].NavigationEndpoint
	title := joinText(card.Title.Runs)
	thumbnails := card.Thumbnail.thumbnails()

	if navigation.isEpisode() {
		return parseEpisodeCardShelf(card, navigation, title, thumbnails), true
	}
	if navigation.WatchEndpoint != nil && navigation.WatchEndpoint.VideoID != "" {
		return parseSongCard(card, navigation.WatchEndpoint.VideoID, title, thumbnails), true
	}
	kind, known := itemKinds[navigation.BrowseEndpoint.pageType()]
	if !known {
		return music.Item{}, false
	}
	return music.Item{
		Kind:       kind,
		ID:         navigation.BrowseEndpoint.BrowseID,
		Title:      title,
		Subtitle:   viewsAsPlays(joinText(card.Subtitle.Runs)),
		Thumbnails: parseThumbnails(thumbnails),
	}, true
}

func parseSongCard(card cardShelf, videoID, title string, thumbnails []thumbnail) music.Item {
	subtitle := dropTypeLabel(card.Subtitle.Runs)
	segments := splitSegments(subtitle)
	track := music.Track{
		VideoID:         videoID,
		Title:           title,
		Artists:         parseArtists(segments),
		Album:           parseAlbum(subtitle),
		DurationSeconds: parseDuration(segments),
		Thumbnails:      parseThumbnails(thumbnails),
		PlayCount:       parsePlayCount(subtitle),
	}
	return songItem(track, joinText(subtitle))
}

func parseEpisodeCardShelf(card cardShelf, navigation navigationEndpoint, title string, thumbnails []thumbnail) music.Item {
	videoID := ""
	if navigation.WatchEndpoint != nil {
		videoID = navigation.WatchEndpoint.VideoID
	}
	if videoID == "" {
		videoID = strings.TrimPrefix(navigation.BrowseEndpoint.BrowseID, episodeBrowsePrefix)
	}
	subtitle := dropTypeLabel(card.Subtitle.Runs)
	track := parseEpisodeTrack(videoID, title, subtitle, thumbnails)
	return episodeItem(track, joinText(subtitle))
}
