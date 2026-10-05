package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/neoworks/otter-music/core/account"
	"github.com/neoworks/otter-music/core/download"
	"github.com/neoworks/otter-music/core/importer"
	"github.com/neoworks/otter-music/core/innertube"
	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/palette"
	"github.com/neoworks/otter-music/core/queue"
	"github.com/neoworks/otter-music/core/recommend"
	"github.com/neoworks/otter-music/core/stream"
)

const shutdownTimeout = 2 * time.Second

// App holds the methods bound to the frontend. Wails generates TypeScript for
// every exported method into apps/desktop/frontend/src/lib/wailsjs.
type App struct {
	ctx       context.Context
	logger    *slog.Logger
	innertube *innertube.Client
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

	streamServer  *http.Server
	streamBaseURL string
}

// NewApp creates the bound application.
func NewApp(logger *slog.Logger) *App {
	innertubeClient := innertube.NewClient()
	resolver := stream.NewResolver(innertubeClient)
	app := &App{
		logger:    logger,
		innertube: innertubeClient,
		resolver:  resolver,
		proxy:     stream.NewProxy(resolver, logger),
		palette:   newPaletteService(),
	}
	app.queue = app.newQueue()
	return app
}

// startup keeps the Wails context so bound calls are cancelled when the app quits,
// and starts the stream proxy.
func (app *App) startup(ctx context.Context) {
	app.ctx = ctx
	app.openLibrary()
	if err := app.startStreamServer(); err != nil {
		app.logger.Error("stream server failed to start", "error", err)
	}
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

func (app *App) shutdown(_ context.Context) {
	if app.downloads != nil {
		app.downloads.Close()
	}
	app.closeLibrary()
	if app.streamServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := app.streamServer.Shutdown(ctx); err != nil {
		app.logger.Warn("stream server shutdown", "error", err)
	}
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

// startStreamServer serves the proxy on a random loopback port. Audio goes over
// plain HTTP rather than the Wails asset server, because WebKitGTK's media
// pipeline does not reliably seek over custom URI schemes.
func (app *App) startStreamServer() error {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(app.ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	app.streamServer = &http.Server{
		Handler:           app.proxy.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	app.streamBaseURL = "http://" + listener.Addr().String()

	go func() {
		err := app.streamServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("stream server stopped", "error", err)
		}
	}()
	return nil
}

// StreamURL returns the URL an <audio> element plays videoID from.
func (app *App) StreamURL(videoID string) string {
	return app.streamBaseURL + stream.Path(videoID)
}
