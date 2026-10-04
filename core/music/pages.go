package music

// SearchResults is what a search finds, grouped like YouTube Music groups it:
// a top result shelf, then songs, albums, artists, playlists, ...
type SearchResults struct {
	Shelves []Shelf `json:"shelves"`
}

// AlbumPage is an album with its tracks. Every track carries the album and
// its cover, which album rows on YouTube Music leave out.
type AlbumPage struct {
	ID string `json:"id"`
	// PlaylistID plays the whole album as a radio-free playlist; may be empty.
	PlaylistID string      `json:"playlistId"`
	Title      string      `json:"title"`
	Subtitle   string      `json:"subtitle"`
	Artists    []Artist    `json:"artists"`
	Year       string      `json:"year"`
	Thumbnails []Thumbnail `json:"thumbnails"`
	Tracks     []Track     `json:"tracks"`
}

// PlaylistPage is a YouTube Music playlist. Long playlists come in pages; see
// Continuation.
type PlaylistPage struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Author      string      `json:"author"`
	Description string      `json:"description"`
	Thumbnails  []Thumbnail `json:"thumbnails"`
	Tracks      []Track     `json:"tracks"`
	// Continuation is an opaque token for more tracks; empty when there are none.
	Continuation string `json:"continuation"`
}

// PlaylistTracks is a further page of a playlist's tracks.
type PlaylistTracks struct {
	Tracks       []Track `json:"tracks"`
	Continuation string  `json:"continuation"`
}

// ArtistPage is an artist with their popular songs and the shelves below
// (albums, singles, videos, related artists, ...).
type ArtistPage struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Thumbnails  []Thumbnail `json:"thumbnails"`
	TopSongs    []Track     `json:"topSongs"`
	Shelves     []Shelf     `json:"shelves"`
}

// PodcastPage is a podcast show with its newest episodes.
type PodcastPage struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Author      Artist      `json:"author"`
	Description string      `json:"description"`
	Thumbnails  []Thumbnail `json:"thumbnails"`
	Episodes    []Episode   `json:"episodes"`
	// Continuation is an opaque token for older episodes; empty when there are none.
	Continuation string `json:"continuation"`
}

// PodcastEpisodes is a further page of a podcast's episodes.
type PodcastEpisodes struct {
	Episodes     []Episode `json:"episodes"`
	Continuation string    `json:"continuation"`
}

// Episode is one podcast episode. Track plays it: its Artists hold the show's
// author and its Album the show.
type Episode struct {
	Track       Track  `json:"track"`
	Description string `json:"description"`
	// Published is YouTube's localized date text, e.g. "Mar 3, 2026" or "3 days ago".
	Published string `json:"published"`
}
