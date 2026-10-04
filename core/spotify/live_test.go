//go:build live

package spotify

import (
	"context"
	"testing"
)

const (
	liveLongPlaylist = "https://open.spotify.com/playlist/37i9dQZF1DX5Ejj0EkURtP" // All Out 2010s, over 100 tracks
	liveAlbum        = "https://open.spotify.com/album/4m2880jivSbbyEGAKfITCa"    // Random Access Memories
)

func TestLiveFetchPlaylist(t *testing.T) {
	library, err := NewFetcher().Fetch(context.Background(), liveLongPlaylist)
	if err != nil {
		t.Fatalf("fetch playlist: %v", err)
	}
	if len(library.Playlists) != 1 || len(library.Playlists[0].Tracks) == 0 {
		t.Fatal("playlist has no tracks; the embed page layout probably changed")
	}
	tracks := library.Playlists[0].Tracks
	t.Logf("%s: %d tracks, first %q by %v (%d ms)", library.Playlists[0].Name, len(tracks), tracks[0].Title, tracks[0].Artists, tracks[0].DurationMillis)
	if len(tracks) > PublicTrackLimit {
		t.Errorf("embed returned %d tracks, more than the assumed limit %d", len(tracks), PublicTrackLimit)
	}
	if tracks[0].DurationMillis == 0 || len(tracks[0].Artists) == 0 {
		t.Errorf("first track lacks duration or artists: %+v", tracks[0])
	}
}

func TestLiveFetchAlbum(t *testing.T) {
	library, err := NewFetcher().Fetch(context.Background(), liveAlbum)
	if err != nil {
		t.Fatalf("fetch album: %v", err)
	}
	if len(library.Albums) != 1 || library.Albums[0].Title == "" || library.Albums[0].Artist == "" {
		t.Errorf("album = %+v", library.Albums)
	}
}
