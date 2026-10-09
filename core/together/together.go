// Package together is Listen Together: people listen to the same queue at the
// same time, each instance streaming the audio itself.
//
// Everyone has an Ed25519 identity key; the fixed user ID friends add is
// derived from its public half, and every relay message is signed with it, so
// nobody can speak for an ID without its key. A room is keyed by its host's
// user ID. Trusted friends join it at once; anyone else waits until the host
// accepts them, on every new connection. Peers meet on public Nostr relays:
// handshake messages are AES-GCM sealed with a key derived from the host's ID
// and tagged with a topic derived from that key. After the WebRTC handshake,
// everything runs peer-to-peer over a data channel; the host is the hub and
// forwards between guests.
//
// Outside rooms, the lobby listens for invites (a topic derived from the
// user's ID) and for friends' presence (a topic derived from each friend's ID):
// while online, everyone shares the song they play and its position.
// Adapted from otter's Watch Together.
package together

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/pion/webrtc/v4"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/queue"
)

const (
	maxNameLength = 40
	defaultName   = "Drift listener"
	// Invites nobody acted on disappear after this.
	inviteLifetime = 30 * time.Minute
)

// ErrInvalidName is returned for an empty display name.
var ErrInvalidName = errors.New("the name is empty")

// ErrOwnUserID is returned when adding oneself as a friend.
var ErrOwnUserID = errors.New("that is your own ID")

// Store keeps the identity and the friends list; *library.Library is one.
type Store interface {
	TogetherIdentity(ctx context.Context) (library.TogetherIdentity, bool, error)
	SaveTogetherIdentity(ctx context.Context, identity library.TogetherIdentity) error
	TogetherFriends(ctx context.Context) ([]library.TogetherFriend, error)
	SaveTogetherFriend(ctx context.Context, friend library.TogetherFriend) error
	RemoveTogetherFriend(ctx context.Context, userID string) error
}

// Identity is a person as others see them.
type Identity struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

// Friend is someone the user listens with.
type Friend struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	// Trusted friends join the user's sessions without being asked for.
	Trusted bool      `json:"trusted"`
	AddedAt time.Time `json:"addedAt"`
	// Presence is nil while the friend is offline or not sharing.
	Presence *Presence `json:"presence"`
}

// Invite asks the user to join the room of From.
type Invite struct {
	From       Identity  `json:"from"`
	ReceivedAt time.Time `json:"receivedAt"`
}

// Room is the open room as apps show it.
type Room struct {
	HostUserID string `json:"hostUserId"`
	Hosting    bool   `json:"hosting"`
	Status     Status `json:"status"`
	SelfPeerID string `json:"selfPeerId"`
	// Peers includes the user.
	Peers []Peer `json:"peers"`
	// Requests are guests waiting for the host to let them in.
	Requests []Peer `json:"requests"`
}

// State is everything apps render.
type State struct {
	Self Identity `json:"self"`
	// SharePresence says whether friends see what the user plays.
	SharePresence bool `json:"sharePresence"`
	// Room is nil without an open room.
	Room    *Room    `json:"room"`
	Friends []Friend `json:"friends"`
	Invites []Invite `json:"invites"`
}

// Option configures New.
type Option func(*config)

type config struct {
	signaling   Signaling
	iceServers  []string
	settings    webrtc.SettingEngine
	logger      *slog.Logger
	onState     func(State)
	onPlayback  func(Playback)
	defaultName string
	now         func() time.Time
}

// WithSignaling replaces the public Nostr relays, for tests.
func WithSignaling(signaling Signaling) Option {
	return func(config *config) { config.signaling = signaling }
}

// withSettingEngine replaces the WebRTC settings, for loopback tests.
func withSettingEngine(settings webrtc.SettingEngine) Option {
	return func(config *config) { config.settings = settings }
}

// WithLogger sets where connection problems are logged.
func WithLogger(logger *slog.Logger) Option {
	return func(config *config) { config.logger = logger }
}

// WithOnState registers a callback for every change of State.
func WithOnState(onState func(State)) Option {
	return func(config *config) { config.onState = onState }
}

// WithOnPlayback registers the callback that makes the local player follow the room.
func WithOnPlayback(onPlayback func(Playback)) Option {
	return func(config *config) { config.onPlayback = onPlayback }
}

