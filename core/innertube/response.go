package innertube

// Response types model only the fields the parsers read. encoding/json matches
// field names case-insensitively, so InnerTube's camelCase keys need no tags.

const (
	pageTypeArtist   = "MUSIC_PAGE_TYPE_ARTIST"
	pageTypeAlbum    = "MUSIC_PAGE_TYPE_ALBUM"
	pageTypePlaylist = "MUSIC_PAGE_TYPE_PLAYLIST"
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
	Continuations []struct {
		NextContinuationData struct {
			Continuation string
		}
	}
}

type section struct {
	MusicShelfRenderer *struct {
		Contents []struct {
			MusicResponsiveListItemRenderer *listItem
		}
	}
	MusicCarouselShelfRenderer *carouselShelf
}

type listItem struct {
	PlaylistItemData struct {
		VideoID string
	}
	Thumbnail struct {
		MusicThumbnailRenderer struct {
			Thumbnail struct {
				Thumbnails []thumbnail
			}
		}
	}
	FlexColumns []struct {
		MusicResponsiveListItemFlexColumnRenderer struct {
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
		VideoID string
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

// listItems flattens the shelves of the first (active) result tab.
func (response searchResponse) listItems() []listItem {
	tabs := response.Contents.TabbedSearchResultsRenderer.Tabs
	if len(tabs) == 0 {
		return nil
	}

	var items []listItem
	for _, section := range tabs[0].TabRenderer.Content.SectionListRenderer.Contents {
		items = append(items, section.listItems()...)
	}
	return items
}

func (section section) listItems() []listItem {
	if section.MusicShelfRenderer == nil {
		return nil
	}

	var items []listItem
	for _, content := range section.MusicShelfRenderer.Contents {
		if content.MusicResponsiveListItemRenderer != nil {
			items = append(items, *content.MusicResponsiveListItemRenderer)
		}
	}
	return items
}

func (item listItem) column(index int) []run {
	if index >= len(item.FlexColumns) {
		return nil
	}
	return item.FlexColumns[index].MusicResponsiveListItemFlexColumnRenderer.Text.Runs
}

func (item listItem) thumbnails() []thumbnail {
	return item.Thumbnail.MusicThumbnailRenderer.Thumbnail.Thumbnails
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
