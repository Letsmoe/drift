package library

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

// PlaylistSummary is a playlist as listed, without its entries.
type PlaylistSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TrackCount  int    `json:"trackCount"`
	// Thumbnails is the cover of the first entry; empty for an empty playlist.
	Thumbnails []music.Thumbnail `json:"thumbnails"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

// PlaylistEntry is one track in a playlist. ID is unique per entry, so the
// same track can be in a playlist twice.
type PlaylistEntry struct {
	ID      int64       `json:"id"`
	Track   music.Track `json:"track"`
	AddedAt time.Time   `json:"addedAt"`
}

// Playlist is a user-made playlist with its entries in order.
type Playlist struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Entries     []PlaylistEntry `json:"entries"`
}

// CreatePlaylist makes an empty playlist.
func (library *Library) CreatePlaylist(ctx context.Context, name string) (Playlist, error) {
	identifier, err := newPlaylistID()
	if err != nil {
		return Playlist{}, err
	}
	now := library.now().UTC().Truncate(time.Millisecond)
	_, err = library.db.ExecContext(ctx,
		`INSERT INTO playlist (id, name, description, created_at, updated_at) VALUES (?, ?, '', ?, ?)`,
		identifier, name, FormatTime(now), FormatTime(now))
	if err != nil {
		return Playlist{}, fmt.Errorf("library: create playlist: %w", err)
	}
	return Playlist{ID: identifier, Name: name, CreatedAt: now, UpdatedAt: now, Entries: []PlaylistEntry{}}, nil
}

func newPlaylistID() (string, error) {
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("library: generate playlist id: %w", err)
	}
	return hex.EncodeToString(randomBytes), nil
}

// RenamePlaylist changes a playlist's name and description.
func (library *Library) RenamePlaylist(ctx context.Context, id, name, description string) error {
	result, err := library.db.ExecContext(ctx, `UPDATE playlist SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, description, library.timestamp(), id)
	if err != nil {
		return fmt.Errorf("library: rename playlist %s: %w", id, err)
	}
	return requireAffected(result, "playlist "+id)
}

// DeletePlaylist removes a playlist and its entries.
func (library *Library) DeletePlaylist(ctx context.Context, id string) error {
	result, err := library.db.ExecContext(ctx, `DELETE FROM playlist WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("library: delete playlist %s: %w", id, err)
	}
	return requireAffected(result, "playlist "+id)
}

func requireAffected(result sql.Result, subject string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("library: rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("library: %s: %w", subject, ErrNotFound)
	}
	return nil
}

// Playlists lists every playlist, the most recently changed first.
func (library *Library) Playlists(ctx context.Context) ([]PlaylistSummary, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT p.id, p.name, p.description, p.updated_at,
		(SELECT COUNT(*) FROM playlist_entry e WHERE e.playlist_id = p.id),
		(SELECT t.thumbnails FROM playlist_entry e JOIN track t ON t.video_id = e.video_id
			WHERE e.playlist_id = p.id ORDER BY e.position LIMIT 1)
		FROM playlist p ORDER BY p.updated_at DESC, p.created_at DESC, p.id`)
	if err != nil {
		return nil, fmt.Errorf("library: query playlists: %w", err)
	}
	defer func() { _ = rows.Close() }()
	summaries := []PlaylistSummary{}
	for rows.Next() {
		summary, err := scanPlaylistSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("library: %w", err)
		}
		summaries = append(summaries, summary)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read playlists: %w", err)
	}
	return summaries, nil
}

func scanPlaylistSummary(row rowScanner) (PlaylistSummary, error) {
	var summary PlaylistSummary
	var updatedAt string
	var thumbnails sql.NullString
	err := row.Scan(&summary.ID, &summary.Name, &summary.Description, &updatedAt, &summary.TrackCount, &thumbnails)
	if err != nil {
		return PlaylistSummary{}, fmt.Errorf("scan playlist: %w", err)
	}
	summary.UpdatedAt, err = ParseTime(updatedAt)
	if err != nil {
		return PlaylistSummary{}, err
	}
	summary.Thumbnails = []music.Thumbnail{}
	if !thumbnails.Valid {
		return summary, nil
	}
	err = json.Unmarshal([]byte(thumbnails.String), &summary.Thumbnails)
	if err != nil {
		return PlaylistSummary{}, fmt.Errorf("decode playlist cover: %w", err)
	}
	return summary, nil
}

// Playlist loads one playlist with its entries.
func (library *Library) Playlist(ctx context.Context, id string) (Playlist, error) {
	var playlist Playlist
	var createdAt, updatedAt string
	err := library.db.QueryRowContext(ctx, `SELECT id, name, description, created_at, updated_at FROM playlist WHERE id = ?`, id).
		Scan(&playlist.ID, &playlist.Name, &playlist.Description, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Playlist{}, fmt.Errorf("library: playlist %s: %w", id, ErrNotFound)
	}
	if err != nil {
		return Playlist{}, fmt.Errorf("library: load playlist %s: %w", id, err)
	}
	playlist.CreatedAt, err = ParseTime(createdAt)
	if err != nil {
		return Playlist{}, err
	}
	playlist.UpdatedAt, err = ParseTime(updatedAt)
	if err != nil {
		return Playlist{}, err
	}
	playlist.Entries, err = library.playlistEntries(ctx, id)
	if err != nil {
		return Playlist{}, err
	}
	return playlist, nil
}

