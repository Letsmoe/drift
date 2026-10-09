package main

import (
	"errors"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

var errLibraryUnavailable = errors.New("the local library could not be opened")

// requireLibrary returns the open library, or an error when startup failed to open it.
func (app *App) requireLibrary() (*library.Library, error) {
	if app.library == nil {
		return nil, errLibraryUnavailable
	}
	return app.library, nil
}

// SetRating rates a track: 1 like, -1 dislike, 0 clears the rating. A dislike
// also drops the track from the current recommendations.
func (app *App) SetRating(track music.Track, kind string, rating int) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	err = lib.SetRating(app.ctx, track, music.ItemKind(kind), library.Rating(rating))
	if err != nil || library.Rating(rating) != library.RatingDislike || app.engine == nil {
		return err
	}
	return app.engine.ApplyExclusions(app.ctx)
}

// Ratings returns the ratings of the given videos; unrated videos are absent.
func (app *App) Ratings(videoIDs []string) (map[string]int, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return nil, err
	}
	ratings, err := lib.Ratings(app.ctx, videoIDs)
	if err != nil {
		return nil, err
	}
	plainRatings := make(map[string]int, len(ratings))
	for videoID, rating := range ratings {
		plainRatings[videoID] = int(rating)
	}
	return plainRatings, nil
}

// LikedTracks returns every liked track, newest like first.
func (app *App) LikedTracks() ([]music.Track, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return nil, err
	}
	return lib.LikedTracks(app.ctx)
}

// Playlists lists the user's playlists.
func (app *App) Playlists() ([]library.PlaylistSummary, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return nil, err
	}
	return lib.Playlists(app.ctx)
}

// LibraryPlaylist returns a user playlist with its entries.
func (app *App) LibraryPlaylist(id string) (library.Playlist, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return library.Playlist{}, err
	}
	return lib.Playlist(app.ctx, id)
}

// CreatePlaylist makes an empty playlist.
func (app *App) CreatePlaylist(name string) (library.Playlist, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return library.Playlist{}, err
	}
	return lib.CreatePlaylist(app.ctx, name)
}

// RenamePlaylist changes a playlist's name and description.
func (app *App) RenamePlaylist(id, name, description string) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.RenamePlaylist(app.ctx, id, name, description)
}

// DeletePlaylist removes a playlist and its entries.
func (app *App) DeletePlaylist(id string) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.DeletePlaylist(app.ctx, id)
}

// AddToPlaylist appends tracks to a playlist.
func (app *App) AddToPlaylist(id string, tracks []music.Track) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.AddToPlaylist(app.ctx, id, tracks)
}

// RemoveFromPlaylist removes one entry from a playlist.
func (app *App) RemoveFromPlaylist(id string, entryID int64) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.RemoveFromPlaylist(app.ctx, id, entryID)
}

// MovePlaylistEntry moves an entry to newIndex within its playlist.
func (app *App) MovePlaylistEntry(id string, entryID int64, newIndex int) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.MovePlaylistEntry(app.ctx, id, entryID, newIndex)
}

// Save adds an album, playlist, artist or podcast to the user's library.
func (app *App) Save(item music.Item) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.Save(app.ctx, item)
}

// Unsave removes an item from the user's library.
func (app *App) Unsave(kind string, id string) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.Unsave(app.ctx, music.ItemKind(kind), id)
}

// IsSaved reports whether an item is in the user's library.
func (app *App) IsSaved(kind string, id string) (bool, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return false, err
	}
	return lib.IsSaved(app.ctx, music.ItemKind(kind), id)
}

// SavedItems lists everything the user saved.
func (app *App) SavedItems() ([]library.SavedItem, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return nil, err
	}
	return lib.SavedItems(app.ctx)
}

// StartPlay records the start of a playback and returns the play ID for UpdatePlay.
func (app *App) StartPlay(track music.Track, kind string, playContext string) (int64, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return 0, err
	}
	return lib.StartPlay(app.ctx, track, music.ItemKind(kind), playContext)
}

// UpdatePlay reports playback progress of a play started with StartPlay.
func (app *App) UpdatePlay(playID int64, progressSeconds int, completed bool) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.UpdatePlay(app.ctx, playID, progressSeconds, completed)
}

// ResumePosition is where an unfinished episode should pick up.
func (app *App) ResumePosition(videoID string) (ResumePosition, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return ResumePosition{}, err
	}
	seconds, found, err := lib.ResumePosition(app.ctx, videoID)
	if err != nil {
		return ResumePosition{}, err
	}
	return ResumePosition{Seconds: seconds, Found: found}, nil
}

// ResumePosition is the result of ResumePosition; Found is false when there is nothing to resume.
type ResumePosition struct {
	Seconds int  `json:"seconds"`
	Found   bool `json:"found"`
}

// RecentTracks returns the most recently played tracks.
func (app *App) RecentTracks(limit int) ([]music.Track, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return nil, err
	}
	return lib.RecentTracks(app.ctx, limit)
}

// RecentSearches returns the latest search queries, newest first.
func (app *App) RecentSearches(limit int) ([]string, error) {
	lib, err := app.requireLibrary()
	if err != nil {
		return nil, err
	}
	return lib.RecentSearches(app.ctx, limit)
}

// ClearSearches forgets the search history.
func (app *App) ClearSearches() error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.ClearSearches(app.ctx)
}

// RecordImpressions stores recommendation items that were shown or clicked.
func (app *App) RecordImpressions(impressions []library.Impression) error {
	lib, err := app.requireLibrary()
	if err != nil {
		return err
	}
	return lib.RecordImpressions(app.ctx, impressions)
}
