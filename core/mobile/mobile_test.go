package mobile

import (
	"encoding/json"
	"regexp"
	"sync"
	"testing"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/queue"
)

type recordingListener struct {
	mutex  sync.Mutex
	states []string
}

func (listener *recordingListener) OnQueueChanged(stateJSON string) {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()
	listener.states = append(listener.states, stateJSON)
}

func (listener *recordingListener) lastState() string {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()
	if len(listener.states) == 0 {
		return ""
	}
	return listener.states[len(listener.states)-1]
}

const twoTracksJSON = `[
	{"videoId":"vid1","title":"One","artists":[{"name":"A","id":"a1"}]},
	{"videoId":"vid2","title":"Two"}
]`

func openTestCore(t *testing.T) (*Core, *recordingListener) {
	t.Helper()
	listener := &recordingListener{}
	core, err := Open(t.TempDir(), listener)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(core.Close)
	return core, listener
}

func decodeInto[T any](t *testing.T, encoded string) T {
	t.Helper()
	var value T
	err := json.Unmarshal([]byte(encoded), &value)
	if err != nil {
		t.Fatalf("decode %q: %v", encoded, err)
	}
	return value
}

func TestStreamURL(t *testing.T) {
	core, _ := openTestCore(t)
	pattern := regexp.MustCompile(`^http://127\.0\.0\.1:\d+/stream/abc$`)
	url := core.StreamURL("abc")
	if !pattern.MatchString(url) {
		t.Fatalf("StreamURL = %q", url)
	}
}

func TestEmptyListsEncodeAsArrays(t *testing.T) {
	core, _ := openTestCore(t)
	liked, err := core.LikedTracks()
	if err != nil {
		t.Fatal(err)
	}
	if liked != "[]" {
		t.Fatalf("LikedTracks = %s", liked)
	}
	searches, err := core.RecentSearches(5)
	if err != nil {
		t.Fatal(err)
	}
	if searches != "[]" {
		t.Fatalf("RecentSearches = %s", searches)
	}
}

func TestNestedNilSlicesEncodeAsArrays(t *testing.T) {
	encoded := mustEncode(queue.State{})
	if encoded != `{"entries":[],"index":0,"shuffle":false,"magicShuffle":false,"repeat":"","autoplay":false,"context":""}` {
		t.Fatalf("state = %s", encoded)
	}
}

func TestPlaylistRoundTrip(t *testing.T) {
	core, _ := openTestCore(t)
	created, err := core.CreatePlaylist("Mix")
	if err != nil {
		t.Fatal(err)
	}
	playlist := decodeInto[library.Playlist](t, created)
	err = core.AddToPlaylist(playlist.ID, twoTracksJSON)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := core.LibraryPlaylist(playlist.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(decodeInto[library.Playlist](t, loaded).Entries); got != 2 {
		t.Fatalf("playlist has %d entries, want 2", got)
	}
	summaries, err := core.Playlists()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(decodeInto[[]library.PlaylistSummary](t, summaries)); got != 1 {
		t.Fatalf("%d playlists, want 1", got)
	}
}

func TestLikeRoundTrip(t *testing.T) {
	core, _ := openTestCore(t)
	err := core.SetRating(`{"videoId":"vid1","title":"One"}`, "song", 1)
	if err != nil {
		t.Fatal(err)
	}
	liked, err := core.LikedTracks()
	if err != nil {
		t.Fatal(err)
	}
	tracks := decodeInto[[]map[string]any](t, liked)
	if len(tracks) != 1 || tracks[0]["videoId"] != "vid1" {
		t.Fatalf("liked = %s", liked)
	}
	ratings, err := core.Ratings(`["vid1","other"]`)
	if err != nil {
		t.Fatal(err)
	}
	if got := decodeInto[map[string]int](t, ratings); got["vid1"] != 1 || len(got) != 1 {
		t.Fatalf("ratings = %s", ratings)
	}
}

func TestSavedItems(t *testing.T) {
	core, _ := openTestCore(t)
	err := core.Save(`{"kind":"album","id":"MPREb_1","title":"Album"}`)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := core.IsSaved("album", "MPREb_1")
	if err != nil || !saved {
		t.Fatalf("IsSaved = %v, %v", saved, err)
	}
	items, err := core.SavedItems()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(decodeInto[[]library.SavedItem](t, items)); got != 1 {
		t.Fatalf("%d saved items, want 1", got)
	}
	err = core.Unsave("album", "MPREb_1")
	if err != nil {
		t.Fatal(err)
	}
	items, err = core.SavedItems()
	if err != nil || items != "[]" {
		t.Fatalf("SavedItems after unsave = %s, %v", items, err)
	}
}

func TestQueuePlayAndStep(t *testing.T) {
	core, listener := openTestCore(t)
	played, err := core.QueuePlay(twoTracksJSON, "song", 0, "test")
	if err != nil {
		t.Fatal(err)
	}
	state := decodeInto[queue.State](t, played)
	if len(state.Entries) != 2 || state.Index != 0 || state.Context != "test" {
		t.Fatalf("state = %s", played)
	}
	if listener.lastState() != played {
		t.Fatalf("listener got %q, want %q", listener.lastState(), played)
	}

	step := decodeInto[struct {
		State queue.State `json:"state"`
		Moved bool        `json:"moved"`
	}](t, core.QueueNext())
	if !step.Moved || step.State.Index != 1 {
		t.Fatalf("next step = %+v", step)
	}
	if core.QueueState() == "" {
		t.Fatal("empty QueueState")
	}
}

func TestInvalidJSONIsAnError(t *testing.T) {
	core, _ := openTestCore(t)
	_, err := core.QueueAdd("not json")
	if err == nil {
		t.Fatal("expected error")
	}
}
