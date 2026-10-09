package together

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/gorilla/websocket"
)

// DefaultRelays are public Nostr relays, the same otter uses.
var DefaultRelays = []string{
	"wss://relay.damus.io",
	"wss://nos.lol",
	"wss://relay.primal.net",
}

// Ephemeral range (20000–29999): relays forward these and never store them.
// otter uses 25050.
const signalKind = 25051

const (
	relayRetryDelay   = 3 * time.Second
	relayDialTimeout  = 10 * time.Second
	relayWriteTimeout = 5 * time.Second
	// Received event IDs remembered for deduplication across relays.
	seenLimit = 512
)

var errNoRelay = errors.New("together: no signaling relay reachable")

// Signaling carries sealed messages between instances over relay topics.
type Signaling interface {
	// Open subscribes to topics; the channel closes when ctx ends.
	Open(ctx context.Context, topics []string) (SignalChannel, error)
}

// SignalChannel receives what is sent to its topics while it is subscribed,
// and sends to any topic (subscribed or not).
type SignalChannel interface {
	Send(ctx context.Context, topic, content string) error
	Receive() <-chan SignalMessage
}

// SignalMessage is one message received on a topic.
type SignalMessage struct {
	Topic   string
	Content string
}

// NostrSignaling publishes and subscribes on public Nostr relays.
type NostrSignaling struct {
	Relays []string
	Logger *slog.Logger
}

// Open connects to every relay in the background; Send fails with no relay connected yet.
func (signaling NostrSignaling) Open(ctx context.Context, topics []string) (SignalChannel, error) {
	key, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("together: signing key: %w", err)
	}
	logger := signaling.Logger
	if logger == nil {
		logger = slog.Default()
	}
	channel := &nostrChannel{
		topics:   topics,
		key:      key,
		logger:   logger,
		received: make(chan SignalMessage, 64),
		seen:     map[string]bool{},
	}
	for _, url := range signaling.Relays {
		relay := &relayConnection{url: url, channel: channel}
		channel.relays = append(channel.relays, relay)
		go relay.run(ctx)
	}
	return channel, nil
}

type nostrChannel struct {
	topics   []string
	key      *btcec.PrivateKey
	logger   *slog.Logger
	relays   []*relayConnection
	received chan SignalMessage

	seenMutex sync.Mutex
	seen      map[string]bool
	seenOrder []string
}

func (channel *nostrChannel) Receive() <-chan SignalMessage {
	return channel.received
}

func (channel *nostrChannel) Send(_ context.Context, topic, content string) error {
	event := nostrEvent{
		CreatedAt: time.Now().Unix(),
		Kind:      signalKind,
		Tags:      [][]string{{"t", topic}},
		Content:   content,
	}
	if err := event.sign(channel.key); err != nil {
		return err
	}
	frame, err := marshalNostr([]any{"EVENT", event})
	if err != nil {
		return err
	}
	sent := 0
	for _, relay := range channel.relays {
		if relay.write(frame) == nil {
			sent++
		}
	}
	if sent == 0 {
		return errNoRelay
	}
	return nil
}

// deliver passes on an event's content once, however many relays echo it.
func (channel *nostrChannel) deliver(event nostrEvent) {
	topic, ok := event.matchingTopic(channel.topics)
	if !ok || channel.markSeen(event.ID) {
		return
	}
	select {
	case channel.received <- SignalMessage{Topic: topic, Content: event.Content}:
	default:
		channel.logger.Warn("together: signaling buffer full, dropped a message")
	}
}

// markSeen records an event ID and reports whether it was already there.
func (channel *nostrChannel) markSeen(eventID string) bool {
	channel.seenMutex.Lock()
	defer channel.seenMutex.Unlock()
	if channel.seen[eventID] {
		return true
	}
	channel.seen[eventID] = true
	channel.seenOrder = append(channel.seenOrder, eventID)
	if len(channel.seenOrder) > seenLimit {
		delete(channel.seen, channel.seenOrder[0])
		channel.seenOrder = channel.seenOrder[1:]
	}
	return false
}

func (channel *nostrChannel) subscribeFrame() ([]byte, error) {
	filter := map[string]any{
		"kinds": []int{signalKind},
		"#t":    channel.topics,
		"since": time.Now().Add(-time.Minute).Unix(),
	}
	return marshalNostr([]any{"REQ", "otter-music-" + subscriptionID(channel.topics), filter})
}

// relayConnection keeps one relay subscribed until the channel's context ends.
type relayConnection struct {
	url     string
	channel *nostrChannel

	mutex      sync.Mutex
	connection *websocket.Conn
}

func (relay *relayConnection) run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := relay.session(ctx); err != nil && ctx.Err() == nil {
			relay.channel.logger.Debug("together: relay", "url", relay.url, "error", err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(relayRetryDelay):
		}
	}
}

