package download

import (
	"time"

	"github.com/letsmoe/drift-music/core/music"
)

// State is where a download is in its life.
type State string

// Download states. StatePaused is reported, never stored: it is what a queued
// or running download shows while the whole queue is paused or blocked.
const (
	StateQueued      State = "queued"
	StateDownloading State = "downloading"
	StateDone        State = "done"
	StateFailed      State = "failed"
	StatePaused      State = "paused"
)

// CollectionKind says where a collection's tracks come from.
type CollectionKind string

// Collection kinds.
const (
	// CollectionAlbum is a YouTube Music album, identified by browse ID.
	CollectionAlbum CollectionKind = "album"
	// CollectionPlaylist is a YouTube Music playlist, identified by browse ID.
	CollectionPlaylist CollectionKind = "playlist"
	// CollectionLocalPlaylist is a playlist in the local library.
	CollectionLocalPlaylist CollectionKind = "local"
	// CollectionLikes is the liked songs; its ID is LikedID.
	CollectionLikes CollectionKind = "likes"
)

// LikedID is the ID of the single CollectionLikes collection.
const LikedID = "liked"

// CollectionRef names a collection.
type CollectionRef struct {
	Kind CollectionKind `json:"kind"`
	ID   string         `json:"id"`
}

// Status is the download state of one video.
type Status struct {
	VideoID    string `json:"videoId"`
	State      State  `json:"state"`
	BytesDone  int64  `json:"bytesDone"`
	BytesTotal int64  `json:"bytesTotal"`
	Error      string `json:"error"`
}

// Item is a queued, running, failed or finished download with its track, so
// offline lists can render it without the network.
type Item struct {
	Track music.Track `json:"track"`
	// Kind is "song" or "episode".
	Kind   music.ItemKind `json:"kind"`
	Status Status         `json:"status"`
	// Explicit is true when the user downloaded the track itself, not only as part of a collection.
	Explicit bool `json:"explicit"`
	// Format is the MIME type of the stored audio; empty until the download finished.
	Format       string    `json:"format"`
	SizeBytes    int64     `json:"sizeBytes"`
	DownloadedAt time.Time `json:"downloadedAt"`
}

// Collection is an album, playlist or the liked songs the user downloaded.
type Collection struct {
	Kind  CollectionKind `json:"kind"`
	ID    string         `json:"id"`
	Title string         `json:"title"`
	// Keep means new tracks are downloaded and removed ones deleted as the collection changes.
	Keep        bool      `json:"keep"`
	TrackCount  int       `json:"trackCount"`
	DoneCount   int       `json:"doneCount"`
	FailedCount int       `json:"failedCount"`
	SyncedAt    time.Time `json:"syncedAt"`
}

// Summary is the queue at a glance.
type Summary struct {
	Queued      int `json:"queued"`
	Downloading int `json:"downloading"`
	Failed      int `json:"failed"`
	Done        int `json:"done"`
	// Paused is true while the user paused the queue.
	Paused bool `json:"paused"`
	// Blocked is true while the app disallowed downloads, e.g. mobile data only.
	Blocked bool `json:"blocked"`
}

// Update is what listeners receive when one download changes.
type Update struct {
	Status  Status  `json:"status"`
	Summary Summary `json:"summary"`
}

// Usage is the storage taken by finished downloads.
type Usage struct {
	Bytes int64 `json:"bytes"`
	Count int   `json:"count"`
}
