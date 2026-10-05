package main

import (
	"net/http"
	"time"

	"github.com/neoworks/otter-music/core/palette"
)

const paletteFetchTimeout = 10 * time.Second

func newPaletteService() *palette.Service {
	return palette.NewService(&http.Client{Timeout: paletteFetchTimeout})
}

// CoverColors returns the dominant and vibrant colors of a cover image.
func (app *App) CoverColors(url string) (palette.Palette, error) {
	return app.palette.Colors(app.ctx, url)
}
