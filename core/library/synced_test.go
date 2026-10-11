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

func TestRestorePlaylistReplacesEntriesAndKeepsTimes(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	var changes []Change
	lib.OnChange(func(change Change) { changes = append(changes, change) })
	for _, videoID := range []string{"a", "b"} {
		if err := lib.SaveTracks(ctx, []music.Track{testTrack(videoID)}, music.ItemKindSong); err != nil {
			t.Fatal(err)
		}
	}
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	restore := func(videoIDs ...string) {
		playlist := Playlist{ID: "p1", Name: "Mix", Description: "d", CreatedAt: createdAt, UpdatedAt: updatedAt}
		for _, videoID := range videoIDs {
			playlist.Entries = append(playlist.Entries, PlaylistEntry{Track: music.Track{VideoID: videoID}, AddedAt: createdAt})
		}
		if err := lib.RestorePlaylist(ctx, playlist); err != nil {
			t.Fatal(err)
		}
	}

	restore("a", "b")
	restore("b", "a", "b")

	playlist, err := lib.Playlist(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	assertVideoIDs(t, []music.Track{playlist.Entries[0].Track, playlist.Entries[1].Track, playlist.Entries[2].Track}, "b", "a", "b")
	if !playlist.CreatedAt.Equal(createdAt) || !playlist.UpdatedAt.Equal(updatedAt) || playlist.Description != "d" {
		t.Errorf("playlist = %+v", playlist)
	}
	if len(changes) != 0 {
		t.Errorf("restoring reported %v", changes)
	}
}

func TestRestoreSavedItemKeepsItsSavingTime(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	savedAt := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	item := music.Item{Kind: music.ItemKindAlbum, ID: "MPREb_1", Title: "Album", Thumbnails: []music.Thumbnail{{URL: "u", Width: 1, Height: 1}}}

	if err := lib.RestoreSavedItem(ctx, SavedItem{Item: item, SavedAt: savedAt}); err != nil {
		t.Fatal(err)
	}
	item.Title = "Album (Deluxe)"
	if err := lib.RestoreSavedItem(ctx, SavedItem{Item: item, SavedAt: savedAt.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	saved, err := lib.SavedItems(ctx)
	if err != nil || len(saved) != 1 || saved[0].Item.Title != "Album (Deluxe)" || !saved[0].SavedAt.Equal(savedAt.Add(time.Hour)) {
		t.Errorf("saved = %+v, %v", saved, err)
	}
}

func TestEveryEditOfASyncedKindReportsAChange(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	var changes []ChangeKind
	lib.OnChange(func(change Change) { changes = append(changes, change.Kind) })

	_ = lib.Save(ctx, music.Item{Kind: music.ItemKindAlbum, ID: "x", Title: "X"})
	_ = lib.Unsave(ctx, music.ItemKindAlbum, "x")
	_ = lib.BlockRecommendation(ctx, BlockArtist, "UC1", "Artist")
	_ = lib.UnblockRecommendation(ctx, BlockArtist, "UC1")
	playlist, _ := lib.CreatePlaylist(ctx, "Mix")
	_ = lib.RenamePlaylist(ctx, playlist.ID, "Mix 2", "")
	playID, _ := lib.StartPlay(ctx, testTrack("a"), music.ItemKindSong, "queue")
	_ = lib.UpdatePlay(ctx, playID, 10, false)
	_ = lib.UpdatePlay(ctx, playID, 200, true)
	_ = lib.RecordSearch(ctx, "daft punk")
	_ = lib.ClearSearches(ctx)

	want := []ChangeKind{
		ChangeSaved, ChangeSaved, ChangeBlocks, ChangeBlocks, ChangePlaylist, ChangePlaylist,
		ChangeHistory, ChangeHistory, ChangeHistory, ChangeHistory,
	}
	if !slices.Equal(changes, want) {
		t.Errorf("changes = %v, want %v", changes, want)
	}
}

func TestRestoredPlaysAndSearchesAreOtherInstallsHistoryAndAddedOnce(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	if err := lib.SaveTracks(ctx, []music.Track{testTrack("a")}, music.ItemKindSong); err != nil {
		t.Fatal(err)
	}
	startedAt := time.Date(2026, 4, 5, 6, 7, 8, 0, time.UTC)
	plays := []PlayRecord{{VideoID: "a", Context: "queue", StartedAt: startedAt, ProgressSeconds: 90, Completed: true}}
	searches := []SearchRecord{{Query: "Daft Punk", SearchedAt: startedAt}}
	for range 2 {
		if err := lib.RestorePlays(ctx, "phone", plays); err != nil {
			t.Fatal(err)
		}
		if err := lib.RestoreSearches(ctx, "phone", searches); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := lib.StartPlay(ctx, testTrack("a"), music.ItemKindSong, "radio"); err != nil {
		t.Fatal(err)
	}

	own, err := lib.OwnPlays(ctx)
	if err != nil || len(own) != 1 || own[0].Context != "radio" {
		t.Errorf("own plays = %v, %v", own, err)
	}
	var total int
	if err = lib.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM play WHERE origin = 'phone' AND duration_seconds = 200 AND completed = 1`).Scan(&total); err != nil || total != 1 {
		t.Errorf("restored plays = %d, %v", total, err)
	}
	if ownSearches, _ := lib.OwnSearches(ctx); len(ownSearches) != 0 {
		t.Errorf("own searches = %v", ownSearches)
	}
	if recent, _ := lib.RecentSearches(ctx, 5); !slices.Equal(recent, []string{"Daft Punk"}) {
		t.Errorf("recent searches = %v", recent)
	}
}

func TestRestoreBlockKeepsItsCreationTime(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)
	createdAt := time.Date(2026, 4, 5, 6, 7, 8, 0, time.UTC)
	if err := lib.RestoreBlock(ctx, RecommendationBlock{Kind: BlockPodcast, ID: "UC1", Name: "Show", CreatedAt: createdAt}); err != nil {
		t.Fatal(err)
	}
	blocks, err := lib.RecommendationBlocks(ctx)
	if err != nil || len(blocks) != 1 || !blocks[0].CreatedAt.Equal(createdAt) || blocks[0].Kind != BlockPodcast {
		t.Errorf("blocks = %+v, %v", blocks, err)
	}
}
