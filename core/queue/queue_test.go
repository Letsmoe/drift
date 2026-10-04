package queue

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math/rand/v2"
	"runtime"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

func tracksNamed(videoIDs ...string) []music.Track {
	tracks := make([]music.Track, len(videoIDs))
	for position, videoID := range videoIDs {
		tracks[position] = music.Track{VideoID: videoID, Title: videoID}
	}
	return tracks
}

func newTestQueue() *Queue {
	return New(WithRandom(rand.New(rand.NewPCG(1, 2))))
}

func videoIDs(state State) []string {
	ids := make([]string, len(state.Entries))
	for position, entry := range state.Entries {
		ids[position] = entry.Track.VideoID
	}
	return ids
}

func assertOrder(t *testing.T, state State, wantIndex int, want ...string) {
	t.Helper()
	if !slices.Equal(videoIDs(state), want) || state.Index != wantIndex {
		t.Fatalf("queue = %v at %d, want %v at %d", videoIDs(state), state.Index, want, wantIndex)
	}
}

func TestPlayNextPrevious(t *testing.T) {
	queue := newTestQueue()
	state := queue.Play(tracksNamed("a", "b", "c"), music.ItemKindSong, 1, "album:x")
	assertOrder(t, state, 1, "a", "b", "c")
	if state.Context != "album:x" {
		t.Fatalf("context = %q", state.Context)
	}
	state, ok := queue.Next()
	if !ok {
		t.Fatalf("Next failed")
	}
	assertOrder(t, state, 2, "a", "b", "c")
	state, ok = queue.Next()
	if ok {
		t.Fatalf("Next past the end succeeded")
	}
	assertOrder(t, state, 2, "a", "b", "c")
	assertOrder(t, queue.Previous(), 1, "a", "b", "c")
	queue.Previous()
	assertOrder(t, queue.Previous(), 0, "a", "b", "c")
}

func TestRepeatModes(t *testing.T) {
	queue := newTestQueue()
	queue.Play(tracksNamed("a", "b"), music.ItemKindSong, 1, "")

	queue.SetRepeat(RepeatAll)
	state, ok := queue.Next()
	if !ok || state.Index != 0 {
		t.Fatalf("RepeatAll Next = %d, %v", state.Index, ok)
	}
	if state = queue.Previous(); state.Index != 1 {
		t.Fatalf("RepeatAll Previous at start = %d", state.Index)
	}

	queue.SetRepeat(RepeatOne)
	state, ok = queue.Advance()
	if !ok || state.Index != 1 {
		t.Fatalf("RepeatOne Advance = %d, %v", state.Index, ok)
	}
	queue.Previous()
	state, ok = queue.Next()
	if !ok || state.Index != 1 {
		t.Fatalf("RepeatOne user Next should skip, got %d, %v", state.Index, ok)
	}

	queue.SetRepeat(RepeatOff)
	if _, ok = queue.Advance(); ok {
		t.Fatalf("Advance at end with repeat off succeeded")
	}
}

func TestShuffleRoundTrip(t *testing.T) {
	queue := newTestQueue()
	original := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	queue.Play(tracksNamed(original...), music.ItemKindSong, 2, "")
	shuffled := queue.SetShuffle(true)
	if shuffled.Index != 2 || shuffled.Entries[2].Track.VideoID != "c" {
		t.Fatalf("current moved by shuffle: %v at %d", videoIDs(shuffled), shuffled.Index)
	}
	if !slices.Equal(videoIDs(shuffled)[:3], original[:3]) {
		t.Fatalf("history changed by shuffle: %v", videoIDs(shuffled))
	}
	if slices.Equal(videoIDs(shuffled), original) {
		t.Fatalf("shuffle left the order unchanged")
	}
	restored := queue.SetShuffle(false)
	assertOrder(t, restored, 2, original...)
}

func TestPlayWithShuffleStartsWithChosenTrack(t *testing.T) {
	queue := newTestQueue()
	queue.SetShuffle(true)
	state := queue.Play(tracksNamed("a", "b", "c", "d", "e"), music.ItemKindSong, 3, "")
	if state.Index != 0 || state.Entries[0].Track.VideoID != "d" || len(state.Entries) != 5 {
		t.Fatalf("queue = %v at %d", videoIDs(state), state.Index)
	}
	assertOrder(t, queue.SetShuffle(false), 0, "d", "a", "b", "c", "e")
}

