package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
)

const (
	catalogConcurrency     = 4
	defaultRequestDelay    = 150 * time.Millisecond
	maxConsecutiveFailures = 8
	catalogCallTimeout     = 20 * time.Second
)

// ErrCatalogUnavailable means too many catalog calls failed in a row.
var ErrCatalogUnavailable = errors.New("importer: YouTube Music kept failing; check the connection and try again")

// Catalog is the part of YouTube Music the importer searches. The innertube
// client satisfies it.
type Catalog interface {
	SearchSongs(ctx context.Context, query string) ([]music.Track, error)
	Search(ctx context.Context, query string) (music.SearchResults, error)
}

// lookup is one Spotify item to resolve to a YouTube Music one.
type lookup struct {
	uri   string
	kind  music.ItemKind
	label string
	phase string
	find  func(ctx context.Context) (found, error)
}

// found is a search result: a zero targetID means nothing matched.
type found struct {
	targetID string
	score    float64
	track    *music.Track
	item     *music.Item
}

// resolved is a lookup's outcome as the apply step needs it.
type resolved struct {
	track *music.Track
	item  *music.Item
}

func (outcome resolved) matched() bool {
	return outcome.track != nil || outcome.item != nil
}

// matcher resolves lookups through stored matches first, then the catalog.
type matcher struct {
	library        *library.Library
	logger         *slog.Logger
	delay          time.Duration
	retryUnmatched bool
	report         func(Progress)

	mutex       sync.Mutex
	done        int
	total       int
	failureRun  int
	failures    int
	successes   int
	outcomes    map[string]resolved
	abortReason error
}

func (importer *Importer) newMatcher(total int, retryUnmatched bool, report func(Progress)) *matcher {
	return &matcher{
		library:        importer.library,
		logger:         importer.logger,
		delay:          importer.requestDelay,
		retryUnmatched: retryUnmatched,
		report:         report,
		total:          total,
		outcomes:       map[string]resolved{},
	}
}

// resolve runs the lookups, at most catalogConcurrency at a time, and returns
// the outcome per lookup URI. It stops early on cancellation or when the
// catalog keeps failing; matches found so far stay stored.
func (runner *matcher) resolve(ctx context.Context, lookups []lookup) (map[string]resolved, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	queue := make(chan lookup)
	var workers sync.WaitGroup
	for range catalogConcurrency {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for pending := range queue {
				runner.handle(ctx, cancel, pending)
			}
		}()
	}
	for _, pending := range lookups {
		if ctx.Err() != nil {
			break
		}
		queue <- pending
	}
	close(queue)
	workers.Wait()

	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	if runner.abortReason == nil && runner.failures > 0 && runner.successes == 0 {
		runner.abortReason = ErrCatalogUnavailable
	}
	if runner.abortReason != nil {
		return runner.outcomes, runner.abortReason
	}
	err := ctx.Err()
	if err != nil {
		return runner.outcomes, fmt.Errorf("importer: %w", err)
	}
	return runner.outcomes, nil
}

func (runner *matcher) handle(ctx context.Context, abort context.CancelFunc, pending lookup) {
	outcome, searched, err := runner.resolveOne(ctx, pending)
	if ctx.Err() != nil {
		return
	}
	if err != nil {
		runner.recordFailure(abort, pending, err)
	} else if searched {
		runner.recordSuccess()
	}
	runner.finish(pending, outcome)
	if searched {
		sleep(ctx, runner.delay)
	}
}

// resolveOne answers from the stored match when there is one; searched says
// whether the catalog was called.
func (runner *matcher) resolveOne(ctx context.Context, pending lookup) (outcome resolved, searched bool, err error) {
	stored, exists, err := runner.library.ImportMatch(ctx, matchSource, pending.uri)
	if err != nil {
		return resolved{}, false, fmt.Errorf("importer: %w", err)
	}
	if exists && (stored.TargetID != "" || !runner.retryUnmatched) {
		return decodeStored(stored), false, nil
	}
	callCtx, cancel := context.WithTimeout(ctx, catalogCallTimeout)
	defer cancel()
	result, err := pending.find(callCtx)
	if err != nil {
		return resolved{}, true, err
	}
	err = runner.store(ctx, pending, result)
	if err != nil {
		return resolved{}, true, err
	}
	return resolved{track: result.track, item: result.item}, true, nil
}

func decodeStored(stored library.ImportMatch) resolved {
	if stored.TargetID == "" {
		return resolved{}
	}
	if stored.Kind == music.ItemKindSong {
		var track music.Track
		if json.Unmarshal([]byte(stored.Payload), &track) != nil {
			return resolved{}
		}
		return resolved{track: &track}
	}
	var item music.Item
	if json.Unmarshal([]byte(stored.Payload), &item) != nil {
		return resolved{}
	}
	return resolved{item: &item}
}

func (runner *matcher) store(ctx context.Context, pending lookup, result found) error {
	payload := ""
	if result.track != nil {
		payload = mustMarshal(result.track)
	}
	if result.item != nil {
		payload = mustMarshal(result.item)
	}
	return runner.library.StoreImportMatch(ctx, library.ImportMatch{
		Source: matchSource, SourceURI: pending.uri, Kind: pending.kind,
		TargetID: result.targetID, Score: result.score, Payload: payload,
	})
}

func mustMarshal(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func (runner *matcher) recordFailure(abort context.CancelFunc, pending lookup, err error) {
	runner.logger.Warn("importer: lookup failed, leaving unmatched", "item", pending.label, "error", err)
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	runner.failureRun++
	runner.failures++
	if runner.failureRun >= maxConsecutiveFailures && runner.abortReason == nil {
		runner.abortReason = ErrCatalogUnavailable
		abort()
	}
}

func (runner *matcher) recordSuccess() {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	runner.failureRun = 0
	runner.successes++
}

func (runner *matcher) finish(pending lookup, outcome resolved) {
	runner.mutex.Lock()
	defer runner.mutex.Unlock()
	runner.outcomes[pending.uri] = outcome
	runner.done++
	runner.report(Progress{Phase: pending.phase, Done: runner.done, Total: runner.total, Current: pending.label})
}

func sleep(ctx context.Context, duration time.Duration) {
	if duration <= 0 {
		return
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
