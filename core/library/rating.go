package library

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/neoworks/otter-music/core/music"
)

// Rating is the user's verdict on a track.
type Rating int

// Ratings. The values are what the track_rating table stores.
const (
	RatingDislike Rating = -1
	RatingNone    Rating = 0
	RatingLike    Rating = 1
)

// SetRating rates a track and caches its metadata; RatingNone removes the rating.
func (library *Library) SetRating(ctx context.Context, track music.Track, kind music.ItemKind, rating Rating) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		err := library.upsertTrack(ctx, tx, track, kind)
		if err != nil {
			return err
		}
		if rating == RatingNone {
			return deleteRating(ctx, tx, track.VideoID)
		}
		return library.storeRating(ctx, tx, track.VideoID, rating)
	})
}

func deleteRating(ctx context.Context, tx *sql.Tx, videoID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM track_rating WHERE video_id = ?`, videoID)
	if err != nil {
		return fmt.Errorf("delete rating of %s: %w", videoID, err)
	}
	return nil
}

func (library *Library) storeRating(ctx context.Context, tx *sql.Tx, videoID string, rating Rating) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO track_rating (video_id, rating, updated_at) VALUES (?, ?, ?)
		ON CONFLICT (video_id) DO UPDATE SET rating = excluded.rating, updated_at = excluded.updated_at`,
		videoID, int(rating), library.timestamp())
	if err != nil {
		return fmt.Errorf("store rating of %s: %w", videoID, err)
	}
	return nil
}

// Ratings returns the ratings of the given videos; unrated videos are absent.
func (library *Library) Ratings(ctx context.Context, videoIDs []string) (map[string]Rating, error) {
	ratings := map[string]Rating{}
	if len(videoIDs) == 0 {
		return ratings, nil
	}
	rows, err := library.db.QueryContext(ctx,
		`SELECT video_id, rating FROM track_rating WHERE video_id IN (`+placeholders(len(videoIDs))+`)`,
		stringArguments(videoIDs)...)
	if err != nil {
		return nil, fmt.Errorf("library: query ratings: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var videoID string
		var rating int
		err = rows.Scan(&videoID, &rating)
		if err != nil {
			return nil, fmt.Errorf("library: scan rating: %w", err)
		}
		ratings[videoID] = Rating(rating)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read ratings: %w", err)
	}
	return ratings, nil
}

// LikedTracks returns every liked track, the most recent like first.
func (library *Library) LikedTracks(ctx context.Context) ([]music.Track, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT `+trackColumns+` FROM track_rating r
		JOIN track t ON t.video_id = r.video_id
		WHERE r.rating = 1
		ORDER BY r.updated_at DESC, r.rowid DESC`)
	if err != nil {
		return nil, fmt.Errorf("library: query liked tracks: %w", err)
	}
	tracks, err := scanTracks(rows)
	if err != nil {
		return nil, fmt.Errorf("library: %w", err)
	}
	return tracks, nil
}
