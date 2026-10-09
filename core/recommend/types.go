package recommend

import (
	"context"
	"strings"

	"github.com/letsmoe/drift/core/music"
)

// Catalog is the part of YouTube Music the engine reads. The innertube client
// satisfies it.
type Catalog interface {
	// Radio is YouTube Music's radio for a track: its collaborative filtering.
	Radio(ctx context.Context, videoID string) ([]music.Track, error)
	// Artist is an artist page: top songs plus shelves (albums, related artists, ...).
	Artist(ctx context.Context, browseID string) (music.ArtistPage, error)
	// SearchSongs searches for songs only.
	SearchSongs(ctx context.Context, query string) ([]music.Track, error)
	// Home is YouTube's own home feed, personalized by visitor.
	Home(ctx context.Context) (music.HomePage, error)
}

// Candidate source names. Every recommended item carries the one it was
// served from, and impressions echo it back.
const (
	SourceRadio      = "related:radio"
	SourceArtist     = "affinity:artist"
	SourceExplore    = "explore:artist"
	SourceRediscover = "rediscover"
	SourceSearch     = "search"
	SourceYouTube    = "youtube"
)

// Reason kinds: why an item was recommended, and what Reason.EntityID points at.
const (
	ReasonRelated    = "related"    // seed track video ID
	ReasonArtist     = "artist"     // artist browse ID
	ReasonExplore    = "explore"    // artist browse ID
	ReasonRediscover = "rediscover" // none
	ReasonSearch     = "search"     // the search query
	ReasonYouTube    = "youtube"    // none
)

// Reason explains one recommendation, e.g. "Because you listened to X".
type Reason struct {
	Kind string `json:"kind"`
	// EntityID is the track, artist or search query behind the reason; empty
	// for kinds without one.
	EntityID string `json:"entityId"`
	// EntityName is the entity's display name (seed title, artist name, query).
	EntityName string `json:"entityName"`
	// Text is the human-readable explanation.
	Text string `json:"text"`
}

// Key identifies the reason's group, used to build rows.
func (reason Reason) Key() string {
	if reason.EntityID == "" {
		return reason.Kind
	}
	return reason.Kind + ":" + reason.EntityID
}

// Item is one ranked recommendation.
type Item struct {
	Track  music.Track `json:"track"`
	Source string      `json:"source"`
	Reason Reason      `json:"reason"`
	Score  float64     `json:"score"`
}

// Row is a titled group of items sharing a reason.
type Row struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Reason Reason `json:"reason"`
	Items  []Item `json:"items"`
}

// isExploration reports whether items from source fill the reserved
// exploration slots.
func isExploration(source string) bool {
	return strings.HasPrefix(source, "explore:")
}

// relatedReason explains a radio track by the seed it came from.
func relatedReason(seed Seed) Reason {
	return Reason{Kind: ReasonRelated, EntityID: seed.VideoID, EntityName: seed.Title, Text: "Because you listened to " + seed.Title}
}

// artistReason explains a track by the artist the user likes.
func artistReason(artistID, artistName string) Reason {
	return Reason{Kind: ReasonArtist, EntityID: artistID, EntityName: artistName, Text: "More from " + artistName}
}

// exploreReason explains a track by a new artist worth trying.
func exploreReason(artistID, artistName string) Reason {
	return Reason{Kind: ReasonExplore, EntityID: artistID, EntityName: artistName, Text: "Explore " + artistName}
}

// rediscoverReason explains a loved track that has not been played lately.
func rediscoverReason() Reason {
	return Reason{Kind: ReasonRediscover, Text: "Rediscover"}
}

// searchReason explains a track by the recent search it answers.
func searchReason(query string) Reason {
	return Reason{Kind: ReasonSearch, EntityID: query, EntityName: query, Text: "From your search “" + query + "”"}
}

// youTubeReason explains a track taken from YouTube Music's own home feed.
func youTubeReason() Reason {
	return Reason{Kind: ReasonYouTube, Text: "Popular on YouTube Music"}
}
