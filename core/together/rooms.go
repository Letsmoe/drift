package together

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// DefaultICEServers are public STUN servers for NAT traversal. Peers behind
// symmetric NAT would need TURN, which no public service offers for free.
var DefaultICEServers = []string{
	"stun:stun.l.google.com:19302",
	"stun:stun.cloudflare.com:3478",
}

// rooms holds the one open room; a person listens in one room at a time.
type rooms struct {
	signaling Signaling
	api       *webrtc.API
	config    webrtc.Configuration
	logger    *slog.Logger
	now       func() time.Time

	// Called without locks held, from any goroutine.
	onChange  func(*roomSnapshot)
	onMessage func(roomMessage)

	mutex   sync.Mutex
	current *room
}

func newRooms(signaling Signaling, iceServers []string, settings webrtc.SettingEngine, logger *slog.Logger, now func() time.Time) *rooms {
	config := webrtc.Configuration{}
	if len(iceServers) > 0 {
		config.ICEServers = []webrtc.ICEServer{{URLs: iceServers}}
	}
	return &rooms{
		signaling: signaling,
		api:       webrtc.NewAPI(webrtc.WithSettingEngine(settings)),
		config:    config,
		logger:    logger,
		now:       now,
		onChange:  func(*roomSnapshot) {},
		onMessage: func(roomMessage) {},
	}
}

// open replaces the current room with the room of hostUserID; self hosts it
// when it is their own. admit decides which guests join without asking.
func (rooms *rooms) open(hostUserID string, self Peer, identity *signer, admit func(Peer) bool) error {
	key, err := deriveRoomKey(hostUserID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	signals, err := rooms.signaling.Open(ctx, []string{key.topic})
	if err != nil {
		cancel()
		return err
	}
	self.ID = uuid.NewString()
	hosting := hostUserID == self.UserID
	opened := &room{
		hostUserID: hostUserID,
		hosting:    hosting,
		self:       self,
		signer:     identity,
		key:        key,
		signals:    signals,
		api:        rooms.api,
		config:     rooms.config,
		logger:     rooms.logger,
		ctx:        ctx,
		cancel:     cancel,
		admit:      admit,
		now:        rooms.now,
		status:     initialStatus(hosting),
		links:      map[string]*link{},
		roster:     []Peer{self},
		requests:   map[string]*joinRequest{},
		declined:   map[string]bool{},
		deadline:   rooms.now().Add(joinTimeout),
	}
	opened.onChange = func() { rooms.roomChanged(opened) }
	opened.onMessage = func(message roomMessage) { rooms.roomMessage(opened, message) }
	rooms.replace(opened)
	opened.start()
	return nil
}

func initialStatus(hosting bool) Status {
	if hosting {
		return StatusWaiting
	}
	return StatusConnecting
}

// leave closes the current room, if any.
func (rooms *rooms) leave() {
	rooms.replace(nil)
}

func (rooms *rooms) replace(next *room) {
	rooms.mutex.Lock()
	previous := rooms.current
	rooms.current = next
	rooms.mutex.Unlock()
	if previous != nil {
		go previous.close()
	}
	rooms.onChange(snapshotOf(next))
}

// send passes an app payload to everyone else in the room.
func (rooms *rooms) send(payload string) error {
	current := rooms.currentRoom()
	if current == nil {
		return errNotConnected
	}
	return current.send(payload)
}

// accept lets a guest waiting for the host's answer in.
func (rooms *rooms) accept(peerID string) error {
	current := rooms.currentRoom()
	if current == nil || !current.hosting {
		return errRequestGone
	}
	return current.accept(peerID)
}

// decline turns a waiting guest away.
func (rooms *rooms) decline(peerID string) {
	current := rooms.currentRoom()
	if current != nil && current.hosting {
		current.decline(peerID)
	}
}

// snapshot is the room's state, or nil without one.
func (rooms *rooms) snapshot() *roomSnapshot {
	return snapshotOf(rooms.currentRoom())
}

func snapshotOf(current *room) *roomSnapshot {
	if current == nil {
		return nil
	}
	snapshot := current.snapshot()
	return &snapshot
}

func (rooms *rooms) currentRoom() *room {
	rooms.mutex.Lock()
	defer rooms.mutex.Unlock()
	return rooms.current
}

// Callbacks from a room that was replaced in the meantime are dropped.
func (rooms *rooms) roomChanged(changed *room) {
	if rooms.currentRoom() != changed {
		return
	}
	rooms.onChange(snapshotOf(changed))
}

func (rooms *rooms) roomMessage(changed *room, message roomMessage) {
	if rooms.currentRoom() != changed {
		return
	}
	rooms.onMessage(message)
}
