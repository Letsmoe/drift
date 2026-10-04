package mobile

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/neoworks/otter-music/core/download"
	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/stream"
)

var errDownloadsUnavailable = errors.New("downloads need the local library, which could not be opened")

// DownloadListener receives download progress. Implemented in Kotlin.
type DownloadListener interface {
	// OnDownloadsChanged receives every download.Update as JSON: one download's
	// status (progress is throttled to ~4 a second) and the queue summary.
	// An empty status.state means the download no longer exists.
	OnDownloadsChanged(updateJSON string)
}

// SetDownloadListener registers the receiver of download updates; nil unregisters it.
func (core *Core) SetDownloadListener(listener DownloadListener) {
	if listener == nil {
		core.downloadListener.Store(nil)
		return
	}
	core.downloadListener.Store(&listener)
}

func (core *Core) emitDownloadUpdate(update download.Update) {
	listener := core.downloadListener.Load()
	if listener == nil {
		return
	}
	(*listener).OnDownloadsChanged(mustEncode(update))
}

// startDownloads creates the download manager once the stream server runs, so
// cover URLs and the proxy's local files are ready before the first download.
func (core *Core) startDownloads(dataDir string) {
	if core.library == nil {
		return
	}
	manager, err := download.New(core.library, core.resolver, filepath.Join(dataDir, "downloads"),
		download.WithCatalog(core.innertube),
		download.WithLogger(core.logger),
		download.WithListener(core.emitDownloadUpdate),
		download.WithCoverURL(core.coverURL),
	)
	if err != nil {
		core.logger.Error("downloads failed to open", "error", err)
		return
	}
	err = manager.Start(core.ctx)
	if err != nil {
		core.logger.Error("downloads failed to start", "error", err)
		return
	}
	core.downloads = manager
	core.proxy.SetLocalFiles(manager)
}

func (core *Core) coverURL(videoID string) string {
	return core.streamBaseURL + stream.CoverPath(videoID)
}

func (core *Core) requireDownloads() (*download.Manager, error) {
	if core.downloads == nil {
		return nil, errDownloadsUnavailable
	}
	return core.downloads, nil
}

func collectionRef(kind, id string) download.CollectionRef {
	return download.CollectionRef{Kind: download.CollectionKind(kind), ID: id}
}

// DownloadTracks downloads songs (music.Track JSON array); kind is "song" or "episode".
func (core *Core) DownloadTracks(tracksJSON string, kind string) error {
	manager, err := core.requireDownloads()
	if err != nil {
		return err
	}
	var tracks []music.Track
	err = decode(tracksJSON, &tracks)
	if err != nil {
		return err
	}
	return manager.DownloadTracks(core.ctx, tracks, music.ItemKind(kind))
}

// DownloadCollection downloads a whole collection. kind is "album", "playlist"
// (YouTube), "local" (a local playlist ID) or "likes" (id "liked"). With keep the
// collection stays in sync as it changes.
func (core *Core) DownloadCollection(kind, id string, keep bool) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.DownloadCollection(ctx, collectionRef(kind, id), keep)
	})
}

// SetKeepDownloaded turns keeping a downloaded collection in sync on or off.
func (core *Core) SetKeepDownloaded(kind, id string, keep bool) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.SetKeep(ctx, collectionRef(kind, id), keep)
	})
}

// RefreshDownloadedCollection re-reads a downloaded collection (call when its page opens).
func (core *Core) RefreshDownloadedCollection(kind, id string) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.RefreshCollection(ctx, collectionRef(kind, id))
	})
}

// RemoveDownloadedCollection forgets a collection and deletes the files nothing else needs.
func (core *Core) RemoveDownloadedCollection(kind, id string) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.RemoveCollection(ctx, collectionRef(kind, id))
	})
}

// CancelDownload stops and forgets an unfinished download.
func (core *Core) CancelDownload(videoID string) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.Cancel(ctx, videoID)
	})
}

// RetryDownload queues a failed download again.
func (core *Core) RetryDownload(videoID string) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.Retry(ctx, videoID)
	})
}

// DeleteDownload removes one downloaded track.
func (core *Core) DeleteDownload(videoID string) error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.Delete(ctx, videoID)
	})
}

// DeleteAllDownloads removes every download and collection.
func (core *Core) DeleteAllDownloads() error {
	return core.withDownloads(func(ctx context.Context, manager *download.Manager) error {
		return manager.DeleteAll(ctx)
	})
}

func (core *Core) withDownloads(run func(ctx context.Context, manager *download.Manager) error) error {
	manager, err := core.requireDownloads()
	if err != nil {
		return err
	}
	return run(core.ctx, manager)
}

// PauseDownloads pauses the whole download queue.
func (core *Core) PauseDownloads() {
	if core.downloads != nil {
		core.downloads.Pause()
	}
}

// ResumeDownloads resumes a paused download queue.
func (core *Core) ResumeDownloads() {
	if core.downloads != nil {
		core.downloads.Resume()
	}
}

// SetAllowDownloads lets the app veto downloading, e.g. off Wi-Fi when the
// "Download only on Wi-Fi" setting is on. Running transfers stop and resume
// from their partial files when allowed again.
func (core *Core) SetAllowDownloads(allowed bool) {
	if core.downloads != nil {
		core.downloads.SetAllowDownloads(allowed)
	}
}

// DownloadStates takes a JSON array of video IDs and returns {"videoId": download.Status};
// videos that were never queued are absent.
func (core *Core) DownloadStates(videoIDsJSON string) (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	var videoIDs []string
	err = decode(videoIDsJSON, &videoIDs)
	if err != nil {
		return "", err
	}
	states, err := manager.States(core.ctx, videoIDs)
	return encodeResult(states, err)
}

// DownloadQueue returns the unfinished downloads as download.Item JSON array.
func (core *Core) DownloadQueue() (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	items, err := manager.Queue(core.ctx)
	return encodeResult(items, err)
}

// DownloadedItems returns the finished downloads as download.Item JSON array, newest first.
func (core *Core) DownloadedItems() (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	items, err := manager.Downloaded(core.ctx)
	return encodeResult(items, err)
}

// DownloadedTracks returns the finished downloads as music.Track JSON array
// (thumbnails point at the downloaded cover), playable offline.
func (core *Core) DownloadedTracks() (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	tracks, err := manager.DownloadedTracks(core.ctx)
	return encodeResult(tracks, err)
}

// DownloadedCollections returns the downloaded collections as download.Collection JSON array.
func (core *Core) DownloadedCollections() (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	collections, err := manager.Collections(core.ctx)
	return encodeResult(collections, err)
}

// DownloadedCollection returns one download.Collection as JSON, or "null" when it is not downloaded.
func (core *Core) DownloadedCollection(kind, id string) (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	collection, found, err := manager.Collection(core.ctx, collectionRef(kind, id))
	if err != nil || !found {
		return "null", err
	}
	return encode(collection)
}

// DownloadSummary returns the queue's download.Summary JSON.
func (core *Core) DownloadSummary() (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	summary, err := manager.Summary(core.ctx)
	return encodeResult(summary, err)
}

// StorageUsage returns download.Usage JSON: bytes and number of downloaded tracks.
func (core *Core) StorageUsage() (string, error) {
	manager, err := core.requireDownloads()
	if err != nil {
		return "", err
	}
	usage, err := manager.StorageUsage(core.ctx)
	return encodeResult(usage, err)
}
