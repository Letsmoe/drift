package mobile

// Home returns the first page of the home feed as music.HomePage JSON.
func (core *Core) Home() (string, error) {
	page, err := core.innertube.Home(core.ctx)
	return encodeResult(page, err)
}

// HomeContinuation returns the home feed page that token points to.
func (core *Core) HomeContinuation(token string) (string, error) {
	page, err := core.innertube.HomeContinuation(core.ctx, token)
	return encodeResult(page, err)
}

// Search returns music.SearchResults JSON and remembers the query in the search history.
func (core *Core) Search(query string) (string, error) {
	results, err := core.innertube.Search(core.ctx, query)
	if err != nil {
		return "", err
	}
	core.recordSearch(query)
	return encode(results)
}

// recordSearch stores the query; a failure only costs history, so it is logged.
func (core *Core) recordSearch(query string) {
	if core.library == nil {
		return
	}
	err := core.library.RecordSearch(core.ctx, query)
	if err != nil {
		core.logger.Warn("record search", "error", err)
	}
}

// SearchSuggestions returns autocomplete suggestions as a JSON string array.
func (core *Core) SearchSuggestions(query string) (string, error) {
	suggestions, err := core.innertube.SearchSuggestions(core.ctx, query)
	return encodeResult(suggestions, err)
}

// Album returns music.AlbumPage JSON.
func (core *Core) Album(browseID string) (string, error) {
	page, err := core.innertube.Album(core.ctx, browseID)
	return encodeResult(page, err)
}

// Playlist returns music.PlaylistPage JSON.
func (core *Core) Playlist(browseID string) (string, error) {
	page, err := core.innertube.Playlist(core.ctx, browseID)
	return encodeResult(page, err)
}

// PlaylistContinuation returns music.PlaylistTracks JSON.
func (core *Core) PlaylistContinuation(token string) (string, error) {
	tracks, err := core.innertube.PlaylistContinuation(core.ctx, token)
	return encodeResult(tracks, err)
}

// Artist returns music.ArtistPage JSON.
func (core *Core) Artist(browseID string) (string, error) {
	page, err := core.innertube.Artist(core.ctx, browseID)
	return encodeResult(page, err)
}

// Radio returns YouTube Music's radio for a track as a music.Track JSON array.
func (core *Core) Radio(videoID string) (string, error) {
	tracks, err := core.innertube.Radio(core.ctx, videoID)
	return encodeResult(tracks, err)
}

// Lyrics returns music.Lyrics JSON, or "null" when the track has no lyrics.
func (core *Core) Lyrics(videoID string) (string, error) {
	lyrics, found, err := core.innertube.Lyrics(core.ctx, videoID)
	if err != nil || !found {
		return encodeResult(nil, err)
	}
	return encodeResult(lyrics, nil)
}

// Podcast returns music.PodcastPage JSON.
func (core *Core) Podcast(browseID string) (string, error) {
	page, err := core.innertube.Podcast(core.ctx, browseID)
	return encodeResult(page, err)
}

// PodcastContinuation returns music.PodcastEpisodes JSON.
func (core *Core) PodcastContinuation(token string) (string, error) {
	episodes, err := core.innertube.PodcastContinuation(core.ctx, token)
	return encodeResult(episodes, err)
}
