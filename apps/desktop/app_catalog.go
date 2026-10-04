package main

import (
	"github.com/neoworks/otter-music/core/music"
)

// SearchSongs returns YouTube Music songs matching query.
func (app *App) SearchSongs(query string) ([]music.Track, error) {
	return app.innertube.SearchSongs(app.ctx, query)
}

// Search returns everything matching query, grouped in shelves, and remembers
// the query in the search history.
func (app *App) Search(query string) (music.SearchResults, error) {
	results, err := app.innertube.Search(app.ctx, query)
	if err != nil {
		return music.SearchResults{}, err
	}
	app.recordSearch(query)
	return results, nil
}

// recordSearch stores the query; a failure only costs history, so it is logged.
func (app *App) recordSearch(query string) {
	if app.library == nil {
		return
	}
	err := app.library.RecordSearch(app.ctx, query)
	if err != nil {
		app.logger.Warn("record search", "error", err)
	}
}

// SearchSuggestions returns autocomplete suggestions for a partial query.
func (app *App) SearchSuggestions(query string) ([]string, error) {
	return app.innertube.SearchSuggestions(app.ctx, query)
}

// Album returns an album page.
func (app *App) Album(browseID string) (music.AlbumPage, error) {
	return app.innertube.Album(app.ctx, browseID)
}

// Playlist returns the first page of a YouTube Music playlist.
func (app *App) Playlist(browseID string) (music.PlaylistPage, error) {
	return app.innertube.Playlist(app.ctx, browseID)
}

// PlaylistContinuation returns the playlist page that token points to.
func (app *App) PlaylistContinuation(token string) (music.PlaylistTracks, error) {
	return app.innertube.PlaylistContinuation(app.ctx, token)
}

// Artist returns an artist page.
func (app *App) Artist(browseID string) (music.ArtistPage, error) {
	return app.innertube.Artist(app.ctx, browseID)
}

// Radio returns YouTube Music's radio for a track.
func (app *App) Radio(videoID string) ([]music.Track, error) {
	return app.innertube.Radio(app.ctx, videoID)
}

// Podcast returns a podcast show with its newest episodes.
func (app *App) Podcast(browseID string) (music.PodcastPage, error) {
	return app.innertube.Podcast(app.ctx, browseID)
}

// PodcastContinuation returns the episodes page that token points to.
func (app *App) PodcastContinuation(token string) (music.PodcastEpisodes, error) {
	return app.innertube.PodcastContinuation(app.ctx, token)
}

// Home returns the first page of the home feed.
func (app *App) Home() (music.HomePage, error) {
	return app.innertube.Home(app.ctx)
}

// HomeContinuation returns the home feed page that continuation points to.
func (app *App) HomeContinuation(continuation string) (music.HomePage, error) {
	return app.innertube.HomeContinuation(app.ctx, continuation)
}
