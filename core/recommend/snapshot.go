package recommend

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
)

// ignoredShows counts the showings of one track that no click or play followed.
type ignoredShows struct {
	count     int
	lastShown time.Time
}

// savedTaste holds the artists and albums the user saved, id to title.
type savedTaste struct {
	artists map[string]string
	albums  map[string]string
}

// snapshot is everything one ranking pass reads about the user, taken once so
// every candidate is scored against the same state.
type snapshot struct {
	now     time.Time
	profile *Profile
	ignored map[string]ignoredShows
	// lastPlayed maps a track to its latest play within the last week.
	lastPlayed map[string]time.Time
}

// Dislikes are read without a time window: a disliked track never comes back.
const dislikedQuery = `SELECT video_id FROM track_rating WHERE rating = -1`

// ignoredImpressionQuery counts, per track, the showings since ? that the
// user never engaged with afterwards: no click and no play.
const ignoredImpressionQuery = `SELECT shown.video_id, COUNT(*), MAX(shown.shown_at) FROM impression shown
	WHERE shown.clicked = 0 AND shown.shown_at >= ?
	  AND NOT EXISTS (SELECT 1 FROM impression click
	    WHERE click.video_id = shown.video_id AND click.clicked = 1 AND click.shown_at >= shown.shown_at)
	  AND NOT EXISTS (SELECT 1 FROM play
	    WHERE play.video_id = shown.video_id AND play.started_at >= shown.shown_at)
	GROUP BY shown.video_id`

const lastPlayedQuery = `SELECT video_id, MAX(started_at) FROM play WHERE started_at >= ? GROUP BY video_id`

const savedTasteQuery = `SELECT kind, id, title FROM saved_item WHERE kind IN ('artist', 'album')`

const blocksQuery = `SELECT kind, id FROM recommendation_block`

// blocks holds the "Don't recommend" tracks (video IDs) and artists (browse IDs).
type blocks struct {
	tracks  map[string]bool
	artists map[string]bool
}

// takeSnapshot reads the profile and the per-track state ranking needs.
func (e *Engine) takeSnapshot(ctx context.Context) (*snapshot, error) {
	now := e.now()
	profile, ignored, err := e.profileAt(ctx, now)
	if err != nil {
		return nil, err
	}
	lastPlayed, err := e.lastPlayed(ctx, now)
	if err != nil {
		return nil, err
	}
	return &snapshot{now: now, profile: profile, ignored: ignored, lastPlayed: lastPlayed}, nil
}

// profileAt builds the taste profile as of now, returning the ignored
// impressions it read along the way.
func (e *Engine) profileAt(ctx context.Context, now time.Time) (*Profile, map[string]ignoredShows, error) {
	ignored, err := e.ignoredImpressions(ctx, now)
	if err != nil {
		return nil, nil, err
	}
	signals, err := e.collectSignals(ctx, now, ignored)
	if err != nil {
		return nil, nil, err
	}
	tracks, err := e.library.Tracks(ctx, signalVideoIDs(signals))
	if err != nil {
		return nil, nil, fmt.Errorf("recommend: load signalled tracks: %w", err)
	}
	disliked, err := e.dislikedTracks(ctx)
	if err != nil {
		return nil, nil, err
	}
	saved, err := e.savedTaste(ctx)
	if err != nil {
		return nil, nil, err
	}
	blocked, err := e.blocks(ctx)
	if err != nil {
		return nil, nil, err
	}
	profile := buildProfile(profileInput{
		signals: signals, tracks: tracks, saved: saved, disliked: disliked, blocked: blocked, now: now,
	})
	return profile, ignored, nil
}

// ignoredImpressions reads the "shown but not clicked" counts per track.
func (e *Engine) ignoredImpressions(ctx context.Context, now time.Time) (map[string]ignoredShows, error) {
	ignored := map[string]ignoredShows{}
	since := library.FormatTime(now.Add(-ignoredImpressionWindow))
	err := e.queryRows(ctx, ignoredImpressionQuery, []any{since}, func(rows *sql.Rows) error {
		var videoID, lastShownText string
		var shows ignoredShows
		err := rows.Scan(&videoID, &shows.count, &lastShownText)
		if err != nil {
			return fmt.Errorf("scan ignored impression: %w", err)
		}
		shows.lastShown, err = library.ParseTime(lastShownText)
		ignored[videoID] = shows
		return err
	})
	return ignored, err
}

// lastPlayed reads when each track was last played within the last week.
func (e *Engine) lastPlayed(ctx context.Context, now time.Time) (map[string]time.Time, error) {
	played := map[string]time.Time{}
	since := library.FormatTime(now.Add(-playedThisWeekWindow))
	err := e.queryRows(ctx, lastPlayedQuery, []any{since}, func(rows *sql.Rows) error {
		var videoID, startedAt string
		err := rows.Scan(&videoID, &startedAt)
		if err != nil {
			return fmt.Errorf("scan last play: %w", err)
		}
		played[videoID], err = library.ParseTime(startedAt)
		return err
	})
	return played, err
}

// dislikedTracks reads every thumbs-down track.
func (e *Engine) dislikedTracks(ctx context.Context) (map[string]bool, error) {
	disliked := map[string]bool{}
	err := e.queryRows(ctx, dislikedQuery, nil, func(rows *sql.Rows) error {
		var videoID string
		err := rows.Scan(&videoID)
		disliked[videoID] = true
		return wrapScan(err)
	})
	return disliked, err
}

// blocks reads every "Don't recommend" choice.
func (e *Engine) blocks(ctx context.Context) (blocks, error) {
	blocked := blocks{tracks: map[string]bool{}, artists: map[string]bool{}}
	err := e.queryRows(ctx, blocksQuery, nil, func(rows *sql.Rows) error {
		var kind, id string
		err := rows.Scan(&kind, &id)
		if err != nil {
			return fmt.Errorf("scan recommendation block: %w", err)
		}
		if library.BlockKind(kind) == library.BlockArtist {
			blocked.artists[id] = true
			return nil
		}
		blocked.tracks[id] = true
		return nil
	})
	return blocked, err
}

// savedTaste reads the followed artists and saved albums.
func (e *Engine) savedTaste(ctx context.Context) (savedTaste, error) {
	saved := savedTaste{artists: map[string]string{}, albums: map[string]string{}}
	err := e.queryRows(ctx, savedTasteQuery, nil, func(rows *sql.Rows) error {
		var kind, id, title string
		err := rows.Scan(&kind, &id, &title)
		if err != nil {
			return fmt.Errorf("scan saved item: %w", err)
		}
		if kind == string(music.ItemKindArtist) {
			saved.artists[id] = title
			return nil
		}
		saved.albums[id] = title
		return nil
	})
	return saved, err
}

// signalVideoIDs lists the distinct tracks the signals are about.
func signalVideoIDs(signals []signal) []string {
	seen := map[string]bool{}
	videoIDs := []string{}
	for _, entry := range signals {
		if !seen[entry.videoID] {
			seen[entry.videoID] = true
			videoIDs = append(videoIDs, entry.videoID)
		}
	}
	return videoIDs
}
