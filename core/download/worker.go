package download

import (
	"context"
	"errors"
	"os"
	"time"
)

// dispatch starts workers for queued downloads whenever there is a free slot.
func (manager *Manager) dispatch() {
	defer manager.group.Done()
	for {
		manager.startQueued()
		select {
		case <-manager.ctx.Done():
			return
		case <-manager.wake:
		}
	}
}

func (manager *Manager) startQueued() {
	for manager.ctx.Err() == nil {
		slot, claimed := manager.claimNext()
		if !claimed {
			return
		}
		manager.group.Add(1)
		go manager.runWorker(slot)
	}
}

// claimNext marks the oldest queued download as downloading and reserves a
// worker slot for it.
func (manager *Manager) claimNext() (claim, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if manager.paused || !manager.allowed || len(manager.active) >= manager.concurrency {
		return claim{}, false
	}
	videoID, found, err := manager.oldestQueued()
	if err != nil {
		manager.logger.Error("download: read queue", "error", err)
		return claim{}, false
	}
	if !found {
		return claim{}, false
	}
	err = manager.markState(manager.ctx, videoID, StateQueued, StateDownloading)
	if err != nil {
		manager.logger.Error("download: claim", "error", err)
		return claim{}, false
	}
	workerContext, cancel := context.WithCancelCause(manager.ctx)
	manager.active[videoID] = cancel
	manager.progress[videoID] = &progressState{}
	return claim{videoID: videoID, ctx: workerContext}, true
}

// claim is a reserved worker slot for one download.
type claim struct {
	videoID string
	ctx     context.Context
}

func (manager *Manager) runWorker(slot claim) {
	defer manager.group.Done()
	manager.publish(slot.videoID, true)

	outcome := manager.downloadWithRetries(slot.ctx, slot.videoID)
	manager.finishWorker(slot.videoID, outcome)
}

// finishWorker releases the slot and records the result of a worker.
func (manager *Manager) finishWorker(videoID string, outcome workerOutcome) {
	manager.mutex.Lock()
	cancel := manager.active[videoID]
	delete(manager.active, videoID)
	manager.mutex.Unlock()
	cancel(nil)

	manager.recordOutcome(videoID, outcome)
	manager.clearProgress(videoID)
	manager.publish(videoID, true)
	manager.wakeUp()
}

func (manager *Manager) recordOutcome(videoID string, outcome workerOutcome) {
	ctx := context.WithoutCancel(manager.ctx)
	var err error
	switch {
	case outcome.interrupted:
		err = manager.markState(ctx, videoID, StateDownloading, StateQueued)
	case outcome.err != nil:
		manager.logger.Warn("download failed", "videoId", videoID, "error", outcome.err)
		err = manager.markFailed(ctx, videoID, outcome.err.Error())
	default:
		err = manager.recordFinished(ctx, videoID, outcome.files)
	}
	if err != nil {
		manager.logger.Error("download: record outcome", "videoId", videoID, "error", err)
	}
}

func (manager *Manager) recordFinished(ctx context.Context, videoID string, files finishedFiles) error {
	stillWanted, err := manager.markDone(ctx, videoID, files)
	if err != nil {
		return err
	}
	if !stillWanted {
		removeFiles(files.audioPath, files.coverPath)
	}
	return nil
}

// workerOutcome is how one download ended: finished, failed, or interrupted
// by pause, removal or shutdown (partial files are kept for resuming).
type workerOutcome struct {
	files       finishedFiles
	err         error
	interrupted bool
}

func (manager *Manager) downloadWithRetries(ctx context.Context, videoID string) workerOutcome {
	var lastErr error
	for attempt := 1; attempt <= manager.maxAttempts; attempt++ {
		files, err := manager.fetchTrack(ctx, videoID)
		if err == nil {
			return workerOutcome{files: files}
		}
		if ctx.Err() != nil {
			return manager.interruptedOutcome(ctx, videoID)
		}
		lastErr = err
		manager.logger.Debug("download attempt failed", "videoId", videoID, "attempt", attempt, "error", err)
		if attempt < manager.maxAttempts && !sleepContext(ctx, manager.backoff(attempt)) {
			return manager.interruptedOutcome(ctx, videoID)
		}
	}
	return workerOutcome{err: lastErr}
}

func (manager *Manager) interruptedOutcome(ctx context.Context, videoID string) workerOutcome {
	if errors.Is(context.Cause(ctx), errRemoved) {
		removeDownloadFiles(manager.directory, videoID)
	}
	return workerOutcome{interrupted: true}
}

func sleepContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func removeFiles(paths ...string) {
	for _, path := range paths {
		if path == "" {
			continue
		}
		_ = os.Remove(path)
	}
}

// cancelActive stops a running worker for videoID, if any.
func (manager *Manager) cancelActive(videoID string, cause error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	cancel, running := manager.active[videoID]
	if running {
		cancel(cause)
	}
}
