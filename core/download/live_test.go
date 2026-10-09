//go:build live

package download

import (
	"path/filepath"
	"testing"

	"github.com/letsmoe/drift-music/core/innertube"
	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
	"github.com/letsmoe/drift-music/core/stream"
)

func TestLiveDownloadsASearchedTrack(t *testing.T) {
	client := innertube.NewClient()
	tracks, err := client.SearchSongs(t.Context(), "daft punk around the world")
	if err != nil || len(tracks) == 0 {
		t.Fatalf("search: %d tracks, error %v", len(tracks), err)
	}
	root := t.TempDir()
	lib, err := library.Open(t.Context(), filepath.Join(root, "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lib.Close() }()
	manager, err := New(lib, stream.NewResolver(client), filepath.Join(root, "downloads"))
	if err != nil {
		t.Fatal(err)
	}
	err = manager.Start(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	err = manager.DownloadTracks(t.Context(), []music.Track{tracks[0]}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, tracks[0].VideoID, StateDone)

	usage, err := manager.StorageUsage(t.Context())
	if err != nil || usage.Count != 1 || usage.Bytes < 100_000 {
		t.Errorf("usage = %+v (err %v)", usage, err)
	}
}
