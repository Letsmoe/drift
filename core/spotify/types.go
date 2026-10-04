// Package spotify reads the user's Spotify data without an account: the data
// export (a zip, a folder or single JSON files) and public playlist and album
// links. It only parses and fetches; matching to YouTube Music lives in
// package importer.
package spotify

import "strings"

// Track is a song as Spotify names it. DurationMillis is 0 when unknown: the
// data export carries no durations.
type Track struct {
	URI            string   `json:"uri"`
	Title          string   `json:"title"`
	Artists        []string `json:"artists"`
	Album          string   `json:"album"`
	DurationMillis int      `json:"durationMillis"`
}

// Key identifies the track across imports: its URI, or its text when the
// export left the URI out.
func (track Track) Key() string {
	if track.URI != "" {
		return track.URI
	}
	return textKey("track", track.Title, strings.Join(track.Artists, ","))
}

// Playlist is a Spotify playlist with its tracks in order.
type Playlist struct {
	// Key identifies the playlist across imports and is unique within a Library.
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Tracks      []Track `json:"tracks"`
}

// Album is a saved album.
type Album struct {
	URI    string `json:"uri"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

// Key identifies the album across imports.
func (album Album) Key() string {
	if album.URI != "" {
		return album.URI
	}
	return textKey("album", album.Title, album.Artist)
}

// Artist is a followed artist.
type Artist struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// Key identifies the artist across imports.
func (artist Artist) Key() string {
	if artist.URI != "" {
		return artist.URI
	}
	return textKey("artist", artist.Name, "")
}

// Show is a saved podcast.
type Show struct {
	URI       string `json:"uri"`
	Name      string `json:"name"`
	Publisher string `json:"publisher"`
}

// Key identifies the show across imports.
func (show Show) Key() string {
	if show.URI != "" {
		return show.URI
	}
	return textKey("show", show.Name, show.Publisher)
}

// Library is everything a Spotify source holds.
type Library struct {
	LikedTracks []Track    `json:"likedTracks"`
	Playlists   []Playlist `json:"playlists"`
	Albums      []Album    `json:"albums"`
	Artists     []Artist   `json:"artists"`
	Shows       []Show     `json:"shows"`
}

func (library Library) isEmpty() bool {
	return len(library.LikedTracks)+len(library.Playlists)+len(library.Albums)+len(library.Artists)+len(library.Shows) == 0
}

func textKey(kind, primary, secondary string) string {
	return "text:" + kind + ":" + strings.ToLower(strings.TrimSpace(primary)) + "|" + strings.ToLower(strings.TrimSpace(secondary))
}
