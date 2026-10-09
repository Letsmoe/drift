package download

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/letsmoe/drift/core/music"
)

// fakeCatalog serves one album whose tracks the test can change.
type fakeCatalog struct {
	mutex  sync.Mutex
	tracks []music.Track
}

func (catalog *fakeCatalog) setTracks(tracks []music.Track) {
	catalog.mutex.Lock()
	defer catalog.mutex.Unlock()
	catalog.tracks = tracks
}

func (catalog *fakeCatalog) Album(context.Context, string) (music.AlbumPage, error) {
	catalog.mutex.Lock()
	defer catalog.mutex.Unlock()
	return music.AlbumPage{ID: "album", Title: "Album", Tracks: catalog.tracks}, nil
}

func (catalog *fakeCatalog) Playlist(context.Context, string) (music.PlaylistPage, error) {
	return music.PlaylistPage{}, nil
}

func (catalog *fakeCatalog) PlaylistContinuation(context.Context, string) (music.PlaylistTracks, error) {
	return music.PlaylistTracks{}, nil
}

func fileExists(t *testing.T, manager *Manager, videoID string) bool {
	t.Helper()
	_, err := os.Stat(manager.audioPath(videoID, "audio/webm"))
	return err == nil
}

func createPlaylist(t *testing.T, fixture *fixture, tracks ...music.Track) string {
	t.Helper()
	playlist, err := fixture.library.CreatePlaylist(t.Context(), "Mix")
	if err != nil {
		t.Fatal(err)
	}
	err = fixture.library.AddToPlaylist(t.Context(), playlist.ID, tracks)
	if err != nil {
		t.Fatal(err)
	}
	return playlist.ID
}

func TestKeptLocalPlaylistFollowsEdits(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	playlistID := createPlaylist(t, fixture, fixture.track("aaa"), fixture.track("bbb"))
	ref := CollectionRef{Kind: CollectionLocalPlaylist, ID: playlistID}

	err := manager.DownloadCollection(t.Context(), ref, true)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "aaa", StateDone)
	waitForState(t, manager, "bbb", StateDone)

	err = fixture.library.AddToPlaylist(t.Context(), playlistID, []music.Track{fixture.track("ccc")})
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "ccc", StateDone)

	playlist, err := fixture.library.Playlist(t.Context(), playlistID)
	if err != nil {
		t.Fatal(err)
	}
	err = fixture.library.RemoveFromPlaylist(t.Context(), playlistID, playlist.Entries[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "removed track to be deleted", func() bool { return stateOf(t, manager, "aaa") == "" })
	if fileExists(t, manager, "aaa") {
		t.Error("the file of a removed track should be deleted")
	}
	if stateOf(t, manager, "bbb") != StateDone {
		t.Error("tracks that stay in the playlist keep their download")
	}
}

func TestSharedTracksSurviveUntilNothingNeedsThem(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	first := createPlaylist(t, fixture, fixture.track("aaa"), fixture.track("shared"))
	second := createPlaylist(t, fixture, fixture.track("shared"), fixture.track("ddd"))
	firstRef := CollectionRef{Kind: CollectionLocalPlaylist, ID: first}
	secondRef := CollectionRef{Kind: CollectionLocalPlaylist, ID: second}
	for _, ref := range []CollectionRef{firstRef, secondRef} {
		err := manager.DownloadCollection(t.Context(), ref, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("aaa")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "ddd", StateDone)
	waitForState(t, manager, "aaa", StateDone)
	waitForState(t, manager, "shared", StateDone)

	err = manager.RemoveCollection(t.Context(), firstRef)
	if err != nil {
		t.Fatal(err)
	}

	if stateOf(t, manager, "shared") != StateDone {
		t.Error("a track still in another collection must stay")
	}
	if stateOf(t, manager, "aaa") != StateDone {
		t.Error("an explicitly downloaded track must stay")
	}

	err = manager.RemoveCollection(t.Context(), secondRef)
	if err != nil {
		t.Fatal(err)
	}
	if stateOf(t, manager, "shared") != "" || stateOf(t, manager, "ddd") != "" {
		t.Error("tracks nothing needs anymore should be deleted")
	}
	if stateOf(t, manager, "aaa") != StateDone {
		t.Error("the explicit download must still stay")
	}
}

func TestRefreshingAnAlbumDownloadsNewAndDropsRemovedTracks(t *testing.T) {
	fixture := newFixture(t)
	catalog := &fakeCatalog{tracks: []music.Track{fixture.track("aaa"), fixture.track("bbb")}}
	manager := fixture.startManager(WithCatalog(catalog))
	ref := CollectionRef{Kind: CollectionAlbum, ID: "album"}
	err := manager.DownloadCollection(t.Context(), ref, true)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "bbb", StateDone)

	catalog.setTracks([]music.Track{fixture.track("bbb"), fixture.track("ccc")})
	err = manager.RefreshCollection(t.Context(), ref)
	if err != nil {
		t.Fatal(err)
	}

	waitForState(t, manager, "ccc", StateDone)
	if stateOf(t, manager, "aaa") != "" {
		t.Error("a track that left the album should be deleted")
	}
	collection, found, err := manager.Collection(t.Context(), ref)
	if err != nil || !found {
		t.Fatalf("collection missing (err %v)", err)
	}
	if !collection.Keep || collection.Title != "Album" || collection.TrackCount != 2 {
		t.Errorf("collection = %+v", collection)
	}
}

func TestKeptCollectionsRefreshAtStartup(t *testing.T) {
	fixture := newFixture(t)
	catalog := &fakeCatalog{tracks: []music.Track{fixture.track("aaa")}}
	first := fixture.startManager(WithCatalog(catalog))
	ref := CollectionRef{Kind: CollectionAlbum, ID: "album"}
	err := first.DownloadCollection(t.Context(), ref, true)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, first, "aaa", StateDone)
	first.Close()

	catalog.setTracks([]music.Track{fixture.track("aaa"), fixture.track("bbb")})
	second := fixture.startManager(WithCatalog(catalog))

	waitForState(t, second, "bbb", StateDone)
}

func TestLikedSongsCollectionFollowsRatings(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	err := fixture.library.SetRating(t.Context(), fixture.track("aaa"), music.ItemKindSong, 1)
	if err != nil {
		t.Fatal(err)
	}
	ref := CollectionRef{Kind: CollectionLikes, ID: LikedID}
	err = manager.DownloadCollection(t.Context(), ref, true)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "aaa", StateDone)

	err = fixture.library.SetRating(t.Context(), fixture.track("bbb"), music.ItemKindSong, 1)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "bbb", StateDone)
	err = fixture.library.SetRating(t.Context(), fixture.track("aaa"), music.ItemKindSong, 0)
	if err != nil {
		t.Fatal(err)
	}

	waitFor(t, "unliked track to be deleted", func() bool { return stateOf(t, manager, "aaa") == "" })
}

func TestDeletedPlaylistRemovesItsCollection(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	playlistID := createPlaylist(t, fixture, fixture.track("aaa"))
	ref := CollectionRef{Kind: CollectionLocalPlaylist, ID: playlistID}
	err := manager.DownloadCollection(t.Context(), ref, true)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "aaa", StateDone)

	err = fixture.library.DeletePlaylist(t.Context(), playlistID)
	if err != nil {
		t.Fatal(err)
	}

	waitFor(t, "collection to disappear", func() bool {
		_, found, _ := manager.Collection(t.Context(), ref)
		return !found
	})
	waitFor(t, "track to be deleted", func() bool { return stateOf(t, manager, "aaa") == "" })
}
