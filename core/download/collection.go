package download

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

// DownloadTracks queues tracks the user downloaded one by one. Such tracks stay
// downloaded until deleted, whatever happens to the collections they are in.
func (manager *Manager) DownloadTracks(ctx context.Context, tracks []music.Track, kind music.ItemKind) error {
	err := validateVideoIDs(tracks)
	if err != nil {
		return err
	}
	err = manager.inTransaction(ctx, func(tx *sql.Tx) error {
		for _, track := range tracks {
			err := manager.upsertDownload(ctx, tx, track, kind, true)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	manager.wakeUp()
	manager.publishTracks(tracks)
	return nil
}

func (manager *Manager) publishTracks(tracks []music.Track) {
	for _, track := range tracks {
		manager.publish(track.VideoID, true)
	}
}

// DownloadCollection downloads every track of a collection. With keep the
// collection stays in sync: tracks added later are downloaded, removed ones deleted.
func (manager *Manager) DownloadCollection(ctx context.Context, ref CollectionRef, keep bool) error {
	return manager.syncCollection(ctx, ref, &keep)
}

// SetKeep turns keeping a downloaded collection in sync on or off. Turning it
// on also syncs right away.
func (manager *Manager) SetKeep(ctx context.Context, ref CollectionRef, keep bool) error {
	_, found, err := manager.Collection(ctx, ref)
	if err != nil {
		return err
	}
	if !found {
		return errCollectionNotFound
	}
	return manager.syncCollection(ctx, ref, &keep)
}

// RefreshCollection re-reads a downloaded collection, downloads what is new and
// deletes what left it (unless another collection or an explicit download needs it).
func (manager *Manager) RefreshCollection(ctx context.Context, ref CollectionRef) error {
	return manager.syncCollection(ctx, ref, nil)
}

// RefreshKept refreshes every kept collection; failures are logged and skipped.
func (manager *Manager) RefreshKept(ctx context.Context) {
	collections, err := manager.Collections(ctx)
	if err != nil {
		manager.logger.Warn("download: list collections", "error", err)
		return
	}
	for _, collection := range collections {
		if !collection.Keep {
			continue
		}
		ref := CollectionRef{Kind: collection.Kind, ID: collection.ID}
		err = manager.RefreshCollection(ctx, ref)
		if err != nil {
			manager.logger.Warn("download: refresh collection", "kind", ref.Kind, "id", ref.ID, "error", err)
		}
	}
}

// syncCollection reads the collection's current tracks and makes the stored
// membership match. keep nil keeps the stored setting.
func (manager *Manager) syncCollection(ctx context.Context, ref CollectionRef, keep *bool) error {
	manager.syncMutex.Lock()
	defer manager.syncMutex.Unlock()

	existing, found, err := manager.Collection(ctx, ref)
	if err != nil {
		return err
	}
	if keep == nil && !found {
		return errCollectionNotFound
	}
	current, err := manager.collectionContents(ctx, ref)
	if errors.Is(err, library.ErrNotFound) {
		return manager.removeCollection(ctx, ref)
	}
	if err != nil {
		return err
	}
	keepValue := existing.Keep
	if keep != nil {
		keepValue = *keep
	}
	return manager.applyCollection(ctx, ref, current, keepValue)
}

func (manager *Manager) applyCollection(ctx context.Context, ref CollectionRef, current contents, keep bool) error {
	tracks := downloadableTracks(current.tracks)
	err := manager.inTransaction(ctx, func(tx *sql.Tx) error {
		err := manager.upsertCollection(ctx, tx, ref, current.title, keep)
		if err != nil {
			return err
		}
		for _, track := range tracks {
			err = manager.upsertDownload(ctx, tx, track, music.ItemKindSong, false)
			if err != nil {
				return err
			}
		}
		return replaceMembership(ctx, tx, ref, tracks)
	})
	if err != nil {
		return err
	}
	manager.wakeUp()
	manager.publishTracks(tracks)
	return manager.removeOrphans(ctx)
}

// downloadableTracks drops tracks without a usable ID and repeated tracks.
func downloadableTracks(tracks []music.Track) []music.Track {
	seen := map[string]bool{}
	usable := make([]music.Track, 0, len(tracks))
	for _, track := range tracks {
		if !videoIDPattern.MatchString(track.VideoID) || seen[track.VideoID] {
			continue
		}
		seen[track.VideoID] = true
		usable = append(usable, track)
	}
	return usable
}

func (manager *Manager) upsertCollection(ctx context.Context, db executor, ref CollectionRef, title string, keep bool) error {
	now := library.FormatTime(manager.now())
	_, err := db.ExecContext(ctx, `INSERT INTO download_collection (kind, id, title, keep, created_at, synced_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (kind, id) DO UPDATE SET title = excluded.title, keep = excluded.keep, synced_at = excluded.synced_at`,
		string(ref.Kind), ref.ID, title, boolToInt(keep), now, now)
	if err != nil {
		return fmt.Errorf("download: save collection %s %s: %w", ref.Kind, ref.ID, err)
	}
	return nil
}

func replaceMembership(ctx context.Context, db executor, ref CollectionRef, tracks []music.Track) error {
	_, err := db.ExecContext(ctx, `DELETE FROM download_collection_track WHERE collection_kind = ? AND collection_id = ?`,
		string(ref.Kind), ref.ID)
	if err != nil {
		return fmt.Errorf("download: clear collection %s %s: %w", ref.Kind, ref.ID, err)
	}
	for position, track := range tracks {
		_, err = db.ExecContext(ctx, `INSERT INTO download_collection_track (collection_kind, collection_id, video_id, position)
			VALUES (?, ?, ?, ?)`, string(ref.Kind), ref.ID, track.VideoID, position)
		if err != nil {
			return fmt.Errorf("download: list %s in collection: %w", track.VideoID, err)
		}
	}
	return nil
}

// RemoveCollection forgets a collection and deletes the downloads nothing else needs.
func (manager *Manager) RemoveCollection(ctx context.Context, ref CollectionRef) error {
	manager.syncMutex.Lock()
	defer manager.syncMutex.Unlock()
	return manager.removeCollection(ctx, ref)
}

func (manager *Manager) removeCollection(ctx context.Context, ref CollectionRef) error {
	_, err := manager.db.ExecContext(ctx, `DELETE FROM download_collection WHERE kind = ? AND id = ?`, string(ref.Kind), ref.ID)
	if err != nil {
		return fmt.Errorf("download: remove collection %s %s: %w", ref.Kind, ref.ID, err)
	}
	return manager.removeOrphans(ctx)
}

// removeOrphans deletes downloads that are neither explicit nor in a collection.
func (manager *Manager) removeOrphans(ctx context.Context) error {
	rows, err := manager.db.QueryContext(ctx, `SELECT video_id FROM download WHERE explicit = 0
		AND NOT EXISTS (SELECT 1 FROM download_collection_track WHERE video_id = download.video_id)`)
	if err != nil {
		return fmt.Errorf("download: find orphans: %w", err)
	}
	orphans, err := scanVideoIDs(rows)
	if err != nil {
		return err
	}
	for _, videoID := range orphans {
		err = manager.deleteDownload(ctx, videoID)
		if err != nil {
			return err
		}
	}
	return nil
}

func scanVideoIDs(rows *sql.Rows) ([]string, error) {
	defer func() { _ = rows.Close() }()
	var videoIDs []string
	for rows.Next() {
		var videoID string
		err := rows.Scan(&videoID)
		if err != nil {
			return nil, fmt.Errorf("download: scan video ID: %w", err)
		}
		videoIDs = append(videoIDs, videoID)
	}
	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("download: read video IDs: %w", err)
	}
	return videoIDs, nil
}

// Collection returns one downloaded collection.
func (manager *Manager) Collection(ctx context.Context, ref CollectionRef) (Collection, bool, error) {
	collections, err := manager.Collections(ctx)
	if err != nil {
		return Collection{}, false, err
	}
	index := slices.IndexFunc(collections, func(collection Collection) bool {
		return collection.Kind == ref.Kind && collection.ID == ref.ID
	})
	if index < 0 {
		return Collection{}, false, nil
	}
	return collections[index], true, nil
}

// Collections lists the downloaded collections, oldest first.
func (manager *Manager) Collections(ctx context.Context) ([]Collection, error) {
	rows, err := manager.db.QueryContext(ctx, `SELECT c.kind, c.id, c.title, c.keep, c.synced_at,
			COUNT(ct.video_id),
			COALESCE(SUM(d.state = 'done'), 0),
			COALESCE(SUM(d.state = 'failed'), 0)
		FROM download_collection c
		LEFT JOIN download_collection_track ct ON ct.collection_kind = c.kind AND ct.collection_id = c.id
		LEFT JOIN download d ON d.video_id = ct.video_id
		GROUP BY c.kind, c.id
		ORDER BY c.created_at, c.rowid`)
	if err != nil {
		return nil, fmt.Errorf("download: list collections: %w", err)
	}
	defer func() { _ = rows.Close() }()
	collections := []Collection{}
	for rows.Next() {
		var collection Collection
		var kind, syncedAt string
		var keep int
		err = rows.Scan(&kind, &collection.ID, &collection.Title, &keep, &syncedAt,
			&collection.TrackCount, &collection.DoneCount, &collection.FailedCount)
		if err != nil {
			return nil, fmt.Errorf("download: scan collection: %w", err)
		}
		collection.Kind = CollectionKind(kind)
		collection.Keep = keep == 1
		collection.SyncedAt = parseOptionalTime(syncedAt)
		collections = append(collections, collection)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("download: read collections: %w", err)
	}
	return collections, nil
}

// handleLibraryChange keeps kept local collections in sync with edits to likes and playlists.
func (manager *Manager) handleLibraryChange(change library.Change) {
	ref := CollectionRef{Kind: CollectionLikes, ID: LikedID}
	if change.Kind == library.ChangePlaylist {
		ref = CollectionRef{Kind: CollectionLocalPlaylist, ID: change.ID}
	}
	manager.group.Add(1)
	go func() {
		defer manager.group.Done()
		manager.refreshIfKept(ref)
	}()
}

func (manager *Manager) refreshIfKept(ref CollectionRef) {
	collection, found, err := manager.Collection(manager.ctx, ref)
	if err != nil || !found || !collection.Keep {
		return
	}
	err = manager.RefreshCollection(manager.ctx, ref)
	if err != nil {
		manager.logger.Warn("download: sync collection", "kind", ref.Kind, "id", ref.ID, "error", err)
	}
}

// syncPeriodically refreshes kept collections at startup and then on an interval.
func (manager *Manager) syncPeriodically() {
	defer manager.group.Done()
	manager.RefreshKept(manager.ctx)
	ticker := time.NewTicker(manager.syncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-manager.ctx.Done():
			return
		case <-ticker.C:
			manager.RefreshKept(manager.ctx)
		}
	}
}
