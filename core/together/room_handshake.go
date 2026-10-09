package together

import (
	"cmp"
	"errors"
	"slices"
	"strings"
	"time"
)

var errRequestGone = errors.New("together: that person stopped waiting to join")

// Signaling ----------------------------------------------------------------

func (room *room) listen() {
	for {
		select {
		case <-room.ctx.Done():
			return
		case received := <-room.signals.Receive():
			room.handleSignal(received.Content)
		}
	}
}

// handleSignal drops everything not signed by the peer it claims to be from,
// old replays and messages for someone else.
func (room *room) handleSignal(content string) {
	var message signal
	signerUserID, err := openSigned(room.key, room.key.topic, content, &message)
	if err != nil || signerUserID != message.From.UserID || message.From.ID == room.self.ID {
		return
	}
	if room.now().Sub(time.UnixMilli(message.SentAtMillis)) > signalMaxAge {
		return
	}
	if message.To != "" && message.To != room.self.ID {
		return
	}
	if room.hosting {
		room.dispatchHostSignal(message)
		return
	}
	if message.From.UserID == room.hostUserID {
		room.dispatchGuestSignal(message)
	}
}

func (room *room) dispatchHostSignal(message signal) {
	switch message.Type {
	case signalJoin:
		room.handleJoin(message.From)
	case signalAnswer:
		room.handleAnswer(message)
	}
}

func (room *room) dispatchGuestSignal(message signal) {
	switch message.Type {
	case signalOffer:
		room.handleOffer(message)
	case signalWaiting:
		room.awaitApproval()
	case signalDeclined:
		room.endJoining(StatusDeclined)
	}
}

func (room *room) publish(message signal) {
	message.From = room.self
	message.SentAtMillis = room.now().UnixMilli()
	content, err := room.signer.seal(room.key, room.key.topic, message)
	if err != nil {
		return
	}
	if err := room.signals.Send(room.ctx, room.key.topic, content); err != nil {
		room.logger.Debug("together: signaling", "error", err)
	}
}

// Host side -----------------------------------------------------------------

// handleJoin answers a guest's (repeated) join: trusted guests get an offer,
// everyone else waits for the host to accept or decline.
func (room *room) handleJoin(guest Peer) {
	room.mutex.Lock()
	if room.declined[guest.ID] {
		room.mutex.Unlock()
		room.publish(signal{Type: signalDeclined, To: guest.ID})
		return
	}
	if existing := room.links[guest.ID]; existing != nil {
		localSDP := existing.localSDP
		room.mutex.Unlock()
		// The guest missed our offer; send it again.
		if localSDP != "" {
			room.publish(signal{Type: signalOffer, To: guest.ID, SDP: localSDP})
		}
		return
	}
	if request := room.requests[guest.ID]; request != nil {
		request.lastSeen = room.now()
		room.mutex.Unlock()
		room.publish(signal{Type: signalWaiting, To: guest.ID})
		return
	}
	room.mutex.Unlock()

	if room.admit(guest) {
		room.startLink(guest)
		return
	}
	room.addRequest(guest)
}

func (room *room) addRequest(guest Peer) {
	room.mutex.Lock()
	if len(room.requests) >= maxRequests {
		room.mutex.Unlock()
		return
	}
	room.requests[guest.ID] = &joinRequest{peer: guest, lastSeen: room.now()}
	room.mutex.Unlock()
	room.publish(signal{Type: signalWaiting, To: guest.ID})
	room.onChange()
}

// startLink creates the link to a guest and sends the offer, once per guest.
func (room *room) startLink(guest Peer) {
	room.mutex.Lock()
	if room.links[guest.ID] != nil {
		room.mutex.Unlock()
		return
	}
	created := &link{peer: guest}
	room.links[guest.ID] = created
	room.mutex.Unlock()
	go room.offer(created)
}

// accept lets a waiting guest in.
func (room *room) accept(peerID string) error {
	request := room.takeRequest(peerID)
	if request == nil {
		return errRequestGone
	}
	room.startLink(request.peer)
	room.onChange()
	return nil
}

// decline turns a waiting guest away for the rest of this room.
func (room *room) decline(peerID string) {
	room.takeRequest(peerID)
	room.mutex.Lock()
	room.declined[peerID] = true
	room.mutex.Unlock()
	room.publish(signal{Type: signalDeclined, To: peerID})
	room.onChange()
}

func (room *room) takeRequest(peerID string) *joinRequest {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	request := room.requests[peerID]
	delete(room.requests, peerID)
	return request
}

func (room *room) requestPeersLocked() []Peer {
	peers := make([]Peer, 0, len(room.requests))
	for _, request := range room.requests {
		peers = append(peers, request.peer)
	}
	slices.SortFunc(peers, func(left, right Peer) int {
		return cmp.Or(strings.Compare(left.Name, right.Name), strings.Compare(left.ID, right.ID))
	})
	return peers
}

