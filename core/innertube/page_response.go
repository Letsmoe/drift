package innertube

// pageResponse is a two-column browse page (album, playlist, podcast) or one of
// its continuations.
type pageResponse struct {
	Contents struct {
		TwoColumnBrowseResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer struct {
							Contents []struct {
								MusicResponsiveHeaderRenderer *pageHeader
								// Playlists the user owns wrap the same header.
								MusicEditablePlaylistDetailHeaderRenderer *struct {
									Header struct {
										MusicResponsiveHeaderRenderer *pageHeader
									}
								}
							}
						}
					}
				}
			}
			SecondaryContents struct {
				SectionListRenderer sectionList
			}
		}
	}
	ContinuationContents struct {
		MusicShelfContinuation listShelf
	}
	OnResponseReceivedActions []struct {
		AppendContinuationItemsAction struct {
			ContinuationItems []shelfItem
		}
	}
}

type pageHeader struct {
	Title struct {
		Runs []run
	}
	Subtitle struct {
		Runs []run
	}
	StraplineTextOne struct {
		Runs []run
	}
	Thumbnail   thumbnailContainer
	Description struct {
		MusicDescriptionShelfRenderer struct {
			Description struct {
				Runs []run
			}
		}
	}
	Facepile struct {
		AvatarStackViewModel struct {
			Text struct {
				Content string
			}
		}
	}
	Buttons []struct {
		MusicPlayButtonRenderer *struct {
			PlayNavigationEndpoint struct {
				WatchPlaylistEndpoint struct {
					PlaylistID string
				}
			}
		}
	}
}

// artistHeader is the immersive or visual header of an artist page.
type artistHeader struct {
	Title struct {
		Runs []run
	}
	Description struct {
		Runs []run
	}
	Thumbnail thumbnailContainer
}

// episodeRow is a podcast episode in a show's list.
type episodeRow struct {
	Title struct {
		Runs []run
	}
	Description struct {
		Runs []run
	}
	Subtitle struct {
		Runs []run
	}
	// SecondTitle links the show; only set outside the show's own page.
	SecondTitle struct {
		Runs []run
	}
	Thumbnail thumbnailContainer
	OnTap     navigationEndpoint
	// PlaybackProgress carries the duration as text, e.g. "1 hr 5 min".
	PlaybackProgress struct {
		MusicPlaybackProgressRenderer struct {
			DurationText struct {
				Runs []run
			}
		}
	}
}

// radioResponse is the answer of the next endpoint.
type radioResponse struct {
	Contents struct {
		SingleColumnMusicWatchNextResultsRenderer struct {
			TabbedRenderer struct {
				WatchNextTabbedResultsRenderer struct {
					Tabs []struct {
						TabRenderer struct {
							Content struct {
								MusicQueueRenderer struct {
									Content struct {
										PlaylistPanelRenderer struct {
											Contents []struct {
												PlaylistPanelVideoRenderer *panelVideo
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

type panelVideo struct {
	VideoID string
	Title   struct {
		Runs []run
	}
	LongBylineText struct {
		Runs []run
	}
	LengthText struct {
		Runs []run
	}
	Thumbnail struct {
		Thumbnails []thumbnail
	}
}

type suggestionsResponse struct {
	Contents []struct {
		SearchSuggestionsSectionRenderer struct {
			Contents []struct {
				SearchSuggestionRenderer *struct {
					Suggestion struct {
						Runs []run
					}
				}
			}
		}
	}
}

func (response pageResponse) header() *pageHeader {
	for _, tab := range response.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		for _, content := range tab.TabRenderer.Content.SectionListRenderer.Contents {
			if content.MusicResponsiveHeaderRenderer != nil {
				return content.MusicResponsiveHeaderRenderer
			}
			if content.MusicEditablePlaylistDetailHeaderRenderer != nil {
				return content.MusicEditablePlaylistDetailHeaderRenderer.Header.MusicResponsiveHeaderRenderer
			}
		}
	}
	return nil
}

func (response pageResponse) sections() []section {
	return response.Contents.TwoColumnBrowseResultsRenderer.SecondaryContents.SectionListRenderer.Contents
}

// listShelf returns the page's row list (album tracks, podcast episodes).
func (response pageResponse) listShelf() listShelf {
	for _, section := range response.sections() {
		if section.MusicShelfRenderer != nil {
			return *section.MusicShelfRenderer
		}
	}
	return listShelf{}
}

func (response pageResponse) playlistShelf() listShelf {
	for _, section := range response.sections() {
		if section.MusicPlaylistShelfRenderer != nil {
			return *section.MusicPlaylistShelfRenderer
		}
	}
	return listShelf{}
}

// appendedItems returns the entries a playlist continuation appends.
func (response pageResponse) appendedItems() []shelfItem {
	if len(response.OnResponseReceivedActions) == 0 {
		return nil
	}
	return response.OnResponseReceivedActions[0].AppendContinuationItemsAction.ContinuationItems
}

func (response browseResponse) artistHeader() *artistHeader {
	if response.Header.MusicImmersiveHeaderRenderer != nil {
		return response.Header.MusicImmersiveHeaderRenderer
	}
	return response.Header.MusicVisualHeaderRenderer
}

// playlistID is the playlist the header's play button starts.
func (header pageHeader) playlistID() string {
	for _, button := range header.Buttons {
		if button.MusicPlayButtonRenderer != nil {
			return button.MusicPlayButtonRenderer.PlayNavigationEndpoint.WatchPlaylistEndpoint.PlaylistID
		}
	}
	return ""
}

func (header pageHeader) description() string {
	return joinText(header.Description.MusicDescriptionShelfRenderer.Description.Runs)
}

// watchTabsResponse is the answer of the next endpoint, read for the tab endpoints only.
type watchTabsResponse struct {
	Contents struct {
		SingleColumnMusicWatchNextResultsRenderer struct {
			TabbedRenderer struct {
				WatchNextTabbedResultsRenderer struct {
					Tabs []struct {
						TabRenderer struct {
							Endpoint struct {
								BrowseEndpoint struct {
									BrowseID string
								}
							}
						}
					}
				}
			}
		}
	}
}

// lyricsResponse is the browse answer for a lyrics page (MPLYt...).
type lyricsResponse struct {
	Contents struct {
		SectionListRenderer struct {
			Contents []struct {
				MusicDescriptionShelfRenderer *struct {
					Description struct {
						Runs []run
					}
					Footer struct {
						Runs []run
					}
				}
			}
		}
	}
}
