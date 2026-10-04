package mobile

// RecommendedRows returns the ranked candidates grouped by reason as recommend.Row JSON.
func (core *Core) RecommendedRows(rowLimit, perRow int) (string, error) {
	if core.engine == nil {
		return "", errRecommendUnavailable
	}
	rows, err := core.engine.Rows(core.ctx, rowLimit, perRow)
	return encodeResult(rows, err)
}

// RecommendedFeed returns one page of the ranked feed as recommend.Item JSON; refresh forces a re-rank.
func (core *Core) RecommendedFeed(limit, offset int, refresh bool) (string, error) {
	if core.engine == nil {
		return "", errRecommendUnavailable
	}
	items, err := core.engine.Feed(core.ctx, limit, offset, refresh)
	return encodeResult(items, err)
}
