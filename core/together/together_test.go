package together

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pion/webrtc/v4"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
	"github.com/letsmoe/drift-music/core/queue"
)

// memorySignaling is an in-process relay: every send reaches every
// subscriber of the topic, the sender included.
type memorySignaling struct {
	mutex       sync.Mutex
	subscribers map[string][]chan SignalMessage
}

type memoryChannel struct {
	bus      *memorySignaling
	received chan SignalMessage
}

func newMemorySignaling() *memorySignaling {
	return &memorySignaling{subscribers: map[string][]chan SignalMessage{}}
}

func (bus *memorySignaling) Open(ctx context.Context, topics []string) (SignalChannel, error) {
	channel := &memoryChannel{bus: bus, received: make(chan SignalMessage, 256)}
	bus.mutex.Lock()
	for _, topic := range topics {
		bus.subscribers[topic] = append(bus.subscribers[topic], channel.received)
	}
	bus.mutex.Unlock()
	context.AfterFunc(ctx, func() { bus.unsubscribe(channel.received) })
	return channel, nil
}

func (bus *memorySignaling) unsubscribe(received chan SignalMessage) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()
	for topic, subscribers := range bus.subscribers {
		bus.subscribers[topic] = slices.DeleteFunc(subscribers, func(other chan SignalMessage) bool { return other == received })
	}
}

func (channel *memoryChannel) Receive() <-chan SignalMessage {
	return channel.received
}

func (channel *memoryChannel) Send(_ context.Context, topic, content string) error {
	channel.bus.mutex.Lock()
	defer channel.bus.mutex.Unlock()
	for _, subscriber := range channel.bus.subscribers[topic] {
		select {
		case subscriber <- SignalMessage{Topic: topic, Content: content}:
		default:
		}
	}
	return nil
}

// memoryStore is a Store without a database.
type memoryStore struct {
	mutex    sync.Mutex
	identity *library.TogetherIdentity
	friends  map[string]library.TogetherFriend
}

func newMemoryStore() *memoryStore {
	return &memoryStore{friends: map[string]library.TogetherFriend{}}
}

func (store *memoryStore) TogetherIdentity(context.Context) (library.TogetherIdentity, bool, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.identity == nil {
		return library.TogetherIdentity{}, false, nil
	}
	return *store.identity, true, nil
}

func (store *memoryStore) SaveTogetherIdentity(_ context.Context, identity library.TogetherIdentity) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.identity = &identity
	return nil
}

func (store *memoryStore) TogetherFriends(context.Context) ([]library.TogetherFriend, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	friends := slices.Collect(maps.Values(store.friends))
	slices.SortFunc(friends, func(left, right library.TogetherFriend) int { return strings.Compare(left.UserID, right.UserID) })
	return friends, nil
}

func (store *memoryStore) SaveTogetherFriend(_ context.Context, friend library.TogetherFriend) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.friends[friend.UserID] = friend
	return nil
}

func (store *memoryStore) RemoveTogetherFriend(_ context.Context, userID string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.friends, userID)
	return nil
}

func (store *memoryStore) friend(userID string) library.TogetherFriend {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.friends[userID]
}

// participant is one app: a queue, Listen Together and what it was told.
type participant struct {
	together *Together
	queue    *queue.Queue
	store    *memoryStore

	mutex     sync.Mutex
	playbacks []Playback
}

func newParticipant(t *testing.T, bus *memorySignaling, name string) *participant {
	t.Helper()
	settings := webrtc.SettingEngine{}
	settings.SetIncludeLoopbackCandidate(true)
	settings.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeUDP4})

	created := &participant{store: newMemoryStore()}
	created.queue = queue.New(queue.WithOnChange(func(state queue.State) {
		if created.together != nil {
			created.together.QueueChanged(state)
		}
	}))
	together, err := New(context.Background(), created.store, created.queue,
		WithSignaling(bus),
		withSettingEngine(settings),
		WithDefaultName(name),
		WithOnPlayback(created.recordPlayback),
	)
	if err != nil {
		t.Fatal(err)
	}
	created.together = together
	t.Cleanup(together.Close)
	return created
}

func (participant *participant) recordPlayback(playback Playback) {
	participant.mutex.Lock()
	defer participant.mutex.Unlock()
	participant.playbacks = append(participant.playbacks, playback)
}

func (participant *participant) lastPlayback() (Playback, bool) {
	participant.mutex.Lock()
	defer participant.mutex.Unlock()
	if len(participant.playbacks) == 0 {
		return Playback{}, false
	}
	return participant.playbacks[len(participant.playbacks)-1], true
}

