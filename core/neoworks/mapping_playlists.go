package neoworks

import (
	"context"
	"errors"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

type playlistEntryContent struct {
	VideoID string `json:"videoId"`
	AddedAt string `json:"addedAt"`
}

type playlistContent struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Entries     []playlistEntryContent `json:"entries,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
	ID          string                 `json:"id"`
}

// playlistMapping maps the user's own playlists, keyed by the playlist ID the
// creating install chose, which every install keeps as its local ID.
type playlistMapping struct {
	library *library.Library
}

func (playlistMapping) name() string { return "playlists" }

func (playlistMapping) optionalFields() []string { return []string{"description", "entries"} }

func (playlistMapping) keepsNodes() bool { return false }

func (playlistMapping) ownsKey(string) bool { return false }

func (mapping playlistMapping) local(ctx context.Context) ([]entity, error) {
	summaries, err := mapping.library.Playlists(ctx)
	if err != nil {
		return nil, err
	}
	entities := make([]entity, 0, len(summaries))
	for _, summary := range summaries {
		playlist, err := mapping.library.Playlist(ctx, summary.ID)
		if err != nil {
			return nil, err
		}
		item, err := playlistEntity(playlistContentOf(playlist), playlist.UpdatedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, item)
	}
	return entities, nil
}

func playlistContentOf(playlist library.Playlist) playlistContent {
	content := playlistContent{
		Name:        playlist.Name,
		Description: playlist.Description,
		CreatedAt:   library.FormatTime(playlist.CreatedAt),
		UpdatedAt:   library.FormatTime(playlist.UpdatedAt),
		ID:          playlist.ID,
	}
	for _, entry := range playlist.Entries {
		content.Entries = append(content.Entries, playlistEntryContent{
			VideoID: entry.Track.VideoID,
			AddedAt: library.FormatTime(entry.AddedAt),
		})
	}
	return content
}

func playlistEntity(content playlistContent, updatedAt time.Time) (entity, error) {
	canonical, err := canonicalJSON(content)
	if err != nil {
		return entity{}, err
	}
	return entity{key: content.ID, canonical: canonical, updatedAt: updatedAt, model: content}, nil
}

func (playlistMapping) decode(_ string, raw []byte) (entity, error) {
	content, err := decodeModel[playlistContent](raw)
	if err != nil {
		return entity{}, err
	}
	if content.ID == "" {
		return entity{}, errors.New("the playlist has no id")
	}
	content.CreatedAt, err = normalizeTime(content.CreatedAt)
	if err != nil {
		return entity{}, err
	}
	content.UpdatedAt, err = normalizeTime(content.UpdatedAt)
	if err != nil {
		return entity{}, err
	}
	for index := range content.Entries {
		content.Entries[index].AddedAt, err = normalizeTime(content.Entries[index].AddedAt)
		if err != nil {
			return entity{}, err
		}
	}
	return playlistEntity(content, parseNodeTime(content.UpdatedAt))
}

func (mapping playlistMapping) apply(ctx context.Context, remote entity) error {
	content := remote.model.(playlistContent)
	videoIDs := make([]string, 0, len(content.Entries))
	for _, entry := range content.Entries {
		videoIDs = append(videoIDs, entry.VideoID)
	}
	err := requireTracks(ctx, mapping.library, videoIDs...)
	if err != nil {
		return err
	}
	playlist := library.Playlist{
		ID:          remote.key,
		Name:        content.Name,
		Description: content.Description,
		CreatedAt:   parseNodeTime(content.CreatedAt),
		UpdatedAt:   parseNodeTime(content.UpdatedAt),
	}
	for _, entry := range content.Entries {
		playlist.Entries = append(playlist.Entries, library.PlaylistEntry{
			Track:   music.Track{VideoID: entry.VideoID},
			AddedAt: parseNodeTime(entry.AddedAt),
		})
	}
	return mapping.library.RestorePlaylist(ctx, playlist)
}

func (mapping playlistMapping) remove(ctx context.Context, key string) error {
	return mapping.library.DeletePlaylist(ctx, key)
}
