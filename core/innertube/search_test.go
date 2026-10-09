package innertube

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/letsmoe/drift/core/music"
)

func TestSearchSongs(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "search", "testdata/search_songs.json")

	tracks, err := server.client().SearchSongs(t.Context(), "daft punk")
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "search")
	if request.body["query"] != "daft punk" || request.body["params"] != songsFilterParams {
		t.Errorf("unexpected request body: %v", request.body)
	}
	if request.contextClient()["clientName"] != WebRemix.Name {
		t.Errorf("client = %v, want %s", request.contextClient()["clientName"], WebRemix.Name)
	}
	if len(tracks) != 3 {
		t.Fatalf("got %d tracks, want 3", len(tracks))
	}

	first := tracks[0]
	first.Thumbnails = nil
	want := music.Track{
		VideoID:         "UM6Cts99XdA",
		Title:           `Outlands (From "TRON: Legacy"/Score)`,
		Artists:         []music.Artist{{ID: "UCRr1xG_2WIDs18a6cIiCxeA", Name: "Daft Punk"}},
		Album:           &music.Album{ID: "MPREb_BQHrCwSBM37", Name: "TRON: Legacy"},
		DurationSeconds: 163,
		PlayCount:       "10M plays",
	}
	if !reflect.DeepEqual(first, want) {
		t.Errorf("first track = %+v, want %+v", first, want)
	}
}

func TestSearchSongsStatusError(t *testing.T) {
	server := newFakeServer(t)

	_, err := server.client().SearchSongs(t.Context(), "anything")

	var statusError *StatusError
	if !errors.As(err, &statusError) || statusError.StatusCode != http.StatusNotFound {
		t.Errorf("err = %v, want StatusError 404", err)
	}
}

func TestVisitorDataIsRequestedOnceAndSent(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "search", "testdata/search_songs.json")
	client := server.client()

	for range 2 {
		if _, err := client.SearchSongs(t.Context(), "daft punk"); err != nil {
			t.Fatal(err)
		}
	}

	if count := server.requestCount("visitor_id"); count != 1 {
		t.Errorf("visitor_id requested %d times, want 1", count)
	}
	request := server.lastRequest(t, "search")
	if request.headers.Get("X-Goog-Visitor-Id") != fakeVisitorData {
		t.Errorf("X-Goog-Visitor-Id = %q", request.headers.Get("X-Goog-Visitor-Id"))
	}
	if request.contextClient()["visitorData"] != fakeVisitorData {
		t.Errorf("context visitorData = %v", request.contextClient()["visitorData"])
	}
}

func TestVisitorDataFailureStopsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"responseContext":{}}`))
	}))
	defer server.Close()

	_, err := NewClient(WithBaseURL(server.URL)).SearchSongs(t.Context(), "anything")

	if !errors.Is(err, errNoVisitorData) {
		t.Errorf("err = %v, want errNoVisitorData", err)
	}
}

func TestParseArtistsWithMultipleArtists(t *testing.T) {
	runs := []run{
		artistRun("Daft Punk", "UC1"),
		{Text: ", "},
		artistRun("Pharrell Williams", "UC2"),
		{Text: " & "},
		{Text: "Nile Rodgers"},
		{Text: segmentSeparator},
		{Text: "6:10"},
	}

	artists := parseArtists(splitSegments(runs))

	want := []music.Artist{
		{ID: "UC1", Name: "Daft Punk"},
		{ID: "UC2", Name: "Pharrell Williams"},
		{ID: "", Name: "Nile Rodgers"},
	}
	if !reflect.DeepEqual(artists, want) {
		t.Errorf("artists = %+v, want %+v", artists, want)
	}
}

func TestParseDuration(t *testing.T) {
	cases := map[string]int{
		"2:43":    163,
		"1:02:03": 3723,
		"10M":     0,
	}
	for text, want := range cases {
		segments := [][]run{{{Text: text}}}
		if got := parseDuration(segments); got != want {
			t.Errorf("parseDuration(%q) = %d, want %d", text, got, want)
		}
	}
}

func artistRun(name, browseID string) run {
	endpoint := &browseEndpoint{BrowseID: browseID}
	endpoint.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType = pageTypeArtist
	return run{Text: name, NavigationEndpoint: &navigationEndpoint{BrowseEndpoint: endpoint}}
}
