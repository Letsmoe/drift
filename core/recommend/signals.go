package recommend

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/neoworks/otter-music/core/library"
)

// signal is one weighted user interaction with a track.
type signal struct {
	videoID string
	weight  float64
	at      time.Time
}

// songsOnly keeps episodes out of every query; they are never recommended and
// carry no taste.
const songsOnly = `video_id NOT IN (SELECT video_id FROM track WHERE kind = 'episode')`

// fixedWeightSources are queries returning (video_id, at) rows since ?; every
// row gets the same weight.
var fixedWeightSources = []struct {
	query  string
	weight float64
}{
	{`SELECT video_id, updated_at FROM track_rating WHERE rating = 1 AND updated_at >= ? AND ` + songsOnly, weightLike},
	{`SELECT video_id, updated_at FROM track_rating WHERE rating = -1 AND updated_at >= ? AND ` + songsOnly, weightDislike},
	{`SELECT video_id, added_at FROM playlist_entry WHERE added_at >= ? AND ` + songsOnly, weightPlaylistAdd},
	{`SELECT video_id, MAX(shown_at) FROM impression WHERE clicked = 1 AND shown_at >= ? AND ` + songsOnly + ` GROUP BY video_id`, weightClick},
}

// playQuery reads the plays since ?.
const playQuery = `SELECT video_id, progress_seconds, duration_seconds, completed, updated_at FROM play
	WHERE started_at >= ? AND ` + songsOnly

// queryRows runs a read query and hands every row to scan.
func (e *Engine) queryRows(ctx context.Context, query string, arguments []any, scan func(rows *sql.Rows) error) error {
	rows, err := e.library.DB().QueryContext(ctx, query, arguments...)
	if err != nil {
		return fmt.Errorf("recommend: query: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		err = scan(rows)
		if err != nil {
			return fmt.Errorf("recommend: read row: %w", err)
		}
	}
	err = rows.Err()
	if err != nil {
		return fmt.Errorf("recommend: read rows: %w", err)
	}
	return nil
}

// collectSignals reads every track interaction inside the signal window and
// converts each into a weighted signal. ignored holds the "shown but not
// clicked" counts, already read for fatigue.
func (e *Engine) collectSignals(ctx context.Context, now time.Time, ignored map[string]ignoredShows) ([]signal, error) {
	since := library.FormatTime(now.Add(-signalWindow))
	signals := ignoredImpressionSignals(ignored)
	for _, source := range fixedWeightSources {
		batch, err := e.fixedWeightSignals(ctx, source.query, since, source.weight)
		if err != nil {
			return nil, err
		}
		signals = append(signals, batch...)
	}
	plays, err := e.playSignals(ctx, since, now)
	if err != nil {
		return nil, err
	}
	return append(signals, plays...), nil
}

// fixedWeightSignals runs a (video_id, at) query and gives every row the same weight.
func (e *Engine) fixedWeightSignals(ctx context.Context, query, since string, weight float64) ([]signal, error) {
	signals := []signal{}
	err := e.queryRows(ctx, query, []any{since}, func(rows *sql.Rows) error {
		var videoID, atText string
		err := rows.Scan(&videoID, &atText)
		if err != nil {
			return fmt.Errorf("scan signal: %w", err)
		}
		at, err := library.ParseTime(atText)
		if err != nil {
			return err
		}
		signals = append(signals, signal{videoID: videoID, weight: weight, at: at})
		return nil
	})
	return signals, err
}

// playSignals turns play rows into completed, partial or skipped signals.
func (e *Engine) playSignals(ctx context.Context, since string, now time.Time) ([]signal, error) {
	signals := []signal{}
	err := e.queryRows(ctx, playQuery, []any{since}, func(rows *sql.Rows) error {
		var videoID, updatedAt string
		var progressSeconds, durationSeconds int
		var completed bool
		err := rows.Scan(&videoID, &progressSeconds, &durationSeconds, &completed, &updatedAt)
		if err != nil {
			return fmt.Errorf("scan play: %w", err)
		}
		at, err := library.ParseTime(updatedAt)
		if err != nil {
			return err
		}
		weight, counts := playWeight(progressSeconds, durationSeconds, completed, now.Sub(at))
		if counts {
			signals = append(signals, signal{videoID: videoID, weight: weight, at: at})
		}
		return nil
	})
	return signals, err
}

// playWeight classifies one play: completed, partial, or skipped early.
// counts is false for a fresh play that may still be starting.
func playWeight(progressSeconds, durationSeconds int, completed bool, sinceUpdate time.Duration) (weight float64, counts bool) {
	if isCompletedPlay(progressSeconds, durationSeconds, completed) {
		return weightCompletedPlay, true
	}
	if !isEarlySkip(progressSeconds, durationSeconds) {
		return weightPartialPlay, true
	}
	if sinceUpdate < skipGrace {
		return 0, false
	}
	return weightEarlySkip, true
}

// isCompletedPlay reports whether the play ended or reached 80% of the track.
func isCompletedPlay(progressSeconds, durationSeconds int, completed bool) bool {
	return completed || playedFraction(progressSeconds, durationSeconds) >= completedFraction
}

// playedFraction is the share of the duration reached; 0 when it is unknown.
func playedFraction(progressSeconds, durationSeconds int) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	return float64(progressSeconds) / float64(durationSeconds)
}

// isEarlySkip reports whether playback stopped before 30 seconds and before
// half the track.
func isEarlySkip(progressSeconds, durationSeconds int) bool {
	if progressSeconds >= earlySkipSeconds {
		return false
	}
	return durationSeconds <= 0 || playedFraction(progressSeconds, durationSeconds) < earlySkipFraction
}

// ignoredImpressionSignals turns "shown but not clicked" counts into weak
// negative signals, capped per track.
func ignoredImpressionSignals(ignored map[string]ignoredShows) []signal {
	signals := make([]signal, 0, len(ignored))
	for videoID, shows := range ignored {
		count := min(shows.count, ignoredImpressionCap)
		signals = append(signals, signal{videoID: videoID, weight: weightIgnoredImpression * float64(count), at: shows.lastShown})
	}
	return signals
}
