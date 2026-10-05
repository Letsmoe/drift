package innertube

import (
	"reflect"
	"testing"

	"github.com/neoworks/otter-music/core/music"
)

func TestAlbum(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/album.json")

	page, err := server.client().Album(t.Context(), "MPREb_BQHrCwSBM37")
	if err != nil {
		t.Fatal(err)
	}

	if server.lastRequest(t, "browse").body["browseId"] != "MPREb_BQHrCwSBM37" {
		t.Errorf("unexpected request body: %v", server.lastRequest(t, "browse").body)
	}
	if page.Title != "TRON: Legacy" || page.Year != "2010" || page.Subtitle != "Album • 2010" {
		t.Errorf("page = %+v", page)
	}
	if page.PlaylistID != "OLAK5uy_k8XkxQFGZhSX8gENqaH6pg1xz8vOTnzQs" {
		t.Errorf("playlistId = %q", page.PlaylistID)
	}
	wantArtists := []music.Artist{{ID: "UCRr1xG_2WIDs18a6cIiCxeA", Name: "Daft Punk"}}
	if !reflect.DeepEqual(page.Artists, wantArtists) {
		t.Errorf("artists = %+v", page.Artists)
	}
	if len(page.Thumbnails) == 0 || len(page.Tracks) != 3 {
		t.Fatalf("thumbnails = %d, tracks = %d", len(page.Thumbnails), len(page.Tracks))
	}

	wantTrack := music.Track{
		VideoID:         "mGCnOLGwzg0",
		Title:           `Overture (From "TRON: Legacy"/Score)`,
		Artists:         wantArtists,
		Album:           &music.Album{ID: "MPREb_BQHrCwSBM37", Name: "TRON: Legacy"},
		DurationSeconds: 149,
		Thumbnails:      page.Thumbnails,
	}
	if !reflect.DeepEqual(page.Tracks[0], wantTrack) {
		t.Errorf("first track = %+v, want %+v", page.Tracks[0], wantTrack)
	}
}

func TestAlbumWithoutHeaderFails(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/home.json")

	_, err := server.client().Album(t.Context(), "MPREb_x")

	if err == nil {
		t.Fatal("expected an error for a page without album header")
	}
}

func TestPlaylist(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/playlist.json")

	page, err := server.client().Playlist(t.Context(), "RDCLAK5uy_k1Wu8QbZASiGVqr1wmie9NIYo38aBqscQ")
	if err != nil {
		t.Fatal(err)
	}

	wantID := "VLRDCLAK5uy_k1Wu8QbZASiGVqr1wmie9NIYo38aBqscQ"
	if server.lastRequest(t, "browse").body["browseId"] != wantID || page.ID != wantID {
		t.Errorf("request = %v, page ID = %q", server.lastRequest(t, "browse").body, page.ID)
	}
	if page.Title != "'80s Pop" || page.Author != "YouTube Music" || page.Description == "" {
		t.Errorf("page = %+v", page)
	}
	if len(page.Tracks) != 3 || page.Tracks[0].VideoID != "lcOxhH8N3Bo" || page.Tracks[0].DurationSeconds != 334 {
		t.Errorf("tracks = %+v", page.Tracks)
	}
	if page.Continuation == "" {
		t.Error("playlist page has no continuation")
	}
}

func TestPlaylistKeepsVLPrefix(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/playlist.json")

	if _, err := server.client().Playlist(t.Context(), "VLPLabc"); err != nil {
		t.Fatal(err)
	}

	if server.lastRequest(t, "browse").body["browseId"] != "VLPLabc" {
		t.Errorf("unexpected request body: %v", server.lastRequest(t, "browse").body)
	}
}

func TestPlaylistContinuation(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/playlist_continuation.json")

	page, err := server.client().PlaylistContinuation(t.Context(), "token")
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "browse")
	if request.body["continuation"] != "token" || request.body["browseId"] != nil {
		t.Errorf("unexpected request body: %v", request.body)
	}
	if len(page.Tracks) != 3 || page.Tracks[0].VideoID == "" {
		t.Errorf("tracks = %+v", page.Tracks)
	}
	if page.Continuation != "" {
		t.Errorf("continuation = %q on the last page", page.Continuation)
	}
}

func TestArtist(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/artist.json")

	page, err := server.client().Artist(t.Context(), "UCRr1xG_2WIDs18a6cIiCxeA")
	if err != nil {
		t.Fatal(err)
	}

	if page.ID != "UCRr1xG_2WIDs18a6cIiCxeA" || page.Name != "Daft Punk" || page.Description == "" || len(page.Thumbnails) == 0 {
		t.Errorf("page = %+v", page)
	}
	if len(page.TopSongs) != 2 {
		t.Fatalf("top songs = %+v", page.TopSongs)
	}
	song := page.TopSongs[0]
	if song.Title != "Instant Crush (feat. Julian Casablancas)" || len(song.Artists) != 2 || song.Album == nil || song.Album.ID != "MPREb_K8qWMWVqXGi" {
		t.Errorf("first top song = %+v", song)
	}
	if titles := pageShelfTitles(page.Shelves); !reflect.DeepEqual(titles, []string{"Albums", "Videos", "Fans might also like"}) {
		t.Fatalf("shelves = %v", titles)
	}
	if kind := page.Shelves[2].Items[0].Kind; kind != music.ItemKindArtist {
		t.Errorf("related kind = %s", kind)
	}
}

