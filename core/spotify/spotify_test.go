package spotify

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestParseExportDirectory(t *testing.T) {
	library, err := ParseExport(filepath.Join("testdata", "export"))
	if err != nil {
		t.Fatalf("parse export: %v", err)
	}
	assertFixtureLibrary(t, library)
}

func TestParseExportZip(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "my_spotify_data.zip")
	writeFixtureZip(t, archivePath)
	library, err := ParseExport(archivePath)
	if err != nil {
		t.Fatalf("parse zip: %v", err)
	}
	assertFixtureLibrary(t, library)
}

func TestParseExportSingleFiles(t *testing.T) {
	library, err := ParseExport(filepath.Join("testdata", "export", "Playlist1.json"))
	if err != nil {
		t.Fatalf("parse playlist file: %v", err)
	}
	if len(library.Playlists) != 2 || len(library.LikedTracks) != 0 {
		t.Errorf("playlist file: %d playlists, %d liked", len(library.Playlists), len(library.LikedTracks))
	}
	library, err = ParseExport(filepath.Join("testdata", "export", "YourLibrary.json"))
	if err != nil {
		t.Fatalf("parse library file: %v", err)
	}
	if len(library.LikedTracks) != 3 || len(library.Playlists) != 0 {
		t.Errorf("library file: %d liked, %d playlists", len(library.LikedTracks), len(library.Playlists))
	}
}

func TestParseExportRejectsOtherData(t *testing.T) {
	_, err := ParseExport(filepath.Join("testdata", "export", "StreamingHistory0.json"))
	if err == nil {
		t.Fatal("streaming history parsed as a library")
	}
}

func assertFixtureLibrary(t *testing.T, library Library) {
	t.Helper()
	if len(library.LikedTracks) != 3 {
		t.Fatalf("liked tracks = %d, want 3 (the nameless one is dropped)", len(library.LikedTracks))
	}
	liked := library.LikedTracks[1]
	if liked.Title != "Hold Up" || liked.Artists[0] != "Beyoncé" || liked.Album != "Lemonade" {
		t.Errorf("liked track = %+v", liked)
	}
	if library.LikedTracks[2].Key() != "text:track:garage tape|local band" {
		t.Errorf("uri-less key = %q", library.LikedTracks[2].Key())
	}
	assertFixtureCollections(t, library)
	assertFixturePlaylists(t, library)
}

func assertFixtureCollections(t *testing.T, library Library) {
	t.Helper()
	if len(library.Albums) != 1 || library.Albums[0].Title != "OK Computer" || library.Albums[0].Artist != "Radiohead" {
		t.Errorf("albums = %+v", library.Albums)
	}
	if len(library.Artists) != 1 || library.Artists[0].Name != "Bonobo" {
		t.Errorf("artists = %+v", library.Artists)
	}
	if len(library.Shows) != 1 || library.Shows[0].Publisher != "Jack Rhysider" {
		t.Errorf("shows = %+v", library.Shows)
	}
}

func assertFixturePlaylists(t *testing.T, library Library) {
	t.Helper()
	if len(library.Playlists) != 2 {
		t.Fatalf("playlists = %d, want 2", len(library.Playlists))
	}
	roadTrip := library.Playlists[0]
	if roadTrip.Name != "Road Trip" || roadTrip.Description != "Windows down" || len(roadTrip.Tracks) != 2 {
		t.Errorf("playlist = %+v (the episode item is dropped)", roadTrip)
	}
	if roadTrip.Tracks[0].Title != "Midnight City" || roadTrip.Tracks[1].URI != "spotify:track:0wXuerDYiBnERgIpbb3JBR" {
		t.Errorf("playlist tracks = %+v", roadTrip.Tracks)
	}
	if library.Playlists[0].Key == library.Playlists[1].Key {
		t.Errorf("playlists with the same name share key %q", library.Playlists[0].Key)
	}
}

