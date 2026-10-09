// Package mobile adapts the core to gomobile. gomobile cannot export slices,
// maps or most structs, so every complex value crosses the boundary as a JSON
// string with the same shape as the Go type's json tags; errors become Java
// exceptions.
package mobile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/letsmoe/drift-music/core/account"
	"github.com/letsmoe/drift-music/core/download"
	"github.com/letsmoe/drift-music/core/importer"
	"github.com/letsmoe/drift-music/core/innertube"
	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/lyrics"
	"github.com/letsmoe/drift-music/core/podcasts"
	"github.com/letsmoe/drift-music/core/queue"
	"github.com/letsmoe/drift-music/core/recommend"
	"github.com/letsmoe/drift-music/core/stream"
)

const shutdownTimeout = 2 * time.Second

var (
	errLibraryUnavailable   = errors.New("the local library could not be opened")
	errAccountUnavailable   = errors.New("the YouTube Music account needs the local library, which could not be opened")
	errRecommendUnavailable = errors.New("recommendations need the local library, which could not be opened")
)

// Listener receives core events. Implemented in Kotlin.
type Listener interface {
	// OnQueueChanged receives every new queue.State as JSON.
	OnQueueChanged(stateJSON string)
}

// Core is the whole otter-music core behind one gomobile-friendly object.
type Core struct {
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
	listener Listener
	// downloadListener is set by SetDownloadListener while downloads run.
	downloadListener atomic.Pointer[DownloadListener]
	resolver         *stream.Resolver
	innertube        *innertube.Client
	podcasts         *podcasts.Directory
	lyrics           *lyrics.Service
	proxy            *stream.Proxy
	// library and engine stay nil when the library database could not be opened.
	library  *library.Library
	engine   *recommend.Engine
	importer *importer.Importer
	account  *account.Service
	queue    *queue.Queue
	// downloads stays nil when the library could not be opened.
	downloads *download.Manager

	streamServer  *http.Server
	streamBaseURL string
}

// Open opens the library at dataDir/library.db (creating dataDir), builds the
// recommendation engine and queue, and starts the loopback stream proxy on
// 127.0.0.1:<random>. A library that fails to open is logged; library and
// recommendation methods then return errors. listener may be nil.
func Open(dataDir string, listener Listener) (*Core, error) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()}))
	ctx, cancel := context.WithCancel(context.Background())
	innertubeClient := innertube.NewClient()
	resolver := stream.NewResolver(innertubeClient)
	core := &Core{
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
		listener:  listener,
		innertube: innertubeClient,
		podcasts:  podcasts.New(innertubeClient, logger),
		lyrics:    lyrics.New(innertubeClient, lyrics.NewLRCLIB(nil, ""), logger),
		resolver:  resolver,
		proxy:     stream.NewProxy(resolver, logger),
	}
	core.queue = core.newQueue()
	core.openLibrary(dataDir)
	err := core.startStreamServer()
	if err != nil {
		core.Close()
		return nil, err
	}
	core.startDownloads(dataDir)
	return core, nil
}

// logLevel enables debug logs with OTTER_MUSIC_LOG=debug.
func logLevel() slog.Level {
	if os.Getenv("OTTER_MUSIC_LOG") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func (core *Core) openLibrary(dataDir string) {
	err := os.MkdirAll(dataDir, 0o700)
	if err != nil {
		core.logger.Error("create data directory", "path", dataDir, "error", err)
		return
	}
	path := filepath.Join(dataDir, "library.db")
	lib, err := library.Open(core.ctx, path)
	if err != nil {
		core.logger.Error("library failed to open", "path", path, "error", err)
		return
	}
	core.library = lib
	core.engine = recommend.New(lib, core.innertube, recommend.WithLogger(core.logger))
	core.account = account.New(lib)
	core.importer = importer.New(lib, core.innertube, core.logger, importer.WithAccount(core.account.Reader))
}

// startStreamServer serves the proxy on a random loopback port.
func (core *Core) startStreamServer() error {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(core.ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	core.streamServer = &http.Server{
		Handler:           core.proxy.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	core.streamBaseURL = "http://" + listener.Addr().String()

	go func() {
		err := core.streamServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			core.logger.Error("stream server stopped", "error", err)
		}
	}()
	return nil
}

// Close stops the stream server, closes the library and cancels running calls.
func (core *Core) Close() {
	if core.downloads != nil {
		core.downloads.Close()
	}
	core.stopStreamServer()
	if core.library != nil {
		err := core.library.Close()
		if err != nil {
			core.logger.Warn("library close", "error", err)
		}
	}
	core.cancel()
}

func (core *Core) stopStreamServer() {
	if core.streamServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err := core.streamServer.Shutdown(ctx)
	if err != nil {
		core.logger.Warn("stream server shutdown", "error", err)
	}
}

// StreamURL returns the URL a player streams videoID from.
func (core *Core) StreamURL(videoID string) string {
	return core.streamBaseURL + stream.Path(videoID)
}

// requireLibrary returns the open library, or an error when Open could not open it.
func (core *Core) requireLibrary() (*library.Library, error) {
	if core.library == nil {
		return nil, errLibraryUnavailable
	}
	return core.library, nil
}
