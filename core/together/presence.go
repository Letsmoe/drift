package together

import (
	"context"
	"math"
	"slices"
	"sync"
	"time"

	"github.com/letsmoe/drift-music/core/music"
	"github.com/letsmoe/drift-music/core/queue"
)

const (
	// Presence goes out at least this often while online.
	presenceHeartbeat = 20 * time.Second
	// A friend is offline after this long without presence (three missed heartbeats).
	presenceTimeout = 3*presenceHeartbeat + 5*time.Second
	// Changes within this window go out as one message.
	presenceDebounce = time.Second
	// A position this far from where friends expect it counts as a seek.
	presenceSeekThresholdSeconds = 3.0
	// The offline message on quit waits at most this long for a relay.
	offlineSendTimeout = 2 * time.Second
)

// NowPlaying is what someone is playing.
type NowPlaying struct {
	Track           music.Track `json:"track"`
	PositionSeconds float64     `json:"positionSeconds"`
	Paused          bool        `json:"paused"`
	// UpdatedAt is when PositionSeconds was current; add the time since while playing.
	UpdatedAt time.Time `json:"updatedAt"`
}

// Presence is a friend who is online.
type Presence struct {
	// NowPlaying is nil when nothing is loaded.
	NowPlaying *NowPlaying `json:"nowPlaying"`
	// SessionOpen is true while they host a session.
	SessionOpen bool      `json:"sessionOpen"`
	SeenAt      time.Time `json:"seenAt"`
}

type presenceMessage struct {
	Type string   `json:"type"`
	From Identity `json:"from"`
	// Online is false in the last message before quitting or stopping to share.
	Online       bool               `json:"online"`
	NowPlaying   *nowPlayingMessage `json:"nowPlaying,omitempty"`
	SessionOpen  bool               `json:"sessionOpen"`
	SentAtMillis int64              `json:"sentAt"`
}

type nowPlayingMessage struct {
	Track           music.Track `json:"track"`
	PositionSeconds float64     `json:"positionSeconds"`
	Paused          bool        `json:"paused"`
}

// presenceTracker holds the user's own presence and what friends shared.
type presenceTracker struct {
	mutex       sync.Mutex
	share       bool
	track       *music.Track
	entryID     int64
	report      PlaybackReport
	reportedAt  time.Time
	sessionOpen bool
	// What friends last got, to tell real changes from progress.
	published nowPlayingState
	pending   bool
	friends   map[string]Presence
}

type nowPlayingState struct {
	videoID         string
	paused          bool
	positionSeconds float64
	at              time.Time
	sessionOpen     bool
}

// presenceQueueChanged shares a new current track.
func (together *Together) presenceQueueChanged(state queue.State) {
	tracker := &together.presence
	tracker.mutex.Lock()
	tracker.track = nil
	tracker.entryID = 0
	if state.Index >= 0 && state.Index < len(state.Entries) {
		entry := state.Entries[state.Index]
		tracker.track = &entry.Track
		tracker.entryID = entry.ID
	}
	changed := tracker.currentVideoIDLocked() != tracker.published.videoID
	tracker.mutex.Unlock()
	if changed {
		together.schedulePresence()
	}
}

// presencePlaybackReported shares a pause, resume or seek; plain progress is
// left to the heartbeat.
func (together *Together) presencePlaybackReported(report PlaybackReport) {
	tracker := &together.presence
	tracker.mutex.Lock()
	tracker.report = report
	tracker.reportedAt = together.now()
	changed := tracker.differsFromPublishedLocked(together.now())
	tracker.mutex.Unlock()
	if changed {
		together.schedulePresence()
	}
}

// presenceSessionChanged shares whether friends can join the user now.
func (together *Together) presenceSessionChanged(snapshot *roomSnapshot) {
	open := snapshot != nil && snapshot.Hosting
	tracker := &together.presence
	tracker.mutex.Lock()
	changed := tracker.sessionOpen != open
	tracker.sessionOpen = open
	tracker.mutex.Unlock()
	if changed {
		together.schedulePresence()
	}
}

func (tracker *presenceTracker) currentVideoIDLocked() string {
	if tracker.track == nil {
		return ""
	}
	return tracker.track.VideoID
}

func (tracker *presenceTracker) differsFromPublishedLocked(now time.Time) bool {
	current, ok := tracker.nowPlayingLocked(now)
	if !ok {
		return tracker.published.videoID != ""
	}
	published := tracker.published
	if current.Track.VideoID != published.videoID || current.Paused != published.paused {
		return true
	}
	expected := published.positionSeconds
	if !published.paused {
		expected += now.Sub(published.at).Seconds()
	}
	return math.Abs(current.PositionSeconds-expected) > presenceSeekThresholdSeconds
}

// nowPlayingLocked extrapolates the last report of the current entry to now.
func (tracker *presenceTracker) nowPlayingLocked(now time.Time) (nowPlayingMessage, bool) {
	if tracker.track == nil {
		return nowPlayingMessage{}, false
	}
	current := nowPlayingMessage{Track: trimmedTrack(*tracker.track), Paused: tracker.report.Paused}
	if tracker.report.EntryID != tracker.entryID || tracker.reportedAt.IsZero() {
		return current, true
	}
	current.PositionSeconds = tracker.report.PositionSeconds
	if !tracker.report.Paused {
		current.PositionSeconds += now.Sub(tracker.reportedAt).Seconds()
	}
	return current, true
}

