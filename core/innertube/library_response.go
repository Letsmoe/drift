package innertube

// libraryResponse is a library browse page (grid or list of the user's saved
// items) or one of its continuations. The shapes follow ytmusicapi's library
// parsers; see CLAUDE.md for how sure each one is.
type libraryResponse struct {
	Contents struct {
		SingleColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer librarySectionList
					}
				}
			}
		}
	}
	ContinuationContents struct {
		GridContinuation       libraryGrid
		MusicShelfContinuation libraryShelf
	}
	OnResponseReceivedActions []struct {
		AppendContinuationItemsAction struct {
			ContinuationItems []libraryEntry
		}
	}
}

type librarySectionList struct {
	Contents      []librarySection
	Continuations continuationList
}

// librarySection holds the grid or list directly or inside an item section.
type librarySection struct {
	ItemSectionRenderer *struct {
		Contents []librarySection
	}
	GridRenderer       *libraryGrid
	MusicShelfRenderer *libraryShelf
}

type libraryGrid struct {
	Items         []libraryEntry
	Continuations continuationList
}

type libraryShelf struct {
	Contents      []libraryEntry
	Continuations continuationList
}

// libraryEntry is a card (playlists, albums, podcasts) or a row (artists).
type libraryEntry struct {
	MusicTwoRowItemRenderer         *twoRowItem
	MusicResponsiveListItemRenderer *listItem
	ContinuationItemRenderer        *struct {
		ContinuationEndpoint struct {
			ContinuationCommand struct {
				Token string
			}
		}
	}
}

// libraryPage is one page of entries and the token of the next, if any.
type libraryPage struct {
	entries      []libraryEntry
	continuation string
}

func (response libraryResponse) page() libraryPage {
	continuation := response.ContinuationContents
	if len(continuation.GridContinuation.Items) > 0 {
		return newLibraryPage(continuation.GridContinuation.Items, continuation.GridContinuation.Continuations.token())
	}
	if len(continuation.MusicShelfContinuation.Contents) > 0 {
		return newLibraryPage(continuation.MusicShelfContinuation.Contents, continuation.MusicShelfContinuation.Continuations.token())
	}
	if len(response.OnResponseReceivedActions) > 0 {
		return newLibraryPage(response.OnResponseReceivedActions[0].AppendContinuationItemsAction.ContinuationItems, "")
	}
	return response.firstPage()
}

func (response libraryResponse) firstPage() libraryPage {
	tabs := response.Contents.SingleColumnBrowseResultsRenderer.Tabs
	if len(tabs) == 0 {
		return libraryPage{}
	}
	list := tabs[0].TabRenderer.Content.SectionListRenderer
	page := findLibraryContainer(list.Contents)
	if page.continuation == "" {
		page.continuation = list.Continuations.token()
	}
	return page
}

// findLibraryContainer returns the first grid or list among sections, looking
// inside item sections.
func findLibraryContainer(sections []librarySection) libraryPage {
	for _, section := range sections {
		if section.GridRenderer != nil {
			return newLibraryPage(section.GridRenderer.Items, section.GridRenderer.Continuations.token())
		}
		if section.MusicShelfRenderer != nil {
			return newLibraryPage(section.MusicShelfRenderer.Contents, section.MusicShelfRenderer.Continuations.token())
		}
		if section.ItemSectionRenderer == nil {
			continue
		}
		nested := findLibraryContainer(section.ItemSectionRenderer.Contents)
		if len(nested.entries) > 0 {
			return nested
		}
	}
	return libraryPage{}
}

// newLibraryPage prefers the older token next to the list and falls back to a
// trailing continuationItemRenderer entry.
func newLibraryPage(entries []libraryEntry, token string) libraryPage {
	if token == "" {
		token = trailingContinuation(entries)
	}
	return libraryPage{entries: entries, continuation: token}
}

func trailingContinuation(entries []libraryEntry) string {
	for _, entry := range entries {
		if entry.ContinuationItemRenderer != nil {
			return entry.ContinuationItemRenderer.ContinuationEndpoint.ContinuationCommand.Token
		}
	}
	return ""
}
