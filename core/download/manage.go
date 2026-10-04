package download

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/stream"
)

// Cancel stops and forgets a download that has not finished. Finished
// downloads stay; use Delete for those.
func (manager *Manager) Cancel(ctx context.Context, videoID string) error {
	statuses, err := manager.States(ctx, []string{videoID})
	if err != nil {
		return err
	}
	status, found := statuses[videoID]
	if !found || status.State == StateDone {
		return nil
	}
	return manager.deleteDownload(ctx, videoID)
}

// Retry queues a failed download again.
func (manager *Manager) Retry(ctx context.Context, videoID string) error {
	_, err := manager.db.ExecContext(ctx, `UPDATE download SET state = 'queued', error = '' WHERE video_id = ? AND state = 'failed'`, videoID)
	if err != nil {
		return fmt.Errorf("download: retry %s: %w", videoID, err)
	}
	manager.wakeUp()
	manager.publish(videoID, true)
	return nil
}

// Delete removes one download and its files, including from any collection.
func (manager *Manager) Delete(ctx context.Context, videoID string) error {
	return manager.deleteDownload(ctx, videoID)
}

// DeleteAll removes every download, collection and file.
func (manager *Manager) DeleteAll(ctx context.Context) error {
	manager.syncMutex.Lock()
	defer manager.syncMutex.Unlock()
	rows, err := manager.db.QueryContext(ctx, `SELECT video_id FROM download`)
	if err != nil {
		return fmt.Errorf("download: list downloads: %w", err)
	}
	videoIDs, err := scanVideoIDs(rows)
	if err != nil {
		return err
	}
	_, err = manager.db.ExecContext(ctx, `DELETE FROM download_collection`)
	if err != nil {
		return fmt.Errorf("download: clear collections: %w", err)
	}
	for _, videoID := range videoIDs {
		err = manager.deleteDownload(ctx, videoID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) deleteDownload(ctx context.Context, videoID string) error {
	manager.cancelActive(videoID, errRemoved)
	_, err := manager.db.ExecContext(ctx, `DELETE FROM download WHERE video_id = ?`, videoID)
	if err != nil {
		return fmt.Errorf("download: delete %s: %w", videoID, err)
	}
	removeDownloadFiles(manager.directory, videoID)
	manager.clearProgress(videoID)
	manager.publish(videoID, true)
	return nil
}

// Queue lists downloads that are not finished: running, queued and failed, oldest first.
func (manager *Manager) Queue(ctx context.Context) ([]Item, error) {
	downloads, err := manager.loadDownloads(ctx, `WHERE state != 'done' ORDER BY queued_at, rowid`)
	return items(downloads), err
}

// Downloaded lists finished downloads, newest first. Their thumbnails point at
// the downloaded cover, so the list renders offline.
func (manager *Manager) Downloaded(ctx context.Context) ([]Item, error) {
	downloads, err := manager.loadDownloads(ctx, `WHERE state = 'done' ORDER BY downloaded_at DESC, rowid DESC`)
	return items(downloads), err
}

// DownloadedTracks lists the tracks of Downloaded.
func (manager *Manager) DownloadedTracks(ctx context.Context) ([]music.Track, error) {
	downloaded, err := manager.Downloaded(ctx)
	if err != nil {
		return nil, err
	}
	tracks := make([]music.Track, 0, len(downloaded))
	for _, item := range downloaded {
		tracks = append(tracks, item.Track)
	}
	return tracks, nil
}

func items(downloads []storedDownload) []Item {
	listed := make([]Item, 0, len(downloads))
	for _, stored := range downloads {
		listed = append(listed, stored.item)
	}
	return listed
}

// StorageUsage totals the audio stored by finished downloads.
func (manager *Manager) StorageUsage(ctx context.Context) (Usage, error) {
	var usage Usage
	err := manager.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(size_bytes), 0), COUNT(*) FROM download WHERE state = 'done'`).Scan(&usage.Bytes, &usage.Count)
	if err != nil {
		return Usage{}, fmt.Errorf("download: storage usage: %w", err)
	}
	return usage, nil
}

// AudioFile implements stream.LocalFiles: the finished download of videoID.
func (manager *Manager) AudioFile(videoID string) (stream.LocalFile, bool) {
	var path, format string
	err := manager.db.QueryRowContext(context.Background(),
		`SELECT file_path, format FROM download WHERE video_id = ? AND state = 'done'`, videoID).Scan(&path, &format)
	if err != nil {
		logUnexpected(manager, err)
		return stream.LocalFile{}, false
	}
	return stream.LocalFile{Path: path, MimeType: format}, path != ""
}

// CoverFile implements stream.LocalFiles: the downloaded cover of videoID.
func (manager *Manager) CoverFile(videoID string) (stream.LocalFile, bool) {
	var path string
	err := manager.db.QueryRowContext(context.Background(),
		`SELECT cover_path FROM download WHERE video_id = ? AND cover_path != ''`, videoID).Scan(&path)
	if err != nil {
		logUnexpected(manager, err)
		return stream.LocalFile{}, false
	}
	return stream.LocalFile{Path: path, MimeType: coverMimeType(path)}, true
}

func logUnexpected(manager *Manager, err error) {
	if !errors.Is(err, sql.ErrNoRows) {
		manager.logger.Warn("download: look up local file", "error", err)
	}
}
