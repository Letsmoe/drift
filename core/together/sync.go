package together

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/neoworks/otter-music/core/queue"
)

const (
	payloadQueue    = "queue"
	payloadPlayback = "playback"
)

const (
	// A play, pause or seek by someone: follow closely.
	userActionToleranceSeconds = 0.75
	// The host's periodic position: only correct real drift.
	heartbeatToleranceSeconds = 2.0
	// Transit time added to a playing position at most, against skewed clocks.
	maxTransitDelay = 3 * time.Second
)

// Queue is the playback queue a room keeps in step.
type Queue interface {
	State() queue.State
	Adopt(state queue.State) queue.State
}

// PlaybackReport is the local player's state, reported by the app.
type PlaybackReport struct {
	EntryID         int64   `json:"entryId"`
	Paused          bool    `json:"paused"`
	PositionSeconds float64 `json:"positionSeconds"`
	// UserAction is true for a play, pause or seek by the user, false for the
	// periodic report while playing.
	UserAction bool `json:"userAction"`
}

// Playback asks the local player to follow someone else's player.
type Playback struct {
	EntryID         int64   `json:"entryId"`
	Paused          bool    `json:"paused"`
	PositionSeconds float64 `json:"positionSeconds"`
	// ToleranceSeconds is how far the local position may be off before seeking.
	ToleranceSeconds float64 `json:"toleranceSeconds"`
	// From is who played, paused or seeked; empty for heartbeats.
	From string `json:"from"`
}

type syncPayload struct {
	Type     string           `json:"type"`
	Version  uint64           `json:"version,omitempty"`
	Origin   string           `json:"origin,omitempty"`
	Queue    *queue.State     `json:"queue,omitempty"`
	Playback *playbackPayload `json:"playback,omitempty"`
}

type playbackPayload struct {
	EntryID         int64   `json:"entryId"`
	Paused          bool    `json:"paused"`
	PositionSeconds float64 `json:"positionSeconds"`
	SentAtMillis    int64   `json:"sentAt"`
	Heartbeat       bool    `json:"heartbeat,omitempty"`
}

// syncer is the room's view of the shared queue and player.
type syncer struct {
	queue      Queue
	send       func(payload string)
	onPlayback func(Playback)
	now        func() time.Time

	mutex sync.Mutex
	// selfPeerID changes with every room; empty without one.
	selfPeerID string
	hosting    bool
	connected  bool
	// A guest shares nothing before it adopted the host's queue.
	synced bool
	// Lamport version of the shared queue and the peer that set it.
	version     uint64
	origin      string
	fingerprint string
	// Peers the host already brought up to date.
	knownPeers   map[string]bool
	lastPlayback PlaybackReport
	reportedAt   time.Time
}

// roomChanged follows the room: a new room starts over, and the host brings
// new guests up to date.
func (syncer *syncer) roomChanged(snapshot *roomSnapshot) {
	syncer.mutex.Lock()
	if snapshot == nil {
		syncer.resetLocked("", false)
		syncer.mutex.Unlock()
		return
	}
	if snapshot.Self.ID != syncer.selfPeerID {
		syncer.resetLocked(snapshot.Self.ID, snapshot.Hosting)
	}
	syncer.connected = snapshot.Status == StatusConnected
	bringUpToDate := syncer.newcomersLocked(snapshot.Peers) && syncer.hosting
	syncer.mutex.Unlock()
	if bringUpToDate {
		syncer.sendSnapshot()
	}
}

func (syncer *syncer) resetLocked(selfPeerID string, hosting bool) {
	syncer.selfPeerID = selfPeerID
	syncer.hosting = hosting
	syncer.connected = false
	syncer.synced = hosting
	syncer.version = 0
	syncer.origin = ""
	syncer.fingerprint = ""
	syncer.knownPeers = map[string]bool{selfPeerID: true}
}

// newcomersLocked records peers and reports whether one of them is new.
func (syncer *syncer) newcomersLocked(peers []Peer) bool {
	newcomer := false
	known := map[string]bool{syncer.selfPeerID: true}
	for _, peer := range peers {
		known[peer.ID] = true
		if !syncer.knownPeers[peer.ID] {
			newcomer = true
		}
	}
	syncer.knownPeers = known
	return newcomer
}

// sendSnapshot sends the host's queue and player to everyone; whoever already
// has this queue version ignores it.
func (syncer *syncer) sendSnapshot() {
	state := syncer.queue.State()
	syncer.mutex.Lock()
	if syncer.origin == "" {
		syncer.origin = syncer.selfPeerID
	}
	syncer.fingerprint = fingerprintOf(state)
	queuePayload := syncPayload{Type: payloadQueue, Version: syncer.version, Origin: syncer.origin, Queue: &state}
	playback, hasPlayback := syncer.currentPlaybackLocked(state)
	syncer.mutex.Unlock()

	syncer.sendPayload(queuePayload)
	if hasPlayback {
		syncer.sendPayload(syncPayload{Type: payloadPlayback, Playback: &playback})
	}
}