// WithDefaultName sets the display name a new identity starts with.
func WithDefaultName(name string) Option {
	return func(config *config) { config.defaultName = name }
}

// storedIdentity is the identity as kept in the Store.
type storedIdentity struct {
	signer        *signer
	name          string
	sharePresence bool
}

// Together is safe for concurrent use. Callbacks run without locks held.
type Together struct {
	store      Store
	signaling  Signaling
	rooms      *rooms
	syncer     *syncer
	logger     *slog.Logger
	onState    func(State)
	now        func() time.Time
	presence   presenceTracker
	lobbyOnce  sync.Once
	stopLobby  context.CancelFunc
	lobbyStart context.Context

	mutex    sync.Mutex
	identity storedIdentity
	friends  []library.TogetherFriend
	invites  []Invite
	lobby    *lobby
}

// New loads (or creates) the user's identity and friends. The lobby (invites
// and presence) starts right away when the user has friends, otherwise on
// first use.
func New(ctx context.Context, store Store, playbackQueue Queue, options ...Option) (*Together, error) {
	settings := newConfig(options)
	together := &Together{
		store:     store,
		signaling: settings.signaling,
		logger:    settings.logger,
		onState:   settings.onState,
		now:       settings.now,
		presence:  presenceTracker{friends: map[string]Presence{}},
	}
	together.lobbyStart, together.stopLobby = context.WithCancel(context.Background())
	together.rooms = newRooms(settings.signaling, settings.iceServers, settings.settings, settings.logger, settings.now)
	together.syncer = &syncer{
		queue:      playbackQueue,
		send:       together.sendToRoom,
		onPlayback: settings.onPlayback,
		now:        settings.now,
	}
	together.syncer.resetLocked("", false)
	together.rooms.onChange = together.roomChanged
	together.rooms.onMessage = together.syncer.receive

	if err := together.loadIdentity(ctx, settings.defaultName); err != nil {
		return nil, err
	}
	if err := together.loadFriends(ctx); err != nil {
		return nil, err
	}
	together.presence.share = together.identity.sharePresence
	together.presenceQueueChanged(playbackQueue.State())
	if len(together.friends) > 0 {
		go together.ensureLobby()
	}
	return together, nil
}

func newConfig(options []Option) config {
	settings := config{
		iceServers:  DefaultICEServers,
		logger:      slog.Default(),
		onState:     func(State) {},
		onPlayback:  func(Playback) {},
		defaultName: defaultName,
		now:         time.Now,
	}
	for _, option := range options {
		option(&settings)
	}
	if settings.signaling == nil {
		settings.signaling = NostrSignaling{Relays: DefaultRelays, Logger: settings.logger}
	}
	return settings
}

// loadIdentity restores the stored identity, or creates one: an identity whose
// key does not match its ID is replaced too.
func (together *Together) loadIdentity(ctx context.Context, fallbackName string) error {
	stored, found, err := together.store.TogetherIdentity(ctx)
	if err != nil {
		return err
	}
	if found {
		restored, err := signerFromSeed(stored.SigningKey)
		if err == nil && restored.userID == stored.UserID {
			together.identity = storedIdentity{signer: restored, name: stored.Name, sharePresence: stored.SharePresence}
			return nil
		}
		together.logger.Warn("together: stored identity unusable, creating a new one")
	}
	created, err := newSigner()
	if err != nil {
		return err
	}
	name, err := cleanName(fallbackName)
	if err != nil {
		name = defaultName
	}
	together.identity = storedIdentity{signer: created, name: name, sharePresence: true}
	return together.store.SaveTogetherIdentity(ctx, libraryIdentity(together.identity))
}

func libraryIdentity(identity storedIdentity) library.TogetherIdentity {
	return library.TogetherIdentity{
		UserID:        identity.signer.userID,
		SigningKey:    identity.signer.seedHex(),
		Name:          identity.name,
		SharePresence: identity.sharePresence,
	}
}

// saveIdentity stores a changed identity, then keeps it.
func (together *Together) saveIdentity(ctx context.Context, change func(*storedIdentity)) error {
	together.mutex.Lock()
	changed := together.identity
	together.mutex.Unlock()
	change(&changed)
	if err := together.store.SaveTogetherIdentity(ctx, libraryIdentity(changed)); err != nil {
		return err
	}
	together.mutex.Lock()
	together.identity = changed
	together.mutex.Unlock()
	return nil
}

