package neoworks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

type savedContent struct {
	Kind         string `json:"kind"`
	BrowseID     string `json:"browseId"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	SavedAt      string `json:"savedAt"`
}

// syncedSavedKinds are the kinds @drift/saved holds.
var syncedSavedKinds = map[music.ItemKind]bool{
	music.ItemKindAlbum:    true,
	music.ItemKindArtist:   true,
	music.ItemKindPlaylist: true,
	music.ItemKindPodcast:  true,
}

// savedMapping maps saved albums, artists, podcasts and YouTube playlists,
// keyed by kind and browse ID.
type savedMapping struct {
	library *library.Library
}

func (savedMapping) name() string { return "saved" }

func (savedMapping) optionalFields() []string { return []string{"subtitle", "thumbnailUrl"} }

func (savedMapping) keepsNodes() bool { return false }

func (savedMapping) ownsKey(string) bool { return false }

func savedKey(kind, browseID string) string {
	return kind + ":" + browseID
}

func (mapping savedMapping) local(ctx context.Context) ([]entity, error) {
	saved, err := mapping.library.SavedItems(ctx)
	if err != nil {
		return nil, err
	}
	entities := []entity{}
	for _, item := range saved {
		if !syncedSavedKinds[item.Item.Kind] {
			continue
		}
		content := savedContent{
			Kind:         string(item.Item.Kind),
			BrowseID:     item.Item.ID,
			Title:        item.Item.Title,
			Subtitle:     item.Item.Subtitle,
			ThumbnailURL: syncedThumbnailURL(item.Item.Thumbnails),
			SavedAt:      library.FormatTime(item.SavedAt),
		}
		entry, err := savedEntity(content, item.SavedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entry)
	}
	return entities, nil
}

func savedEntity(content savedContent, savedAt time.Time) (entity, error) {
	canonical, err := canonicalJSON(content)
	if err != nil {
		return entity{}, err
	}
	return entity{key: savedKey(content.Kind, content.BrowseID), canonical: canonical, updatedAt: savedAt, model: content}, nil
}

func (savedMapping) decode(_ string, raw []byte) (entity, error) {
	content, err := decodeModel[savedContent](raw)
	if err != nil {
		return entity{}, err
	}
	content.SavedAt, err = normalizeTime(content.SavedAt)
	if err != nil {
		return entity{}, err
	}
	return savedEntity(content, parseNodeTime(content.SavedAt))
}

func (mapping savedMapping) apply(ctx context.Context, remote entity) error {
	content := remote.model.(savedContent)
	kind := music.ItemKind(content.Kind)
	if !syncedSavedKinds[kind] {
		return fmt.Errorf("unknown saved kind %q", content.Kind)
	}
	return mapping.library.RestoreSavedItem(ctx, library.SavedItem{
		Item: music.Item{
			Kind:       kind,
			ID:         content.BrowseID,
			Title:      content.Title,
			Subtitle:   content.Subtitle,
			Thumbnails: rebuildThumbnails(content.ThumbnailURL),
		},
		SavedAt: remote.updatedAt,
	})
}

func (mapping savedMapping) remove(ctx context.Context, key string) error {
	kind, browseID, err := splitKey(key)
	if err != nil {
		return err
	}
	return mapping.library.Unsave(ctx, music.ItemKind(kind), browseID)
}

// splitKey splits a "kind:id" key; the id may contain colons.
func splitKey(key string) (kind, id string, err error) {
	kind, id, found := strings.Cut(key, ":")
	if !found {
		return "", "", fmt.Errorf("malformed key %q", key)
	}
	return kind, id, nil
}