func writeFixtureZip(t *testing.T, archivePath string) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	writer := zip.NewWriter(file)
	for _, name := range []string{"YourLibrary.json", "Playlist1.json", "StreamingHistory0.json"} {
		contents, err := os.ReadFile(filepath.Join("testdata", "export", name))
		if err != nil {
			t.Fatal(err)
		}
		entry, err := writer.Create("Spotify Account Data/" + name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = entry.Write(contents)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseURL(t *testing.T) {
	cases := map[string]Reference{
		"https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M?si=abc123": {Kind: "playlist", ID: "37i9dQZF1DXcBWIGoYBM5M"},
		"https://open.spotify.com/intl-de/album/4m2880jivSbbyEGAKfITCa":      {Kind: "album", ID: "4m2880jivSbbyEGAKfITCa"},
		"https://open.spotify.com/embed/playlist/37i9dQZF1DXcBWIGoYBM5M":     {Kind: "playlist", ID: "37i9dQZF1DXcBWIGoYBM5M"},
		"spotify:playlist:37i9dQZF1DXcBWIGoYBM5M":                            {Kind: "playlist", ID: "37i9dQZF1DXcBWIGoYBM5M"},
		"spotify:user:someone:playlist:37i9dQZF1DXcBWIGoYBM5M":               {Kind: "playlist", ID: "37i9dQZF1DXcBWIGoYBM5M"},
		"  https://open.spotify.com/album/4m2880jivSbbyEGAKfITCa  ":          {Kind: "album", ID: "4m2880jivSbbyEGAKfITCa"},
	}
	for link, want := range cases {
		got, err := ParseURL(link)
		if err != nil || got != want {
			t.Errorf("ParseURL(%q) = %+v, %v; want %+v", link, got, err, want)
		}
	}
	for _, link := range []string{"", "https://example.com/playlist/abc", "https://open.spotify.com/track/abc", "spotify:track:abc", "not a url"} {
		_, err := ParseURL(link)
		if err == nil {
			t.Errorf("ParseURL(%q) succeeded", link)
		}
	}
}

func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fixtures := map[string]string{
			"/embed/playlist/37i9dQZF1DXcBWIGoYBM5M": "embed_playlist.html",
			"/embed/album/4m2880jivSbbyEGAKfITCa":    "embed_album.html",
		}
		name, known := fixtures[request.URL.Path]
		if !known || request.Header.Get("User-Agent") == "" {
			http.NotFound(writer, request)
			return
		}
		contents, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = writer.Write(contents)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestFetchPlaylist(t *testing.T) {
	server := fixtureServer(t)
	library, err := NewFetcherFor(server.Client(), server.URL).Fetch(context.Background(), "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M")
	if err != nil {
		t.Fatalf("fetch playlist: %v", err)
	}
	if len(library.Playlists) != 1 {
		t.Fatalf("playlists = %d", len(library.Playlists))
	}
	playlist := library.Playlists[0]
	if playlist.Name != "Today’s Top Hits" || playlist.Key != "spotify:playlist:37i9dQZF1DXcBWIGoYBM5M" || len(playlist.Tracks) != 6 {
		t.Errorf("playlist = %q key %q, %d tracks", playlist.Name, playlist.Key, len(playlist.Tracks))
	}
	first := playlist.Tracks[0]
	if first.Title != "Patient Zero" || first.Artists[0] != "Taylor Swift" || first.DurationMillis != 225868 {
		t.Errorf("first track = %+v", first)
	}
}

func TestFetchAlbum(t *testing.T) {
	server := fixtureServer(t)
	library, err := NewFetcherFor(server.Client(), server.URL).Fetch(context.Background(), "spotify:album:4m2880jivSbbyEGAKfITCa")
	if err != nil {
		t.Fatalf("fetch album: %v", err)
	}
	if len(library.Albums) != 1 || library.Albums[0].Title != "Random Access Memories" || library.Albums[0].Artist != "Daft Punk" {
		t.Errorf("albums = %+v", library.Albums)
	}
}

func TestFetchUnavailable(t *testing.T) {
	server := fixtureServer(t)
	_, err := NewFetcherFor(server.Client(), server.URL).Fetch(context.Background(), "spotify:playlist:doesNotExist1")
	if err == nil {
		t.Fatal("missing playlist fetched")
	}
}

func TestParseEmbedPageWithoutData(t *testing.T) {
	_, err := parseEmbedPage([]byte("<html>nothing</html>"))
	if err == nil {
		t.Fatal("page without data parsed")
	}
}

func TestSplitSubtitleArtists(t *testing.T) {
	got := splitSubtitleArtists("KAROL G, Judeline, rusowsky")
	if len(got) != 3 || got[0] != "KAROL G" || got[1] != "Judeline" || got[2] != "rusowsky" {
		t.Errorf("artists = %q", got)
	}
}
