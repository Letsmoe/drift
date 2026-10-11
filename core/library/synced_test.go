package library

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/letsmoe/drift/core/music"
)

func TestNodeLinksAreStoredPerCollectionAndReplacedByNode(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	first := NodeLink{Collection: "ratings", Key: "a", NodeID: "node-1", ContentHash: "h1", NodeUpdatedAt: "t1"}
	other := NodeLink{Collection: "saved", Key: "a", NodeID: "node-2", ContentHash: "h2", NodeUpdatedAt: "t2"}
	for _, link := range []NodeLink{first, other} {
		if err := lib.SaveNodeLink(ctx, link); err != nil {
			t.Fatal(err)
		}
	}
	links, err := lib.NodeLinks(ctx, "ratings")
	if err != nil || len(links) != 1 || links["a"] != first {
		t.Fatalf("ratings links = %v, %v", links, err)
	}

	first.ContentHash = "h3"
	if err = lib.SaveNodeLink(ctx, first); err != nil {
		t.Fatal(err)
	}
	moved := NodeLink{Collection: "ratings", Key: "b", NodeID: "node-1", ContentHash: "h4"}
	if err = lib.SaveNodeLink(ctx, moved); err != nil {
		t.Fatal(err)
	}
	links, _ = lib.NodeLinks(ctx, "ratings")
	if len(links) != 1 || links["b"] != moved {
		t.Errorf("a node belongs to one entity, links = %v", links)
	}

	if err = lib.DeleteNodeLink(ctx, "ratings", "b"); err != nil {
		t.Fatal(err)
	}
	if err = lib.ClearNodeLinks(ctx); err != nil {
		t.Fatal(err)
	}
	for _, collection := range []string{"ratings", "saved"} {
		if links, _ = lib.NodeLinks(ctx, collection); len(links) != 0 {
			t.Errorf("%s links after clearing = %v", collection, links)
		}
	}
}

func TestOnChangeSupportsSeveralListenersAndQuietWrites(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	var first, second []Change
	lib.OnChange(func(change Change) { first = append(first, change) })
	lib.OnChange(func(change Change) { second = append(second, change) })

	if err := lib.SetRating(ctx, testTrack("a"), music.ItemKindSong, RatingLike); err != nil {
		t.Fatal(err)
	}
	quiet := WithoutNotifications(ctx)
	if err := lib.SetRating(quiet, testTrack("b"), music.ItemKindSong, RatingLike); err != nil {
		t.Fatal(err)
	}
	want := []Change{{Kind: ChangeLikes}}
	if !slices.Equal(first, want) || !slices.Equal(second, want) {
		t.Errorf("listeners saw %v and %v, want %v each", first, second, want)
	}
}

func TestRestoreRatingKeepsTheGivenTimeAndReportsNothing(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	var changes []Change
	lib.OnChange(func(change Change) { changes = append(changes, change) })
	if err := lib.SaveTracks(ctx, []music.Track{testTrack("a")}, music.ItemKindSong); err != nil {
		t.Fatal(err)
	}
	moment := time.Date(2026, 5, 6, 7, 8, 9, 123_000_000, time.UTC)

	if err := lib.RestoreRating(ctx, RatingRecord{VideoID: "a", Rating: RatingDislike, UpdatedAt: moment}); err != nil {
		t.Fatal(err)
	}
	records, err := lib.RatingRecords(ctx)
	if err != nil || len(records) != 1 || records[0].Rating != RatingDislike || !records[0].UpdatedAt.Equal(moment) {
		t.Fatalf("records = %v, %v", records, err)
	}
	if err = lib.RestoreRating(ctx, RatingRecord{VideoID: "a", Rating: RatingNone}); err != nil {
		t.Fatal(err)
	}
	if records, _ = lib.RatingRecords(ctx); len(records) != 0 {
		t.Errorf("records after removal = %v", records)
	}
	if len(changes) != 0 {
		t.Errorf("restoring reported %v", changes)
	}
}

func TestReferencedTracksAreThoseARatingPlaylistPlayOrBlockNames(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	for _, videoID := range []string{"rated", "listed", "played", "blocked", "lonely"} {
		if err := lib.SaveTracks(ctx, []music.Track{testTrack(videoID)}, music.ItemKindSong); err != nil {
			t.Fatal(err)
		}
	}
	_ = lib.SetRating(ctx, testTrack("rated"), music.ItemKindSong, RatingLike)
	playlist, _ := lib.CreatePlaylist(ctx, "Mix")
	_ = lib.AddToPlaylist(ctx, playlist.ID, []music.Track{testTrack("listed")})
	_, _ = lib.StartPlay(ctx, testTrack("played"), music.ItemKindSong, "queue")
	_ = lib.BlockRecommendation(ctx, BlockTrack, "blocked", "Blocked")

	records, err := lib.ReferencedTracks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var videoIDs []string
	for _, record := range records {
		videoIDs = append(videoIDs, record.Track.VideoID)
	}
	want := []string{"blocked", "listed", "played", "rated"}
	if !slices.Equal(videoIDs, want) {
		t.Errorf("referenced = %v, want %v", videoIDs, want)
	}
}