func (participant *participant) lobbyFollows(userID string) bool {
	opened := participant.together.currentLobby()
	if opened == nil {
		return false
	}
	opened.mutex.Lock()
	defer opened.mutex.Unlock()
	return slices.Contains(slices.Collect(maps.Values(opened.topicOwners)), userID)
}

func (participant *participant) lobbyOpen() bool {
	return participant.together.currentLobby() != nil
}

func (participant *participant) friendState(userID string) (Friend, bool) {
	for _, friend := range participant.together.State().Friends {
		if friend.UserID == userID {
			return friend, true
		}
	}
	return Friend{}, false
}

func (participant *participant) requests() []Peer {
	room := participant.together.State().Room
	if room == nil {
		return nil
	}
	return room.Requests
}

func eventually(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", description)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func roomStatus(participant *participant) Status {
	room := participant.together.State().Room
	if room == nil {
		return ""
	}
	return room.Status
}

func currentVideoID(participant *participant) string {
	state := participant.queue.State()
	if state.Index < 0 {
		return ""
	}
	return state.Entries[state.Index].Track.VideoID
}

func tracksNamed(videoIDs ...string) []music.Track {
	tracks := make([]music.Track, len(videoIDs))
	for position, videoID := range videoIDs {
		tracks[position] = music.Track{VideoID: videoID, Title: videoID}
	}
	return tracks
}

func TestListenTogether(t *testing.T) {
	bus := newMemorySignaling()
	ana := newParticipant(t, bus, "Ana")
	ben := newParticipant(t, bus, "Ben")
	anaID := ana.together.State().Self.UserID

	if err := ben.together.AddFriend(context.Background(), FormatUserID(anaID), ""); err != nil {
		t.Fatal(err)
	}
	eventually(t, "Ben's lobby", ben.lobbyOpen)

	ana.queue.Play(tracksNamed("a", "b", "c"), music.ItemKindSong, 0, "album:x")
	if err := ana.together.Invite(ben.together.State().Self.UserID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the invite", func() bool {
		invites := ben.together.State().Invites
		return len(invites) == 1 && invites[0].From.UserID == anaID && invites[0].From.Name == "Ana"
	})

	joinAndAccept(t, ana, ben, anaID)
	if len(ben.together.State().Invites) != 0 {
		t.Fatal("joining kept the invite")
	}
	if !ben.together.Following() || ana.together.Following() {
		t.Fatal("only the guest follows")
	}

	eventually(t, "Ben adopting Ana's queue", func() bool { return currentVideoID(ben) == "a" })
	eventually(t, "Ben learning Ana's name", func() bool { return ben.store.friend(anaID).Name == "Ana" })

	ben.queue.Next()
	eventually(t, "Ana following Ben's skip", func() bool { return currentVideoID(ana) == "b" })
	if ana.queue.State().Entries[1].ID != ben.queue.State().Entries[1].ID {
		t.Fatal("entry IDs differ between the queues")
	}

	ana.queue.Add(tracksNamed("d"))
	eventually(t, "Ben seeing Ana's addition", func() bool {
		return len(ben.queue.State().Entries) == 4
	})

	entryID := ben.queue.State().Entries[1].ID
	ben.together.ReportPlayback(PlaybackReport{EntryID: entryID, Paused: true, PositionSeconds: 42, UserAction: true})
	eventually(t, "Ana pausing", func() bool {
		playback, ok := ana.lastPlayback()
		return ok && playback.Paused && playback.PositionSeconds == 42 && playback.From == "Ben" && playback.EntryID == entryID
	})

	ana.together.ReportPlayback(PlaybackReport{EntryID: entryID, PositionSeconds: 50})
	eventually(t, "Ben receiving the heartbeat", func() bool {
		playback, ok := ben.lastPlayback()
		return ok && !playback.Paused && playback.PositionSeconds >= 50 && playback.ToleranceSeconds == heartbeatToleranceSeconds
	})

	ben.together.Leave()
	eventually(t, "Ana waiting again", func() bool {
		room := ana.together.State().Room
		return room != nil && room.Status == StatusWaiting && len(room.Peers) == 1
	})
	if ben.together.Following() {
		t.Fatal("Ben still follows after leaving")
	}
	if currentVideoID(ben) != "b" {
		t.Fatal("leaving changed the queue")
	}
}