// trimmedTrack keeps a presence message small: the smallest cover only.
func trimmedTrack(track music.Track) music.Track {
	track.PlayCount = ""
	// Apps read these as lists; nil would arrive as null.
	track.Artists = append([]music.Artist{}, track.Artists...)
	track.Thumbnails = append([]music.Thumbnail{}, track.Thumbnails...)
	if len(track.Thumbnails) > 1 {
		smallest := slices.MinFunc(track.Thumbnails, func(left, right music.Thumbnail) int {
			return left.Width - right.Width
		})
		track.Thumbnails = []music.Thumbnail{smallest}
	}
	return track
}

// schedulePresence publishes soon, merging changes that come in a burst.
func (together *Together) schedulePresence() {
	tracker := &together.presence
	tracker.mutex.Lock()
	if tracker.pending || !tracker.share {
		tracker.mutex.Unlock()
		return
	}
	tracker.pending = true
	tracker.mutex.Unlock()
	time.AfterFunc(presenceDebounce, together.publishPresence)
}

func (together *Together) publishPresence() {
	message, ok := together.nextPresenceMessage()
	if !ok {
		return
	}
	together.sendPresence(message, lobbySendTimeout)
}

// nextPresenceMessage builds the presence to send and records it as published.
func (together *Together) nextPresenceMessage() (presenceMessage, bool) {
	tracker := &together.presence
	now := together.now()
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	tracker.pending = false
	if !tracker.share {
		return presenceMessage{}, false
	}
	message := presenceMessage{
		Type:         presenceMessageType,
		From:         together.self(),
		Online:       true,
		SessionOpen:  tracker.sessionOpen,
		SentAtMillis: now.UnixMilli(),
	}
	tracker.published = nowPlayingState{sessionOpen: tracker.sessionOpen, at: now}
	if current, ok := tracker.nowPlayingLocked(now); ok {
		message.NowPlaying = &current
		tracker.published.videoID = current.Track.VideoID
		tracker.published.paused = current.Paused
		tracker.published.positionSeconds = current.PositionSeconds
	}
	return message, true
}

func (together *Together) offlineMessage() presenceMessage {
	return presenceMessage{
		Type:         presenceMessageType,
		From:         together.self(),
		Online:       false,
		SentAtMillis: together.now().UnixMilli(),
	}
}

func (together *Together) sendPresence(message presenceMessage, timeout time.Duration) {
	opened := together.currentLobby()
	if opened == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := opened.publishPresence(ctx, message); err != nil {
		together.logger.Debug("together: presence", "error", err)
	}
}

// runPresence sends the heartbeat and drops friends that went quiet, until ctx ends.
func (together *Together) runPresence(ctx context.Context) {
	together.schedulePresence()
	ticker := time.NewTicker(presenceHeartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go together.publishPresence()
			if together.dropSilentFriends() {
				together.notify()
			}
		}
	}
}

// receivePresence records a friend's presence.
func (together *Together) receivePresence(userID string, message presenceMessage) {
	tracker := &together.presence
	now := together.now()
	tracker.mutex.Lock()
	if message.Online {
		tracker.friends[userID] = presenceOf(message, now)
	} else {
		delete(tracker.friends, userID)
	}
	tracker.mutex.Unlock()
	together.notify()
}

// presenceOf turns a message into Presence, moving a playing position on by
// the transit time (at most maxTransitDelay, against skewed clocks).
func presenceOf(message presenceMessage, now time.Time) Presence {
	presence := Presence{SessionOpen: message.SessionOpen, SeenAt: now}
	if message.NowPlaying == nil {
		return presence
	}
	playing := message.NowPlaying
	position := playing.PositionSeconds
	if !playing.Paused {
		transit := min(max(now.Sub(time.UnixMilli(message.SentAtMillis)), 0), maxTransitDelay)
		position += transit.Seconds()
	}
	presence.NowPlaying = &NowPlaying{
		Track:           playing.Track,
		PositionSeconds: position,
		Paused:          playing.Paused,
		UpdatedAt:       now,
	}
	return presence
}

func (together *Together) dropSilentFriends() bool {
	tracker := &together.presence
	cutoff := together.now().Add(-presenceTimeout)
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	dropped := false
	for userID, presence := range tracker.friends {
		if presence.SeenAt.Before(cutoff) {
			delete(tracker.friends, userID)
			dropped = true
		}
	}
	return dropped
}

func (together *Together) friendPresence(userID string) *Presence {
	tracker := &together.presence
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	presence, ok := tracker.friends[userID]
	if !ok {
		return nil
	}
	return &presence
}

func (together *Together) sharingPresence() bool {
	tracker := &together.presence
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()
	return tracker.share
}

// SetSharePresence turns sharing what the user plays on or off; turning it off
// tells friends the user went offline.
func (together *Together) SetSharePresence(ctx context.Context, share bool) error {
	if err := together.saveIdentity(ctx, func(identity *storedIdentity) { identity.sharePresence = share }); err != nil {
		return err
	}
	tracker := &together.presence
	tracker.mutex.Lock()
	tracker.share = share
	tracker.published = nowPlayingState{}
	tracker.mutex.Unlock()
	if share {
		together.schedulePresence()
	} else {
		go together.sendPresence(together.offlineMessage(), lobbySendTimeout)
	}
	together.notify()
	return nil
}
