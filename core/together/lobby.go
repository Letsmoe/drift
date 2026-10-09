package together

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"
)

const (
	inviteMessageType   = "invite"
	presenceMessageType = "presence"
	// Relays replay the last minute to new subscribers; older messages are stale.
	lobbyMessageMaxAge = 2 * time.Minute
	// How long an invite or presence waits for a relay before giving up.
	lobbySendTimeout = 15 * time.Second
	// An invite goes out a second time for relays that connected late.
	inviteRepeatDelay = 3 * time.Second
)

type inviteMessage struct {
	Type         string   `json:"type"`
	From         Identity `json:"from"`
	SentAtMillis int64    `json:"sentAt"`
}

// lobbySettings is what a lobby is opened with.
type lobbySettings struct {
	signaling  Signaling
	signer     *signer
	logger     *slog.Logger
	now        func() time.Time
	onInvite   func(Identity)
	onPresence func(userID string, message presenceMessage)
}

// lobby receives invites for the user and presence of the user's friends.
type lobby struct {
	lobbySettings
	inboxKey    *sealKey
	presenceKey *sealKey

	mutex sync.Mutex
	// Derived keys, cached: deriving one costs about 100 ms.
	friendPresenceKeys map[string]*sealKey
	friendInboxKeys    map[string]*sealKey
	// topicOwners maps a friend's presence topic to their user ID.
	topicOwners map[string]string
	channel     SignalChannel
	cancel      context.CancelFunc
	closed      bool
}

func openLobby(settings lobbySettings, friendUserIDs []string) (*lobby, error) {
	opened := &lobby{lobbySettings: settings}
	var err error
	if opened.inboxKey, err = deriveInboxKey(opened.signer.userID); err != nil {
		return nil, err
	}
	if opened.presenceKey, err = derivePresenceKey(opened.signer.userID); err != nil {
		return nil, err
	}
	opened.friendPresenceKeys = map[string]*sealKey{}
	opened.friendInboxKeys = map[string]*sealKey{}
	opened.setFriends(friendUserIDs)
	return opened, nil
}

// setFriends subscribes to the presence of exactly these friends.
func (lobby *lobby) setFriends(friendUserIDs []string) {
	keys := lobby.presenceKeysFor(friendUserIDs)
	topicOwners := map[string]string{}
	for userID, key := range keys {
		topicOwners[key.topic] = userID
	}
	lobby.mutex.Lock()
	unchanged := lobby.channel != nil && maps.Equal(topicOwners, lobby.topicOwners)
	lobby.mutex.Unlock()
	if unchanged {
		return
	}
	lobby.resubscribe(topicOwners)
}

func (lobby *lobby) presenceKeysFor(friendUserIDs []string) map[string]*sealKey {
	keys := map[string]*sealKey{}
	for _, userID := range friendUserIDs {
		key, err := lobby.cachedKey(lobby.friendPresenceKeys, userID, derivePresenceKey)
		if err != nil {
			lobby.logger.Warn("together: presence key", "error", err)
			continue
		}
		keys[userID] = key
	}
	return keys
}

func (lobby *lobby) cachedKey(cache map[string]*sealKey, userID string, derive func(string) (*sealKey, error)) (*sealKey, error) {
	lobby.mutex.Lock()
	key := cache[userID]
	lobby.mutex.Unlock()
	if key != nil {
		return key, nil
	}
	key, err := derive(userID)
	if err != nil {
		return nil, err
	}
	lobby.mutex.Lock()
	cache[userID] = key
	lobby.mutex.Unlock()
	return key, nil
}

