package main

import (
	"errors"

	"github.com/neoworks/otter-music/core/recommend"
)

var errRecommendUnavailable = errors.New("recommendations need the local library, which could not be opened")

// RecommendedRows returns the ranked candidates grouped by reason.
func (app *App) RecommendedRows(rowLimit, perRow int) ([]recommend.Row, error) {
	if app.engine == nil {
		return nil, errRecommendUnavailable
	}
	return app.engine.Rows(app.ctx, rowLimit, perRow)
}

// RecommendedFeed returns one page of the ranked feed; refresh forces a re-rank.
func (app *App) RecommendedFeed(limit, offset int, refresh bool) ([]recommend.Item, error) {
	if app.engine == nil {
		return nil, errRecommendUnavailable
	}
	return app.engine.Feed(app.ctx, limit, offset, refresh)
}