func (library *Library) playlistEntries(ctx context.Context, playlistID string) ([]PlaylistEntry, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT e.id, e.added_at, `+trackColumns+` FROM playlist_entry e
		JOIN track t ON t.video_id = e.video_id
		WHERE e.playlist_id = ? ORDER BY e.position`, playlistID)
	if err != nil {
		return nil, fmt.Errorf("library: query playlist entries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	entries := []PlaylistEntry{}
	for rows.Next() {
		entry, err := scanPlaylistEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("library: %w", err)
		}
		entries = append(entries, entry)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read playlist entries: %w", err)
	}
	return entries, nil
}

func scanPlaylistEntry(row rowScanner) (PlaylistEntry, error) {
	var entry PlaylistEntry
	var addedAt string
	track, err := scanTrack(row, &entry.ID, &addedAt)
	if err != nil {
		return PlaylistEntry{}, err
	}
	entry.Track = track
	entry.AddedAt, err = ParseTime(addedAt)
	if err != nil {
		return PlaylistEntry{}, err
	}
	return entry, nil
}

// AddToPlaylist appends tracks to a playlist and caches their metadata.
func (library *Library) AddToPlaylist(ctx context.Context, id string, tracks []music.Track) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		err := library.touchPlaylist(ctx, tx, id)
		if err != nil {
			return err
		}
		return library.appendEntries(ctx, tx, id, tracks)
	})
}

func (library *Library) appendEntries(ctx context.Context, tx *sql.Tx, id string, tracks []music.Track) error {
	var nextPosition int
	err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position) + 1, 0) FROM playlist_entry WHERE playlist_id = ?`, id).Scan(&nextPosition)
	if err != nil {
		return fmt.Errorf("find end of playlist %s: %w", id, err)
	}
	err = library.upsertTracks(ctx, tx, tracks, music.ItemKindSong)
	if err != nil {
		return err
	}
	for _, track := range tracks {
		if track.VideoID == "" {
			continue
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO playlist_entry (playlist_id, video_id, position, added_at) VALUES (?, ?, ?, ?)`,
			id, track.VideoID, nextPosition, library.timestamp())
		if err != nil {
			return fmt.Errorf("add %s to playlist %s: %w", track.VideoID, id, err)
		}
		nextPosition++
	}
	return nil
}

// RemoveFromPlaylist deletes one entry and closes the gap it leaves.
func (library *Library) RemoveFromPlaylist(ctx context.Context, id string, entryID int64) error {
	return library.reorder(ctx, id, func(entryIDs []int64) ([]int64, error) {
		position := indexOfEntry(entryIDs, entryID)
		if position < 0 {
			return nil, fmt.Errorf("playlist entry %d: %w", entryID, ErrNotFound)
		}
		return slices.Delete(slices.Clone(entryIDs), position, position+1), nil
	}, entryID)
}

// MovePlaylistEntry moves an entry to newIndex (clamped to the playlist).
func (library *Library) MovePlaylistEntry(ctx context.Context, id string, entryID int64, newIndex int) error {
	return library.reorder(ctx, id, func(entryIDs []int64) ([]int64, error) {
		position := indexOfEntry(entryIDs, entryID)
		if position < 0 {
			return nil, fmt.Errorf("playlist entry %d: %w", entryID, ErrNotFound)
		}
		remaining := slices.Delete(slices.Clone(entryIDs), position, position+1)
		target := max(0, min(newIndex, len(remaining)))
		return slices.Insert(remaining, target, entryID), nil
	}, 0)
}

// reorder loads the playlist's entry order, lets arrange produce the new
// order and rewrites dense positions; removedEntryID (if not 0) is deleted.
func (library *Library) reorder(ctx context.Context, id string, arrange func([]int64) ([]int64, error), removedEntryID int64) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		err := library.touchPlaylist(ctx, tx, id)
		if err != nil {
			return err
		}
		current, err := entryOrder(ctx, tx, id)
		if err != nil {
			return err
		}
		arranged, err := arrange(current)
		if err != nil {
			return err
		}
		if removedEntryID != 0 {
			_, err = tx.ExecContext(ctx, `DELETE FROM playlist_entry WHERE id = ?`, removedEntryID)
			if err != nil {
				return fmt.Errorf("delete playlist entry %d: %w", removedEntryID, err)
			}
		}
		return writePositions(ctx, tx, arranged)
	})
}

func entryOrder(ctx context.Context, tx *sql.Tx, playlistID string) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id FROM playlist_entry WHERE playlist_id = ? ORDER BY position, id`, playlistID)
	if err != nil {
		return nil, fmt.Errorf("query entry order: %w", err)
	}
	defer func() { _ = rows.Close() }()
	entryIDs := []int64{}
	for rows.Next() {
		var entryID int64
		err = rows.Scan(&entryID)
		if err != nil {
			return nil, fmt.Errorf("scan entry id: %w", err)
		}
		entryIDs = append(entryIDs, entryID)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read entry order: %w", err)
	}
	return entryIDs, nil
}

func writePositions(ctx context.Context, tx *sql.Tx, entryIDs []int64) error {
	for position, entryID := range entryIDs {
		_, err := tx.ExecContext(ctx, `UPDATE playlist_entry SET position = ? WHERE id = ?`, position, entryID)
		if err != nil {
			return fmt.Errorf("write position of entry %d: %w", entryID, err)
		}
	}
	return nil
}

func indexOfEntry(entryIDs []int64, entryID int64) int {
	for position, candidate := range entryIDs {
		if candidate == entryID {
			return position
		}
	}
	return -1
}

// touchPlaylist bumps updated_at and doubles as the existence check.
func (library *Library) touchPlaylist(ctx context.Context, tx *sql.Tx, id string) error {
	result, err := tx.ExecContext(ctx, `UPDATE playlist SET updated_at = ? WHERE id = ?`, library.timestamp(), id)
	if err != nil {
		return fmt.Errorf("touch playlist %s: %w", id, err)
	}
	return requireAffected(result, "playlist "+id)
}
