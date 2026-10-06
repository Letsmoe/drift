package innertube

// Response types model only the fields the parsers read. encoding/json matches
// field names case-insensitively, so InnerTube's camelCase keys need no tags.

const (
	pageTypeArtist        = "MUSIC_PAGE_TYPE_ARTIST"
	pageTypeAlbum         = "MUSIC_PAGE_TYPE_ALBUM"
	pageTypePlaylist      = "MUSIC_PAGE_TYPE_PLAYLIST"
	pageTypePodcast       = "MUSIC_PAGE_TYPE_PODCAST_SHOW_DETAIL_PAGE"
	pageTypeEpisode       = "MUSIC_PAGE_TYPE_NON_MUSIC_AUDIO_TRACK_PAGE"
	musicVideoTypeEpisode = "MUSIC_VIDEO_TYPE_PODCAST_EPISODE"
)

type searchResponse struct {
	Contents struct {
		TabbedSearchResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer sectionList
					}
				}
			}
		}
	}
}

type sectionList struct {
	Contents      []section
	Continuations continuationList
}

// continuationList is the older continuation form: a token next to the list.
type continuationList []struct {
	NextContinuationData struct {
		Continuation string
	}
}

type section struct {
	MusicShelfRenderer         *listShelf
	MusicPlaylistShelfRenderer *listShelf
	MusicCarouselShelfRenderer *carouselShelf
	MusicCardShelfRenderer     *cardShelf
	ItemSectionRenderer        *struct {
		Contents []shelfItem
	}
}

// listShelf is a vertical list of rows: search results, album tracks, podcast episodes.
type listShelf struct {
	Title struct {
		Runs []run
	}
	Contents      []shelfItem
	Continuations continuationList
}

// shelfItem is one entry of a listShelf. The last entry of a playlist page is a
// continuationItemRenderer holding the token for the next page.
type shelfItem struct {
	MusicResponsiveListItemRenderer *listItem
	MusicMultiRowListItemRenderer   *episodeRow
	ContinuationItemRenderer        *struct {
		ContinuationEndpoint struct {
			ContinuationCommand struct {
				Token string
			}
		}
	}
}

// cardShelf is the search "top result" card with its listed songs.
type cardShelf struct {
	Title struct {
		Runs []run
	}
	Subtitle struct {
		Runs []run
	}
	Thumbnail thumbnailContainer
	Contents  []shelfItem
}

type thumbnailContainer struct {
	MusicThumbnailRenderer struct {
		Thumbnail struct {
			Thumbnails []thumbnail
		}
	}
}

type listItem struct {
	PlaylistItemData struct {
		VideoID string
	}
	NavigationEndpoint navigationEndpoint
	Thumbnail          thumbnailContainer
	FlexColumns        []struct {
		MusicResponsiveListItemFlexColumnRenderer struct {
			Text struct {
				Runs []run
			}
		}
	}
	FixedColumns []struct {
		MusicResponsiveListItemFixedColumnRenderer struct {
			Text struct {
				Runs []run
			}
		}
	}
}

type thumbnail struct {
	URL    string
	Width  int
	Height int
}

type run struct {
	Text               string
	NavigationEndpoint *navigationEndpoint
}

type navigationEndpoint struct {
	BrowseEndpoint *browseEndpoint
	WatchEndpoint  *struct {
		VideoID                            string
		WatchEndpointMusicSupportedConfigs struct {
			WatchEndpointMusicConfig struct {
				MusicVideoType string
			}
		}
	}
}

type browseEndpoint struct {
	BrowseID                              string
	BrowseEndpointContextSupportedConfigs struct {
		BrowseEndpointContextMusicConfig struct {
			PageType string
		}
	}
}

// sections returns the sections of the first (active) result tab.
func (response searchResponse) sections() []section {
	tabs := response.Contents.TabbedSearchResultsRenderer.Tabs
	if len(tabs) == 0 {
		return nil
	}
	return tabs[0].TabRenderer.Content.SectionListRenderer.Contents
}

// listItems flattens the shelves of the first (active) result tab.
func (response searchResponse) listItems() []listItem {
	var items []listItem
	for _, section := range response.sections() {
		if section.MusicShelfRenderer != nil {
			items = append(items, rowsOf(section.MusicShelfRenderer.Contents)...)
		}
	}
	return items
}

// rowsOf picks the song rows out of shelf entries.
func rowsOf(entries []shelfItem) []listItem {
	var rows []listItem
	for _, entry := range entries {
		if entry.MusicResponsiveListItemRenderer != nil {
			rows = append(rows, *entry.MusicResponsiveListItemRenderer)
		}
	}
	return rows
}

// continuationTokenOf returns the token of the continuationItemRenderer entry, if any.
func continuationTokenOf(entries []shelfItem) string {
	for _, entry := range entries {
		if entry.ContinuationItemRenderer != nil {
			return entry.ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand.Token
		}
	}
	return ""
}

func (list continuationList) token() string {
	if len(list) == 0 {
		return ""
	}
	return list[0].NextContinuationData.Continuation
}

func (item listItem) column(index int) []run {
	if index >= len(item.FlexColumns) {
		return nil
	}
	return item.FlexColumns[index].MusicResponsiveListItemFlexColumnRenderer.Text.Runs
}

// allColumnRuns concatenates the runs of every flex column.
func (item listItem) allColumnRuns() []run {
	var runs []run
	for index := range item.FlexColumns {
		runs = append(runs, item.column(index)...)
	}
	return runs
}

func (item listItem) fixedColumn(index int) []run {
	if index >= len(item.FixedColumns) {
		return nil
	}
	return item.FixedColumns[index].MusicResponsiveListItemFixedColumnRenderer.Text.Runs
}

func (item listItem) thumbnails() []thumbnail {
	return item.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails
}

// browseEndpoint is the page the row opens: its own endpoint, else the title's.
func (item listItem) browseEndpoint() *browseEndpoint {
	if item.NavigationEndpoint.BrowseEndpoint != nil {
		return item.NavigationEndpoint.BrowseEndpoint
	}
	title := item.column(0)
	if len(title) == 0 {
		return nil
	}
	return title[0].browseEndpoint()
}

// videoID is the video the row plays; episode and album rows without one open a page.
func (item listItem) videoID() string {
	return item.PlaylistItemData.VideoID
}

func (navigation navigationEndpoint) isEpisode() bool {
	if navigation.BrowseEndpoint.pageType() == pageTypeEpisode {
		return true
	}
	if navigation.WatchEndpoint == nil {
		return false
	}
	config := navigation.WatchEndpoint.WatchEndpointMusicSupportedConfigs.WatchEndpointMusicConfig
	return config.MusicVideoType == musicVideoTypeEpisode
}

func (textRun run) browseEndpoint() *browseEndpoint {
	if textRun.NavigationEndpoint == nil {
		return nil
	}
	return textRun.NavigationEndpoint.BrowseEndpoint
}

func (textRun run) browseID() string {
	endpoint := textRun.browseEndpoint()
	if endpoint == nil {
		return ""
	}
	return endpoint.BrowseID
}

func (textRun run) pageType() string {
	return textRun.browseEndpoint().pageType()
}

func (endpoint *browseEndpoint) pageType() string {
	if endpoint == nil {
		return ""
	}
	return endpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType
}

func (container thumbnailContainer) thumbnails() []thumbnail {
	return container.MusicThumbnailRenderer.Thumbnail.Thumbnails
}
