package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/neoworks/otter-music/core/account"
	"github.com/neoworks/otter-music/core/download"
	"github.com/neoworks/otter-music/core/importer"
	"github.com/neoworks/otter-music/core/innertube"
	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/lyrics"
	"github.com/neoworks/otter-music/core/palette"
	"github.com/neoworks/otter-music/core/podcasts"
	"github.com/neoworks/otter-music/core/queue"
	"github.com/neoworks/otter-music/core/recommend"
	"github.com/neoworks/otter-music/core/stream"
)

// eventEmitter pushes a named event with its payload to every connected frontend.
type eventEmitter func(name string, data any)

// App holds the methods the frontend calls over RPC (see rpc.go). `bun run bindings`
// generates TypeScript for every exported method into
// apps/desktop/frontend/src/lib/backend/generated.
type App struct {
	ctx       context.Context
	emit      eventEmitter
	logger    *slog.Logger
	innertube *innertube.Client
	podcasts  *podcasts.Directory
	lyrics    *lyrics.Service
	resolver  *stream.Resolver
	proxy     *stream.Proxy
	// library and engine stay nil when the library database could not be opened.
	library  *library.Library
	engine   *recommend.Engine
	importer *importer.Importer
	account  *account.Service
	queue    *queue.Queue
	// downloads stays nil when the library could not be opened.
	downloads     *download.Manager
	palette       *palette.Service
	dataDirectory string

	// streamBaseURL is the backend server's origin; the proxy routes live under it.
	streamBaseURL string
}

// NewApp creates the application. streamBaseURL is the origin the backend
// server listens on; emit receives every event for the frontend.
func NewApp(logger *slog.Logger, streamBaseURL string, emit eventEmitter) *App {
	innertubeClient := innertube.NewClient()
	resolver := stream.NewResolver(innertubeClient)
	app := &App{
		logger:        logger,
		emit:          emit,
		innertube:     innertubeClient,
		podcasts:      podcasts.New(innertubeClient, logger),
		lyrics:        lyrics.New(innertubeClient, lyrics.NewLRCLIB(nil, ""), logger),
		resolver:      resolver,
		proxy:         stream.NewProxy(resolver, logger),
		palette:       newPaletteService(),
		streamBaseURL: streamBaseURL,
	}
	app.queue = app.newQueue()
	return app
}

// startup keeps ctx so long-running work stops when the backend quits, then opens
// the library and resumes downloads.
func (app *App) startup(ctx context.Context) {
	app.ctx = ctx
	app.openLibrary()
	app.startDownloads()
}

// openLibrary opens the local database and builds the recommendation engine on
// it. Failure is logged loudly; library methods then return errors.
func (app *App) openLibrary() {
	path, err := libraryPath()
	if err != nil {
		app.logger.Error("library path", "error", err)
		return
	}
	app.dataDirectory = filepath.Dir(path)
	lib, err := library.Open(app.ctx, path)
	if err != nil {
		app.logger.Error("library failed to open", "path", path, "error", err)
		return
	}
	app.library = lib
	app.engine = recommend.New(lib, app.innertube, recommend.WithLogger(app.logger))
	app.account = account.New(lib)
	app.importer = importer.New(lib, app.innertube, app.logger, importer.WithAccount(app.account.Reader))
}

func libraryPath() (string, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find config directory: %w", err)
	}
	directory := filepath.Join(configDirectory, "otter-music")
	err = os.MkdirAll(directory, 0o700)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", directory, err)
	}
	return filepath.Join(directory, "library.db"), nil
}

func (app *App) shutdown() {
	if app.downloads != nil {
		app.downloads.Close()
	}
	app.closeLibrary()
}

func (app *App) closeLibrary() {
	if app.library == nil {
		return
	}
	err := app.library.Close()
	if err != nil {
		app.logger.Warn("library close", "error", err)
	}
}

// proxyHandler serves the stream proxy and downloaded covers.
func (app *App) proxyHandler() http.Handler {
	return app.proxy.Handler()
}

// StreamURL returns the URL an <audio> element plays videoID from.
func (app *App) StreamURL(videoID string) string {
	return app.streamBaseURL + stream.Path(videoID)
}
