//go:build live

package innertube

import (
	"os"
	"testing"
)

// TestLiveAccountLibrary reads the library of a real signed-in account. It
// runs only when DRIFT_TEST_COOKIES holds the Cookie header of a
// music.youtube.com session; the cookies are never logged.
func TestLiveAccountLibrary(t *testing.T) {
	cookieHeader := os.Getenv("DRIFT_TEST_COOKIES")
	if cookieHeader == "" {
		t.Skip("DRIFT_TEST_COOKIES is not set")
	}
	client := NewAccountClient(Session{CookieHeader: cookieHeader})

	account, err := client.Account(t.Context())
	if err != nil {
		t.Fatalf("account: %v", err)
	}
	t.Logf("signed in as %q (%s)", account.Name, account.Handle)

	liked, err := client.LikedSongs(t.Context())
	if err != nil {
		t.Fatalf("liked songs: %v", err)
	}
	playlists, err := client.LibraryPlaylists(t.Context())
	if err != nil {
		t.Fatalf("playlists: %v", err)
	}
	albums, err := client.LibraryAlbums(t.Context())
	if err != nil {
		t.Fatalf("albums: %v", err)
	}
	artists, err := client.LibraryArtists(t.Context())
	if err != nil {
		t.Fatalf("artists: %v", err)
	}
	podcasts, err := client.LibraryPodcasts(t.Context())
	if err != nil {
		t.Fatalf("podcasts: %v", err)
	}
	t.Logf("liked %d, playlists %d, albums %d, artists %d, podcasts %d", len(liked), len(playlists), len(albums), len(artists), len(podcasts))
}