func (together *Together) loadFriends(ctx context.Context) error {
	friends, err := together.store.TogetherFriends(ctx)
	if err != nil {
		return err
	}
	together.mutex.Lock()
	together.friends = friends
	together.mutex.Unlock()
	return nil
}

// State returns what apps render.
func (together *Together) State() State {
	room := roomOf(together.rooms.snapshot())
	share := together.sharingPresence()
	together.mutex.Lock()
	friends := slices.Clone(together.friends)
	state := State{
		Self:          together.selfLocked(),
		SharePresence: share,
		Room:          room,
		Invites:       together.liveInvitesLocked(),
	}
	together.mutex.Unlock()
	state.Friends = make([]Friend, len(friends))
	for index, friend := range friends {
		state.Friends[index] = Friend{
			UserID:   friend.UserID,
			Name:     friend.Name,
			Trusted:  friend.Trusted,
			AddedAt:  friend.AddedAt,
			Presence: together.friendPresence(friend.UserID),
		}
	}
	return state
}

func roomOf(snapshot *roomSnapshot) *Room {
	if snapshot == nil {
		return nil
	}
	return &Room{
		HostUserID: snapshot.HostUserID,
		Hosting:    snapshot.Hosting,
		Status:     snapshot.Status,
		SelfPeerID: snapshot.Self.ID,
		Peers:      snapshot.Peers,
		Requests:   snapshot.Requests,
	}
}

func (together *Together) liveInvitesLocked() []Invite {
	cutoff := together.now().Add(-inviteLifetime)
	together.invites = slices.DeleteFunc(together.invites, func(invite Invite) bool {
		return invite.ReceivedAt.Before(cutoff)
	})
	// Not slices.Clone: it keeps nil, which apps receive as null instead of [].
	return append([]Invite{}, together.invites...)
}

func (together *Together) notify() {
	together.onState(together.State())
}

// SetName changes the name others see; rooms opened afterwards use it.
func (together *Together) SetName(ctx context.Context, name string) error {
	cleaned, err := cleanName(name)
	if err != nil {
		return err
	}
	if err := together.saveIdentity(ctx, func(identity *storedIdentity) { identity.name = cleaned }); err != nil {
		return err
	}
	together.schedulePresence()
	together.notify()
	return nil
}

func cleanName(name string) (string, error) {
	cleaned := collapseSpaces(name)
	if cleaned == "" {
		return "", ErrInvalidName
	}
	if utf8.RuneCountInString(cleaned) > maxNameLength {
		cleaned = string([]rune(cleaned)[:maxNameLength])
	}
	return cleaned, nil
}

