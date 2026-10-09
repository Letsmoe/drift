package main

import "time"

// savedItemsEvent tells the frontend to reload the saved items.
const savedItemsEvent = "saved-items"

// podcastCoverRefreshInterval is well inside the few days YouTube serves a show cover URL.
const podcastCoverRefreshInterval = 12 * time.Hour

// refreshPodcastCovers renews saved podcasts' expiring cover URLs now and every
// podcastCoverRefreshInterval until the backend quits.
func (app *App) refreshPodcastCovers() {
	if app.library == nil {
		return
	}
	ticker := time.NewTicker(podcastCoverRefreshInterval)
	defer ticker.Stop()
	for {
		app.refreshPodcastCoversOnce()
		select {
		case <-app.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (app *App) refreshPodcastCoversOnce() {
	refreshed, err := app.podcasts.RefreshSavedCovers(app.ctx, app.library)
	if err != nil {
		app.logger.Warn("podcast covers refresh", "error", err)
	}
	if refreshed > 0 {
		app.emit(savedItemsEvent, nil)
	}
}
