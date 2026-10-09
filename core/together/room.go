package together

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

// Status is how far a room got.
type Status string

// Room states.
const (
	// StatusWaiting is a host alone in the room, waiting for someone to join.
	StatusWaiting Status = "waiting"
	// StatusConnecting is a guest looking for the host.
	StatusConnecting Status = "connecting"
	// StatusAwaiting is a guest the host has to let in first.
	StatusAwaiting  Status = "awaiting"
	StatusConnected Status = "connected"
	// StatusClosed means the host left; the guest's room is over.
	StatusClosed Status = "closed"
	// StatusFailed means the guest never reached the host.
	StatusFailed Status = "failed"
	// StatusDeclined means the host did not let the guest in.
	StatusDeclined Status = "declined"
)

const (
	// A guest repeats its join until the host answers; the relays forward
	// ephemeral events only to whoever is subscribed at that moment.
	joinInterval = 2 * time.Second
	joinTimeout  = 60 * time.Second
	// How long a guest waits for the host to let it in.
	approvalTimeout = 3 * time.Minute
	// A join request the guest stopped repeating is dropped after this.
	requestTimeout = 3 * joinInterval
	// More pending requests than this are ignored, against join floods.
	maxRequests = 8
	// Signed handshake messages older than this are replays and dropped.
	signalMaxAge = 2 * time.Minute
	// Without a STUN answer, ICE gathering still ends after this.
	gatherTimeout = 8 * time.Second
	byeFlushDelay = 200 * time.Millisecond
)

var errNotConnected = errors.New("together: not connected to anyone")

// roomSnapshot is a room's state at one moment.
type roomSnapshot struct {
	HostUserID string
	Hosting    bool
	Status     Status
	Self       Peer
	Peers      []Peer
	// Requests are guests waiting for the host to let them in.
	Requests []Peer
}

// roomMessage is an app payload from another participant.
type roomMessage struct {
	From    Peer
	Payload string
}

// link is the WebRTC connection to one other participant: every guest for
// the host, the host for a guest.
type link struct {
	peer       Peer
	connection *webrtc.PeerConnection
	// Set once the data channel is open.
	channel *webrtc.DataChannel
	// Our offer or answer, sent again when the other side repeats itself.
	localSDP  string
	remoteSDP string
}

// room is a star: guests connect to the host, who forwards between them.
type room struct {
	hostUserID string
	hosting    bool
	self       Peer
	signer     *signer
	key        *sealKey
	signals    SignalChannel
	api        *webrtc.API
	config     webrtc.Configuration
	logger     *slog.Logger
	ctx        context.Context
	cancel     context.CancelFunc

	// admit tells the host whether a guest may join without being asked for.
	admit     func(Peer) bool
	onChange  func()
	onMessage func(roomMessage)
	now       func() time.Time

	mutex  sync.Mutex
	status Status
	links  map[string]*link
	roster []Peer
	// Host: guests waiting to be let in, and guests turned away, by peer ID.
	requests map[string]*joinRequest
	declined map[string]bool
	// Guest: when to give up joining.
	deadline time.Time
}

// joinRequest is a guest waiting for the host's answer.
type joinRequest struct {
	peer     Peer
	lastSeen time.Time
}

func (room *room) start() {
	go room.listen()
	if room.hosting {
		go room.pruneRequests()
		return
	}
	go room.announce()
}

func (room *room) snapshot() roomSnapshot {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	return roomSnapshot{
		HostUserID: room.hostUserID,
		Hosting:    room.hosting,
		Status:     room.status,
		Self:       room.self,
		Peers:      slices.Clone(room.roster),
		Requests:   room.requestPeersLocked(),
	}
}

func (room *room) close() {
	if room.forward(wireMessage{Type: wireBye}, "") > 0 {
		// Let the bye leave before the connections close.
		time.Sleep(byeFlushDelay)
	}
	room.mutex.Lock()
	connections := make([]*webrtc.PeerConnection, 0, len(room.links))
	for _, peerLink := range room.links {
		if peerLink.connection != nil {
			connections = append(connections, peerLink.connection)
		}
	}
	room.links = map[string]*link{}
	room.mutex.Unlock()
	for _, connection := range connections {
		_ = connection.Close()
	}
	room.cancel()
}

// WebRTC ---------------------------------------------------------------------

func (room *room) setConnection(peerLink *link, connection *webrtc.PeerConnection) {
	room.mutex.Lock()
	peerLink.connection = connection
	room.mutex.Unlock()
	room.watchConnection(peerLink)
}

func (room *room) setLocalSDP(peerLink *link, localSDP string) {
	room.mutex.Lock()
	peerLink.localSDP = localSDP
	room.mutex.Unlock()
}

func (room *room) setRemoteAnswer(peerLink *link, sdp string) {
	answer := webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: sdp}
	if err := peerLink.connection.SetRemoteDescription(answer); err != nil {
		room.logger.Warn("together: answer", "error", err)
	}
}

