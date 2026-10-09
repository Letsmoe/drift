//go:build live

package together

import (
	"context"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/queue"
)

// Uses the public relays and STUN servers.
func TestLiveListenTogetherOverPublicRelays(t *testing.T) {
	host, err := New(context.Background(), newMemoryStore(), queue.New(), WithDefaultName("Ana"))
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()
	guestQueue := queue.New()
	guest, err := New(context.Background(), newMemoryStore(), guestQueue, WithDefaultName("Ben"))
	if err != nil {
		t.Fatal(err)
	}
	defer guest.Close()

	guest.ensureLobby()
	if err := host.Invite(guest.State().Self.UserID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the invite", func() bool { return len(guest.State().Invites) == 1 })

	if err := guest.Join(host.State().Self.UserID); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the join request", func() bool {
		room := host.State().Room
		return room != nil && len(room.Requests) == 1
	})
	if err := host.AcceptJoin(context.Background(), host.State().Room.Requests[0].ID, false); err != nil {
		t.Fatal(err)
	}
	eventually(t, "both connected", func() bool {
		hostRoom := host.State().Room
		guestRoom := guest.State().Room
		return hostRoom != nil && guestRoom != nil &&
			hostRoom.Status == StatusConnected && guestRoom.Status == StatusConnected
	})

	// Presence: the host follows the guest, who plays something.
	guestID := guest.State().Self.UserID
	if err := host.AddFriend(context.Background(), guestID, ""); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second) // the host's lobby resubscribes on the relays
	state := guestQueue.Play([]music.Track{{VideoID: "live", Title: "Live"}}, music.ItemKindSong, 0, "")
	guest.QueueChanged(state)
	guest.ReportPlayback(PlaybackReport{EntryID: state.Entries[0].ID, PositionSeconds: 12, UserAction: true})
	eventually(t, "the guest's presence", func() bool {
		friends := host.State().Friends
		return len(friends) == 1 && friends[0].Presence != nil && friends[0].Presence.NowPlaying != nil &&
			friends[0].Presence.NowPlaying.Track.VideoID == "live"
	})
}
