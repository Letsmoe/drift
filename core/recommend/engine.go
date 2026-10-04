// Package recommend is the recommendation engine: candidate generation from
// YouTube Music's radio, artist pages, searches and home feed plus the local
// library, a transparent weighted ranking against a time-decayed taste
// profile, and diversity-aware re-ranking.
//
// The taste profile is computed on demand rather than stored: the signal
// tables hold one row per user action inside a bounded window, so the
// aggregation stays small and there is no cache to invalidate on every like
// or play. The ranked feed itself is memoized for pagination (see
// feedCacheTTL). Catalog responses are cached in the library.
package recommend

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/queue"
)

// The engine plugs into the queue's autoplay.
var _ queue.Extender = (*Engine)(nil)

// Engine ranks recommendations. Safe for concurrent use.
type Engine struct {
	library *library.Library
	catalog *cachedCatalog
	logger  *slog.Logger
	// now is the clock, swappable in tests.
	now func() time.Time

	mutex sync.Mutex
	cache *rankedFeed
}

// Option configures New.
type Option func(*Engine)

// WithLogger sets the logger failing sources are reported to.
func WithLogger(logger *slog.Logger) Option {
	return func(engine *Engine) {
		engine.logger = logger
	}
}

// WithClock replaces the wall clock, for tests.
func WithClock(now func() time.Time) Option {
	return func(engine *Engine) {
		engine.now = now
	}
}

// rankedFeed is one full ranking pass.
type rankedFeed struct {
	builtAt time.Time
	// scored holds every eligible candidate, best score first.
	scored []Item
	// ranked is the re-ranked feed order.
	ranked []Item
}

// New returns an engine reading the user's signals from lib and the catalog
// from catalog.
func New(lib *library.Library, catalog Catalog, options ...Option) *Engine {
	engine := &Engine{library: lib, logger: slog.New(slog.DiscardHandler), now: time.Now}
	for _, option := range options {
		option(engine)
	}
	engine.catalog = &cachedCatalog{library: lib, source: catalog, logger: engine.logger}
	return engine
}

// Feed returns one page of the ranked feed. Offset 0 re-ranks once the
// ranking is older than feedFirstPageTTL (a reload shows the same feed), or
// always when refresh is set; later pages read the same ranking while it is
// fresh, so pages neither overlap nor skip as impressions shift the scores.
func (e *Engine) Feed(ctx context.Context, limit, offset int, refresh bool) ([]Item, error) {
	maxAge := feedCacheTTL
	if offset == 0 {
		maxAge = feedFirstPageTTL
	}
	if refresh {
		maxAge = 0
	}
	feed, err := e.rankedFeed(ctx, maxAge)
	if err != nil {
		return nil, err
	}
	return page(feed.ranked, limit, offset), nil
}

// Rows groups the ranked candidates by reason into titled rows ("Because you
// listened to X", "More from <artist>", ...), strongest row first.
func (e *Engine) Rows(ctx context.Context, rowLimit, perRow int) ([]Row, error) {
	feed, err := e.rankedFeed(ctx, feedCacheTTL)
	if err != nil {
		return nil, err
	}
	return groupRows(feed.scored, rowLimit, perRow), nil
}

// Profile returns the current taste profile.
func (e *Engine) Profile(ctx context.Context) (*Profile, error) {
	profile, _, err := e.profileAt(ctx, e.now())
	return profile, err
}

// rankedFeed returns the memoized ranking, rebuilding it when it is older than
// maxAge.
func (e *Engine) rankedFeed(ctx context.Context, maxAge time.Duration) (*rankedFeed, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.cache.freshFor(e.now(), maxAge) {
		return e.cache, nil
	}
	feed, err := e.buildFeed(ctx)
	if err != nil {
		return nil, err
	}
	e.cache = feed
	return feed, nil
}

// freshFor reports whether the cached ranking can serve a request that
// accepts rankings up to maxAge old.
func (feed *rankedFeed) freshFor(now time.Time, maxAge time.Duration) bool {
	return feed != nil && now.Sub(feed.builtAt) < maxAge
}

// buildFeed runs the full pipeline: snapshot, candidates, scoring, re-ranking.
func (e *Engine) buildFeed(ctx context.Context) (*rankedFeed, error) {
	state, err := e.takeSnapshot(ctx)
	if err != nil {
		return nil, err
	}
	pool, err := e.generateCandidates(ctx, state)
	if err != nil {
		return nil, err
	}
	scored := scoreCandidates(pool.list(), state)
	return &rankedFeed{builtAt: state.now, scored: scored, ranked: rerank(slices.Clone(scored))}, nil
}

// page slices one page out of the feed.
func page(items []Item, limit, offset int) []Item {
	if offset >= len(items) || limit <= 0 || offset < 0 {
		return []Item{}
	}
	end := min(offset+limit, len(items))
	return slices.Clone(items[offset:end])
}
