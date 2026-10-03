package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/neoworks/otter-music/core/innertube"
	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/stream"
)

const shutdownTimeout = 2 * time.Second

// App holds the methods bound to the frontend. Wails generates TypeScript for
// every exported method into apps/desktop/frontend/src/lib/wailsjs.
type App struct {
	ctx       context.Context
	logger    *slog.Logger
	innertube *innertube.Client
	proxy     *stream.Proxy

	streamServer  *http.Server
	streamBaseURL string
}

// NewApp creates the bound application.
func NewApp(logger *slog.Logger) *App {
	innertubeClient := innertube.NewClient()
	return &App{
		logger:    logger,
		innertube: innertubeClient,
		proxy:     stream.NewProxy(stream.NewResolver(innertubeClient), logger),
	}
}

// startup keeps the Wails context so bound calls are cancelled when the app quits,
// and starts the stream proxy.
func (app *App) startup(ctx context.Context) {
	app.ctx = ctx
	if err := app.startStreamServer(); err != nil {
		app.logger.Error("stream server failed to start", "error", err)
	}
}

func (app *App) shutdown(_ context.Context) {
	if app.streamServer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := app.streamServer.Shutdown(ctx); err != nil {
		app.logger.Warn("stream server shutdown", "error", err)
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

// SearchSongs returns YouTube Music songs matching query.
func (app *App) SearchSongs(query string) ([]music.Track, error) {
	return app.innertube.SearchSongs(app.ctx, query)
}

// Home returns the first page of the home feed.
func (app *App) Home() (music.HomePage, error) {
	return app.innertube.Home(app.ctx)
}

// HomeContinuation returns the home feed page that continuation points to.
func (app *App) HomeContinuation(continuation string) (music.HomePage, error) {
	return app.innertube.HomeContinuation(app.ctx, continuation)
}

// StreamURL returns the URL an <audio> element plays videoID from.
func (app *App) StreamURL(videoID string) string {
	return app.streamBaseURL + stream.Path(videoID)
}
