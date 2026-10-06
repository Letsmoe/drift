package main

import (
	"github.com/neoworks/otter-music/core/lyrics"
	"github.com/neoworks/otter-music/core/music"
)

// Lyrics returns the lyrics of a track, or nil when no source has any. allowLRCLIB
// lets the core ask lrclib.net when YouTube has no synced lyrics.
func (app *App) Lyrics(track music.Track, allowLRCLIB bool) (*music.Lyrics, error) {
	found, ok, err := app.lyrics.Lyrics(app.ctx, track, lyrics.Options{AllowLRCLIB: allowLRCLIB})
	if err != nil || !ok {
		return nil, err
	}
	return &found, nil
}