func TestRadio(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "next", "testdata/radio.json")

	tracks, err := server.client().Radio(t.Context(), "UM6Cts99XdA")
	if err != nil {
		t.Fatal(err)
	}

	body := server.lastRequest(t, "next").body
	if body["videoId"] != "UM6Cts99XdA" || body["playlistId"] != "RDAMVMUM6Cts99XdA" {
		t.Errorf("unexpected request body: %v", body)
	}
	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3", len(tracks))
	}
	want := music.Track{
		VideoID:         "UM6Cts99XdA",
		Title:           `Outlands (From "TRON: Legacy"/Score)`,
		Artists:         []music.Artist{{ID: "UCRr1xG_2WIDs18a6cIiCxeA", Name: "Daft Punk"}},
		Album:           &music.Album{ID: "MPREb_BQHrCwSBM37", Name: "TRON: Legacy"},
		DurationSeconds: 163,
	}
	if got := withoutTrackThumbnails(tracks[0]); !reflect.DeepEqual(got, want) {
		t.Errorf("seed track = %+v, want %+v", got, want)
	}
	if len(tracks[0].Thumbnails) == 0 {
		t.Error("seed track has no thumbnails")
	}
}

func TestPodcast(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/podcast.json")

	page, err := server.client().Podcast(t.Context(), "MPSPPLOrNU6aQc_Z4")
	if err != nil {
		t.Fatal(err)
	}

	wantAuthor := music.Artist{ID: "UCkQeESHaeH4g-gzFK-_G5Mg", Name: "Joel Garcia"}
	if page.Title != "The Pop Culture Shuffle" || !reflect.DeepEqual(page.Author, wantAuthor) || page.Description == "" {
		t.Errorf("page = %+v", page)
	}
	if len(page.Episodes) != 3 || page.Continuation == "" {
		t.Fatalf("episodes = %d, continuation = %q", len(page.Episodes), page.Continuation)
	}

	episode := page.Episodes[0]
	wantTrack := music.Track{
		VideoID:         "M6NrZuph2fU",
		Title:           "Revenge of Potter in the Box",
		Artists:         []music.Artist{wantAuthor},
		Album:           &music.Album{ID: "MPSPPLOrNU6aQc_Z4", Name: "The Pop Culture Shuffle"},
		DurationSeconds: 34 * 60,
	}
	if got := withoutTrackThumbnails(episode.Track); !reflect.DeepEqual(got, wantTrack) {
		t.Errorf("track = %+v, want %+v", got, wantTrack)
	}
	if episode.Published != "Aug 30" || episode.Description == "" {
		t.Errorf("episode = %+v", episode)
	}
}

func TestPodcastContinuationKeepsShow(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/podcast.json")
	page, err := server.client().Podcast(t.Context(), "MPSPPLOrNU6aQc_Z4")
	if err != nil {
		t.Fatal(err)
	}
	server.respondWithFixture(t, "browse", "testdata/podcast_continuation.json")

	more, err := server.client().PodcastContinuation(t.Context(), page.Continuation)
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "browse")
	if request.body["continuation"] == page.Continuation || request.body["continuation"] == nil {
		t.Errorf("sent the wrapped token instead of YouTube's: %v", request.body)
	}
	if len(more.Episodes) != 2 {
		t.Fatalf("episodes = %+v", more.Episodes)
	}
	track := more.Episodes[0].Track
	if track.Album == nil || track.Album.ID != "MPSPPLOrNU6aQc_Z4" || track.Artists[0].Name != "Joel Garcia" {
		t.Errorf("episode lost its show: %+v", track)
	}
}

func TestPodcastContinuationRejectsForeignToken(t *testing.T) {
	_, err := newFakeServer(t).client().PodcastContinuation(t.Context(), "not a token")

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseSpokenDuration(t *testing.T) {
	tests := map[string]int{
		" • 34 min":     34 * 60,
		"1 hr 5 min":    3900,
		" • 2 hr":       7200,
		"45 sec":        45,
		"":              0,
		"about an hour": 0,
	}
	for text, want := range tests {
		if got := parseSpokenDuration(text); got != want {
			t.Errorf("parseSpokenDuration(%q) = %d, want %d", text, got, want)
		}
	}
}

func pageShelfTitles(shelves []music.Shelf) []string {
	titles := []string{}
	for _, shelf := range shelves {
		titles = append(titles, shelf.Title)
	}
	return titles
}

func TestLyrics(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "next", "testdata/next_tabs.json")
	server.respondWithFixture(t, "browse", "testdata/lyrics.json")

	lyrics, found, err := server.client().Lyrics(t.Context(), "dQw4w9WgXcQ")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("lyrics not found")
	}
	if body := server.lastRequest(t, "browse").body; body["browseId"] != "MPLYt_6AEOVn4k62q" {
		t.Errorf("browse body = %v", body)
	}
	if lyrics.Source != "Musixmatch" || lyrics.Synced {
		t.Errorf("source %q synced %v", lyrics.Source, lyrics.Synced)
	}
	wantTexts := []string{"First line of the song", "Second line of the song", "", "Chorus line one", "Chorus line two"}
	var gotTexts []string
	for _, line := range lyrics.Lines {
		gotTexts = append(gotTexts, line.Text)
	}
	if !reflect.DeepEqual(gotTexts, wantTexts) {
		t.Errorf("lines = %q", gotTexts)
	}
}

func TestLyricsWithoutLyricsTab(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "next", "testdata/next_tabs_no_lyrics.json")

	_, found, err := server.client().Lyrics(t.Context(), "dQw4w9WgXcQ")
	if err != nil || found {
		t.Fatalf("found %v, err %v", found, err)
	}
	if server.requestCount("browse") != 0 {
		t.Error("browsed a lyrics page that does not exist")
	}
}

func TestLyricsUnavailable(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "next", "testdata/next_tabs.json")
	server.respondWithFixture(t, "browse", "testdata/lyrics_unavailable.json")

	_, found, err := server.client().Lyrics(t.Context(), "dQw4w9WgXcQ")
	if err != nil || found {
		t.Fatalf("found %v, err %v", found, err)
	}
}
