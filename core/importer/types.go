// Package importer brings a Spotify library into the local one: it matches
// Spotify tracks, albums, artists and podcasts to YouTube Music through the
// catalog, remembers the matches in the library, and writes likes, playlists
// and saved items. Package spotify supplies the source data.
package importer

import "github.com/letsmoe/drift/core/spotify"

// Source kinds.
const (
	// SourceExport is a Spotify data export: a zip, a folder or a JSON file.
	SourceExport = "export"
	// SourceURL is the link of a public playlist or album.
	SourceURL = "url"
	// SourceYouTubeAccount is the signed-in YouTube Music account; its items
	// already are YouTube Music items, so nothing is matched.
	SourceYouTubeAccount = "youtube-account"
)

// matchSource is the import_match/import_playlist source name of Spotify data.
const matchSource = "spotify"

// Source says where the Spotify data comes from.
type Source struct {
	Kind string `json:"kind"`
	// Location is a file or folder path for SourceExport, a link for SourceURL.
	Location string `json:"location"`
}

// PlaylistSummary is one importable playlist.
type PlaylistSummary struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	TrackCount int    `json:"trackCount"`
}

// Summary counts what a source holds, without any matching.
type Summary struct {
	Liked     int               `json:"liked"`
	Playlists []PlaylistSummary `json:"playlists"`
	Albums    int               `json:"albums"`
	Artists   int               `json:"artists"`
	Shows     int               `json:"shows"`
	// TrackLimit is set when the source shows at most that many tracks per
	// playlist (public links), so a playlist may be longer on Spotify.
	TrackLimit int `json:"trackLimit"`
}

// Options select what Run imports.
type Options struct {
	Liked bool `json:"liked"`
	// Playlists holds the keys of the playlists to import (PlaylistSummary.Key).
	Playlists []string `json:"playlists"`
	Albums    bool     `json:"albums"`
	Artists   bool     `json:"artists"`
	Shows     bool     `json:"shows"`
	// RetryUnmatched searches again for items an earlier import did not find.
	RetryUnmatched bool `json:"retryUnmatched"`
}

// Progress phases.
const (
	PhaseTracks  = "tracks"
	PhaseAlbums  = "albums"
	PhaseArtists = "artists"
	PhaseShows   = "shows"
	PhaseSaving  = "saving"
)

// Progress reports a running import. Done and Total count the items to
// match across all phases; Current names the item last handled.
type Progress struct {
	Phase   string `json:"phase"`
	Done    int    `json:"done"`
	Total   int    `json:"total"`
	Current string `json:"current"`
}

// CategoryCount is how many items of a category were matched.
type CategoryCount struct {
	Matched   int `json:"matched"`
	Unmatched int `json:"unmatched"`
}

// PlaylistResult is the outcome for one imported playlist.
type PlaylistResult struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	PlaylistID string `json:"playlistId"`
	Matched    int    `json:"matched"`
	Unmatched  int    `json:"unmatched"`
}

// UnmatchedItem is an item YouTube Music had no match for.
type UnmatchedItem struct {
	// Kind is "track", "album", "artist", "show" or "playlist".
	Kind    string   `json:"kind"`
	Title   string   `json:"title"`
	Artists []string `json:"artists"`
	// Playlist names the playlist a track belongs to; empty for liked tracks and other kinds.
	Playlist string `json:"playlist"`
}

// Result is what an import did.
type Result struct {
	Liked     CategoryCount    `json:"liked"`
	Albums    CategoryCount    `json:"albums"`
	Artists   CategoryCount    `json:"artists"`
	Shows     CategoryCount    `json:"shows"`
	Playlists []PlaylistResult `json:"playlists"`
	Unmatched []UnmatchedItem  `json:"unmatched"`
}

// selection is the part of a source library an import acts on.
type selection struct {
	liked     []spotify.Track
	playlists []spotify.Playlist
	albums    []spotify.Album
	artists   []spotify.Artist
	shows     []spotify.Show
}
