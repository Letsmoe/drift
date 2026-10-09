package recommend

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

// cachedCatalog reads the catalog through the library's response cache, with
// a timeout per network call.
type cachedCatalog struct {
	library *library.Library
	source  Catalog
	logger  *slog.Logger
}

// radio returns the radio of a track, cached for a day.
func (catalog *cachedCatalog) radio(ctx context.Context, videoID string) ([]music.Track, error) {
	return cached(ctx, catalog, "recommend:radio:"+videoID, radioCacheTTL, func(fetchCtx context.Context) ([]music.Track, error) {
		return catalog.source.Radio(fetchCtx, videoID)
	})
}

// artist returns an artist page, cached for a day.
func (catalog *cachedCatalog) artist(ctx context.Context, browseID string) (music.ArtistPage, error) {
	return cached(ctx, catalog, "recommend:artist:"+browseID, artistCacheTTL, func(fetchCtx context.Context) (music.ArtistPage, error) {
		return catalog.source.Artist(fetchCtx, browseID)
	})
}

// searchSongs returns the songs a query finds, cached for a day.
func (catalog *cachedCatalog) searchSongs(ctx context.Context, query string) ([]music.Track, error) {
	return cached(ctx, catalog, "recommend:search:"+query, searchCacheTTL, func(fetchCtx context.Context) ([]music.Track, error) {
		return catalog.source.SearchSongs(fetchCtx, query)
	})
}

// home returns the first page of YouTube's home feed, cached for an hour.
func (catalog *cachedCatalog) home(ctx context.Context) (music.HomePage, error) {
	return cached(ctx, catalog, "recommend:home", homeCacheTTL, catalog.source.Home)
}

// cached returns the library-cached value for key, or fetches and caches it.
// Cache failures are logged and never fail the call.
func cached[T any](ctx context.Context, catalog *cachedCatalog, key string, ttl time.Duration, fetch func(ctx context.Context) (T, error)) (T, error) {
	var value T
	found, err := catalog.library.CacheGet(ctx, key, &value)
	if err != nil {
		catalog.logger.Warn("recommend: cache read failed", "key", key, "error", err)
	}
	if found {
		return value, nil
	}
	fetchCtx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	value, err = fetch(fetchCtx)
	if err != nil {
		var empty T
		return empty, fmt.Errorf("recommend: fetch %s: %w", key, err)
	}
	err = catalog.library.CachePut(ctx, key, value, ttl)
	if err != nil {
		catalog.logger.Warn("recommend: cache write failed", "key", key, "error", err)
	}
	return value, nil
}

// fetchTask is one catalog request of a ranking pass.
type fetchTask struct {
	name string
	run  func(ctx context.Context) error
}

// logSourceFailure keeps being offline quiet; any other failure is a warning.
func logSourceFailure(logger *slog.Logger, source string, err error) {
	var networkError net.Error
	if errors.As(err, &networkError) {
		logger.Debug("recommend: source offline, skipping", "source", source, "error", err)
		return
	}
	logger.Warn("recommend: source failed, skipping", "source", source, "error", err)
}

// runLimited runs the tasks concurrently, at most fetchConcurrency at a time.
// A failing task is logged and skipped.
func runLimited(ctx context.Context, logger *slog.Logger, tasks []fetchTask) {
	slots := make(chan struct{}, fetchConcurrency)
	var group sync.WaitGroup
	for _, task := range tasks {
		group.Add(1)
		go func() {
			defer group.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			err := task.run(ctx)
			if err != nil {
				logSourceFailure(logger, task.name, err)
			}
		}()
	}
	group.Wait()
}