// answerOffer accepts the host's offer and returns the gathered answer.
func (room *room) answerOffer(peerLink *link, connection *webrtc.PeerConnection) (string, error) {
	connection.OnDataChannel(func(dataChannel *webrtc.DataChannel) {
		room.attachChannel(peerLink, dataChannel)
	})
	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: peerLink.remoteSDP}
	if err := connection.SetRemoteDescription(offer); err != nil {
		return "", fmt.Errorf("together: remote offer: %w", err)
	}
	description, err := connection.CreateAnswer(nil)
	if err != nil {
		return "", fmt.Errorf("together: create answer: %w", err)
	}
	return room.setLocalAndGather(connection, description)
}

// setLocalAndGather waits for every ICE candidate so the description can be
// sent in one signaling message (no trickle ICE).
func (room *room) setLocalAndGather(connection *webrtc.PeerConnection, description webrtc.SessionDescription) (string, error) {
	gathered := webrtc.GatheringCompletePromise(connection)
	if err := connection.SetLocalDescription(description); err != nil {
		return "", fmt.Errorf("together: local description: %w", err)
	}
	select {
	case <-gathered:
	case <-time.After(gatherTimeout):
	case <-room.ctx.Done():
		return "", fmt.Errorf("together: room closed: %w", room.ctx.Err())
	}
	return connection.LocalDescription().SDP, nil
}

func (room *room) watchConnection(peerLink *link) {
	peerLink.connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			room.dropLink(peerLink)
		}
	})
}

func (room *room) attachChannel(peerLink *link, dataChannel *webrtc.DataChannel) {
	dataChannel.OnOpen(func() {
		room.linkOpened(peerLink, dataChannel)
	})
	dataChannel.OnMessage(func(message webrtc.DataChannelMessage) {
		room.receive(peerLink, message.Data)
	})
}

func (room *room) linkOpened(peerLink *link, dataChannel *webrtc.DataChannel) {
	room.mutex.Lock()
	peerLink.channel = dataChannel
	room.status = StatusConnected
	if room.hosting {
		room.roster = append(room.roster, peerLink.peer)
	}
	room.mutex.Unlock()
	if room.hosting {
		room.broadcastRoster()
	}
	room.onChange()
}

func (room *room) dropLink(peerLink *link) {
	room.mutex.Lock()
	if room.links[peerLink.peer.ID] != peerLink {
		room.mutex.Unlock()
		return
	}
	delete(room.links, peerLink.peer.ID)
	room.roster = slices.DeleteFunc(room.roster, func(peer Peer) bool { return peer.ID == peerLink.peer.ID })
	room.status = statusAfterDrop(room.hosting, len(room.links))
	connection := peerLink.connection
	room.mutex.Unlock()
	if connection != nil {
		_ = connection.Close()
	}
	if room.hosting {
		room.broadcastRoster()
	}
	room.onChange()
}

func statusAfterDrop(hosting bool, remainingLinks int) Status {
	if !hosting {
		return StatusClosed
	}
	if remainingLinks == 0 {
		return StatusWaiting
	}
	return StatusConnected
}

// Data channel ---------------------------------------------------------------

func (room *room) receive(peerLink *link, data []byte) {
	var message wireMessage
	if json.Unmarshal(data, &message) != nil {
		return
	}
	switch message.Type {
	case wireBye:
		room.dropLink(peerLink)
	case wireRoster:
		room.receiveRoster(message.Peers)
	case wireApp:
		room.receiveApp(peerLink, message)
	}
}

func (room *room) receiveRoster(peers []Peer) {
	if room.hosting {
		return
	}
	room.mutex.Lock()
	room.roster = peers
	room.mutex.Unlock()
	room.onChange()
}

func (room *room) receiveApp(peerLink *link, message wireMessage) {
	// The host vouches for senders: a guest can't speak for someone else.
	sender := peerLink.peer
	if !room.hosting && message.From != nil {
		sender = *message.From
	}
	if room.hosting {
		room.forward(wireMessage{Type: wireApp, From: &sender, Payload: message.Payload}, sender.ID)
	}
	room.onMessage(roomMessage{From: sender, Payload: message.Payload})
}

// send passes an app payload to everyone else in the room.
func (room *room) send(payload string) error {
	message := wireMessage{Type: wireApp, From: &room.self, Payload: payload}
	if room.forward(message, "") == 0 {
		return errNotConnected
	}
	return nil
}

func (room *room) broadcastRoster() {
	room.mutex.Lock()
	peers := slices.Clone(room.roster)
	room.mutex.Unlock()
	room.forward(wireMessage{Type: wireRoster, Peers: peers}, "")
}

// forward sends to every open link except the one with exceptPeerID and
// returns how many it reached.
func (room *room) forward(message wireMessage, exceptPeerID string) int {
	data, err := json.Marshal(message)
	if err != nil {
		return 0
	}
	reached := 0
	for _, dataChannel := range room.openChannels(exceptPeerID) {
		if dataChannel.SendText(string(data)) == nil {
			reached++
		}
	}
	return reached
}

func (room *room) openChannels(exceptPeerID string) []*webrtc.DataChannel {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	channels := make([]*webrtc.DataChannel, 0, len(room.links))
	for peerID, peerLink := range room.links {
		if peerID != exceptPeerID && peerLink.channel != nil {
			channels = append(channels, peerLink.channel)
		}
	}
	return channels
}
