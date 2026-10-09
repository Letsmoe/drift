package podcasts

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
)

// expiringCoverPath marks show covers served as signed URLs (".../pl_c/<id>/studio_square_thumbnail.jpg?sqp=…&rs=…");
// YouTube answers 404 for them a few days after they were issued.
const expiringCoverPath = "/pl_c/"

// SavedShows is the part of the library holding saved podcasts.
type SavedShows interface {
	SavedItems(ctx context.Context) ([]library.SavedItem, error)
	Save(ctx context.Context, item music.Item) error
}

// RefreshSavedCovers replaces the expiring cover URLs of saved podcasts with fresh ones from
// their show pages and reports how many items changed. Shows that fail to load keep their cover.
func (directory *Directory) RefreshSavedCovers(ctx context.Context, saved SavedShows) (int, error) {
	items, err := saved.SavedItems(ctx)
	if err != nil {
		return 0, fmt.Errorf("podcasts: refresh covers: %w", err)
	}
	stale := staleShows(items)
	if len(stale) == 0 {
		return 0, nil
	}
	pages := directory.fetchShows(ctx, showIDs(stale))
	refreshed := 0
	for _, item := range stale {
		page, found := pages[item.ID]
		if !found || len(page.Thumbnails) == 0 || slices.Equal(page.Thumbnails, item.Thumbnails) {
			continue
		}
		item.Thumbnails = page.Thumbnails
		err := saved.Save(ctx, item)
		if err != nil {
			return refreshed, fmt.Errorf("podcasts: refresh cover of %s: %w", item.ID, err)
		}
		refreshed++
	}
	return refreshed, nil
}

// staleShows lists the saved podcasts whose cover URLs expire.
func staleShows(items []library.SavedItem) []music.Item {
	stale := []music.Item{}
	for _, saved := range items {
		if saved.Item.Kind == music.ItemKindPodcast && hasExpiringCover(saved.Item.Thumbnails) {
			stale = append(stale, saved.Item)
		}
	}
	return stale
}

func hasExpiringCover(thumbnails []music.Thumbnail) bool {
	return slices.ContainsFunc(thumbnails, func(thumbnail music.Thumbnail) bool {
		return strings.Contains(thumbnail.URL, expiringCoverPath)
	})
}

func showIDs(items []music.Item) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
