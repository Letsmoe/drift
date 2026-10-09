package mobile

import (
	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

// resumePosition is where an unfinished episode should pick up.
type resumePosition struct {
	Seconds int  `json:"seconds"`
	Found   bool `json:"found"`
}

// SetRating rates a track (music.Track JSON): 1 like, -1 dislike, 0 clears the rating.
func (core *Core) SetRating(trackJSON string, kind string, rating int) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	var track music.Track
	err = decode(trackJSON, &track)
	if err != nil {
		return err
	}
	return lib.SetRating(core.ctx, track, music.ItemKind(kind), library.Rating(rating))
}

// Ratings takes a JSON array of video IDs and returns {"videoId": rating}; unrated videos are absent.
func (core *Core) Ratings(videoIDsJSON string) (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	var videoIDs []string
	err = decode(videoIDsJSON, &videoIDs)
	if err != nil {
		return "", err
	}
	ratings, err := lib.Ratings(core.ctx, videoIDs)
	if err != nil {
		return "", err
	}
	plainRatings := make(map[string]int, len(ratings))
	for videoID, rating := range ratings {
		plainRatings[videoID] = int(rating)
	}
	return encode(plainRatings)
}

// LikedTracks returns every liked track, newest like first.
func (core *Core) LikedTracks() (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	tracks, err := lib.LikedTracks(core.ctx)
	return encodeResult(tracks, err)
}

// Playlists lists the user's playlists as library.PlaylistSummary JSON.
func (core *Core) Playlists() (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	playlists, err := lib.Playlists(core.ctx)
	return encodeResult(playlists, err)
}

// LibraryPlaylist returns a user playlist with its entries.
func (core *Core) LibraryPlaylist(id string) (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	playlist, err := lib.Playlist(core.ctx, id)
	return encodeResult(playlist, err)
}

// CreatePlaylist makes an empty playlist.
func (core *Core) CreatePlaylist(name string) (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	playlist, err := lib.CreatePlaylist(core.ctx, name)
	return encodeResult(playlist, err)
}

// RenamePlaylist changes a playlist's name and description.
func (core *Core) RenamePlaylist(id, name, description string) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.RenamePlaylist(core.ctx, id, name, description)
}

// DeletePlaylist removes a playlist and its entries.
func (core *Core) DeletePlaylist(id string) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.DeletePlaylist(core.ctx, id)
}

// AddToPlaylist appends tracks (music.Track JSON array) to a playlist.
func (core *Core) AddToPlaylist(id string, tracksJSON string) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	var tracks []music.Track
	err = decode(tracksJSON, &tracks)
	if err != nil {
		return err
	}
	return lib.AddToPlaylist(core.ctx, id, tracks)
}

// RemoveFromPlaylist removes one entry from a playlist.
func (core *Core) RemoveFromPlaylist(id string, entryID int64) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.RemoveFromPlaylist(core.ctx, id, entryID)
}

// MovePlaylistEntry moves an entry to newIndex within its playlist.
func (core *Core) MovePlaylistEntry(id string, entryID int64, newIndex int) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.MovePlaylistEntry(core.ctx, id, entryID, newIndex)
}

// Save adds an album, playlist, artist or podcast (music.Item JSON) to the user's library.
func (core *Core) Save(itemJSON string) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	var item music.Item
	err = decode(itemJSON, &item)
	if err != nil {
		return err
	}
	return lib.Save(core.ctx, item)
}

// Unsave removes an item from the user's library.
func (core *Core) Unsave(kind string, id string) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.Unsave(core.ctx, music.ItemKind(kind), id)
}

// IsSaved reports whether an item is in the user's library.
func (core *Core) IsSaved(kind string, id string) (bool, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return false, err
	}
	return lib.IsSaved(core.ctx, music.ItemKind(kind), id)
}

// SavedItems lists everything the user saved as library.SavedItem JSON.
func (core *Core) SavedItems() (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	items, err := lib.SavedItems(core.ctx)
	return encodeResult(items, err)
}

// StartPlay records the start of a playback and returns the play ID for UpdatePlay.
func (core *Core) StartPlay(trackJSON string, kind string, playContext string) (int64, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return 0, err
	}
	var track music.Track
	err = decode(trackJSON, &track)
	if err != nil {
		return 0, err
	}
	return lib.StartPlay(core.ctx, track, music.ItemKind(kind), playContext)
}

// UpdatePlay reports playback progress of a play started with StartPlay.
func (core *Core) UpdatePlay(playID int64, progressSeconds int, completed bool) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.UpdatePlay(core.ctx, playID, progressSeconds, completed)
}

// ResumePosition returns {"seconds", "found"} JSON; found is false when there is nothing to resume.
func (core *Core) ResumePosition(videoID string) (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	seconds, found, err := lib.ResumePosition(core.ctx, videoID)
	if err != nil {
		return "", err
	}
	return encode(resumePosition{Seconds: seconds, Found: found})
}

// RecentTracks returns the most recently played tracks.
func (core *Core) RecentTracks(limit int) (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	tracks, err := lib.RecentTracks(core.ctx, limit)
	return encodeResult(tracks, err)
}

// RecentSearches returns the latest search queries as a JSON string array, newest first.
func (core *Core) RecentSearches(limit int) (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	searches, err := lib.RecentSearches(core.ctx, limit)
	return encodeResult(searches, err)
}

// ClearSearches forgets the search history.
func (core *Core) ClearSearches() error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	return lib.ClearSearches(core.ctx)
}

// RecordImpressions stores recommendation items (library.Impression JSON array) that were shown or clicked.
func (core *Core) RecordImpressions(impressionsJSON string) error {
	lib, err := core.requireLibrary()
	if err != nil {
		return err
	}
	var impressions []library.Impression
	err = decode(impressionsJSON, &impressions)
	if err != nil {
		return err
	}
	return lib.RecordImpressions(core.ctx, impressions)
}
