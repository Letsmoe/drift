package recommend

import (
	"slices"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

func TestMixMergesSeedRadiosWithoutQueuedOrDislikedTracks(t *testing.T) {
	environment := newTestEnvironment(t)
	first := makeTrack("first", "artist:first", "")
	second := makeTrack("second", "artist:second", "")
	environment.catalog.radios["first"] = []music.Track{
		first,
		makeTrack("fromFirst", "artist:one", ""),
		makeTrack("queued", "artist:queued", ""),
	}
	environment.catalog.radios["second"] = []music.Track{
		second,
		makeTrack("fromSecond", "artist:two", ""),
		makeTrack("hated", "artist:hated", ""),
	}
	environment.dislike(t, makeTrack("hated", "artist:hated", ""), time.Hour)

	mixed, err := environment.engine.Mix(t.Context(), []music.Track{first, second},
		[]music.Track{first, second, makeTrack("queued", "artist:queued", "")}, 5)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, track := range mixed {
		ids = append(ids, track.VideoID)
	}
	slices.Sort(ids)
	if !slices.Equal(ids, []string{"fromFirst", "fromSecond"}) {
		t.Errorf("mix = %v", ids)
	}

	mixed, err = environment.engine.Mix(t.Context(), []music.Track{first, second}, nil, 1)
	if err != nil || len(mixed) != 1 {
		t.Errorf("count 1: %d tracks, %v", len(mixed), err)
	}
}

func TestMixFailsOnlyWhenEveryRadioFails(t *testing.T) {
	environment := newTestEnvironment(t)
	environment.catalog.failing["radio"] = true
	if _, err := environment.engine.Mix(t.Context(), []music.Track{makeTrack("seed", "artist:seed", "")}, nil, 3); err == nil {
		t.Error("expected an error when no seed radio loads")
	}
}
