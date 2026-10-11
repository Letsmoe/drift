package library

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/music"
)

// RatingRecord is a stored rating with the time it was last set.
type RatingRecord struct {
	VideoID   string
	Rating    Rating
	UpdatedAt time.Time
}

// RatingRecords lists every rating, ordered by video ID.
func (library *Library) RatingRecords(ctx context.Context) ([]RatingRecord, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT video_id, rating, updated_at FROM track_rating ORDER BY video_id`)
	if err != nil {
		return nil, fmt.Errorf("library: query rating records: %w", err)
	}
	defer func() { _ = rows.Close() }()
	records := []RatingRecord{}
	for rows.Next() {
		var record RatingRecord
		var rating int
		var updatedAt string
		err = rows.Scan(&record.VideoID, &rating, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("library: scan rating record: %w", err)
		}
		record.Rating = Rating(rating)
		record.UpdatedAt, err = ParseTime(updatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read rating records: %w", err)
	}
	return records, nil
}

// RestoreRating writes a rating that came from a synced copy, keeping its time
// and reporting no change. The track must be cached already. RatingNone removes
// the rating.
func (library *Library) RestoreRating(ctx context.Context, record RatingRecord) error {
	if record.Rating == RatingNone {
		_, err := library.db.ExecContext(ctx, `DELETE FROM track_rating WHERE video_id = ?`, record.VideoID)
		if err != nil {
			return fmt.Errorf("library: remove rating of %s: %w", record.VideoID, err)
		}
		return nil
	}
	_, err := library.db.ExecContext(ctx, `INSERT INTO track_rating (video_id, rating, updated_at) VALUES (?, ?, ?)
		ON CONFLICT (video_id) DO UPDATE SET rating = excluded.rating, updated_at = excluded.updated_at`,
		record.VideoID, int(record.Rating), FormatTime(record.UpdatedAt))
	if err != nil {
		return fmt.Errorf("library: restore rating of %s: %w", record.VideoID, err)
	}
	return nil
}

// TrackRecord is a cached track with the time it was last cached.
type TrackRecord struct {
	Track     music.Track
	UpdatedAt time.Time
}

// ReferencedTracks lists the cached tracks that a rating, a playlist entry, a
// play or a track block refers to, ordered by video ID.
func (library *Library) ReferencedTracks(ctx context.Context) ([]TrackRecord, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT t.updated_at, `+trackColumns+` FROM track t
		WHERE t.video_id IN (SELECT video_id FROM track_rating)
			OR t.video_id IN (SELECT video_id FROM playlist_entry)
			OR t.video_id IN (SELECT video_id FROM play)
			OR t.video_id IN (SELECT id FROM recommendation_block WHERE kind = 'track')
		ORDER BY t.video_id`)
	if err != nil {
		return nil, fmt.Errorf("library: query referenced tracks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	records := []TrackRecord{}
	for rows.Next() {
		var updatedAt string
		track, err := scanTrack(rows, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("library: %w", err)
		}
		record := TrackRecord{Track: track}
		record.UpdatedAt, err = ParseTime(updatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read referenced tracks: %w", err)
	}
	return records, nil
}

// RestoreSavedItem writes a saved item that came from a synced copy, keeping its
// saving time and reporting no change.
func (library *Library) RestoreSavedItem(ctx context.Context, saved SavedItem) error {
	thumbnails, err := marshalList(saved.Item.Thumbnails)
	if err != nil {
		return fmt.Errorf("library: %w", err)
	}
	_, err = library.db.ExecContext(ctx, `INSERT INTO saved_item (kind, id, title, subtitle, thumbnails, saved_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (kind, id) DO UPDATE SET
			title = excluded.title, subtitle = excluded.subtitle, thumbnails = excluded.thumbnails, saved_at = excluded.saved_at`,
		string(saved.Item.Kind), saved.Item.ID, saved.Item.Title, saved.Item.Subtitle, thumbnails, FormatTime(saved.SavedAt))
	if err != nil {
		return fmt.Errorf("library: restore saved %s %s: %w", saved.Item.Kind, saved.Item.ID, err)
	}
	return nil
}

// RestorePlaylist writes a playlist that came from a synced copy: it creates
// the playlist or replaces its name, description, times and entries, and
// reports no change. The entries' tracks must be cached already.
func (library *Library) RestorePlaylist(ctx context.Context, playlist Playlist) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO playlist (id, name, description, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (id) DO UPDATE SET name = excluded.name, description = excluded.description,
				created_at = excluded.created_at, updated_at = excluded.updated_at`,
			playlist.ID, playlist.Name, playlist.Description, FormatTime(playlist.CreatedAt), FormatTime(playlist.UpdatedAt))
		if err != nil {
			return fmt.Errorf("restore playlist %s: %w", playlist.ID, err)
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM playlist_entry WHERE playlist_id = ?`, playlist.ID)
		if err != nil {
			return fmt.Errorf("clear entries of playlist %s: %w", playlist.ID, err)
		}
		return restoreEntries(ctx, tx, playlist)
	})
}

func restoreEntries(ctx context.Context, tx *sql.Tx, playlist Playlist) error {
	for position, entry := range playlist.Entries {
		_, err := tx.ExecContext(ctx, `INSERT INTO playlist_entry (playlist_id, video_id, position, added_at) VALUES (?, ?, ?, ?)`,
			playlist.ID, entry.Track.VideoID, position, FormatTime(entry.AddedAt))
		if err != nil {
			return fmt.Errorf("restore entry %s of playlist %s: %w", entry.Track.VideoID, playlist.ID, err)
		}
	}
	return nil
}