// currentPlaybackLocked extrapolates the last report to now, for the current entry only.
func (syncer *syncer) currentPlaybackLocked(state queue.State) (playbackPayload, bool) {
	if state.Index < 0 || syncer.reportedAt.IsZero() {
		return playbackPayload{}, false
	}
	report := syncer.lastPlayback
	if report.EntryID != state.Entries[state.Index].ID {
		return playbackPayload{}, false
	}
	position := report.PositionSeconds
	if !report.Paused {
		position += syncer.now().Sub(syncer.reportedAt).Seconds()
	}
	return playbackPayload{
		EntryID:         report.EntryID,
		Paused:          report.Paused,
		PositionSeconds: position,
		SentAtMillis:    syncer.now().UnixMilli(),
	}, true
}

// queueChanged shares a local queue change, unless it is the state just adopted.
func (syncer *syncer) queueChanged(state queue.State) {
	syncer.mutex.Lock()
	if !syncer.sharingLocked() {
		syncer.mutex.Unlock()
		return
	}
	fingerprint := fingerprintOf(state)
	if fingerprint == syncer.fingerprint {
		syncer.mutex.Unlock()
		return
	}
	syncer.fingerprint = fingerprint
	syncer.version++
	syncer.origin = syncer.selfPeerID
	payload := syncPayload{Type: payloadQueue, Version: syncer.version, Origin: syncer.origin, Queue: &state}
	syncer.mutex.Unlock()
	syncer.sendPayload(payload)
}

func (syncer *syncer) sharingLocked() bool {
	return syncer.connected && syncer.synced
}

// reportPlayback remembers the local player and shares user actions; the host
// also shares its position as a heartbeat while playing.
func (syncer *syncer) reportPlayback(report PlaybackReport) {
	syncer.mutex.Lock()
	syncer.lastPlayback = report
	syncer.reportedAt = syncer.now()
	share := syncer.sharingLocked() && (report.UserAction || (syncer.hosting && !report.Paused))
	syncer.mutex.Unlock()
	if !share {
		return
	}
	syncer.sendPayload(syncPayload{Type: payloadPlayback, Playback: &playbackPayload{
		EntryID:         report.EntryID,
		Paused:          report.Paused,
		PositionSeconds: report.PositionSeconds,
		SentAtMillis:    syncer.now().UnixMilli(),
		Heartbeat:       !report.UserAction,
	}})
}

func (syncer *syncer) sendPayload(payload syncPayload) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}
	syncer.send(string(encoded))
}

// receive applies a payload from someone else in the room.
func (syncer *syncer) receive(message roomMessage) {
	var payload syncPayload
	if json.Unmarshal([]byte(message.Payload), &payload) != nil {
		return
	}
	if payload.Type == payloadQueue && payload.Queue != nil {
		syncer.receiveQueue(payload)
		return
	}
	if payload.Type == payloadPlayback && payload.Playback != nil {
		syncer.receivePlayback(message.From, *payload.Playback)
	}
}

// receiveQueue adopts a newer queue; concurrent changes resolve the same way
// everywhere because the version ties break by origin.
func (syncer *syncer) receiveQueue(payload syncPayload) {
	syncer.mutex.Lock()
	if !isNewer(payload.Version, payload.Origin, syncer.version, syncer.origin) {
		syncer.mutex.Unlock()
		return
	}
	syncer.version = payload.Version
	syncer.origin = payload.Origin
	syncer.fingerprint = fingerprintOf(*payload.Queue)
	syncer.synced = true
	syncer.mutex.Unlock()
	syncer.queue.Adopt(*payload.Queue)
}

func isNewer(version uint64, origin string, currentVersion uint64, currentOrigin string) bool {
	if version != currentVersion {
		return version > currentVersion
	}
	return origin > currentOrigin
}

func (syncer *syncer) receivePlayback(from Peer, payload playbackPayload) {
	position := payload.PositionSeconds
	if !payload.Paused {
		position += syncer.transitDelay(payload.SentAtMillis).Seconds()
	}
	playback := Playback{
		EntryID:          payload.EntryID,
		Paused:           payload.Paused,
		PositionSeconds:  position,
		ToleranceSeconds: userActionToleranceSeconds,
		From:             from.Name,
	}
	if payload.Heartbeat {
		playback.ToleranceSeconds = heartbeatToleranceSeconds
		playback.From = ""
	}
	syncer.onPlayback(playback)
}

func (syncer *syncer) transitDelay(sentAtMillis int64) time.Duration {
	delay := syncer.now().Sub(time.UnixMilli(sentAtMillis))
	return min(max(delay, 0), maxTransitDelay)
}

// following reports whether this instance is a guest, so automatic queue
// extensions come from the host only.
func (syncer *syncer) following() bool {
	syncer.mutex.Lock()
	defer syncer.mutex.Unlock()
	return syncer.selfPeerID != "" && !syncer.hosting
}

func fingerprintOf(state queue.State) string {
	encoded, err := json.Marshal(state)
	if err != nil {
		return ""
	}
	return string(encoded)
}
