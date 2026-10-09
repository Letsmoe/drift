package recommend

import (
	"context"
	"testing"
	"time"

	"github.com/letsmoe/drift-music/core/music"
)

// TestExtendFollowsTheLastTracksRadio checks filtering, ranking and the limit.
func TestExtendFollowsTheLastTracksRadio(t *testing.T) {
	environment := newTestEnvironment(t)
	environment.like(t, makeTrack("fav", "artist:fav", ""), time.Hour)
	last := makeTrack("last", "artist:last", "")
	environment.catalog.radios["last"] = []music.Track{
		makeTrack("neutral", "artist:neutral", ""),
		makeTrack("queued", "artist:queued", ""),
		makeTrack("played", "artist:played", ""),
		makeTrack("hated", "artist:hated", ""),
		makeTrack("favorite", "artist:fav", ""),
		last,
	}
	environment.dislike(t, makeTrack("hated", "artist:hated", ""), time.Hour)

	extension, err := environment.engine.Extend(context.Background(),
		[]music.Track{makeTrack("queued", "artist:queued", "")},
		[]music.Track{makeTrack("played", "artist:played", ""), last})
	if err != nil {
		t.Fatalf("extend: %v", err)
	}
	ids := []string{}
	for _, track := range extension {
		ids = append(ids, track.VideoID)
	}
	if len(ids) != 2 || ids[0] != "favorite" || ids[1] != "neutral" {
		t.Errorf("extension = %v, want [favorite neutral]", ids)
	}
}

// TestExtendCapsAndDiversifies checks the 20 track limit and the artist cap.
func TestExtendCapsAndDiversifies(t *testing.T) {
	environment := newTestEnvironment(t)
	last := makeTrack("last", "artist:last", "")
	radio := []music.Track{}
	for index := range 40 {
		artist := "artist:" + string(rune('a'+index))
		if index < 5 {
			artist = "artist:dominant"
		}
		radio = append(radio, makeTrack("radio"+string(rune('a'+index)), artist, ""))
	}
	environment.catalog.radios["last"] = radio

	extension, err := environment.engine.Extend(context.Background(), nil, []music.Track{last})
	if err != nil {
		t.Fatalf("extend: %v", err)
	}
	if len(extension) != extendLimit {
		t.Fatalf("got %d tracks, want %d", len(extension), extendLimit)
	}
	dominant := 0
	for _, track := range extension {
		if track.Artists[0].ID == "artist:dominant" {
			dominant++
		}
	}
	if dominant > maxPerArtistPerPage {
		t.Errorf("%d tracks of one artist in the extension", dominant)
	}
}

// TestExtendWithoutHistoryOrCatalogReturnsWhatItCan checks the edge cases.
func TestExtendWithoutHistoryOrCatalogReturnsWhatItCan(t *testing.T) {
	environment := newTestEnvironment(t)
	extension, err := environment.engine.Extend(context.Background(), nil, nil)
	if err != nil || len(extension) != 0 {
		t.Errorf("no history: got %v, %v", extension, err)
	}
	environment.catalog.failing["radio"] = true
	_, err = environment.engine.Extend(context.Background(), nil, []music.Track{makeTrack("last", "artist:last", "")})
	if err == nil {
		t.Error("a failing radio must surface as an error so the queue can retry")
	}
}