func collapseSpaces(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// AddFriend stores a friend by the ID they shared; name may be empty and is
// then learned once they are online or in a session with the user.
func (together *Together) AddFriend(ctx context.Context, userIDText, name string) error {
	userID, err := NormalizeUserID(userIDText)
	if err != nil {
		return err
	}
	if userID == together.self().UserID {
		return ErrOwnUserID
	}
	friend, _ := together.findFriend(userID)
	friend.UserID = userID
	if cleaned := collapseSpaces(name); cleaned != "" {
		friend.Name = cleaned
	}
	return together.saveFriend(ctx, friend)
}

// SetFriendTrusted decides whether a friend joins the user's sessions without asking.
func (together *Together) SetFriendTrusted(ctx context.Context, userID string, trusted bool) error {
	friend, found := together.findFriend(userID)
	if !found {
		return nil
	}
	friend.Trusted = trusted
	return together.saveFriend(ctx, friend)
}

// RemoveFriend deletes a friend.
func (together *Together) RemoveFriend(ctx context.Context, userID string) error {
	if err := together.store.RemoveTogetherFriend(ctx, userID); err != nil {
		return err
	}
	return together.friendsChanged(ctx)
}

func (together *Together) saveFriend(ctx context.Context, friend library.TogetherFriend) error {
	if err := together.store.SaveTogetherFriend(ctx, friend); err != nil {
		return err
	}
	return together.friendsChanged(ctx)
}

// friendsChanged reloads the friends, follows their presence and tells the app.
func (together *Together) friendsChanged(ctx context.Context) error {
	if err := together.loadFriends(ctx); err != nil {
		return err
	}
	go together.ensureLobby()
	go together.followFriends()
	together.notify()
	return nil
}

func (together *Together) findFriend(userID string) (library.TogetherFriend, bool) {
	together.mutex.Lock()
	defer together.mutex.Unlock()
	for _, friend := range together.friends {
		if friend.UserID == userID {
			return friend, true
		}
	}
	return library.TogetherFriend{}, false
}

func (together *Together) friendUserIDs() []string {
	together.mutex.Lock()
	defer together.mutex.Unlock()
	userIDs := make([]string, len(together.friends))
	for index, friend := range together.friends {
		userIDs[index] = friend.UserID
	}
	return userIDs
}

func (together *Together) self() Identity {
	together.mutex.Lock()
	defer together.mutex.Unlock()
	return together.selfLocked()
}

func (together *Together) selfLocked() Identity {
	return Identity{UserID: together.identity.signer.userID, Name: together.identity.name}
}

func (together *Together) currentSigner() *signer {
	together.mutex.Lock()
	defer together.mutex.Unlock()
	return together.identity.signer
}

func (together *Together) selfPeer() Peer {
	identity := together.self()
	return Peer{UserID: identity.UserID, Name: identity.Name}
}

// Host opens the user's own room, which friends join by the user's ID. An
// open own room is kept as it is.
func (together *Together) Host() error {
	go together.ensureLobby()
	snapshot := together.rooms.snapshot()
	if snapshot != nil && snapshot.Hosting {
		return nil
	}
	return together.rooms.open(together.self().UserID, together.selfPeer(), together.currentSigner(), together.admits)
}

// admits lets trusted friends in without asking.
func (together *Together) admits(guest Peer) bool {
	friend, found := together.findFriend(guest.UserID)
	return found && friend.Trusted
}

// AcceptJoin lets a waiting guest into the user's room. With trust, the guest
// becomes a trusted friend and joins without asking from then on.
func (together *Together) AcceptJoin(ctx context.Context, peerID string, trust bool) error {
	if trust {
		if err := together.trustRequester(ctx, peerID); err != nil {
			return err
		}
	}
	return together.rooms.accept(peerID)
}

func (together *Together) trustRequester(ctx context.Context, peerID string) error {
	snapshot := together.rooms.snapshot()
	if snapshot == nil {
		return errRequestGone
	}
	index := slices.IndexFunc(snapshot.Requests, func(peer Peer) bool { return peer.ID == peerID })
	if index < 0 {
		return errRequestGone
	}
	requester := snapshot.Requests[index]
	friend, found := together.findFriend(requester.UserID)
	if !found {
		friend = library.TogetherFriend{UserID: requester.UserID, Name: requester.Name}
	}
	friend.Trusted = true
	return together.saveFriend(ctx, friend)
}

// DeclineJoin turns a waiting guest away for the rest of the session.
func (together *Together) DeclineJoin(peerID string) {
	together.rooms.decline(peerID)
}

// Join enters the room of the person with the given ID, leaving any other
// room. Unless the host trusts the user, the host has to accept first.
func (together *Together) Join(userIDText string) error {
	hostUserID, err := NormalizeUserID(userIDText)
	if err != nil {
		return err
	}
	if hostUserID == together.self().UserID {
		return together.Host()
	}
	go together.ensureLobby()
	together.removeInvite(hostUserID)
	return together.rooms.open(hostUserID, together.selfPeer(), together.currentSigner(), together.admits)
}

// Leave closes the open room; the queue stays as it is.
func (together *Together) Leave() {
	together.rooms.leave()
}

// Invite hosts the user's room (when not already) and asks the friend to join.
// It returns once a relay accepted the invite. Invited friends still need to
// be let in unless they are trusted.
func (together *Together) Invite(friendUserID string) error {
	userID, err := NormalizeUserID(friendUserID)
	if err != nil {
		return err
	}
	if err := together.Host(); err != nil {
		return err
	}
	opened := together.ensureLobby()
	if opened == nil {
		return errNoRelay
	}
	return opened.sendInvite(userID, together.self())
}

// DismissInvite drops the invite from the person with userID.
func (together *Together) DismissInvite(userID string) {
	together.removeInvite(userID)
	together.notify()
}

func (together *Together) removeInvite(userID string) {
	together.mutex.Lock()
	defer together.mutex.Unlock()
	together.invites = slices.DeleteFunc(together.invites, func(invite Invite) bool {
		return invite.From.UserID == userID
	})
}

// QueueChanged shares a queue change with the room and the current track with
// friends; call it with every queue state.
func (together *Together) QueueChanged(state queue.State) {
	together.syncer.queueChanged(state)
	together.presenceQueueChanged(state)
}

// ReportPlayback tells the room and friends about the local player: on every
// play, pause and seek by the user, and every few seconds otherwise.
func (together *Together) ReportPlayback(report PlaybackReport) {
	together.syncer.reportPlayback(report)
	together.presencePlaybackReported(report)
}

// Following reports whether the user is a guest in someone's room; automatic
// queue extensions (autoplay) should then come from the host only.
func (together *Together) Following() bool {
	return together.syncer.following()
}

// Close leaves the room, tells friends the user went offline and stops the lobby.
func (together *Together) Close() {
	together.rooms.leave()
	if together.sharingPresence() {
		together.sendPresence(together.offlineMessage(), offlineSendTimeout)
	}
	together.stopLobby()
	together.mutex.Lock()
	opened := together.lobby
	together.mutex.Unlock()
	if opened != nil {
		opened.close()
	}
}

func (together *Together) sendToRoom(payload string) {
	if err := together.rooms.send(payload); err != nil && !errors.Is(err, errNotConnected) {
		together.logger.Debug("together: send", "error", err)
	}
}

func (together *Together) roomChanged(snapshot *roomSnapshot) {
	together.syncer.roomChanged(snapshot)
	together.presenceSessionChanged(snapshot)
	if snapshot != nil {
		together.learnFriendNames(snapshot.Peers)
	}
	together.notify()
}

// learnFriendNames fills in names of friends that were added by ID only.
func (together *Together) learnFriendNames(peers []Peer) {
	for _, peer := range peers {
		together.learnFriendName(peer.UserID, peer.Name)
	}
}

func (together *Together) learnFriendName(userID, name string) {
	friend, found := together.findFriend(userID)
	if !found || friend.Name != "" || name == "" {
		return
	}
	friend.Name = name
	if err := together.saveFriend(context.Background(), friend); err != nil {
		together.logger.Warn("together: save friend name", "error", err)
	}
}

// ensureLobby starts listening for invites and presence once; it derives
// keys, so callers run it off their own goroutine when they can.
func (together *Together) ensureLobby() *lobby {
	together.lobbyOnce.Do(together.openLobby)
	return together.currentLobby()
}

func (together *Together) currentLobby() *lobby {
	together.mutex.Lock()
	defer together.mutex.Unlock()
	return together.lobby
}

func (together *Together) openLobby() {
	opened, err := openLobby(lobbySettings{
		signaling:  together.signaling,
		signer:     together.currentSigner(),
		logger:     together.logger,
		now:        together.now,
		onInvite:   together.receiveInvite,
		onPresence: together.receiveFriendPresence,
	}, together.friendUserIDs())
	if err != nil {
		together.logger.Warn("together: lobby", "error", err)
		return
	}
	together.mutex.Lock()
	together.lobby = opened
	together.mutex.Unlock()
	go together.runPresence(together.lobbyStart)
}

// followFriends subscribes to the presence of the current friends.
func (together *Together) followFriends() {
	if opened := together.ensureLobby(); opened != nil {
		opened.setFriends(together.friendUserIDs())
	}
}

func (together *Together) receiveFriendPresence(userID string, message presenceMessage) {
	together.learnFriendName(userID, collapseSpaces(message.From.Name))
	together.receivePresence(userID, message)
}

func (together *Together) receiveInvite(from Identity) {
	current := together.rooms.snapshot()
	if current != nil && current.HostUserID == from.UserID {
		return
	}
	together.mutex.Lock()
	together.invites = slices.DeleteFunc(together.invites, func(invite Invite) bool {
		return invite.From.UserID == from.UserID
	})
	together.invites = append(together.invites, Invite{From: from, ReceivedAt: together.now()})
	together.mutex.Unlock()
	together.notify()
}
