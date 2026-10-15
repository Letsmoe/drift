package spotify

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

var exportifyFixtureDirectory = filepath.Join("testdata", "exportify")

func TestParseExportifyDirectory(t *testing.T) {
	library, err := ParseExport(exportifyFixtureDirectory)
	if err != nil {
		t.Fatalf("parse folder: %v", err)
	}
	assertExportifyLibrary(t, library)
}

func TestParseExportifyZip(t *testing.T) {
	archivePath := filepath.Join(t.TempDir(), "spotify_playlists.zip")
	writeExportifyZip(t, archivePath)
	library, err := ParseExport(archivePath)
	if err != nil {
		t.Fatalf("parse zip: %v", err)
	}
	assertExportifyLibrary(t, library)
}

func TestParseExportifySingleFile(t *testing.T) {
	library, err := ParseExport(filepath.Join(exportifyFixtureDirectory, "Road_Trip.csv"))
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(library.Playlists) != 1 || len(library.LikedTracks) != 0 || len(library.Albums) != 0 {
		t.Fatalf("library = %+v", library)
	}
	if library.Playlists[0].Name != "Road Trip" {
		t.Errorf("playlist name = %q", library.Playlists[0].Name)
	}
}

func TestParseExportifyIgnoresOtherCSV(t *testing.T) {
	_, err := ParseExport(filepath.Join(exportifyFixtureDirectory, "contacts.csv"))
	if !errors.Is(err, ErrNoLibraryData) {
		t.Fatalf("error = %v, want ErrNoLibraryData", err)
	}
}

func TestParseExportifyByteOrderMarkAndColumnOrder(t *testing.T) {
	contents := "\xEF\xBB\xBFArtist Name(s),track name , Track URI\nA;B,Song,spotify:track:x\n,  ,spotify:track:y\n"
	path := filepath.Join(t.TempDir(), "Mixed_Up.csv")
	err := os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	library, err := ParseExport(path)
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	tracks := library.Playlists[0].Tracks
	if len(tracks) != 1 || tracks[0].URI != "spotify:track:x" || !slices.Equal(tracks[0].Artists, []string{"A", "B"}) {
		t.Errorf("tracks = %+v", tracks)
	}
	if tracks[0].DurationMillis != 0 || tracks[0].Album != "" {
		t.Errorf("absent columns should stay empty: %+v", tracks[0])
	}
}

func TestParseExportifyKeepsJSONKeys(t *testing.T) {
	jsonOnly, err := ParseExport(filepath.Join("testdata", "export", "Playlist1.json"))
	if err != nil {
		t.Fatalf("parse json: %v", err)
	}
	directory := t.TempDir()
	copyFixtureFile(t, filepath.Join("testdata", "export", "Playlist1.json"), filepath.Join(directory, "Playlist1.json"))
	copyFixtureFile(t, filepath.Join(exportifyFixtureDirectory, "Road_Trip.csv"), filepath.Join(directory, "Road_Trip.csv"))
	mixed, err := ParseExport(directory)
	if err != nil {
		t.Fatalf("parse folder: %v", err)
	}
	keys := []string{}
	for _, playlist := range mixed.Playlists {
		keys = append(keys, playlist.Key)
	}
	want := []string{jsonOnly.Playlists[0].Key, jsonOnly.Playlists[1].Key, "exportify:playlist:Road_Trip"}
	if !slices.Equal(keys, want) {
		t.Errorf("keys = %q, want %q", keys, want)
	}
}

func assertExportifyLibrary(t *testing.T, library Library) {
	t.Helper()
	assertExportifyLiked(t, library)
	assertExportifyAlbums(t, library)
	assertExportifyPlaylists(t, library)
}

func assertExportifyLiked(t *testing.T, library Library) {
	t.Helper()
	if len(library.LikedTracks) != 2 {
		t.Fatalf("liked tracks = %d, want 2", len(library.LikedTracks))
	}
	liked := library.LikedTracks[0]
	if liked.Title != "Mr. Brightside" || liked.DurationMillis != 222973 || liked.Album != "Hot Fuss" {
		t.Errorf("liked track = %+v", liked)
	}
}

func assertExportifyAlbums(t *testing.T, library Library) {
	t.Helper()
	if len(library.Albums) != 2 {
		t.Fatalf("albums = %+v, want 2", library.Albums)
	}
	if library.Albums[0].Title != "OK Computer" || library.Albums[0].Artist != "Radiohead" {
		t.Errorf("first album = %+v", library.Albums[0])
	}
	if library.Albums[1].Title != "Mezzanine" || library.Albums[1].Artist != "Massive Attack" {
		t.Errorf("second album = %+v", library.Albums[1])
	}
	if library.Albums[0].Key() != "text:album:ok computer|radiohead" {
		t.Errorf("album key = %q", library.Albums[0].Key())
	}
}

func assertExportifyPlaylists(t *testing.T, library Library) {
	t.Helper()
	playlists := map[string]Playlist{}
	for _, playlist := range library.Playlists {
		playlists[playlist.Key] = playlist
	}
	if len(playlists) != 4 {
		t.Fatalf("playlists = %d, want 4 (special files and other csv excluded)", len(library.Playlists))
	}
	assertRoadTrip(t, playlists["exportify:playlist:Road_Trip"])
	assertExportifyNames(t, playlists)
}

func assertRoadTrip(t *testing.T, roadTrip Playlist) {
	t.Helper()
	if roadTrip.Name != "Road Trip" || len(roadTrip.Tracks) != 2 {
		t.Fatalf("road trip = %+v (the nameless row is dropped)", roadTrip)
	}
	first := roadTrip.Tracks[0]
	if first.URI != "spotify:track:4cOdK2wGLETKBW3PvgPWqT" || first.DurationMillis != 213573 {
		t.Errorf("first track = %+v", first)
	}
	second := roadTrip.Tracks[1]
	if second.Title != "Hello, World (Live; Remastered)" || second.Album != "Songs, Vol. 1" {
		t.Errorf("quoted fields = %+v", second)
	}
	if !slices.Equal(second.Artists, []string{"Duo A", "Band / Crew"}) {
		t.Errorf("artists = %q", second.Artists)
	}
}

func assertExportifyNames(t *testing.T, playlists map[string]Playlist) {
	t.Helper()
	empty, found := playlists["exportify:playlist:Empty_One"]
	if !found || empty.Name != "Empty One" || len(empty.Tracks) != 0 || empty.Tracks == nil {
		t.Errorf("empty playlist = %+v (kept, with no tracks)", empty)
	}
	plain, plainFound := playlists["exportify:playlist:🩷"]
	underscored, underscoredFound := playlists["exportify:playlist:🩷_"]
	if !plainFound || !underscoredFound {
		t.Fatalf("emoji playlists missing: %v", playlists)
	}
	if plain.Name != "🩷" || underscored.Name != "🩷" {
		t.Errorf("names = %q, %q, want both 🩷", plain.Name, underscored.Name)
	}
}

func writeExportifyZip(t *testing.T, archivePath string) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	writer := zip.NewWriter(file)
	entries, err := os.ReadDir(exportifyFixtureDirectory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		contents, err := os.ReadFile(filepath.Join(exportifyFixtureDirectory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		zipEntry, err := writer.Create(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		_, err = zipEntry.Write(contents)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func copyFixtureFile(t *testing.T, source, destination string) {
	t.Helper()
	contents, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(destination, contents, 0o600)
	if err != nil {
		t.Fatal(err)
	}
}
