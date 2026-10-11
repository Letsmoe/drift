package neoworks

import (
	"slices"
	"testing"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

func testAlbum(browseID string) music.Item {
	return music.Item{
		Kind:     music.ItemKindAlbum,
		ID:       browseID,
		Title:    "Album " + browseID,
		Subtitle: "Album • Artist • 2024",
		Thumbnails: []music.Thumbnail{
			{URL: "https://lh3.googleusercontent.com/" + browseID + "=w226-h226-l90-rj", Width: 226, Height: 226},
		},
	}
}

func TestASavedAlbumReachesAnotherInstallAndUnsavingRemovesIt(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	ctx := t.Context()
	if err := laptop.library.Save(ctx, testAlbum("MPREb_1")); err != nil {
		t.Fatal(err)
	}
	laptop.reconcile(t)

	nodes := store.liveItems("saved")
	if len(nodes) != 1 || nodes[0]["kind"] != "album" || nodes[0]["browseId"] != "MPREb_1" ||
		nodes[0]["thumbnailUrl"] != "https://lh3.googleusercontent.com/MPREb_1=w544-h544-l90-rj" {
		t.Fatalf("saved nodes = %v", nodes)
	}
	phone.reconcile(t)
	saved, err := phone.library.SavedItems(ctx)
	if err != nil || len(saved) != 1 || saved[0].Item.Title != "Album MPREb_1" || len(saved[0].Item.Thumbnails) != 4 {
		t.Fatalf("phone saved = %+v, %v", saved, err)
	}
	if !saved[0].SavedAt.Equal(*laptop.clock) {
		t.Errorf("saved time = %v, want the laptop's %v", saved[0].SavedAt, *laptop.clock)
	}
	if report := phone.reconcile(t); report.NodesWritten != 0 || report.LibraryWritten != 0 {
		t.Errorf("phone must settle, wrote %+v", report)
	}

	_ = laptop.library.Unsave(ctx, music.ItemKindAlbum, "MPREb_1")
	laptop.reconcile(t)
	phone.reconcile(t)
	if saved, _ = phone.library.SavedItems(ctx); len(saved) != 0 {
		t.Errorf("phone still has %+v", saved)
	}
}

func TestSavedItemsOfTwoKindsWithTheSameIDStayApart(t *testing.T) {
	store := newFakeNodeStore()
	install := newTestInstall(t, store)
	ctx := t.Context()
	album := testAlbum("same")
	artist := music.Item{Kind: music.ItemKindArtist, ID: "same", Title: "Artist"}
	_ = install.library.Save(ctx, album)
	_ = install.library.Save(ctx, artist)

	install.reconcile(t)

	if nodes := store.liveItems("saved"); len(nodes) != 2 {
		t.Errorf("saved nodes = %v", nodes)
	}
	if nodes := store.liveItems("saved"); slices.ContainsFunc(nodes, func(node map[string]any) bool { _, has := node["thumbnailUrl"]; return has && node["kind"] == "artist" }) {
		t.Error("an artist without a cover must not carry a thumbnailUrl")
	}
}

func orderedPlaylist(t *testing.T, install *testInstall, name string, videoIDs ...string) library.Playlist {
	t.Helper()
	ctx := t.Context()
	playlist, err := install.library.CreatePlaylist(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	tracks := []music.Track{}
	for _, videoID := range videoIDs {
		tracks = append(tracks, syncedTestTrack(videoID))
	}
	if err = install.library.AddToPlaylist(ctx, playlist.ID, tracks); err != nil {
		t.Fatal(err)
	}
	return playlist
}

func playlistVideoIDs(t *testing.T, install *testInstall, playlistID string) []string {
	t.Helper()
	playlist, err := install.library.Playlist(t.Context(), playlistID)
	if err != nil {
		t.Fatal(err)
	}
	videoIDs := []string{}
	for _, entry := range playlist.Entries {
		videoIDs = append(videoIDs, entry.Track.VideoID)
	}
	return videoIDs
}

func TestAPlaylistKeepsItsOrderOnAnotherInstall(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	playlist := orderedPlaylist(t, laptop, "Road trip", "c", "a", "b")
	laptop.reconcile(t)

	nodes := store.liveItems("playlists")
	if len(nodes) != 1 || nodes[0]["name"] != "Road trip" {
		t.Fatalf("playlist nodes = %v", nodes)
	}
	entries, _ := nodes[0]["entries"].([]any)
	if len(entries) != 3 || entries[0].(map[string]any)["videoId"] != "c" || entries[2].(map[string]any)["videoId"] != "b" {
		t.Fatalf("entries = %v", entries)
	}
	if len(store.liveItems("tracks")) != 3 {
		t.Errorf("track nodes = %d, want one per entry", len(store.liveItems("tracks")))
	}

	phone.reconcile(t)
	summaries, _ := phone.library.Playlists(t.Context())
	if len(summaries) != 1 {
		t.Fatalf("phone playlists = %v", summaries)
	}
	if got := playlistVideoIDs(t, phone, summaries[0].ID); !slices.Equal(got, []string{"c", "a", "b"}) {
		t.Errorf("phone order = %v", got)
	}
	if again := phone.reconcile(t); again.NodesWritten != 0 || again.LibraryWritten != 0 {
		t.Errorf("phone must settle, wrote %+v", again)
	}

	laptop.advance(time.Hour)
	entriesBefore, _ := laptop.library.Playlist(t.Context(), playlist.ID)
	_ = laptop.library.MovePlaylistEntry(t.Context(), playlist.ID, entriesBefore.Entries[2].ID, 0)
	_ = laptop.library.RenamePlaylist(t.Context(), playlist.ID, "Road trip 2", "Long drive")
	laptop.reconcile(t)
	phone.reconcile(t)
	renamed, _ := phone.library.Playlist(t.Context(), summaries[0].ID)
	if renamed.Name != "Road trip 2" || renamed.Description != "Long drive" {
		t.Errorf("phone playlist = %+v", renamed)
	}
	if got := playlistVideoIDs(t, phone, summaries[0].ID); !slices.Equal(got, []string{"b", "c", "a"}) {
		t.Errorf("phone order after move = %v", got)
	}
}

func TestEmptyingAPlaylistClearsItsEntriesAndDescriptionOnTheNode(t *testing.T) {
	store := newFakeNodeStore()
	install := newTestInstall(t, store)
	playlist := orderedPlaylist(t, install, "Mix", "a")
	_ = install.library.RenamePlaylist(t.Context(), playlist.ID, "Mix", "About")
	install.reconcile(t)

	install.advance(time.Minute)
	current, _ := install.library.Playlist(t.Context(), playlist.ID)
	_ = install.library.RemoveFromPlaylist(t.Context(), playlist.ID, current.Entries[0].ID)
	_ = install.library.RenamePlaylist(t.Context(), playlist.ID, "Mix", "")
	install.reconcile(t)

	nodes := store.liveItems("playlists")
	if _, has := nodes[0]["entries"]; has {
		t.Errorf("entries stayed on the node: %v", nodes[0])
	}
	if _, has := nodes[0]["description"]; has {
		t.Errorf("description stayed on the node: %v", nodes[0])
	}
}

func TestDeletingAPlaylistDeletesItEverywhere(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	playlist := orderedPlaylist(t, laptop, "Mix", "a")
	laptop.reconcile(t)
	phone.reconcile(t)

	_ = laptop.library.DeletePlaylist(t.Context(), playlist.ID)
	laptop.reconcile(t)
	phone.reconcile(t)

	if summaries, _ := phone.library.Playlists(t.Context()); len(summaries) != 0 {
		t.Errorf("phone playlists = %v", summaries)
	}
	if len(store.liveItems("playlists")) != 0 {
		t.Errorf("playlist nodes remain: %v", store.liveItems("playlists"))
	}
}

func TestAPlaylistWaitsForTracksThatHaveNotArrived(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	orderedPlaylist(t, laptop, "Mix", "a")
	laptop.reconcile(t)
	for _, node := range store.nodes {
		if node.summary.Collection == "@drift/tracks" {
			node.summary.Deleted = true
		}
	}

	if _, err := phone.sync.Reconcile(t.Context()); err == nil {
		t.Fatal("applying a playlist whose track is missing must report it")
	}
	if summaries, _ := phone.library.Playlists(t.Context()); len(summaries) != 0 {
		t.Errorf("phone playlists = %v", summaries)
	}
}

func TestTwoOfflineSavesOfOneAlbumMergeIntoTheNewest(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	ctx := t.Context()
	_ = laptop.library.Save(ctx, testAlbum("MPREb_1"))
	laptop.reconcile(t)
	phone.nodes.goOffline()
	phone.advance(time.Hour)
	renamed := testAlbum("MPREb_1")
	renamed.Title = "Album (Deluxe)"
	_ = phone.library.Save(ctx, renamed)
	phone.reconcile(t)
	phone.nodes.goOnline()
	if len(store.liveItems("saved")) != 2 {
		t.Fatalf("setup must leave two nodes, got %v", store.liveItems("saved"))
	}

	phone.reconcile(t)
	laptop.reconcile(t)

	nodes := store.liveItems("saved")
	if len(nodes) != 1 || nodes[0]["title"] != "Album (Deluxe)" {
		t.Fatalf("saved nodes = %v, want the phone's newer one", nodes)
	}
	saved, _ := laptop.library.SavedItems(ctx)
	if len(saved) != 1 || saved[0].Item.Title != "Album (Deluxe)" {
		t.Errorf("laptop saved = %+v", saved)
	}
}
