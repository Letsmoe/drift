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
	Kind      music.ItemKind
	UpdatedAt time.Time
}

// ReferencedTracks lists the cached tracks that a rating, a playlist entry, a
// play or a track block refers to, ordered by video ID.
func (library *Library) ReferencedTracks(ctx context.Context) ([]TrackRecord, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT t.updated_at, t.kind, `+trackColumns+` FROM track t
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
		var updatedAt, kind string
		track, err := scanTrack(rows, &updatedAt, &kind)
		if err != nil {
			return nil, fmt.Errorf("library: %w", err)
		}
		record := TrackRecord{Track: track, Kind: music.ItemKind(kind)}
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

// PlayRecord is one play as the synced history holds it.
type PlayRecord struct {
	VideoID         string
	Context         string
	StartedAt       time.Time
	ProgressSeconds int
	Completed       bool
}

// SearchRecord is one search as the synced history holds it.
type SearchRecord struct {
	Query      string
	SearchedAt time.Time
}

// OwnPlays lists the plays made on this install, oldest first. Plays that came
// from other installs are left out.
func (library *Library) OwnPlays(ctx context.Context) ([]PlayRecord, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT video_id, context, started_at, progress_seconds, completed
		FROM play WHERE origin = '' ORDER BY started_at, id`)
	if err != nil {
		return nil, fmt.Errorf("library: query own plays: %w", err)
	}
	defer func() { _ = rows.Close() }()
	plays := []PlayRecord{}
	for rows.Next() {
		var play PlayRecord
		var startedAt string
		err = rows.Scan(&play.VideoID, &play.Context, &startedAt, &play.ProgressSeconds, &play.Completed)
		if err != nil {
			return nil, fmt.Errorf("library: scan play: %w", err)
		}
		play.StartedAt, err = ParseTime(startedAt)
		if err != nil {
			return nil, err
		}
		plays = append(plays, play)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read own plays: %w", err)
	}
	return plays, nil
}

// OwnSearches lists the searches made on this install, oldest first.
func (library *Library) OwnSearches(ctx context.Context) ([]SearchRecord, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT query, created_at FROM search_history
		WHERE origin = '' ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("library: query own searches: %w", err)
	}
	defer func() { _ = rows.Close() }()
	searches := []SearchRecord{}
	for rows.Next() {
		var search SearchRecord
		var searchedAt string
		err = rows.Scan(&search.Query, &searchedAt)
		if err != nil {
			return nil, fmt.Errorf("library: scan search: %w", err)
		}
		search.SearchedAt, err = ParseTime(searchedAt)
		if err != nil {
			return nil, err
		}
		searches = append(searches, search)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read own searches: %w", err)
	}
	return searches, nil
}

// RestorePlays adds the plays of another install's synced history, reporting no
// change. A play already present (same video and start time from that origin)
// is left as it is. The tracks must be cached already.
func (library *Library) RestorePlays(ctx context.Context, origin string, plays []PlayRecord) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		for _, play := range plays {
			_, err := tx.ExecContext(ctx, `INSERT INTO play
				(video_id, context, started_at, updated_at, progress_seconds, duration_seconds, completed, origin)
				SELECT ?, ?, ?, ?, ?, t.duration_seconds, ?, ? FROM track t
				WHERE t.video_id = ? AND NOT EXISTS (
					SELECT 1 FROM play WHERE origin = ? AND video_id = ? AND started_at = ?)`,
				play.VideoID, play.Context, FormatTime(play.StartedAt), FormatTime(play.StartedAt), play.ProgressSeconds,
				play.Completed, origin, play.VideoID, origin, play.VideoID, FormatTime(play.StartedAt))
			if err != nil {
				return fmt.Errorf("restore play of %s: %w", play.VideoID, err)
			}
		}
		return nil
	})
}

// RestoreSearches adds the searches of another install's synced history,
// reporting no change; one already present is left as it is.
func (library *Library) RestoreSearches(ctx context.Context, origin string, searches []SearchRecord) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		for _, search := range searches {
			_, err := tx.ExecContext(ctx, `INSERT INTO search_history (query, normalized_query, created_at, origin)
				SELECT ?, ?, ?, ? WHERE NOT EXISTS (
					SELECT 1 FROM search_history WHERE origin = ? AND normalized_query = ? AND created_at = ?)`,
				search.Query, normalizeQuery(search.Query), FormatTime(search.SearchedAt), origin,
				origin, normalizeQuery(search.Query), FormatTime(search.SearchedAt))
			if err != nil {
				return fmt.Errorf("restore search %q: %w", search.Query, err)
			}
		}
		return nil
	})
}

// RestoreBlock writes a recommendation block that came from a synced copy,
// keeping its creation time and reporting no change.
func (library *Library) RestoreBlock(ctx context.Context, block RecommendationBlock) error {
	_, err := library.db.ExecContext(ctx, `INSERT INTO recommendation_block (kind, id, name, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (kind, id) DO UPDATE SET name = excluded.name, created_at = excluded.created_at`,
		string(block.Kind), block.ID, block.Name, FormatTime(block.CreatedAt))
	if err != nil {
		return fmt.Errorf("library: restore block %s %s: %w", block.Kind, block.ID, err)
	}
	return nil
}
