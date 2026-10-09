package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/letsmoe/drift/core/music"
)

// minimumResumeSeconds is how far into an episode a play must have got to be worth resuming.
const minimumResumeSeconds = 30

// StartPlay records the start of a playback and caches the track. playContext says
// where it was started e.g. "album:MPREb_x", "playlist:abc", "radio",
// "search" or "queue". The returned ID goes to UpdatePlay.
func (library *Library) StartPlay(ctx context.Context, track music.Track, kind music.ItemKind, playContext string) (int64, error) {
	var playID int64
	err := library.inTransaction(ctx, func(tx *sql.Tx) error {
		err := library.upsertTrack(ctx, tx, track, kind)
		if err != nil {
			return err
		}
		now := library.timestamp()
		result, err := tx.ExecContext(ctx, `INSERT INTO play (video_id, context, started_at, updated_at, duration_seconds)
			VALUES (?, ?, ?, ?, ?)`, track.VideoID, playContext, now, now, track.DurationSeconds)
		if err != nil {
			return fmt.Errorf("start play of %s: %w", track.VideoID, err)
		}
		playID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("read play id: %w", err)
		}
		return nil
	})
	return playID, err
}

// UpdatePlay reports playback progress. The furthest progress is kept and
// completed, once true, stays true.
func (library *Library) UpdatePlay(ctx context.Context, playID int64, progressSeconds int, completed bool) error {
	result, err := library.db.ExecContext(ctx, `UPDATE play SET
		progress_seconds = MAX(progress_seconds, ?),
		completed = MAX(completed, ?),
		updated_at = ?
		WHERE id = ?`, progressSeconds, completed, library.timestamp(), playID)
	if err != nil {
		return fmt.Errorf("library: update play %d: %w", playID, err)
	}
	return requireAffected(result, fmt.Sprintf("play %d", playID))
}

// RecentTracks lists distinct played tracks, the most recently played first.
func (library *Library) RecentTracks(ctx context.Context, limit int) ([]music.Track, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT `+trackColumns+` FROM track t
		JOIN (SELECT video_id AS played_video_id, MAX(started_at) AS last_played FROM play GROUP BY video_id) p
			ON p.played_video_id = t.video_id
		ORDER BY p.last_played DESC, t.video_id LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("library: query recent tracks: %w", err)
	}
	tracks, err := scanTracks(rows)
	if err != nil {
		return nil, fmt.Errorf("library: %w", err)
	}
	return tracks, nil
}

// ResumePosition returns where to pick an episode back up: the progress of its
// latest play, if that play was not completed and got past 30 seconds. Songs
// never resume.
func (library *Library) ResumePosition(ctx context.Context, videoID string) (int, bool, error) {
	var progressSeconds int
	var completed bool
	err := library.db.QueryRowContext(ctx, `SELECT p.progress_seconds, p.completed FROM play p
		JOIN track t ON t.video_id = p.video_id
		WHERE p.video_id = ? AND t.kind = 'episode'
		ORDER BY p.started_at DESC, p.id DESC LIMIT 1`, videoID).Scan(&progressSeconds, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("library: query resume position of %s: %w", videoID, err)
	}
	return progressSeconds, !completed && progressSeconds > minimumResumeSeconds, nil
}
