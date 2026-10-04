package innertube

// browseResponse is a browse page (contents) or its continuation (continuationContents).
type browseResponse struct {
	Header struct {
		MusicImmersiveHeaderRenderer *artistHeader
		MusicVisualHeaderRenderer    *artistHeader
	}
	Contents struct {
		SingleColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer sectionList
					}
				}
			}
		}
	}
	ContinuationContents struct {
		SectionListContinuation sectionList
	}
}

type carouselShelf struct {
	Header struct {
		MusicCarouselShelfBasicHeaderRenderer struct {
			Title struct {
				Runs []run
			}
		}
	}
	Contents []carouselItem
}

// carouselItem holds a card (album, playlist, artist, video) or a song row.
type carouselItem struct {
	MusicTwoRowItemRenderer         *twoRowItem
	MusicResponsiveListItemRenderer *listItem
}

type twoRowItem struct {
	Title struct {
		Runs []run
	}
	Subtitle struct {
		Runs []run
	}
	NavigationEndpoint navigationEndpoint
	ThumbnailRenderer  struct {
		MusicThumbnailRenderer struct {
			Thumbnail struct {
				Thumbnails []thumbnail
			}
		}
	}
}

// sectionList prefers the continuation: continuation responses also carry the
// page's tab skeleton, with an empty section list.
func (response browseResponse) sectionList() sectionList {
	continuation := response.ContinuationContents.SectionListContinuation
	if len(continuation.Contents) > 0 {
		return continuation
	}
	tabs := response.Contents.SingleColumnBrowseResultsRenderer.Tabs
	if len(tabs) == 0 {
		return sectionList{}
	}
	return tabs[0].TabRenderer.Content.SectionListRenderer
}

func (list sectionList) continuation() string {
	return list.Continuations.token()
}

func (item twoRowItem) thumbnails() []thumbnail {
	return item.ThumbnailRenderer.MusicThumbnailRenderer.Thumbnail.Thumbnails
}

func (item twoRowItem) videoID() string {
	if item.NavigationEndpoint.WatchEndpoint == nil {
		return ""
	}
	return item.NavigationEndpoint.WatchEndpoint.VideoID
}