// session connects, subscribes and reads until the connection drops.
func (relay *relayConnection) session(ctx context.Context) error {
	dialContext, cancel := context.WithTimeout(ctx, relayDialTimeout)
	defer cancel()
	connection, response, err := websocket.DefaultDialer.DialContext(dialContext, relay.url, nil)
	if response != nil && response.Body != nil {
		_ = response.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("together: dial %s: %w", relay.url, err)
	}
	// Unblocks ReadMessage when the channel closes.
	stop := context.AfterFunc(ctx, func() { _ = connection.Close() })
	defer stop()

	subscribe, err := relay.channel.subscribeFrame()
	if err != nil {
		_ = connection.Close()
		return err
	}
	if err := writeFrame(connection, subscribe); err != nil {
		_ = connection.Close()
		return err
	}
	// Publish only once subscribed, so a sender also hears answers to what it sent.
	relay.setConnection(connection)
	defer relay.setConnection(nil)
	return relay.readLoop(connection)
}

func (relay *relayConnection) readLoop(connection *websocket.Conn) error {
	for {
		_, frame, err := connection.ReadMessage()
		if err != nil {
			return fmt.Errorf("together: read from %s: %w", relay.url, err)
		}
		event, ok := parseEventFrame(frame)
		if ok {
			relay.channel.deliver(event)
		}
	}
}

func (relay *relayConnection) setConnection(connection *websocket.Conn) {
	relay.mutex.Lock()
	defer relay.mutex.Unlock()
	if connection == nil && relay.connection != nil {
		_ = relay.connection.Close()
	}
	relay.connection = connection
}

func (relay *relayConnection) write(frame []byte) error {
	relay.mutex.Lock()
	defer relay.mutex.Unlock()
	if relay.connection == nil {
		return errNoRelay
	}
	return writeFrame(relay.connection, frame)
}

func writeFrame(connection *websocket.Conn, frame []byte) error {
	if err := connection.SetWriteDeadline(time.Now().Add(relayWriteTimeout)); err != nil {
		return fmt.Errorf("together: write deadline: %w", err)
	}
	if err := connection.WriteMessage(websocket.TextMessage, frame); err != nil {
		return fmt.Errorf("together: write to relay: %w", err)
	}
	return nil
}

// parseEventFrame reads ["EVENT", subscriptionId, event]; other frames
// (EOSE, OK, NOTICE) are ignored.
func parseEventFrame(frame []byte) (nostrEvent, bool) {
	var parts []json.RawMessage
	if json.Unmarshal(frame, &parts) != nil || len(parts) < 3 {
		return nostrEvent{}, false
	}
	var label string
	if json.Unmarshal(parts[0], &label) != nil || label != "EVENT" {
		return nostrEvent{}, false
	}
	var event nostrEvent
	if json.Unmarshal(parts[2], &event) != nil {
		return nostrEvent{}, false
	}
	return event, true
}

// nostrEvent follows NIP-01. Signatures of received events are not checked:
// the AES-GCM seal already proves a message comes from someone with the ID.
type nostrEvent struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

func (event *nostrEvent) sign(key *btcec.PrivateKey) error {
	event.PubKey = hex.EncodeToString(schnorr.SerializePubKey(key.PubKey()))
	serialized, err := marshalNostr([]any{0, event.PubKey, event.CreatedAt, event.Kind, event.Tags, event.Content})
	if err != nil {
		return err
	}
	hash := sha256.Sum256(serialized)
	signature, err := schnorr.Sign(key, hash[:])
	if err != nil {
		return fmt.Errorf("together: sign event: %w", err)
	}
	event.ID = hex.EncodeToString(hash[:])
	event.Sig = hex.EncodeToString(signature.Serialize())
	return nil
}

// matchingTopic returns the first of topics the event is tagged with.
func (event nostrEvent) matchingTopic(topics []string) (string, bool) {
	for _, tag := range event.Tags {
		if len(tag) >= 2 && tag[0] == "t" && slices.Contains(topics, tag[1]) {
			return tag[1], true
		}
	}
	return "", false
}

// subscriptionID names a subscription after its topics, short as relays like it.
func subscriptionID(topics []string) string {
	hash := sha256.Sum256([]byte(strings.Join(topics, ",")))
	return hex.EncodeToString(hash[:4])
}

// marshalNostr encodes without HTML escaping, as NIP-01 hashing requires.
func marshalNostr(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("together: encode event: %w", err)
	}
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}

// sendWhenReady retries Send until a relay is connected, for messages sent
// right after Open.
func sendWhenReady(ctx context.Context, channel SignalChannel, topic, content string) error {
	ticker := time.NewTicker(sendRetryInterval)
	defer ticker.Stop()
	for {
		err := channel.Send(ctx, topic, content)
		if !errors.Is(err, errNoRelay) {
			return err
		}
		select {
		case <-ctx.Done():
			return errNoRelay
		case <-ticker.C:
		}
	}
}

const sendRetryInterval = 300 * time.Millisecond
