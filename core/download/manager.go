// Package download keeps audio offline: a persistent, resumable queue that
// stores each track's best stream under a downloads directory, plus
// collections (albums, playlists, liked songs) that can be kept in sync.
package download

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
	"github.com/letsmoe/drift/core/stream"
)

const (
	defaultConcurrency  = 2
	defaultMaxAttempts  = 4
	defaultSyncInterval = 6 * time.Hour
	progressInterval    = 250 * time.Millisecond
)

// ErrInvalidVideoID is returned for IDs that are not safe to use as file names.
var ErrInvalidVideoID = errors.New("download: invalid video ID")

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

var (
	errPaused  = errors.New("download paused")
	errRemoved = errors.New("download removed")
)

// Resolver finds the stream to download; *stream.Resolver implements it.
type Resolver interface {
	Resolve(ctx context.Context, videoID string) (stream.Stream, error)
	Invalidate(videoID string)
}

// Catalog reads YouTube Music collections; *innertube.Client implements it.
type Catalog interface {
	Album(ctx context.Context, browseID string) (music.AlbumPage, error)
	Playlist(ctx context.Context, browseID string) (music.PlaylistPage, error)
	PlaylistContinuation(ctx context.Context, token string) (music.PlaylistTracks, error)
}

// Manager owns the download queue. It is safe for concurrent use.
type Manager struct {
	library    *library.Library
	db         *sql.DB
	resolver   Resolver
	catalog    Catalog
	directory  string
	httpClient *http.Client
	logger     *slog.Logger
	now        func() time.Time

	concurrency  int
	maxAttempts  int
	syncInterval time.Duration
	backoff      func(attempt int) time.Duration
	coverURL     func(videoID string) string
	listener     func(Update)

	// syncMutex serializes collection syncs, which read, diff and rewrite membership.
	syncMutex sync.Mutex

	mutex    sync.Mutex
	paused   bool
	allowed  bool
	active   map[string]context.CancelCauseFunc
	progress map[string]*progressState

	ctx    context.Context
	cancel context.CancelFunc
	wake   chan struct{}
	group  sync.WaitGroup
}

// Option configures New.
type Option func(*Manager)

// WithCatalog lets the manager read and refresh album and playlist collections.
func WithCatalog(catalog Catalog) Option {
	return func(manager *Manager) { manager.catalog = catalog }
}

// WithLogger sets the logger; the default discards.
func WithLogger(logger *slog.Logger) Option {
	return func(manager *Manager) { manager.logger = logger }
}

// WithHTTPClient replaces the client that fetches audio and covers.
func WithHTTPClient(client *http.Client) Option {
	return func(manager *Manager) { manager.httpClient = client }
}

// WithConcurrency sets how many tracks download at once (default 2).
func WithConcurrency(concurrency int) Option {
	return func(manager *Manager) { manager.concurrency = concurrency }
}

// WithBackoff sets the wait before retry number attempt (1-based).
func WithBackoff(backoff func(attempt int) time.Duration) Option {
	return func(manager *Manager) { manager.backoff = backoff }
}

// WithMaxAttempts sets how often a track is tried before it is marked failed.
func WithMaxAttempts(attempts int) Option {
	return func(manager *Manager) { manager.maxAttempts = attempts }
}

// WithSyncInterval sets how often kept collections are refreshed (default 6 h).
func WithSyncInterval(interval time.Duration) Option {
	return func(manager *Manager) { manager.syncInterval = interval }
}

// WithCoverURL tells the manager how players reach a downloaded cover, so the
// tracks it lists carry that URL as their thumbnail.
func WithCoverURL(coverURL func(videoID string) string) Option {
	return func(manager *Manager) { manager.coverURL = coverURL }
}

// WithListener receives every status change; progress is throttled to ~4 a
// second per download. The listener must not block.
func WithListener(listener func(Update)) Option {
	return func(manager *Manager) { manager.listener = listener }
}

// New creates a manager that stores files in directory (created when missing).
// Call Start to begin downloading.
func New(lib *library.Library, resolver Resolver, directory string, options ...Option) (*Manager, error) {
	err := os.MkdirAll(directory, 0o700)
	if err != nil {
		return nil, fmt.Errorf("download: create %s: %w", directory, err)
	}
	manager := &Manager{
		library:      lib,
		db:           lib.DB(),
		resolver:     resolver,
		directory:    directory,
		httpClient:   &http.Client{},
		logger:       slog.New(slog.DiscardHandler),
		now:          time.Now,
		concurrency:  defaultConcurrency,
		maxAttempts:  defaultMaxAttempts,
		syncInterval: defaultSyncInterval,
		backoff:      exponentialBackoff,
		allowed:      true,
		active:       map[string]context.CancelCauseFunc{},
		progress:     map[string]*progressState{},
		wake:         make(chan struct{}, 1),
	}
	for _, option := range options {
		option(manager)
	}
	return manager, nil
}

func exponentialBackoff(attempt int) time.Duration {
	return time.Duration(1<<attempt) * time.Second
}

// Start resumes the persisted queue and begins the periodic collection sync.
// Downloads that were running when the app last stopped continue from their
// partial files.
func (manager *Manager) Start(ctx context.Context) error {
	_, err := manager.db.ExecContext(ctx, `UPDATE download SET state = 'queued' WHERE state = 'downloading'`)
	if err != nil {
		return fmt.Errorf("download: reset interrupted downloads: %w", err)
	}
	manager.ctx, manager.cancel = context.WithCancel(ctx)
	manager.library.OnChange(manager.handleLibraryChange)
	manager.group.Add(2)
	go manager.dispatch()
	go manager.syncPeriodically()
	manager.wakeUp()
	return nil
}

// Close stops all transfers and waits for them; queued items stay queued.
func (manager *Manager) Close() {
	if manager.cancel == nil {
		return
	}
	manager.cancel()
	manager.group.Wait()
}

func (manager *Manager) wakeUp() {
	select {
	case manager.wake <- struct{}{}:
	default:
	}
}

// Pause stops all transfers; partial files are kept and Resume continues them.
func (manager *Manager) Pause() {
	manager.mutex.Lock()
	manager.paused = true
	manager.mutex.Unlock()
	manager.interruptActive(errPaused)
	manager.publishAll()
}

// Resume continues a paused queue.
func (manager *Manager) Resume() {
	manager.mutex.Lock()
	manager.paused = false
	manager.mutex.Unlock()
	manager.wakeUp()
	manager.publishAll()
}

// SetAllowDownloads lets the app veto downloading, e.g. off Wi-Fi. Disallowing
// interrupts running transfers like Pause does, without touching the user's pause.
func (manager *Manager) SetAllowDownloads(allowed bool) {
	manager.mutex.Lock()
	manager.allowed = allowed
	manager.mutex.Unlock()
	if allowed {
		manager.wakeUp()
	} else {
		manager.interruptActive(errPaused)
	}
	manager.publishAll()
}

func (manager *Manager) interruptActive(cause error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	for _, cancel := range manager.active {
		cancel(cause)
	}
}

func validateVideoIDs(tracks []music.Track) error {
	for _, track := range tracks {
		if !videoIDPattern.MatchString(track.VideoID) {
			return fmt.Errorf("%w: %q", ErrInvalidVideoID, track.VideoID)
		}
	}
	return nil
}
