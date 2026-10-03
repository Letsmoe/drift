// Package music holds the domain types shared by every client of the core.
package music

// Track is a playable song, identified by its YouTube video ID.
type Track struct {
	VideoID         string      `json:"videoId"`
	Title           string      `json:"title"`
	Artists         []Artist    `json:"artists"`
	Album           *Album      `json:"album,omitempty"`
	DurationSeconds int         `json:"durationSeconds"`
	Thumbnails      []Thumbnail `json:"thumbnails"`
}

// Artist is a track's credited artist. ID is empty when YouTube Music does not link the artist.
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Album is the release a track belongs to.
type Album struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Thumbnail is one size of a cover image.
type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