// resubscribe opens a channel for the new topics, then closes the old one.
func (lobby *lobby) resubscribe(topicOwners map[string]string) {
	topics := append([]string{lobby.inboxKey.topic}, slices.Sorted(maps.Keys(topicOwners))...)
	ctx, cancel := context.WithCancel(context.Background())
	channel, err := lobby.signaling.Open(ctx, topics)
	if err != nil {
		cancel()
		lobby.logger.Warn("together: lobby", "error", err)
		return
	}
	lobby.mutex.Lock()
	if lobby.closed {
		lobby.mutex.Unlock()
		cancel()
		return
	}
	previousCancel := lobby.cancel
	lobby.channel = channel
	lobby.cancel = cancel
	lobby.topicOwners = topicOwners
	lobby.mutex.Unlock()
	if previousCancel != nil {
		previousCancel()
	}
	go lobby.listen(ctx, channel)
}

func (lobby *lobby) close() {
	lobby.mutex.Lock()
	lobby.closed = true
	cancel := lobby.cancel
	lobby.mutex.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (lobby *lobby) listen(ctx context.Context, channel SignalChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case received := <-channel.Receive():
			lobby.handle(received)
		}
	}
}

func (lobby *lobby) handle(received SignalMessage) {
	if received.Topic == lobby.inboxKey.topic {
		lobby.handleInvite(received.Content)
		return
	}
	lobby.mutex.Lock()
	owner := lobby.topicOwners[received.Topic]
	key := lobby.friendPresenceKeys[owner]
	lobby.mutex.Unlock()
	if key != nil {
		lobby.handlePresence(owner, key, received)
	}
}

func (lobby *lobby) handleInvite(content string) {
	var message inviteMessage
	signerUserID, err := openSigned(lobby.inboxKey, lobby.inboxKey.topic, content, &message)
	if err != nil || message.Type != inviteMessageType || signerUserID != message.From.UserID {
		return
	}
	if !lobby.fresh(message.SentAtMillis) {
		return
	}
	lobby.onInvite(message.From)
}

func (lobby *lobby) handlePresence(owner string, key *sealKey, received SignalMessage) {
	var message presenceMessage
	signerUserID, err := openSigned(key, received.Topic, received.Content, &message)
	if err != nil || message.Type != presenceMessageType || signerUserID != owner {
		return
	}
	if !lobby.fresh(message.SentAtMillis) {
		return
	}
	lobby.onPresence(owner, message)
}

func (lobby *lobby) fresh(sentAtMillis int64) bool {
	return lobby.now().Sub(time.UnixMilli(sentAtMillis)) <= lobbyMessageMaxAge
}

func (lobby *lobby) currentChannel() SignalChannel {
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()
	return lobby.channel
}

// sendInvite drops an invite from sender into the inbox of recipientUserID. It
// returns once a relay took it and repeats it for relays that connect later.
func (lobby *lobby) sendInvite(recipientUserID string, sender Identity) error {
	key, err := lobby.cachedKey(lobby.friendInboxKeys, recipientUserID, deriveInboxKey)
	if err != nil {
		return err
	}
	message := inviteMessage{Type: inviteMessageType, From: sender, SentAtMillis: lobby.now().UnixMilli()}
	content, err := lobby.signer.seal(key, key.topic, message)
	if err != nil {
		return err
	}
	channel := lobby.currentChannel()
	if channel == nil {
		return errNoRelay
	}
	ctx, cancel := context.WithTimeout(context.Background(), lobbySendTimeout)
	defer cancel()
	if err := sendWhenReady(ctx, channel, key.topic, content); err != nil {
		return err
	}
	go repeatAfter(inviteRepeatDelay, channel, key.topic, content)
	return nil
}

func repeatAfter(delay time.Duration, channel SignalChannel, topic, content string) {
	time.Sleep(delay)
	_ = channel.Send(context.Background(), topic, content)
}

// publishPresence sends the user's presence to everyone who knows their ID.
func (lobby *lobby) publishPresence(ctx context.Context, message presenceMessage) error {
	content, err := lobby.signer.seal(lobby.presenceKey, lobby.presenceKey.topic, message)
	if err != nil {
		return err
	}
	channel := lobby.currentChannel()
	if channel == nil {
		return errNoRelay
	}
	return sendWhenReady(ctx, channel, lobby.presenceKey.topic, content)
}
