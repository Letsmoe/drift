package recommend

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

// candidate is a track some source proposes, before scoring.
type candidate struct {
	track  music.Track
	source string
	// prior is the source's trust in this track, in [0, 1].
	prior float64
	// sourceCount is how many sources proposed the track.
	sourceCount int
	reason      Reason
}

// candidatePool merges candidates by track, keeping the strongest source's
// prior and reason and counting how many sources agreed.
type candidatePool struct {
	byVideoID map[string]*candidate
	order     []string
}

// newCandidatePool returns an empty pool.
func newCandidatePool() *candidatePool {
	return &candidatePool{byVideoID: map[string]*candidate{}}
}

// add merges candidates into the pool; tracks without a video ID are dropped.
func (pool *candidatePool) add(proposals ...candidate) {
	for _, proposed := range proposals {
		pool.addOne(proposed)
	}
}

// addOne merges one candidate into the pool.
func (pool *candidatePool) addOne(proposed candidate) {
	videoID := proposed.track.VideoID
	if videoID == "" {
		return
	}
	existing, found := pool.byVideoID[videoID]
	if !found {
		proposed.sourceCount = 1
		pool.byVideoID[videoID] = &proposed
		pool.order = append(pool.order, videoID)
		return
	}
	if existing.source != proposed.source {
		existing.sourceCount++
	}
	if proposed.prior > existing.prior {
		existing.prior = proposed.prior
		existing.source = proposed.source
		existing.reason = proposed.reason
	}
}

// list returns the pooled candidates in first-seen order.
func (pool *candidatePool) list() []*candidate {
	candidates := make([]*candidate, 0, len(pool.order))
	for _, videoID := range pool.order {
		candidates = append(candidates, pool.byVideoID[videoID])
	}
	return candidates
}

// rankedPrior lowers a source prior by the item's position within the source.
func rankedPrior(prior float64, rank int) float64 {
	return prior / (1 + float64(rank)*rankDecay)
}

// trackCandidates turns a ranked track list from one source into candidates.
func trackCandidates(tracks []music.Track, source string, prior float64, reason Reason) []candidate {
	candidates := make([]candidate, 0, len(tracks))
	for rank, track := range tracks {
		candidates = append(candidates, candidate{track: track, source: source, prior: rankedPrior(prior, rank), reason: reason})
	}
	return candidates
}

// generateCandidates runs every source and pools the results. Sources that
// fail are skipped; only a cancelled context or a library error fails the pass.
func (e *Engine) generateCandidates(ctx context.Context, state *snapshot) (*candidatePool, error) {
	queries, err := e.recentQueries(ctx, state.now)
	if err != nil {
		return nil, err
	}
	plan := newSourcePlan(state.profile, queries)
	runLimited(ctx, e.logger, plan.primaryTasks(e.catalog))
	plan.chooseExploreArtists(state.profile)
	runLimited(ctx, e.logger, plan.exploreTasks(e.catalog))
	err = ctx.Err()
	if err != nil {
		return nil, fmt.Errorf("recommend: %w", err)
	}
	rediscovered, err := e.rediscoverCandidates(ctx, state.now)
	if err != nil {
		return nil, err
	}
	pool := newCandidatePool()
	pool.add(plan.candidates(state.profile)...)
	pool.add(rediscovered...)
	return pool, nil
}

// recentQueries returns the distinct searches of the last two weeks, newest
// first.
func (e *Engine) recentQueries(ctx context.Context, now time.Time) ([]string, error) {
	since := library.FormatTime(now.Add(-recentSearchWindow))
	queries := []string{}
	err := e.queryRows(ctx, `SELECT query FROM search_history
		WHERE id IN (SELECT MAX(id) FROM search_history GROUP BY normalized_query) AND created_at >= ?
		ORDER BY id DESC LIMIT ?`, []any{since, recentSearchLimit}, func(rows *sql.Rows) error {
		var query string
		err := rows.Scan(&query)
		queries = append(queries, query)
		return wrapScan(err)
	})
	return queries, err
}

// rediscoverQuery lists liked or completed songs that were not played since ?,
// the most recently played first.
const rediscoverQuery = `SELECT t.video_id FROM track t
	WHERE t.kind = 'song'
	  AND (EXISTS (SELECT 1 FROM track_rating r WHERE r.video_id = t.video_id AND r.rating = 1)
	    OR EXISTS (SELECT 1 FROM play p WHERE p.video_id = t.video_id
	      AND (p.completed = 1 OR (p.duration_seconds > 0 AND p.progress_seconds >= p.duration_seconds * ?))))
	  AND NOT EXISTS (SELECT 1 FROM track_rating r WHERE r.video_id = t.video_id AND r.rating = -1)
	  AND NOT EXISTS (SELECT 1 FROM play p WHERE p.video_id = t.video_id AND p.started_at >= ?)
	ORDER BY COALESCE((SELECT MAX(p.started_at) FROM play p WHERE p.video_id = t.video_id), '') DESC, t.video_id
	LIMIT ?`

// rediscoverCandidates proposes loved tracks that have not been played lately.
func (e *Engine) rediscoverCandidates(ctx context.Context, now time.Time) ([]candidate, error) {
	videoIDs := []string{}
	since := library.FormatTime(now.Add(-rediscoverWindow))
	err := e.queryRows(ctx, rediscoverQuery, []any{completedFraction, since, rediscoverLimit}, func(rows *sql.Rows) error {
		var videoID string
		err := rows.Scan(&videoID)
		videoIDs = append(videoIDs, videoID)
		return wrapScan(err)
	})
	if err != nil {
		return nil, err
	}
	stored, err := e.library.Tracks(ctx, videoIDs)
	if err != nil {
		return nil, fmt.Errorf("recommend: load rediscover tracks: %w", err)
	}
	tracks := make([]music.Track, 0, len(videoIDs))
	for _, videoID := range videoIDs {
		tracks = append(tracks, stored[videoID])
	}
	return trackCandidates(tracks, SourceRediscover, sourcePriors[SourceRediscover], rediscoverReason()), nil
}

// wrapScan adds context to a row scan error.
func wrapScan(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("scan row: %w", err)
}
