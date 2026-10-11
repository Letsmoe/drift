package library

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/music"
)

// SavedItem is an album, playlist, artist or podcast the user saved.
type SavedItem struct {
	Item    music.Item `json:"item"`
	SavedAt time.Time  `json:"savedAt"`
}

// Save stores an item in the user's library; saving it again refreshes its details.
func (library *Library) Save(ctx context.Context, item music.Item) error {
	thumbnails, err := marshalList(item.Thumbnails)
	if err != nil {
		return fmt.Errorf("library: %w", err)
	}
	_, err = library.db.ExecContext(ctx, `INSERT INTO saved_item (kind, id, title, subtitle, thumbnails, saved_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (kind, id) DO UPDATE SET
			title = excluded.title, subtitle = excluded.subtitle, thumbnails = excluded.thumbnails`,
		string(item.Kind), item.ID, item.Title, item.Subtitle, thumbnails, library.timestamp())
	if err != nil {
		return fmt.Errorf("library: save %s %s: %w", item.Kind, item.ID, err)
	}
	library.emitChange(ctx, Change{Kind: ChangeSaved})
	return nil
}

// Unsave removes an item; removing one that is not saved is not an error.
func (library *Library) Unsave(ctx context.Context, kind music.ItemKind, id string) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM saved_item WHERE kind = ? AND id = ?`, string(kind), id)
	if err != nil {
		return fmt.Errorf("library: unsave %s %s: %w", kind, id, err)
	}
	library.emitChange(ctx, Change{Kind: ChangeSaved})
	return nil
}

// IsSaved says whether the item is in the user's library.
func (library *Library) IsSaved(ctx context.Context, kind music.ItemKind, id string) (bool, error) {
	var count int
	err := library.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM saved_item WHERE kind = ? AND id = ?`, string(kind), id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("library: check saved %s %s: %w", kind, id, err)
	}
	return count > 0, nil
}

// SavedItems lists everything saved, the newest first.
func (library *Library) SavedItems(ctx context.Context) ([]SavedItem, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT kind, id, title, subtitle, thumbnails, saved_at FROM saved_item
		ORDER BY saved_at DESC, rowid DESC`)
	if err != nil {
		return nil, fmt.Errorf("library: query saved items: %w", err)
	}
	defer func() { _ = rows.Close() }()
	saved := []SavedItem{}
	for rows.Next() {
		item, err := scanSavedItem(rows)
		if err != nil {
			return nil, fmt.Errorf("library: %w", err)
		}
		saved = append(saved, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read saved items: %w", err)
	}
	return saved, nil
}

func scanSavedItem(row rowScanner) (SavedItem, error) {
	var saved SavedItem
	var kind, thumbnails, savedAt string
	err := row.Scan(&kind, &saved.Item.ID, &saved.Item.Title, &saved.Item.Subtitle, &thumbnails, &savedAt)
	if err != nil {
		return SavedItem{}, fmt.Errorf("scan saved item: %w", err)
	}
	saved.Item.Kind = music.ItemKind(kind)
	err = json.Unmarshal([]byte(thumbnails), &saved.Item.Thumbnails)
	if err != nil {
		return SavedItem{}, fmt.Errorf("decode thumbnails of %s: %w", saved.Item.ID, err)
	}
	saved.SavedAt, err = ParseTime(savedAt)
	if err != nil {
		return SavedItem{}, err
	}
	return saved, nil
}