// joinAndAccept has guest join host's room and host accept the request once.
func joinAndAccept(t *testing.T, host, guest *participant, hostUserID string) {
	t.Helper()
	if err := guest.together.Join(hostUserID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the join request", func() bool {
		return len(host.requests()) == 1 && roomStatus(guest) == StatusAwaiting
	})
	if err := host.together.AcceptJoin(context.Background(), host.requests()[0].ID, false); err != nil {
		t.Fatal(err)
	}
	eventually(t, "both connected", func() bool {
		return roomStatus(host) == StatusConnected && roomStatus(guest) == StatusConnected
	})
}

func TestJoinNeedsApprovalUnlessTrusted(t *testing.T) {
	bus := newMemorySignaling()
	ana := newParticipant(t, bus, "Ana")
	ben := newParticipant(t, bus, "Ben")
	anaID := ana.together.State().Self.UserID
	benID := ben.together.State().Self.UserID
	if err := ana.together.Host(); err != nil {
		t.Fatal(err)
	}

	if err := ben.together.Join(anaID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the join request", func() bool { return len(ana.requests()) == 1 })
	if request := ana.requests()[0]; request.UserID != benID || request.Name != "Ben" {
		t.Fatalf("request = %+v", request)
	}
	ana.together.DeclineJoin(ana.requests()[0].ID)
	eventually(t, "Ben declined", func() bool { return roomStatus(ben) == StatusDeclined })
	if len(ana.requests()) != 0 {
		t.Fatal("declined request kept")
	}

	// Every new connection asks again; accepting with trust makes Ben a trusted friend.
	if err := ben.together.Join(anaID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the second request", func() bool { return len(ana.requests()) == 1 })
	if err := ana.together.AcceptJoin(context.Background(), ana.requests()[0].ID, true); err != nil {
		t.Fatal(err)
	}
	eventually(t, "Ben connected", func() bool { return roomStatus(ben) == StatusConnected })
	if friend, ok := ana.friendState(benID); !ok || !friend.Trusted || friend.Name != "Ben" {
		t.Fatalf("Ben as Ana's friend = %+v, %v", friend, ok)
	}

	ben.together.Leave()
	eventually(t, "Ana alone", func() bool { return roomStatus(ana) == StatusWaiting })
	if err := ben.together.Join(anaID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "trusted Ben connected without asking", func() bool {
		return roomStatus(ben) == StatusConnected && len(ana.requests()) == 0
	})
}

func TestJoinUnderSomeoneElsesIDIsIgnored(t *testing.T) {
	bus := newMemorySignaling()
	ana := newParticipant(t, bus, "Ana")
	ben := newParticipant(t, bus, "Ben")
	anaID := ana.together.State().Self.UserID
	benID := ben.together.State().Self.UserID
	if err := ana.together.AddFriend(context.Background(), benID, "Ben"); err != nil {
		t.Fatal(err)
	}
	if err := ana.together.SetFriendTrusted(context.Background(), benID, true); err != nil {
		t.Fatal(err)
	}
	if err := ana.together.Host(); err != nil {
		t.Fatal(err)
	}

	impostor, err := newSigner()
	if err != nil {
		t.Fatal(err)
	}
	key, err := deriveRoomKey(anaID)
	if err != nil {
		t.Fatal(err)
	}
	forged := signal{Type: signalJoin, From: Peer{ID: "forged", UserID: benID, Name: "Ben"}, SentAtMillis: time.Now().UnixMilli()}
	content, err := impostor.seal(key, key.topic, forged)
	if err != nil {
		t.Fatal(err)
	}
	channel, err := bus.Open(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := channel.Send(context.Background(), key.topic, content); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	room := ana.together.State().Room
	if room == nil || room.Status != StatusWaiting || len(room.Requests) != 0 {
		t.Fatalf("forged join changed the room: %+v", room)
	}
}

func TestFriendsSeeWhatYouPlay(t *testing.T) {
	bus := newMemorySignaling()
	ana := newParticipant(t, bus, "Ana")
	ben := newParticipant(t, bus, "Ben")
	anaID := ana.together.State().Self.UserID
	benID := ben.together.State().Self.UserID
	ctx := context.Background()

	if err := ana.together.AddFriend(ctx, benID, ""); err != nil {
		t.Fatal(err)
	}
	eventually(t, "Ana following Ben", func() bool { return ana.lobbyFollows(benID) })

	state := ben.queue.Play(tracksNamed("a", "b"), music.ItemKindSong, 1, "")
	ben.together.ReportPlayback(PlaybackReport{EntryID: state.Entries[1].ID, PositionSeconds: 30})
	if err := ben.together.AddFriend(ctx, anaID, ""); err != nil {
		t.Fatal(err)
	}
	eventually(t, "Ben's presence", func() bool {
		friend, _ := ana.friendState(benID)
		presence := friend.Presence
		return presence != nil && presence.NowPlaying != nil && presence.NowPlaying.Track.VideoID == "b" &&
			presence.NowPlaying.PositionSeconds >= 30 && !presence.NowPlaying.Paused && friend.Name == "Ben"
	})

	if err := ben.together.Host(); err != nil {
		t.Fatal(err)
	}
	eventually(t, "Ben's session showing", func() bool {
		friend, _ := ana.friendState(benID)
		return friend.Presence != nil && friend.Presence.SessionOpen
	})

	ben.together.ReportPlayback(PlaybackReport{EntryID: state.Entries[1].ID, Paused: true, PositionSeconds: 31, UserAction: true})
	eventually(t, "Ben's pause", func() bool {
		friend, _ := ana.friendState(benID)
		return friend.Presence != nil && friend.Presence.NowPlaying != nil && friend.Presence.NowPlaying.Paused
	})

	if err := ben.together.SetSharePresence(ctx, false); err != nil {
		t.Fatal(err)
	}
	eventually(t, "Ben offline", func() bool {
		friend, _ := ana.friendState(benID)
		return friend.Presence == nil
	})
	if ben.together.State().SharePresence {
		t.Fatal("sharing still on")
	}
}

func TestAddFriendRejectsOwnAndInvalidIDs(t *testing.T) {
	ana := newParticipant(t, newMemorySignaling(), "Ana")
	ctx := context.Background()
	if err := ana.together.AddFriend(ctx, ana.together.State().Self.UserID, "me"); err == nil {
		t.Fatal("added own ID")
	}
	if err := ana.together.AddFriend(ctx, "nope", ""); err == nil {
		t.Fatal("added an invalid ID")
	}
	if err := ana.together.AddFriend(ctx, "K7QD-9MZA-X2PT", "  Cleo   M "); err != nil {
		t.Fatal(err)
	}
	friends := ana.together.State().Friends
	if len(friends) != 1 || friends[0].UserID != "K7QD9MZAX2PT" || friends[0].Name != "Cleo M" {
		t.Fatalf("friends = %+v", friends)
	}
	if err := ana.together.SetFriendTrusted(ctx, "K7QD9MZAX2PT", true); err != nil {
		t.Fatal(err)
	}
	if friends := ana.together.State().Friends; !friends[0].Trusted {
		t.Fatal("friend not trusted")
	}
	if err := ana.together.RemoveFriend(ctx, "K7QD9MZAX2PT"); err != nil {
		t.Fatal(err)
	}
	if len(ana.together.State().Friends) != 0 {
		t.Fatal("friend not removed")
	}
}

func TestIdentityPersists(t *testing.T) {
	store := newMemoryStore()
	first, err := New(context.Background(), store, queue.New(), WithSignaling(newMemorySignaling()), WithDefaultName("  "))
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if first.State().Self.Name != defaultName {
		t.Fatalf("blank default name gave %q", first.State().Self.Name)
	}
	if err := first.SetName(context.Background(), "Dana"); err != nil {
		t.Fatal(err)
	}
	second, err := New(context.Background(), store, queue.New(), WithSignaling(newMemorySignaling()))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if !slices.Equal([]string{second.State().Self.UserID, second.State().Self.Name},
		[]string{first.State().Self.UserID, "Dana"}) {
		t.Fatalf("identity not kept: %+v vs %+v", second.State().Self, first.State().Self)
	}
}

// Apps index into every list of State; nil slices would reach them as null.
func TestStateHasNoNullLists(t *testing.T) {
	ana := newParticipant(t, newMemorySignaling(), "Ana")
	if err := ana.together.Host(); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(ana.together.State())
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Friends []any `json:"friends"`
		Invites []any `json:"invites"`
		Room    struct {
			Peers    []any `json:"peers"`
			Requests []any `json:"requests"`
		} `json:"room"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Friends == nil || decoded.Invites == nil || decoded.Room.Peers == nil || decoded.Room.Requests == nil {
		t.Fatalf("state has null lists: %s", encoded)
	}
}

func TestTrimmedTrackHasNoNullLists(t *testing.T) {
	trimmed := trimmedTrack(music.Track{VideoID: "a"})
	if trimmed.Artists == nil || trimmed.Thumbnails == nil {
		t.Fatalf("trimmed track = %+v", trimmed)
	}
}
