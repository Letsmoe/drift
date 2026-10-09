package main

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/letsmoe/drift-music/core/download"
	"github.com/letsmoe/drift-music/core/music"
	"github.com/letsmoe/drift-music/core/stream"
)

// downloadsEvent is the event carrying every download.Update.
const downloadsEvent = "downloads"

var errDownloadsUnavailable = errors.New("downloads need the local library, which could not be opened")

// startDownloads starts the download manager once the library and the stream
// server exist: cover URLs need the proxy's address.
func (app *App) startDownloads() {
	if app.library == nil {
		return
	}
	manager, err := download.New(app.library, app.resolver, filepath.Join(app.dataDirectory, "downloads"),
		download.WithCatalog(app.innertube),
		download.WithLogger(app.logger),
		download.WithListener(app.emitDownloadUpdate),
		download.WithCoverURL(func(videoID string) string { return app.streamBaseURL + stream.CoverPath(videoID) }),
	)
	if err != nil {
		app.logger.Error("downloads failed to open", "error", err)
		return
	}
	err = manager.Start(app.ctx)
	if err != nil {
		app.logger.Error("downloads failed to start", "error", err)
		return
	}
	app.downloads = manager
	app.proxy.SetLocalFiles(manager)
}

func (app *App) emitDownloadUpdate(update download.Update) {
	app.emit(downloadsEvent, update)
}

func (app *App) requireDownloads() (*download.Manager, error) {
	if app.downloads == nil {
		return nil, errDownloadsUnavailable
	}
	return app.downloads, nil
}

func (app *App) withDownloads(run func(ctx context.Context, manager *download.Manager) error) error {
	manager, err := app.requireDownloads()
	if err != nil {
		return err
	}
	return run(app.ctx, manager)
}

func downloadRef(kind, id string) download.CollectionRef {
	return download.CollectionRef{Kind: download.CollectionKind(kind), ID: id}
}

// DownloadTracks downloads songs or episodes; kind is "song" or "episode".
func (app *App) DownloadTracks(tracks []music.Track, kind string) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.DownloadTracks(ctx, tracks, music.ItemKind(kind))
	})
}

// DownloadCollection downloads a whole collection: kind is "album" or
// "playlist" (YouTube), "local" (local playlist ID) or "likes" (id "liked").
// With keep it stays in sync as the collection changes.
func (app *App) DownloadCollection(kind, id string, keep bool) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.DownloadCollection(ctx, downloadRef(kind, id), keep)
	})
}

// SetKeepDownloaded turns keeping a downloaded collection in sync on or off.
func (app *App) SetKeepDownloaded(kind, id string, keep bool) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.SetKeep(ctx, downloadRef(kind, id), keep)
	})
}

// RefreshDownloadedCollection re-reads a downloaded collection; the UI calls it
// when the collection's page opens.
func (app *App) RefreshDownloadedCollection(kind, id string) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.RefreshCollection(ctx, downloadRef(kind, id))
	})
}

// RemoveDownloadedCollection forgets a collection and deletes the files nothing else needs.
func (app *App) RemoveDownloadedCollection(kind, id string) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.RemoveCollection(ctx, downloadRef(kind, id))
	})
}

// CancelDownload stops and forgets an unfinished download.
func (app *App) CancelDownload(videoID string) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.Cancel(ctx, videoID)
	})
}

// RetryDownload queues a failed download again.
func (app *App) RetryDownload(videoID string) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.Retry(ctx, videoID)
	})
}

// DeleteDownload removes one downloaded track.
func (app *App) DeleteDownload(videoID string) error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.Delete(ctx, videoID)
	})
}

// DeleteAllDownloads removes every download and collection.
func (app *App) DeleteAllDownloads() error {
	return app.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.DeleteAll(ctx)
	})
}

// PauseDownloads pauses the whole download queue.
func (app *App) PauseDownloads() {
	if app.downloads != nil {
		app.downloads.Pause()
	}
}

// ResumeDownloads resumes a paused download queue.
func (app *App) ResumeDownloads() {
	if app.downloads != nil {
		app.downloads.Resume()
	}
}

// DownloadStates returns the download status of each video; never queued videos are absent.
func (app *App) DownloadStates(videoIDs []string) (map[string]download.Status, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return nil, err
	}
	return manager.States(app.ctx, videoIDs)
}

// DownloadQueue lists unfinished downloads: running, queued and failed.
func (app *App) DownloadQueue() ([]download.Item, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return nil, err
	}
	return manager.Queue(app.ctx)
}

// DownloadedItems lists finished downloads, newest first.
func (app *App) DownloadedItems() ([]download.Item, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return nil, err
	}
	return manager.Downloaded(app.ctx)
}

// DownloadedTracks lists the finished downloads as tracks that play offline.
func (app *App) DownloadedTracks() ([]music.Track, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return nil, err
	}
	return manager.DownloadedTracks(app.ctx)
}

// DownloadedCollections lists the downloaded collections.
func (app *App) DownloadedCollections() ([]download.Collection, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return nil, err
	}
	return manager.Collections(app.ctx)
}

// DownloadSummary counts the downloads by state.
func (app *App) DownloadSummary() (download.Summary, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return download.Summary{}, err
	}
	return manager.Summary(app.ctx)
}

// StorageUsage totals the storage taken by finished downloads.
func (app *App) StorageUsage() (download.Usage, error) {
	manager, err := app.requireDownloads()
	if err != nil {
		return download.Usage{}, err
	}
	return manager.StorageUsage(app.ctx)
}
