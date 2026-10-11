package neoworks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

// errTrackNotSynced means an entity refers to a track whose @drift/tracks node
// has not arrived. The entity is applied on a later pass.
var errTrackNotSynced = errors.New("the track is not in the library yet")

type artistContent struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type albumContent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type trackContent struct {
	VideoID         string          `json:"videoId"`
	Title           string          `json:"title"`
	Artists         []artistContent `json:"artists,omitempty"`
	Album           *albumContent   `json:"album,omitempty"`
	DurationSeconds int             `json:"durationSeconds"`
	ThumbnailURL    string          `json:"thumbnailUrl,omitempty"`
	Kind            string          `json:"kind"`
}

// trackMapping maps every cached track that another collection refers to.
// Tracks are a shared cache: their nodes are never deleted.
type trackMapping struct {
	library *library.Library
}

func (trackMapping) name() string { return "tracks" }

func (trackMapping) optionalFields() []string {
	return []string{"artists", "album", "thumbnailUrl"}
}

func (trackMapping) keepsNodes() bool { return true }

func (trackMapping) ownsKey(string) bool { return false }

func (mapping trackMapping) local(ctx context.Context) ([]entity, error) {
	records, err := mapping.library.ReferencedTracks(ctx)
	if err != nil {
		return nil, err
	}
	entities := make([]entity, 0, len(records))
	for _, record := range records {
		item, err := trackEntity(trackContentOf(record), record.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, item)
	}
	return entities, nil
}

func trackContentOf(record library.TrackRecord) trackContent {
	track := record.Track
	content := trackContent{
		VideoID:         track.VideoID,
		Title:           track.Title,
		DurationSeconds: track.DurationSeconds,
		ThumbnailURL:    syncedThumbnailURL(track.Thumbnails),
		Kind:            string(syncedTrackKind(record.Kind)),
	}
	for _, artist := range track.Artists {
		content.Artists = append(content.Artists, artistContent{ID: artist.ID, Name: artist.Name})
	}
	if track.Album != nil {
		content.Album = &albumContent{ID: track.Album.ID, Name: track.Album.Name}
	}
	return content
}

func trackEntity(content trackContent, updatedAt time.Time) (entity, error) {
	canonical, err := canonicalJSON(content)
	if err != nil {
		return entity{}, err
	}
	return entity{key: content.VideoID, canonical: canonical, updatedAt: updatedAt, model: content}, nil
}

func (trackMapping) decode(_ string, raw []byte) (entity, error) {
	content, err := decodeModel[trackContent](raw)
	if err != nil {
		return entity{}, err
	}
	return trackEntity(content, time.Time{})
}

func (mapping trackMapping) apply(ctx context.Context, remote entity) error {
	content := remote.model.(trackContent)
	track := music.Track{
		VideoID:         content.VideoID,
		Title:           content.Title,
		DurationSeconds: content.DurationSeconds,
		Thumbnails:      rebuildThumbnails(content.ThumbnailURL),
		Artists:         []music.Artist{},
	}
	for _, artist := range content.Artists {
		track.Artists = append(track.Artists, music.Artist{ID: artist.ID, Name: artist.Name})
	}
	if content.Album != nil {
		track.Album = &music.Album{ID: content.Album.ID, Name: content.Album.Name}
	}
	return mapping.library.SaveTracks(ctx, []music.Track{track}, syncedTrackKind(music.ItemKind(content.Kind)))
}

// syncedTrackKind narrows a kind to the two the tracks collection knows.
func syncedTrackKind(kind music.ItemKind) music.ItemKind {
	if kind == music.ItemKindEpisode {
		return music.ItemKindEpisode
	}
	return music.ItemKindSong
}

// remove keeps the cached track: other rows may still refer to it.
func (trackMapping) remove(context.Context, string) error { return nil }

// requireTracks fails with errTrackNotSynced when a track is not cached.
func requireTracks(ctx context.Context, store *library.Library, videoIDs ...string) error {
	cached, err := store.Tracks(ctx, videoIDs)
	if err != nil {
		return err
	}
	for _, videoID := range videoIDs {
		if _, found := cached[videoID]; !found {
			return fmt.Errorf("%s: %w", videoID, errTrackNotSynced)
		}
	}
	return nil
}
