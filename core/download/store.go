package download

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

const itemColumns = `video_id, track, kind, state, explicit, file_path, cover_path, format, size_bytes, error, downloaded_at`

// storedDownload is one download row.
type storedDownload struct {
	item      Item
	state     State
	filePath  string
	coverPath string
	errorText string
}

func (manager *Manager) inTransaction(ctx context.Context, change func(tx *sql.Tx) error) error {
	tx, err := manager.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("download: begin transaction: %w", err)
	}
	err = change(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("download: commit: %w", err)
	}
	return nil
}

// upsertDownload queues a track, or refreshes the metadata of a known one. A
// failed download is queued again; explicit is never cleared by an upsert.
func (manager *Manager) upsertDownload(ctx context.Context, db executor, track music.Track, kind music.ItemKind, explicit bool) error {
	encoded, err := json.Marshal(track)
	if err != nil {
		return fmt.Errorf("download: encode track %s: %w", track.VideoID, err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO download (video_id, track, kind, state, explicit, queued_at)
		VALUES (?, ?, ?, 'queued', ?, ?)
		ON CONFLICT (video_id) DO UPDATE SET
			track = excluded.track,
			explicit = MAX(download.explicit, excluded.explicit),
			error = CASE WHEN download.state = 'failed' THEN '' ELSE download.error END,
			state = CASE WHEN download.state = 'failed' THEN 'queued' ELSE download.state END`,
		track.VideoID, string(encoded), string(kind), boolToInt(explicit), library.FormatTime(manager.now()))
	if err != nil {
		return fmt.Errorf("download: queue %s: %w", track.VideoID, err)
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (manager *Manager) loadDownloads(ctx context.Context, where string, args ...any) ([]storedDownload, error) {
	rows, err := manager.db.QueryContext(ctx, `SELECT `+itemColumns+` FROM download `+where, args...)
	if err != nil {
		return nil, fmt.Errorf("download: query: %w", err)
	}
	defer func() { _ = rows.Close() }()
	downloads := []storedDownload{}
	for rows.Next() {
		stored, err := manager.scanDownload(rows)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, stored)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("download: read rows: %w", err)
	}
	return downloads, nil
}

func (manager *Manager) scanDownload(row interface{ Scan(dest ...any) error }) (storedDownload, error) {
	var stored storedDownload
	var trackJSON, kind, state, downloadedAt string
	var explicit int
	err := row.Scan(&stored.item.Status.VideoID, &trackJSON, &kind, &state, &explicit, &stored.filePath,
		&stored.coverPath, &stored.item.Format, &stored.item.SizeBytes, &stored.errorText, &downloadedAt)
	if err != nil {
		return storedDownload{}, fmt.Errorf("download: scan: %w", err)
	}
	err = json.Unmarshal([]byte(trackJSON), &stored.item.Track)
	if err != nil {
		return storedDownload{}, fmt.Errorf("download: decode track %s: %w", stored.item.Status.VideoID, err)
	}
	stored.item.Kind = music.ItemKind(kind)
	stored.item.Explicit = explicit == 1
	stored.state = State(state)
	stored.item.DownloadedAt = parseOptionalTime(downloadedAt)
	manager.fillStatus(&stored)
	manager.attachCover(&stored)
	return stored, nil
}

func parseOptionalTime(text string) time.Time {
	moment, err := library.ParseTime(text)
	if err != nil {
		return time.Time{}
	}
	return moment
}

// attachCover replaces the remote thumbnails with the downloaded cover.
func (manager *Manager) attachCover(stored *storedDownload) {
	if stored.coverPath == "" || manager.coverURL == nil {
		return
	}
	stored.item.Track.Thumbnails = []music.Thumbnail{{URL: manager.coverURL(stored.item.Status.VideoID)}}
}

func (manager *Manager) markState(ctx context.Context, videoID string, from, to State) error {
	_, err := manager.db.ExecContext(ctx, `UPDATE download SET state = ? WHERE video_id = ? AND state = ?`, string(to), videoID, string(from))
	if err != nil {
		return fmt.Errorf("download: mark %s %s: %w", videoID, to, err)
	}
	return nil
}

// markDone records the finished files; it reports false when the row was
// removed meanwhile, in which case the caller deletes the files.
func (manager *Manager) markDone(ctx context.Context, videoID string, file finishedFiles) (bool, error) {
	result, err := manager.db.ExecContext(ctx, `UPDATE download SET state = 'done', file_path = ?, cover_path = ?,
		format = ?, size_bytes = ?, error = '', downloaded_at = ? WHERE video_id = ?`,
		file.audioPath, file.coverPath, file.mimeType, file.sizeBytes, library.FormatTime(manager.now()), videoID)
	if err != nil {
		return false, fmt.Errorf("download: mark %s done: %w", videoID, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("download: mark %s done: %w", videoID, err)
	}
	return affected > 0, nil
}

func (manager *Manager) markFailed(ctx context.Context, videoID, message string) error {
	_, err := manager.db.ExecContext(ctx, `UPDATE download SET state = 'failed', error = ? WHERE video_id = ?`, message, videoID)
	if err != nil {
		return fmt.Errorf("download: mark %s failed: %w", videoID, err)
	}
	return nil
}

// finishedFiles describes what a completed download left on disk.
type finishedFiles struct {
	audioPath string
	coverPath string
	mimeType  string
	sizeBytes int64
}

// oldestQueued returns the next download to run. It must not take manager.mutex.
func (manager *Manager) oldestQueued() (string, bool, error) {
	var videoID string
	err := manager.db.QueryRowContext(manager.ctx,
		`SELECT video_id FROM download WHERE state = 'queued' ORDER BY queued_at, rowid LIMIT 1`).Scan(&videoID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("download: find next: %w", err)
	}
	return videoID, true, nil
}
