// Package podcasts builds the podcast directory: shelves of podcast shows by
// topic. YouTube's podcasts feed (the home feed's "Podcasts" chip) lists trending
// episodes per topic; the directory keeps the topics and their order but shows
// the shows behind the episodes, each completed from its own show page.
package podcasts

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/letsmoe/drift/core/music"
)

const (
	// showFetchConcurrency caps parallel show-page requests.
	showFetchConcurrency = 4
	// showCacheTTL keeps show pages, so paging and reopening the tab stay fast.
	showCacheTTL     = time.Hour
	showFetchTimeout = 15 * time.Second
)

// Catalog is the part of the YouTube Music client the directory reads.
type Catalog interface {
	PodcastsHome(ctx context.Context) (music.HomePage, error)
	HomeContinuation(ctx context.Context, continuation string) (music.HomePage, error)
	Podcast(ctx context.Context, browseID string) (music.PodcastPage, error)
}

// Directory serves podcast shows by topic. Safe for concurrent use.
type Directory struct {
	catalog Catalog
	logger  *slog.Logger
	now     func() time.Time

	mutex sync.Mutex
	shows map[string]cachedShow
}

type cachedShow struct {
	page      music.PodcastPage
	fetchedAt time.Time
}

// New returns a directory reading from catalog; failing show pages are logged to logger.
func New(catalog Catalog, logger *slog.Logger) *Directory {
	return &Directory{catalog: catalog, logger: logger, now: time.Now, shows: map[string]cachedShow{}}
}

// Shows returns the first page of topic shelves of podcast shows.
func (directory *Directory) Shows(ctx context.Context) (music.HomePage, error) {
	feed, err := directory.catalog.PodcastsHome(ctx)
	if err != nil {
		return music.HomePage{}, err
	}
	return directory.showPage(ctx, feed), nil
}

// MoreShows returns the shelves after the page that carried continuation.
func (directory *Directory) MoreShows(ctx context.Context, continuation string) (music.HomePage, error) {
	feed, err := directory.catalog.HomeContinuation(ctx, continuation)
	if err != nil {
		return music.HomePage{}, err
	}
	return directory.showPage(ctx, feed), nil
}

// showPage replaces each shelf's episodes with their shows; shelves left empty are dropped.
func (directory *Directory) showPage(ctx context.Context, feed music.HomePage) music.HomePage {
	pages := directory.fetchShows(ctx, showIDsOf(feed.Shelves))
	shelves := []music.Shelf{}
	for _, shelf := range feed.Shelves {
		shows := showItems(shelf, pages)
		if len(shows) > 0 {
			shelves = append(shelves, music.Shelf{Title: shelf.Title, Items: shows})
		}
	}
	return music.HomePage{Shelves: shelves, Continuation: feed.Continuation}
}

// showIDsOf lists the distinct shows of the feed's episodes (an episode's album is its show).
func showIDsOf(shelves []music.Shelf) []string {
	seen := map[string]bool{}
	ids := []string{}
	for _, shelf := range shelves {
		for _, item := range shelf.Items {
			id := showIDOf(item)
			if id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func showIDOf(item music.Item) string {
	if item.Kind == music.ItemKindPodcast {
		return item.ID
	}
	if item.Track == nil || item.Track.Album == nil {
		return ""
	}
	return item.Track.Album.ID
}

// showItems lists a shelf's shows once each, in the order their episodes appear.
func showItems(shelf music.Shelf, pages map[string]music.PodcastPage) []music.Item {
	seen := map[string]bool{}
	items := []music.Item{}
	for _, item := range shelf.Items {
		id := showIDOf(item)
		page, found := pages[id]
		if !found || seen[id] {
			continue
		}
		seen[id] = true
		items = append(items, showItem(id, page))
	}
	return items
}

func showItem(id string, page music.PodcastPage) music.Item {
	return music.Item{
		Kind:        music.ItemKindPodcast,
		ID:          id,
		Title:       page.Title,
		Subtitle:    page.Author.Name,
		Thumbnails:  page.Thumbnails,
		Description: page.Description,
	}
}

// fetchShows returns the show pages it could load, from the cache or the catalog.
func (directory *Directory) fetchShows(ctx context.Context, ids []string) map[string]music.PodcastPage {
	pages := map[string]music.PodcastPage{}
	var pagesMutex sync.Mutex
	var group sync.WaitGroup
	slots := make(chan struct{}, showFetchConcurrency)
	for _, id := range ids {
		group.Go(func() {
			slots <- struct{}{}
			defer func() { <-slots }()
			page, ok := directory.show(ctx, id)
			if !ok {
				return
			}
			pagesMutex.Lock()
			pages[id] = page
			pagesMutex.Unlock()
		})
	}
	group.Wait()
	return pages
}

// show returns one show page; false when it could not be loaded.
func (directory *Directory) show(ctx context.Context, id string) (music.PodcastPage, bool) {
	if page, found := directory.cachedShow(id); found {
		return page, true
	}
	fetchCtx, cancel := context.WithTimeout(ctx, showFetchTimeout)
	defer cancel()
	page, err := directory.catalog.Podcast(fetchCtx, id)
	if err != nil {
		directory.logger.Warn("podcast directory: show page failed", "show", id, "error", err)
		return music.PodcastPage{}, false
	}
	directory.mutex.Lock()
	directory.shows[id] = cachedShow{page: page, fetchedAt: directory.now()}
	directory.mutex.Unlock()
	return page, true
}

func (directory *Directory) cachedShow(id string) (music.PodcastPage, bool) {
	directory.mutex.Lock()
	defer directory.mutex.Unlock()
	cached, found := directory.shows[id]
	if !found || directory.now().Sub(cached.fetchedAt) >= showCacheTTL {
		return music.PodcastPage{}, false
	}
	return cached.page, true
}
