package queue

import (
	"context"
	"errors"
	"net"

	"github.com/neoworks/otter-music/core/music"
)

// extensionRequest is what one Extender call is given, plus the queue
// generation it was made for.
type extensionRequest struct {
	generation int
	upcoming   []music.Track
	history    []music.Track
}

// extensionRequestLocked returns a request when the queue should be extended,
// and marks an extension as running. It returns nil otherwise.
func (queue *Queue) extensionRequestLocked() *extensionRequest {
	if queue.extender == nil || queue.extending || !queue.autoplay || queue.repeat != RepeatOff {
		return nil
	}
	if queue.index < 0 || len(queue.items)-queue.index-1 >= autoplayThreshold {
		return nil
	}
	queue.extending = true
	historyStart := max(0, queue.index+1-historyLimit)
	return &extensionRequest{
		generation: queue.generation,
		upcoming:   queue.tracksLocked(queue.index+1, len(queue.items)),
		history:    queue.tracksLocked(historyStart, queue.index+1),
	}
}

func (queue *Queue) tracksLocked(from, to int) []music.Track {
	tracks := make([]music.Track, 0, to-from)
	for _, queued := range queue.items[from:to] {
		tracks = append(tracks, queued.entry.Track)
	}
	return tracks
}

func (queue *Queue) startExtension(request *extensionRequest) {
	if request == nil {
		return
	}
	go queue.runExtension(*request)
}

func (queue *Queue) runExtension(request extensionRequest) {
	tracks, err := queue.extender.Extend(context.Background(), request.upcoming, request.history)
	if err != nil {
		queue.logExtensionFailure(err)
	}
	queue.finishExtension(request.generation, tracks, err == nil)
}

// logExtensionFailure keeps being offline quiet: autoplay simply has nothing to add then.
func (queue *Queue) logExtensionFailure(err error) {
	var networkError net.Error
	if errors.As(err, &networkError) {
		queue.logger.Debug("queue: autoplay extension skipped, offline", "error", err)
		return
	}
	queue.logger.Error("queue: autoplay extension failed", "error", err)
}

// finishExtension appends the extension to the queue it was made for. When the
// queue was replaced meanwhile the result is dropped and the new queue is
// checked for needing an extension of its own.
func (queue *Queue) finishExtension(generation int, tracks []music.Track, succeeded bool) {
	queue.mutex.Lock()
	queue.extending = false
	if generation != queue.generation {
		request := queue.extensionRequestLocked()
		queue.mutex.Unlock()
		queue.startExtension(request)
		return
	}
	if !succeeded {
		queue.mutex.Unlock()
		return
	}
	queue.appendAutoplayLocked(tracks)
	queue.renumberLocked()
	state := queue.stateLocked()
	queue.mutex.Unlock()
	queue.notify(state)
}

func (queue *Queue) appendAutoplayLocked(tracks []music.Track) {
	known := queue.knownVideoIDsLocked()
	for _, track := range tracks {
		if known[track.VideoID] {
			continue
		}
		known[track.VideoID] = true
		queue.insertLocked(len(queue.items), []music.Track{track}, true, queue.endOrderLocked())
	}
}

// knownVideoIDsLocked covers the current and upcoming entries.
func (queue *Queue) knownVideoIDsLocked() map[string]bool {
	known := map[string]bool{}
	for position := max(queue.index, 0); position < len(queue.items); position++ {
		known[queue.items[position].entry.Track.VideoID] = true
	}
	return known
}
