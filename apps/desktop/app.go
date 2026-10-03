package main

import (
	"context"

	"github.com/neoworks/otter-music/core/innertube"
	"github.com/neoworks/otter-music/core/music"
)

// App holds the methods bound to the frontend. Wails generates TypeScript for
// every exported method into apps/desktop/frontend/src/lib/wailsjs.
type App struct {
	ctx       context.Context
	innertube *innertube.Client
}

// NewApp creates the bound application.
func NewApp() *App {
	return &App{innertube: innertube.NewClient()}
}

// startup keeps the Wails context so bound calls are cancelled when the app quits.
func (app *App) startup(ctx context.Context) {
	app.ctx = ctx
}

// SearchSongs returns YouTube Music songs matching query.
func (app *App) SearchSongs(query string) ([]music.Track, error) {
	return app.innertube.SearchSongs(app.ctx, query)
}