func TestAddAndPlayNextKeepAutoplayAtTail(t *testing.T) {
	queue := newTestQueue()
	queue.Play(tracksNamed("a", "b"), music.ItemKindSong, 0, "")
	queue.mutex.Lock()
	queue.insertLocked(2, tracksNamed("auto1", "auto2"), true, queue.endOrderLocked())
	queue.mutex.Unlock()

	assertOrder(t, queue.Add(tracksNamed("x")), 0, "a", "b", "x", "auto1", "auto2")
	assertOrder(t, queue.PlayNext(tracksNamed("n")), 0, "a", "n", "b", "x", "auto1", "auto2")
	state := queue.Add(tracksNamed("y"))
	assertOrder(t, state, 0, "a", "n", "b", "x", "y", "auto1", "auto2")
	if !state.Entries[5].Autoplay || state.Entries[4].Autoplay {
		t.Fatalf("autoplay flags wrong: %+v", state.Entries)
	}
}

func TestShuffleOffRestoresAddedAndPlayNext(t *testing.T) {
	queue := newTestQueue()
	queue.Play(tracksNamed("a", "b", "c", "d"), music.ItemKindSong, 0, "")
	queue.SetShuffle(true)
	queue.PlayNext(tracksNamed("n"))
	queue.Add(tracksNamed("x"))
	assertOrder(t, queue.SetShuffle(false), 0, "a", "n", "b", "c", "d", "x")
}

func TestAddToEmptyQueueMakesFirstCurrent(t *testing.T) {
	queue := newTestQueue()
	assertOrder(t, queue.Add(tracksNamed("a", "b")), 0, "a", "b")
}

func TestSameTrackTwiceHasDistinctEntryIDs(t *testing.T) {
	queue := newTestQueue()
	queue.Play(tracksNamed("a"), music.ItemKindSong, 0, "")
	state := queue.Add(tracksNamed("a"))
	if state.Entries[0].ID == state.Entries[1].ID {
		t.Fatalf("entry IDs collide")
	}
}

func TestMoveAndRemove(t *testing.T) {
	queue := newTestQueue()
	state := queue.Play(tracksNamed("a", "b", "c", "d"), music.ItemKindSong, 1, "")
	idOf := func(position int) int64 { return state.Entries[position].ID }

	assertOrder(t, queue.Move(idOf(3), 0), 1, "a", "b", "d", "c")
	assertOrder(t, queue.Move(idOf(0), 3), 1, "a", "b", "d", "c")
	assertOrder(t, queue.Move(idOf(1), 3), 1, "a", "b", "d", "c")
	assertOrder(t, queue.Move(idOf(3), 99), 1, "a", "b", "c", "d")
	assertOrder(t, queue.Remove(idOf(0)), 0, "b", "c", "d")
	assertOrder(t, queue.Remove(idOf(3)), 0, "b", "c")
	assertOrder(t, queue.Remove(idOf(1)), 0, "c")
	assertOrder(t, queue.Remove(idOf(2)), -1)
}

func TestJumpAndClear(t *testing.T) {
	queue := newTestQueue()
	state := queue.Play(tracksNamed("a", "b", "c"), music.ItemKindSong, 0, "")
	assertOrder(t, queue.Jump(state.Entries[2].ID), 2, "a", "b", "c")
	assertOrder(t, queue.Clear(), 0, "c")
}

type fakeExtender struct {
	requests chan extendCall
	release  chan struct{}
	result   []music.Track
	err      error
}

type extendCall struct {
	upcoming []string
	history  []string
}

func newFakeExtender(result ...string) *fakeExtender {
	return &fakeExtender{requests: make(chan extendCall, 10), result: tracksNamed(result...)}
}

func (extender *fakeExtender) Extend(_ context.Context, upcoming, history []music.Track) ([]music.Track, error) {
	call := extendCall{}
	for _, track := range upcoming {
		call.upcoming = append(call.upcoming, track.VideoID)
	}
	for _, track := range history {
		call.history = append(call.history, track.VideoID)
	}
	extender.requests <- call
	if extender.release != nil {
		<-extender.release
	}
	return extender.result, extender.err
}

func waitFor[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting")
		panic("unreachable")
	}
}

