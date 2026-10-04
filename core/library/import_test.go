package library

import (
	"context"
	"testing"

	"github.com/neoworks/otter-music/core/music"
)

func TestImportMatchRoundTrip(t *testing.T) {
	library, _ := openTestLibrary(t)
	ctx := context.Background()
	_, found, err := library.ImportMatch(ctx, "spotify", "spotify:track:1")
	if err != nil || found {
		t.Fatalf("unknown match: found=%v err=%v", found, err)
	}
	stored := ImportMatch{Source: "spotify", SourceURI: "spotify:track:1", Kind: music.ItemKindSong, TargetID: "vid", Score: 0.9, Payload: `{"videoId":"vid"}`}
	if err = library.StoreImportMatch(ctx, stored); err != nil {
		t.Fatalf("store: %v", err)
	}
	stored.TargetID = ""
	stored.SourceURI = "spotify:track:2"
	if err = library.StoreImportMatch(ctx, stored); err != nil {
		t.Fatalf("store no-match: %v", err)
	}
	match, found, err := library.ImportMatch(ctx, "spotify", "spotify:track:1")
	if err != nil || !found || match.TargetID != "vid" || match.Score != 0.9 || match.Payload != `{"videoId":"vid"}` {
		t.Errorf("match = %+v found=%v err=%v", match, found, err)
	}
	match, found, err = library.ImportMatch(ctx, "spotify", "spotify:track:2")
	if err != nil || !found || match.TargetID != "" {
		t.Errorf("no-match = %+v found=%v err=%v", match, found, err)
	}
}

func TestImportPlaylistReplacesEntries(t *testing.T) {
	library, _ := openTestLibrary(t)
	ctx := context.Background()
	imported := ImportedPlaylist{Source: "spotify", SourceKey: "key", Name: "Mix", Description: "first", Tracks: []music.Track{testTrack("a"), testTrack("b")}}
	firstID, err := library.ImportPlaylist(ctx, imported)
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	imported.Name = "Mix renamed"
	imported.Tracks = []music.Track{testTrack("b"), testTrack("c")}
	secondID, err := library.ImportPlaylist(ctx, imported)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if secondID != firstID {
		t.Fatalf("re-import made a new playlist %s, first was %s", secondID, firstID)
	}
	playlist, err := library.Playlist(ctx, firstID)
	if err != nil {
		t.Fatal(err)
	}
	if playlist.Name != "Mix renamed" || len(playlist.Entries) != 2 || playlist.Entries[0].Track.VideoID != "b" || playlist.Entries[1].Track.VideoID != "c" {
		t.Errorf("playlist = %+v", playlist)
	}
}

func TestImportPlaylistRecreatedAfterDeletion(t *testing.T) {
	library, _ := openTestLibrary(t)
	ctx := context.Background()
	imported := ImportedPlaylist{Source: "spotify", SourceKey: "key", Name: "Mix", Tracks: []music.Track{testTrack("a")}}
	firstID, err := library.ImportPlaylist(ctx, imported)
	if err != nil {
		t.Fatal(err)
	}
	if err = library.DeletePlaylist(ctx, firstID); err != nil {
		t.Fatal(err)
	}
	secondID, err := library.ImportPlaylist(ctx, imported)
	if err != nil || secondID == firstID {
		t.Fatalf("re-import after deletion: id %s (first %s), err %v", secondID, firstID, err)
	}
}
