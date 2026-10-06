package innertube

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/neoworks/otter-music/core/music"
)

func TestHome(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/home.json")

	page, err := server.client().Home(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "browse")
	if request.body["browseId"] != homeBrowseID || request.body["continuation"] != nil {
		t.Errorf("unexpected request body: %v", request.body)
	}
	if page.Continuation != "fixture-continuation-token" {
		t.Errorf("continuation = %q", page.Continuation)
	}
	if titles := shelfTitles(page); !reflect.DeepEqual(titles, []string{"Throwback", "Quick picks"}) {
		t.Fatalf("shelves = %v", titles)
	}

	playlist := withoutThumbnails(page.Shelves[0].Items[0])
	wantPlaylist := music.Item{
		Kind:     music.ItemKindPlaylist,
		ID:       "VLRDCLAK5uy_k1Wu8QbZASiGVqr1wmie9NIYo38aBqscQ",
		Title:    "'80s Pop",
		Subtitle: "George Michael, Madonna, Michael Jackson, Erasure",
	}
	if !reflect.DeepEqual(playlist, wantPlaylist) {
		t.Errorf("playlist = %+v, want %+v", playlist, wantPlaylist)
	}

	song := withoutThumbnails(page.Shelves[1].Items[1])
	wantSong := music.Item{
		Kind:     music.ItemKindSong,
		ID:       "eJFi2TK5174",
		Title:    "Solar Eclipse",
		Subtitle: "Drake & Don Toliver • 6.2M plays",
		Track: &music.Track{
			VideoID: "eJFi2TK5174",
			Title:   "Solar Eclipse",
			Artists: []music.Artist{
				{ID: "UCU6cE7pdJPc6DU2jSrKEsdQ", Name: "Drake"},
				{ID: "UCSzWQmDsKG37iKN2vw1G-2Q", Name: "Don Toliver"},
			},
			Album:     &music.Album{ID: "MPREb_L2anhKI1Knd", Name: "HABIBTI (FOMO)"},
			PlayCount: "6.2M plays",
		},
	}
	if !reflect.DeepEqual(song, wantSong) {
		t.Errorf("song = %+v %+v, want %+v %+v", song, song.Track, wantSong, wantSong.Track)
	}
}

func TestHomeContinuation(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/home_continuation.json")

	page, err := server.client().HomeContinuation(t.Context(), "token")
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "browse")
	if request.body["continuation"] != "token" || request.body["browseId"] != nil {
		t.Errorf("unexpected request body: %v", request.body)
	}
	if page.Continuation != "" {
		t.Errorf("continuation = %q on the last page", page.Continuation)
	}
	if titles := shelfTitles(page); !reflect.DeepEqual(titles, []string{"New releases", "Trending songs for you"}) {
		t.Fatalf("shelves = %v", titles)
	}

	album := page.Shelves[0].Items[0]
	if album.Kind != music.ItemKindAlbum || album.ID != "MPREb_zJXCvnQsnxE" || album.Subtitle != "Single • Bruno Mars" {
		t.Errorf("album = %+v", album)
	}
	if len(album.Thumbnails) != 2 || album.Thumbnails[1].Width != 544 {
		t.Errorf("album thumbnails = %+v", album.Thumbnails)
	}
}

func TestPodcastsHome(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "browse", "testdata/home_podcasts.json")

	page, err := server.client().PodcastsHome(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "browse")
	if request.body["browseId"] != homeBrowseID || request.body["params"] != podcastsChipParams {
		t.Errorf("unexpected request body: %v", request.body)
	}
	if page.Continuation != "fixture-continuation-token" {
		t.Errorf("continuation = %q", page.Continuation)
	}
	if titles := shelfTitles(page); !reflect.DeepEqual(titles, []string{"History"}) {
		t.Fatalf("shelves = %v", titles)
	}

	episode := withoutThumbnails(page.Shelves[0].Items[0])
	show := music.Artist{Name: "Deep Dive History"}
	wantEpisode := music.Item{
		Kind:     music.ItemKindEpisode,
		ID:       "BrIoPtoWims",
		Title:    "5. THE EPIC OF GILGAMESH - The Entire Story",
		Subtitle: "Deep Dive History",
		Track: &music.Track{
			VideoID:         "BrIoPtoWims",
			Title:           "5. THE EPIC OF GILGAMESH - The Entire Story",
			Artists:         []music.Artist{show},
			Album:           &music.Album{ID: "MPSPPLc-YUx4BNhJZFkHcZT-TPTR_6X8GhTApW", Name: "Deep Dive History"},
			DurationSeconds: 111 * 60,
			PlayCount:       "922K views",
		},
	}
	if !reflect.DeepEqual(episode, wantEpisode) {
		t.Errorf("episode = %+v %+v, want %+v %+v", episode, episode.Track, wantEpisode, wantEpisode.Track)
	}
}

func TestParseCardWithWatchEndpointIsSong(t *testing.T) {
	var card twoRowItem
	cardJSON := `{
		"title": {"runs": [{"text": "Video"}]},
		"subtitle": {"runs": [
			{"text": "Song"}, {"text": " • "},
			{"text": "Artist", "navigationEndpoint": {"browseEndpoint": {"browseId": "UC1",
				"browseEndpointContextSupportedConfigs": {"browseEndpointContextMusicConfig": {"pageType": "MUSIC_PAGE_TYPE_ARTIST"}}}}}
		]},
		"navigationEndpoint": {"watchEndpoint": {"videoId": "video1"}}
	}`
	if err := json.Unmarshal([]byte(cardJSON), &card); err != nil {
		t.Fatal(err)
	}

	item, ok := parseCard(card)

	if !ok || item.Kind != music.ItemKindSong || item.ID != "video1" {
		t.Fatalf("item = %+v, ok = %v", item, ok)
	}
	wantArtists := []music.Artist{{ID: "UC1", Name: "Artist"}}
	if !reflect.DeepEqual(item.Track.Artists, wantArtists) {
		t.Errorf("artists = %+v, want %+v", item.Track.Artists, wantArtists)
	}
}

func TestParseCardSkipsUnknownPages(t *testing.T) {
	card := twoRowItem{}
	card.NavigationEndpoint.BrowseEndpoint = &browseEndpoint{BrowseID: "FEmusic_moods"}

	if item, ok := parseCard(card); ok {
		t.Errorf("parsed %+v, want skipped", item)
	}
}

func shelfTitles(page music.HomePage) []string {
	titles := []string{}
	for _, shelf := range page.Shelves {
		titles = append(titles, shelf.Title)
	}
	return titles
}

func withoutThumbnails(item music.Item) music.Item {
	item.Thumbnails = nil
	if item.Track != nil {
		track := *item.Track
		track.Thumbnails = nil
		item.Track = &track
	}
	return item
}