func newAutoplayQueue(extender Extender) (*Queue, chan State) {
	changes := make(chan State, 20)
	queue := New(
		WithRandom(rand.New(rand.NewPCG(1, 2))),
		WithExtender(extender),
		WithOnChange(func(state State) { changes <- state }),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	return queue, changes
}

// waitForChange returns the first state for which matches is true.
func waitForChange(t *testing.T, changes <-chan State, matches func(State) bool) State {
	t.Helper()
	for {
		state := waitFor(t, changes)
		if matches(state) {
			return state
		}
	}
}

func hasAutoplayEntries(state State) bool {
	return slices.ContainsFunc(state.Entries, func(entry Entry) bool { return entry.Autoplay })
}

func TestAutoplayExtendsWhenQueueRunsLow(t *testing.T) {
	extender := newFakeExtender("r1", "r2", "b", "r2")
	queue, changes := newAutoplayQueue(extender)
	queue.Play(tracksNamed("a", "b"), music.ItemKindSong, 0, "search")

	call := waitFor(t, extender.requests)
	if !slices.Equal(call.upcoming, []string{"b"}) || !slices.Equal(call.history, []string{"a"}) {
		t.Fatalf("extender called with %+v", call)
	}
	state := waitForChange(t, changes, hasAutoplayEntries)
	assertOrder(t, state, 0, "a", "b", "r1", "r2")
	if !state.Entries[2].Autoplay || state.Entries[1].Autoplay {
		t.Fatalf("autoplay flags wrong: %+v", state.Entries)
	}
}

func TestAutoplayNotTriggeredWhenDisabledOrRepeating(t *testing.T) {
	extender := newFakeExtender("r1")
	queue, _ := newAutoplayQueue(extender)
	queue.SetAutoplay(false)
	queue.Play(tracksNamed("a"), music.ItemKindSong, 0, "")
	queue.Add(tracksNamed("b"))
	queue.SetRepeat(RepeatAll)
	queue.SetAutoplay(true)
	queue.Add(tracksNamed("c"))
	select {
	case call := <-extender.requests:
		t.Fatalf("unexpected extension: %+v", call)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestAutoplayStaleExtensionIgnoredAfterPlay(t *testing.T) {
	extender := newFakeExtender("stale1", "stale2", "stale3")
	extender.release = make(chan struct{})
	queue, changes := newAutoplayQueue(extender)
	queue.Play(tracksNamed("a"), music.ItemKindSong, 0, "")
	waitFor(t, extender.requests)

	queue.Play(tracksNamed("x", "y", "z", "w"), music.ItemKindSong, 0, "")
	close(extender.release)

	// The stale result must not land; the new queue is long enough to need none.
	time.Sleep(50 * time.Millisecond)
	state := queue.State()
	assertOrder(t, state, 0, "x", "y", "z", "w")
	for len(changes) > 0 {
		if hasAutoplayEntries(<-changes) {
			t.Fatalf("stale extension was applied")
		}
	}
}

func TestAutoplayRetriesForNewQueueAfterStaleResult(t *testing.T) {
	extender := newFakeExtender("r1")
	extender.release = make(chan struct{})
	queue, changes := newAutoplayQueue(extender)
	queue.Play(tracksNamed("a"), music.ItemKindSong, 0, "")
	waitFor(t, extender.requests)
	queue.Play(tracksNamed("x"), music.ItemKindSong, 0, "")
	close(extender.release)

	call := waitFor(t, extender.requests)
	if !slices.Equal(call.history, []string{"x"}) {
		t.Fatalf("second extension was for %+v", call)
	}
	state := waitForChange(t, changes, hasAutoplayEntries)
	assertOrder(t, state, 0, "x", "r1")
}

func TestAutoplayErrorIsLoggedAndQueueUnchanged(t *testing.T) {
	extender := newFakeExtender()
	extender.err = errors.New("offline")
	queue, _ := newAutoplayQueue(extender)
	queue.Play(tracksNamed("a"), music.ItemKindSong, 0, "")
	waitFor(t, extender.requests)
	waitUntilIdle(t, queue)
	queue.Add(tracksNamed("b"))
	waitFor(t, extender.requests)
	assertOrder(t, queue.State(), 0, "a", "b")
}

func TestShuffleRoundTripKeepsLongQueueOrder(t *testing.T) {
	ids := make([]string, 50)
	for position := range ids {
		ids[position] = strconv.Itoa(position)
	}
	queue := newTestQueue()
	queue.Play(tracksNamed(ids...), music.ItemKindSong, 10, "")
	queue.SetShuffle(true)
	if state := queue.SetShuffle(false); !slices.Equal(videoIDs(state), ids) {
		t.Fatalf("round trip changed order")
	}
}

func waitUntilIdle(t *testing.T, queue *Queue) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		queue.mutex.Lock()
		idle := !queue.extending
		queue.mutex.Unlock()
		if idle {
			return
		}
		runtime.Gosched()
	}
	t.Fatalf("extension never finished")
}
