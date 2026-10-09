package together

// Peer is one participant of a room.
type Peer struct {
	// ID is new for every room, so the same person on two devices stays apart.
	ID string `json:"id"`
	// UserID is the person's fixed ID, the one friends know them by.
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

// Handshake messages, signed by their sender, sealed and sent over the relays.
const (
	signalJoin  = "join"
	signalOffer = "offer"
	// Host → guest: the host was asked whether to let the guest in.
	signalWaiting = "waiting"
	// Host → guest: the host turned the guest away.
	signalDeclined = "declined"
	signalAnswer   = "answer"
)

type signal struct {
	Type string `json:"type"`
	From Peer   `json:"from"`
	// Receiving peer ID for everything but joins.
	To  string `json:"to,omitempty"`
	SDP string `json:"sdp,omitempty"`
	// SentAtMillis lets receivers drop replayed old messages.
	SentAtMillis int64 `json:"sentAt"`
}

// Messages on the WebRTC data channel.
const (
	// Host → guests: everyone in the room.
	wireRoster = "roster"
	// App payload; the host forwards it to the other guests.
	wireApp = "app"
	// Sent before leaving so others don't wait for the connection to time out.
	wireBye = "bye"
)

type wireMessage struct {
	Type    string `json:"type"`
	From    *Peer  `json:"from,omitempty"`
	Peers   []Peer `json:"peers,omitempty"`
	Payload string `json:"payload,omitempty"`
}
