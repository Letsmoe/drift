package music

// ItemKind says what a shelf item is: a song plays, the others open a page.
type ItemKind string

// Item kinds.
const (
	ItemKindSong     ItemKind = "song"
	ItemKindAlbum    ItemKind = "album"
	ItemKindPlaylist ItemKind = "playlist"
	ItemKindArtist   ItemKind = "artist"
)

// Item is one entry on a shelf.
type Item struct {
	Kind ItemKind `json:"kind"`
	// ID is the video ID for songs and the browse ID for albums, playlists and artists.
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Subtitle   string      `json:"subtitle"`
	Thumbnails []Thumbnail `json:"thumbnails"`
	// Track is set for songs only.
	Track *Track `json:"track,omitempty"`
}

// Shelf is a titled row of items, e.g. "Quick picks" or "New releases".
type Shelf struct {
	Title string `json:"title"`
	Items []Item `json:"items"`
}

// HomePage is one page of the home feed.
type HomePage struct {
	Shelves []Shelf `json:"shelves"`
	// Continuation is an opaque token for the next page; empty on the last page.
	Continuation string `json:"continuation"`
}
