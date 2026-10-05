package main

import (
	"github.com/neoworks/otter-music/core/music"
)

// Lyrics returns YouTube Music's lyrics for a track, or nil when it has none.
func (app *App) Lyrics(videoID string) (*music.Lyrics, error) {
	lyrics, found, err := app.innertube.Lyrics(app.ctx, videoID)
	if err != nil || !found {
		return nil, err
	}
	return &lyrics, nil
}
