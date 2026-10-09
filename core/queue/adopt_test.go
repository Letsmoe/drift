package queue

import (
	"reflect"
	"testing"

	"github.com/letsmoe/drift/core/music"
)

func TestAdoptTakesTheStateAsItIs(t *testing.T) {
	source := newTestQueue()
	source.SetRepeat(RepeatAll)
	source.Play(tracksNamed("a", "b", "c"), music.ItemKindSong, 1, "album:x")
	sourceState := source.SetAutoplay(false)

	var notified State
	target := New(WithOnChange(func(state State) { notified = state }))
	adopted := target.Adopt(sourceState)

	if !reflect.DeepEqual(adopted, sourceState) {
		t.Fatalf("adopted %+v, want %+v", adopted, sourceState)
	}
	if !reflect.DeepEqual(notified, sourceState) {
		t.Fatalf("onChange got %+v", notified)
	}
}

func TestAdoptKeepsNewEntryIDsUnique(t *testing.T) {
	source := newTestQueue()
	sourceState := source.Play(tracksNamed("a", "b", "c"), music.ItemKindSong, 0, "")

	target := newTestQueue()
	target.Play(tracksNamed("x"), music.ItemKindSong, 0, "")
	target.Adopt(sourceState)
	state := target.Add(tracksNamed("d"))

	seen := map[int64]bool{}
	for _, entry := range state.Entries {
		if seen[entry.ID] {
			t.Fatalf("entry ID %d used twice in %+v", entry.ID, state.Entries)
		}
		seen[entry.ID] = true
	}
}

func TestAdoptEmptyQueue(t *testing.T) {
	target := newTestQueue()
	target.Play(tracksNamed("a"), music.ItemKindSong, 0, "")
	state := target.Adopt(State{Index: 3, Repeat: "bogus"})
	if state.Index != -1 || len(state.Entries) != 0 || state.Repeat != RepeatOff {
		t.Fatalf("adopted empty queue = %+v", state)
	}
}
