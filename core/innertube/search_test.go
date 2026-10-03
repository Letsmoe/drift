package innertube

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/neoworks/otter-music/core/music"
)

func TestSearchSongs(t *testing.T) {
	fixture, err := os.ReadFile("testdata/search_songs.json")
	if err != nil {
		t.Fatal(err)
	}

	var received searchRequest
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Errorf("decode request: %v", err)
		}
		_, _ = writer.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	tracks, err := client.SearchSongs(t.Context(), "daft punk")
	if err != nil {
		t.Fatal(err)
	}

	if received.Query != "daft punk" || received.Params != songsFilterParams {
		t.Errorf("unexpected request: %+v", received)
	}
	if received.Context.Client.ClientName != webRemixClientName {
		t.Errorf("client name = %q, want %q", received.Context.Client.ClientName, webRemixClientName)
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
	}
	if !reflect.DeepEqual(first, want) {
		t.Errorf("first track = %+v, want %+v", first, want)
	}
}

func TestSearchSongsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	_, err := NewClient(WithBaseURL(server.URL)).SearchSongs(t.Context(), "anything")

	var statusError *StatusError
	if !errors.As(err, &statusError) || statusError.StatusCode != http.StatusForbidden {
		t.Errorf("err = %v, want StatusError 403", err)
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
