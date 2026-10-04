package download

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type progressState struct {
	done     int64
	total    int64
	lastEmit time.Time
}

// fillStatus derives the reported status from the stored state, the live
// progress and whether the queue is paused or blocked.
func (manager *Manager) fillStatus(stored *storedDownload) {
	status := &stored.item.Status
	status.State = stored.state
	status.Error = stored.errorText
	if stored.state == StateDone {
		status.BytesDone = stored.item.SizeBytes
		status.BytesTotal = stored.item.SizeBytes
		return
	}
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	progress, found := manager.progress[status.VideoID]
	if found {
		status.BytesDone = progress.done
		status.BytesTotal = progress.total
	}
	inFlight := stored.state == StateQueued || stored.state == StateDownloading
	if inFlight && (manager.paused || !manager.allowed) {
		status.State = StatePaused
	}
}

// States returns the status of each given video; videos never queued are absent.
func (manager *Manager) States(ctx context.Context, videoIDs []string) (map[string]Status, error) {
	statuses := map[string]Status{}
	if len(videoIDs) == 0 {
		return statuses, nil
	}
	arguments := make([]any, len(videoIDs))
	for index, videoID := range videoIDs {
		arguments[index] = videoID
	}
	downloads, err := manager.loadDownloads(ctx, `WHERE video_id IN (`+placeholders(len(videoIDs))+`)`, arguments...)
	if err != nil {
		return nil, err
	}
	for _, stored := range downloads {
		statuses[stored.item.Status.VideoID] = stored.item.Status
	}
	return statuses, nil
}

func placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

// Summary counts the downloads by state.
func (manager *Manager) Summary(ctx context.Context) (Summary, error) {
	rows, err := manager.db.QueryContext(ctx, `SELECT state, COUNT(*) FROM download GROUP BY state`)
	if err != nil {
		return Summary{}, fmt.Errorf("download: summarize: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var summary Summary
	for rows.Next() {
		var state string
		var count int
		err = rows.Scan(&state, &count)
		if err != nil {
			return Summary{}, fmt.Errorf("download: scan summary: %w", err)
		}
		addToSummary(&summary, State(state), count)
	}
	err = rows.Err()
	if err != nil {
		return Summary{}, fmt.Errorf("download: read summary: %w", err)
	}
	manager.mutex.Lock()
	summary.Paused = manager.paused
	summary.Blocked = !manager.allowed
	manager.mutex.Unlock()
	return summary, nil
}

func addToSummary(summary *Summary, state State, count int) {
	switch state {
	case StateQueued:
		summary.Queued += count
	case StateDownloading:
		summary.Downloading += count
	case StateFailed:
		summary.Failed += count
	case StateDone:
		summary.Done += count
	case StatePaused:
	}
}

// publish tells the listener about one download. Progress-only updates are
// throttled per download; force skips the throttle.
func (manager *Manager) publish(videoID string, force bool) {
	if manager.listener == nil {
		return
	}
	if !force && !manager.progressDue(videoID) {
		return
	}
	ctx := manager.listenerContext()
	downloads, err := manager.loadDownloads(ctx, `WHERE video_id = ?`, videoID)
	if err != nil || len(downloads) == 0 {
		manager.publishRemoved(videoID)
		return
	}
	summary, err := manager.Summary(ctx)
	if err != nil {
		manager.logger.Warn("download summary", "error", err)
		return
	}
	manager.listener(Update{Status: downloads[0].item.Status, Summary: summary})
}

// publishRemoved reports a download that no longer exists as "not downloaded".
func (manager *Manager) publishRemoved(videoID string) {
	summary, err := manager.Summary(manager.listenerContext())
	if err != nil {
		manager.logger.Warn("download summary", "error", err)
		return
	}
	manager.listener(Update{Status: Status{VideoID: videoID}, Summary: summary})
}

// publishAll reports every in-flight download, after pause, resume or a policy change.
func (manager *Manager) publishAll() {
	if manager.listener == nil {
		return
	}
	downloads, err := manager.loadDownloads(manager.listenerContext(), `WHERE state IN ('queued', 'downloading')`)
	if err != nil {
		manager.logger.Warn("download publish", "error", err)
		return
	}
	for _, stored := range downloads {
		manager.publish(stored.item.Status.VideoID, true)
	}
	if len(downloads) == 0 {
		manager.publishSummaryOnly()
	}
}

func (manager *Manager) publishSummaryOnly() {
	summary, err := manager.Summary(manager.listenerContext())
	if err != nil {
		return
	}
	manager.listener(Update{Summary: summary})
}

func (manager *Manager) listenerContext() context.Context {
	if manager.ctx != nil {
		return manager.ctx
	}
	return context.Background()
}

func (manager *Manager) progressDue(videoID string) bool {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	progress, found := manager.progress[videoID]
	if !found {
		return true
	}
	now := manager.now()
	if now.Sub(progress.lastEmit) < progressInterval {
		return false
	}
	progress.lastEmit = now
	return true
}

func (manager *Manager) setProgress(videoID string, done, total int64) {
	manager.mutex.Lock()
	progress, found := manager.progress[videoID]
	if !found {
		progress = &progressState{}
		manager.progress[videoID] = progress
	}
	progress.done = done
	progress.total = total
	manager.mutex.Unlock()
	manager.publish(videoID, false)
}

func (manager *Manager) clearProgress(videoID string) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	delete(manager.progress, videoID)
}
