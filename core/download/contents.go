package download

import (
	"context"
	"errors"
	"fmt"

	"github.com/letsmoe/drift-music/core/music"
)

// maxPlaylistPages bounds continuation fetching of a runaway playlist.
const maxPlaylistPages = 100

var (
	errNoCatalog          = errors.New("download: no catalog configured for albums and playlists")
	errUnknownCollection  = errors.New("download: unknown collection kind")
	errCollectionNotFound = errors.New("download: collection not downloaded")
)

// contents is a collection's current title and tracks.
type contents struct {
	title  string
	tracks []music.Track
}

func (manager *Manager) collectionContents(ctx context.Context, ref CollectionRef) (contents, error) {
	switch ref.Kind {
	case CollectionLikes:
		return manager.likedContents(ctx)
	case CollectionLocalPlaylist:
		return manager.localPlaylistContents(ctx, ref.ID)
	case CollectionAlbum:
		return manager.albumContents(ctx, ref.ID)
	case CollectionPlaylist:
		return manager.playlistContents(ctx, ref.ID)
	default:
		return contents{}, fmt.Errorf("%w: %q", errUnknownCollection, ref.Kind)
	}
}

func (manager *Manager) likedContents(ctx context.Context) (contents, error) {
	tracks, err := manager.library.LikedTracks(ctx)
	return contents{title: "Liked Songs", tracks: tracks}, err
}

func (manager *Manager) localPlaylistContents(ctx context.Context, id string) (contents, error) {
	playlist, err := manager.library.Playlist(ctx, id)
	if err != nil {
		return contents{}, err
	}
	tracks := make([]music.Track, 0, len(playlist.Entries))
	for _, entry := range playlist.Entries {
		tracks = append(tracks, entry.Track)
	}
	return contents{title: playlist.Name, tracks: tracks}, nil
}

func (manager *Manager) albumContents(ctx context.Context, id string) (contents, error) {
	if manager.catalog == nil {
		return contents{}, errNoCatalog
	}
	page, err := manager.catalog.Album(ctx, id)
	if err != nil {
		return contents{}, fmt.Errorf("download: read album %s: %w", id, err)
	}
	return contents{title: page.Title, tracks: page.Tracks}, nil
}

func (manager *Manager) playlistContents(ctx context.Context, id string) (contents, error) {
	if manager.catalog == nil {
		return contents{}, errNoCatalog
	}
	page, err := manager.catalog.Playlist(ctx, id)
	if err != nil {
		return contents{}, fmt.Errorf("download: read playlist %s: %w", id, err)
	}
	tracks, err := manager.remainingPlaylistTracks(ctx, page)
	return contents{title: page.Title, tracks: tracks}, err
}

func (manager *Manager) remainingPlaylistTracks(ctx context.Context, page music.PlaylistPage) ([]music.Track, error) {
	tracks := page.Tracks
	token := page.Continuation
	for range maxPlaylistPages {
		if token == "" {
			return tracks, nil
		}
		next, err := manager.catalog.PlaylistContinuation(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("download: read playlist %s: %w", page.ID, err)
		}
		tracks = append(tracks, next.Tracks...)
		token = next.Continuation
	}
	return tracks, nil
}
