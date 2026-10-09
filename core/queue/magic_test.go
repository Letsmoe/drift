package queue

import (
	"context"
	"io"
	"log/slog"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/letsmoe/drift-music/core/music"
)

type mixCall struct {
	seeds   []string
	exclude []string
	count   int
}

type fakeRecommender struct {
	result   []music.Track
	requests chan mixCall
	release  chan struct{}
}

func newFakeRecommender(videoIDs ...string) *fakeRecommender {
	return &fakeRecommender{result: tracksNamed(videoIDs...), requests: make(chan mixCall, 10)}
}

func (recommender *fakeRecommender) Mix(_ context.Context, seeds, exclude []music.Track, count int) ([]music.Track, error) {
	recommender.requests <- mixCall{seeds: videoIDsOf(seeds), exclude: videoIDsOf(exclude), count: count}
	if recommender.release != nil {
		<-recommender.release
	}
	return recommender.result, nil
}

func videoIDsOf(tracks []music.Track) []string {
	ids := make([]string, len(tracks))
	for position, track := range tracks {
		ids[position] = track.VideoID
	}
	return ids
}

func newMagicQueue(recommender Recommender) (*Queue, chan State) {
	changes := make(chan State, 20)
	queue := New(
		WithRandom(rand.New(rand.NewPCG(1, 2))),
		WithRecommender(recommender),
		WithOnChange(func(state State) { changes <- state }),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
	queue.SetAutoplay(false)
	return queue, changes
}

func magicCount(state State) int {
	count := 0
	for _, entry := range state.Entries {
		if entry.Magic {
			count++
		}
	}
	return count
}

// magicPattern renders the upcoming entries as "u" (user) and "M" (magic).
func magicPattern(state State) string {
	pattern := ""
	for _, entry := range state.Entries[state.Index+1:] {
		if entry.Magic {
			pattern += "M"
		} else {
			pattern += "u"
		}
	}
	return pattern
}

func TestMagicShuffleMixesOneRecommendationPerThreeEntries(t *testing.T) {
	recommender := newFakeRecommender("r1", "r2", "r3")
	queue, changes := newMagicQueue(recommender)
	queue.Play(tracksNamed("a", "b", "c", "d", "e", "f", "g", "h"), music.ItemKindSong, 0, "album:x")

	state := queue.SetMagicShuffle(true)
	if !state.Shuffle || !state.MagicShuffle {
		t.Fatalf("shuffle %v, magic %v", state.Shuffle, state.MagicShuffle)
	}
	call := waitFor(t, recommender.requests)
	if call.count != 2 || len(call.seeds) == 0 || len(call.exclude) != 8 {
		t.Fatalf("mix called with %+v", call)
	}
	state = waitForChange(t, changes, func(state State) bool { return magicCount(state) > 0 })
	if pattern := magicPattern(state); pattern != "uuuMuuuMu" {
		t.Errorf("upcoming = %s", pattern)
	}
	if state.Entries[4].Track.VideoID != "r1" || state.Entries[8].Track.VideoID != "r2" {
		t.Errorf("queue = %v", videoIDs(state))
	}
}

func TestMagicShuffleOffRemovesUpcomingRecommendations(t *testing.T) {
	recommender := newFakeRecommender("r1", "r2")
	queue, changes := newMagicQueue(recommender)
	queue.Play(tracksNamed("a", "b", "c", "d", "e", "f", "g"), music.ItemKindSong, 0, "")
	queue.SetMagicShuffle(true)
	waitForChange(t, changes, func(state State) bool { return magicCount(state) == 2 })

	// Play up to and past the first recommendation, so it is history.
	for range 5 {
		queue.Next()
	}
	state := queue.SetMagicShuffle(false)
	if state.MagicShuffle || !state.Shuffle {
		t.Fatalf("magic %v, shuffle %v", state.MagicShuffle, state.Shuffle)
	}
	if magicCount(state) != 1 || magicPattern(state) != "uu" {
		t.Errorf("queue = %v at %d", videoIDs(state), state.Index)
	}
}

func TestShuffleOffEndsMagicShuffle(t *testing.T) {
	recommender := newFakeRecommender("r1", "r2")
	queue, changes := newMagicQueue(recommender)
	queue.Play(tracksNamed("a", "b", "c", "d", "e", "f", "g"), music.ItemKindSong, 0, "")
	queue.SetMagicShuffle(true)
	waitForChange(t, changes, func(state State) bool { return magicCount(state) == 2 })

	state := queue.SetShuffle(false)
	if state.MagicShuffle || magicCount(state) != 0 {
		t.Fatalf("magic %v, entries %v", state.MagicShuffle, videoIDs(state))
	}
	assertOrder(t, state, 0, "a", "b", "c", "d", "e", "f", "g")
}

func TestMagicShuffleSkipsQueuedTracks(t *testing.T) {
	recommender := newFakeRecommender("b", "r1")
	queue, changes := newMagicQueue(recommender)
	queue.Play(tracksNamed("a", "b", "c", "d"), music.ItemKindSong, 0, "")
	queue.SetMagicShuffle(true)

	state := waitForChange(t, changes, func(state State) bool { return magicCount(state) > 0 })
	if magicCount(state) != 1 || !slices.Contains(videoIDs(state), "r1") {
		t.Errorf("queue = %v", videoIDs(state))
	}
}

func TestMagicShuffleMixesNewlyAddedTracks(t *testing.T) {
	recommender := newFakeRecommender("r1")
	queue, changes := newMagicQueue(recommender)
	queue.Play(tracksNamed("a", "b"), music.ItemKindSong, 0, "")
	queue.SetMagicShuffle(true)
	select {
	case call := <-recommender.requests:
		t.Fatalf("mixed a queue too short for recommendations: %+v", call)
	case <-time.After(50 * time.Millisecond):
	}

	queue.Add(tracksNamed("c", "d"))
	if call := waitFor(t, recommender.requests); call.count != 1 {
		t.Fatalf("mix called with %+v", call)
	}
	waitForChange(t, changes, func(state State) bool { return magicCount(state) == 1 })
}

func TestMagicShuffleDropsMixForReplacedQueue(t *testing.T) {
	recommender := newFakeRecommender("r1")
	recommender.release = make(chan struct{})
	queue, changes := newMagicQueue(recommender)
	queue.Play(tracksNamed("a", "b", "c", "d"), music.ItemKindSong, 0, "")
	queue.SetMagicShuffle(true)
	waitFor(t, recommender.requests)

	queue.Play(tracksNamed("x", "y", "z", "w"), music.ItemKindSong, 0, "")
	recommender.release <- struct{}{}
	// The replaced queue asks again once the stale mix finished.
	waitFor(t, recommender.requests)
	recommender.release <- struct{}{}

	state := waitForChange(t, changes, func(state State) bool { return magicCount(state) == 1 })
	if state.Entries[0].Track.VideoID != "x" || len(state.Entries) != 5 {
		t.Errorf("queue = %v", videoIDs(state))
	}
}
