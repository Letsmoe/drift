//go:build live

package innertube

import (
	"testing"

	"github.com/neoworks/otter-music/core/music"
)

func TestLiveHome(t *testing.T) {
	client := NewClient()

	page, err := client.Home(t.Context())
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	if len(page.Shelves) == 0 || len(page.Shelves[0].Items) == 0 {
		t.Fatal("home returned no shelves; the response layout probably changed")
	}
	if page.Continuation == "" {
		t.Fatal("home returned no continuation")
	}

	next, err := client.HomeContinuation(t.Context(), page.Continuation)
	if err != nil {
		t.Fatalf("home continuation: %v", err)
	}
	if len(next.Shelves) == 0 {
		t.Error("home continuation returned no shelves")
	}
	for _, shelf := range append(page.Shelves, next.Shelves...) {
		t.Logf("%s: %d items, first %s %q", shelf.Title, len(shelf.Items), shelf.Items[0].Kind, shelf.Items[0].Title)
	}
}

func TestLivePodcastsHome(t *testing.T) {
	page, err := NewClient().PodcastsHome(t.Context())
	if err != nil {
		t.Fatalf("podcasts home: %v", err)
	}
	if len(page.Shelves) == 0 || len(page.Shelves[0].Items) == 0 {
		t.Fatal("podcasts home returned no shelves; the chip params or layout probably changed")
	}
	first := page.Shelves[0].Items[0]
	if first.Kind != music.ItemKindEpisode || first.Track == nil || first.Track.Album == nil || first.Track.Album.ID == "" {
		t.Errorf("first item = %+v, want an episode linked to its show", first)
	}
}

const (
	liveArtistID   = "UCRr1xG_2WIDs18a6cIiCxeA" // Daft Punk
	liveAlbumID    = "MPREb_BQHrCwSBM37"        // TRON: Legacy
	liveVideoID    = "UM6Cts99XdA"
	longPlaylistID = "RDCLAK5uy_k1Wu8QbZASiGVqr1wmie9NIYo38aBqscQ" // 100+ tracks, so it paginates
)

func TestLiveSearch(t *testing.T) {
	results, err := NewClient().Search(t.Context(), "daft punk")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results.Shelves) < 3 {
		t.Fatalf("search returned %d shelves; the response layout probably changed", len(results.Shelves))
	}
	for _, shelf := range results.Shelves {
		t.Logf("%s: %d items, first %s %q", shelf.Title, len(shelf.Items), shelf.Items[0].Kind, shelf.Items[0].Title)
	}
}

func TestLiveSearchSuggestions(t *testing.T) {
	suggestions, err := NewClient().SearchSuggestions(t.Context(), "daft p")
	if err != nil {
		t.Fatalf("suggestions: %v", err)
	}
	if len(suggestions) == 0 {
		t.Fatal("no suggestions")
	}
}

func TestLiveAlbum(t *testing.T) {
	page, err := NewClient().Album(t.Context(), liveAlbumID)
	if err != nil {
		t.Fatalf("album: %v", err)
	}
	if page.Title == "" || len(page.Artists) == 0 || len(page.Thumbnails) == 0 || len(page.Tracks) == 0 {
		t.Fatalf("incomplete album page: %+v", page)
	}
	for _, track := range page.Tracks {
		if track.Album == nil || track.DurationSeconds == 0 || len(track.Artists) == 0 || len(track.Thumbnails) == 0 {
			t.Errorf("incomplete album track: %+v", track)
		}
	}
}

func TestLivePlaylistPaginates(t *testing.T) {
	client := NewClient()
	page, err := client.Playlist(t.Context(), longPlaylistID)
	if err != nil {
		t.Fatalf("playlist: %v", err)
	}
	if page.Title == "" || len(page.Tracks) == 0 {
		t.Fatalf("incomplete playlist page: %+v", page)
	}
	if page.Continuation == "" {
		t.Fatal("playlist has no continuation")
	}

	more, err := client.PlaylistContinuation(t.Context(), page.Continuation)
	if err != nil {
		t.Fatalf("playlist continuation: %v", err)
	}
	if len(more.Tracks) == 0 {
		t.Error("playlist continuation returned no tracks")
	}
}

func TestLiveArtist(t *testing.T) {
	page, err := NewClient().Artist(t.Context(), liveArtistID)
	if err != nil {
		t.Fatalf("artist: %v", err)
	}
	if page.Name == "" || len(page.Thumbnails) == 0 || len(page.TopSongs) == 0 || len(page.Shelves) == 0 {
		t.Fatalf("incomplete artist page: %+v", page)
	}
}

func TestLiveRadio(t *testing.T) {
	tracks, err := NewClient().Radio(t.Context(), liveVideoID)
	if err != nil {
		t.Fatalf("radio: %v", err)
	}
	if len(tracks) < 10 {
		t.Fatalf("radio returned %d tracks", len(tracks))
	}
}

func TestLivePodcast(t *testing.T) {
	client := NewClient()
	podcastID := findLivePodcast(t, client)

	page, err := client.Podcast(t.Context(), podcastID)
	if err != nil {
		t.Fatalf("podcast: %v", err)
	}
	if page.Title == "" || page.Author.Name == "" || len(page.Episodes) == 0 {
		t.Fatalf("incomplete podcast page: %+v", page)
	}
	for _, episode := range page.Episodes {
		if episode.Track.VideoID == "" || episode.Track.DurationSeconds == 0 || episode.Track.Album == nil {
			t.Errorf("incomplete episode: %+v", episode)
		}
	}
	if page.Continuation == "" {
		return
	}

	more, err := client.PodcastContinuation(t.Context(), page.Continuation)
	if err != nil {
		t.Fatalf("podcast continuation: %v", err)
	}
	if len(more.Episodes) == 0 {
		t.Error("podcast continuation returned no episodes")
	}
}

// findLivePodcast looks a podcast up through search, as the app does.
func findLivePodcast(t *testing.T, client *Client) string {
	t.Helper()
	results, err := client.Search(t.Context(), "joe rogan podcast")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	for _, shelf := range results.Shelves {
		for _, item := range shelf.Items {
			if item.Kind == music.ItemKindPodcast {
				return item.ID
			}
		}
	}
	t.Fatal("search found no podcast")
	return ""
}

func TestLiveLyrics(t *testing.T) {
	client := NewClient()

	lyrics, found, err := client.Lyrics(t.Context(), "60ItHLz5WEA") // Alan Walker, Faded
	if err != nil {
		t.Fatalf("lyrics: %v", err)
	}
	if !found || len(lyrics.Lines) < 4 {
		t.Fatalf("expected lyrics, got found=%v %+v", found, lyrics)
	}
	t.Logf("%d lines, source %q", len(lyrics.Lines), lyrics.Source)

	_, found, err = client.Lyrics(t.Context(), liveVideoID) // an instrumental score
	if err != nil {
		t.Fatalf("lyrics of instrumental: %v", err)
	}
	if found {
		t.Error("instrumental track unexpectedly has lyrics")
	}
}