// pruneRequests drops requests of guests that gave up (stopped repeating their join).
func (room *room) pruneRequests() {
	ticker := time.NewTicker(joinInterval)
	defer ticker.Stop()
	for {
		select {
		case <-room.ctx.Done():
			return
		case <-ticker.C:
			if room.dropStaleRequests() {
				room.onChange()
			}
		}
	}
}

func (room *room) dropStaleRequests() bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	cutoff := room.now().Add(-requestTimeout)
	dropped := false
	for peerID, request := range room.requests {
		if request.lastSeen.Before(cutoff) {
			delete(room.requests, peerID)
			dropped = true
		}
	}
	return dropped
}

func (room *room) offer(peerLink *link) {
	connection, err := room.api.NewPeerConnection(room.config)
	if err != nil {
		room.logger.Warn("together: peer connection", "error", err)
		return
	}
	room.setConnection(peerLink, connection)
	dataChannel, err := connection.CreateDataChannel("together", nil)
	if err != nil {
		room.logger.Warn("together: data channel", "error", err)
		return
	}
	room.attachChannel(peerLink, dataChannel)
	description, err := connection.CreateOffer(nil)
	if err != nil {
		room.logger.Warn("together: offer", "error", err)
		return
	}
	localSDP, err := room.setLocalAndGather(connection, description)
	if err != nil {
		room.logger.Warn("together: offer", "error", err)
		return
	}
	room.setLocalSDP(peerLink, localSDP)
	room.publish(signal{Type: signalOffer, To: peerLink.peer.ID, SDP: localSDP})
}

// handleAnswer takes the first answer of the guest the offer went to; the
// signature check already made sure it is that guest's key.
func (room *room) handleAnswer(message signal) {
	room.mutex.Lock()
	peerLink := room.links[message.From.ID]
	accepted := peerLink != nil && peerLink.connection != nil && peerLink.remoteSDP == "" &&
		peerLink.peer.UserID == message.From.UserID
	if accepted {
		peerLink.remoteSDP = message.SDP
	}
	room.mutex.Unlock()
	if !accepted {
		return
	}
	room.setRemoteAnswer(peerLink, message.SDP)
}

// Guest side ----------------------------------------------------------------

// announce repeats the guest's join until the host's offer arrives, the host
// declines, or the deadline passes.
func (room *room) announce() {
	ticker := time.NewTicker(joinInterval)
	defer ticker.Stop()
	for room.joining() {
		room.publish(signal{Type: signalJoin})
		select {
		case <-room.ctx.Done():
			return
		case <-ticker.C:
		}
		if room.deadlinePassed() {
			room.endJoining(StatusFailed)
			return
		}
	}
}

func (room *room) joining() bool {
	status := room.currentStatus()
	return status == StatusConnecting || status == StatusAwaiting
}

func (room *room) currentStatus() Status {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	return room.status
}

func (room *room) deadlinePassed() bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	return room.now().After(room.deadline)
}

// awaitApproval extends the guest's patience while the host decides.
func (room *room) awaitApproval() {
	room.mutex.Lock()
	changed := room.status == StatusConnecting
	if room.status == StatusConnecting || room.status == StatusAwaiting {
		room.status = StatusAwaiting
		room.deadline = room.now().Add(approvalTimeout)
	}
	room.mutex.Unlock()
	if changed {
		room.onChange()
	}
}

// endJoining stops a guest that is still trying to join.
func (room *room) endJoining(status Status) {
	room.mutex.Lock()
	ended := room.status == StatusConnecting || room.status == StatusAwaiting
	if ended {
		room.status = status
	}
	room.mutex.Unlock()
	if ended {
		room.onChange()
	}
}

func (room *room) handleOffer(message signal) {
	room.mutex.Lock()
	existing := room.links[message.From.ID]
	if existing != nil {
		repeat := existing.remoteSDP == message.SDP && existing.localSDP != ""
		localSDP := existing.localSDP
		room.mutex.Unlock()
		// Same offer again: our answer was likely lost.
		if repeat {
			room.publish(signal{Type: signalAnswer, To: existing.peer.ID, SDP: localSDP})
		}
		return
	}
	// A guest connects to one host only, and only while joining.
	joining := room.status == StatusConnecting || room.status == StatusAwaiting
	if len(room.links) > 0 || !joining {
		room.mutex.Unlock()
		return
	}
	created := &link{peer: message.From, remoteSDP: message.SDP}
	room.links[message.From.ID] = created
	room.mutex.Unlock()
	go room.answer(created)
}

func (room *room) answer(peerLink *link) {
	connection, err := room.api.NewPeerConnection(room.config)
	if err != nil {
		room.logger.Warn("together: peer connection", "error", err)
		return
	}
	room.setConnection(peerLink, connection)
	localSDP, err := room.answerOffer(peerLink, connection)
	if err != nil {
		room.logger.Warn("together: answer", "error", err)
		return
	}
	room.setLocalSDP(peerLink, localSDP)
	room.publish(signal{Type: signalAnswer, To: peerLink.peer.ID, SDP: localSDP})
}
