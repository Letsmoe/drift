package innertube

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/neoworks/otter-music/core/music"
)

func TestSearch(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "search", "testdata/search_all.json")

	results, err := server.client().Search(t.Context(), "daft punk")
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "search")
	if request.body["query"] != "daft punk" || request.body["params"] != nil {
		t.Errorf("unexpected request body: %v", request.body)
	}
	wantTitles := []string{"Top result", "Albums", "Songs", "Artists", "Playlists", "Videos", "Episodes", "Podcasts"}
	if titles := searchShelfTitles(results); !reflect.DeepEqual(titles, wantTitles) {
		t.Fatalf("shelves = %v, want %v", titles, wantTitles)
	}
}

func TestSearchTopResultHoldsCardAndItsSongs(t *testing.T) {
	results := searchFixture(t, "testdata/search_all.json")

	items := results.Shelves[0].Items
	if len(items) != 3 {
		t.Fatalf("top result has %d items, want 3", len(items))
	}
	if items[0].Kind != music.ItemKindArtist || items[0].ID != "UCRr1xG_2WIDs18a6cIiCxeA" || items[0].Title != "Daft Punk" {
		t.Errorf("card = %+v", items[0])
	}

	want := music.Track{
		VideoID:         "OWiVJMgms9E",
		Title:           "Voyager",
		Artists:         []music.Artist{{ID: "UCRr1xG_2WIDs18a6cIiCxeA", Name: "Daft Punk"}},
		DurationSeconds: 228,
		PlayCount:       "51M plays",
	}
	if got := withoutTrackThumbnails(*items[1].Track); !reflect.DeepEqual(got, want) {
		t.Errorf("listed song = %+v, want %+v", got, want)
	}
}

func TestSearchMapsRowsByKind(t *testing.T) {
	results := searchFixture(t, "testdata/search_all.json")

	album := withoutThumbnails(results.Shelves[1].Items[0])
	wantAlbum := music.Item{Kind: music.ItemKindAlbum, ID: "MPREb_K8qWMWVqXGi", Title: "Random Access Memories", Subtitle: "Album • Daft Punk • 2013"}
	if !reflect.DeepEqual(album, wantAlbum) {
		t.Errorf("album = %+v, want %+v", album, wantAlbum)
	}

	video := results.Shelves[5].Items[0]
	if video.Kind != music.ItemKindSong || video.Track.VideoID != "IJSK_R0ZkGk" || video.Track.Artists[0].Name != "TheStruckFernVEVO" {
		t.Errorf("video = %+v %+v", video, video.Track)
	}

	episode := results.Shelves[6].Items[0]
	if episode.Kind != music.ItemKindEpisode || episode.ID != "hHq0aMWRbZQ" || episode.Track == nil {
		t.Fatalf("episode = %+v", episode)
	}
	if episode.Track.Album == nil || episode.Track.Album.ID != "MPSPPLZr1D7YmlHJw" {
		t.Errorf("episode show = %+v", episode.Track.Album)
	}

	podcast := results.Shelves[7].Items[0]
	if podcast.Kind != music.ItemKindPodcast || podcast.ID != "MPSPPLanozi--QaflHWgV4hgvvA8hGMKC4IVKW" {
		t.Errorf("podcast = %+v", podcast)
	}
}

func TestSearchSkipsProfiles(t *testing.T) {
	results := searchFixture(t, "testdata/search_all.json")

	for _, shelf := range results.Shelves {
		for _, item := range shelf.Items {
			if item.Title == "subjekt03" {
				t.Errorf("profile %+v was not skipped", item)
			}
		}
	}
}

func TestSearchKeepsTitledShelves(t *testing.T) {
	results := searchFixture(t, "testdata/search_songs.json")

	if len(results.Shelves) != 1 || len(results.Shelves[0].Items) != 3 {
		t.Fatalf("shelves = %+v", results.Shelves)
	}
	song := results.Shelves[0].Items[0]
	if song.Kind != music.ItemKindSong || song.Track.Album == nil || song.Track.DurationSeconds != 163 {
		t.Errorf("song = %+v %+v", song, song.Track)
	}
}

func TestSearchSuggestions(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "music/get_search_suggestions", "testdata/search_suggestions.json")

	suggestions, err := server.client().SearchSuggestions(t.Context(), "daft p")
	if err != nil {
		t.Fatal(err)
	}

	if server.lastRequest(t, "music/get_search_suggestions").body["input"] != "daft p" {
		t.Errorf("unexpected request body: %v", server.lastRequest(t, "music/get_search_suggestions").body)
	}
	want := []string{"daft punk", "daft punk one more time", "daft punk veridis quo"}
	if !reflect.DeepEqual(suggestions, want) {
		t.Errorf("suggestions = %v, want %v", suggestions, want)
	}
}

func TestParseCardPodcastAndEpisode(t *testing.T) {
	podcastCard := unmarshalCard(t, `{
		"title": {"runs": [{"text": "The Show"}]},
		"subtitle": {"runs": [{"text": "Podcast"}]},
		"navigationEndpoint": {"browseEndpoint": {"browseId": "MPSP1",
			"browseEndpointContextSupportedConfigs": {"browseEndpointContextMusicConfig": {"pageType": "MUSIC_PAGE_TYPE_PODCAST_SHOW_DETAIL_PAGE"}}}}
	}`)
	item, ok := parseCard(podcastCard)
	if !ok || item.Kind != music.ItemKindPodcast || item.ID != "MPSP1" {
		t.Errorf("podcast = %+v, ok = %v", item, ok)
	}

	episodeCard := unmarshalCard(t, `{
		"title": {"runs": [{"text": "Episode one"}]},
		"subtitle": {"runs": [{"text": "Aug 30"}, {"text": " • "},
			{"text": "The Show", "navigationEndpoint": {"browseEndpoint": {"browseId": "MPSP1",
				"browseEndpointContextSupportedConfigs": {"browseEndpointContextMusicConfig": {"pageType": "MUSIC_PAGE_TYPE_PODCAST_SHOW_DETAIL_PAGE"}}}}}]},
		"navigationEndpoint": {"browseEndpoint": {"browseId": "MPEDvideo1",
			"browseEndpointContextSupportedConfigs": {"browseEndpointContextMusicConfig": {"pageType": "MUSIC_PAGE_TYPE_NON_MUSIC_AUDIO_TRACK_PAGE"}}}}
	}`)
	item, ok = parseCard(episodeCard)
	if !ok || item.Kind != music.ItemKindEpisode || item.ID != "video1" || item.Track == nil {
		t.Fatalf("episode = %+v, ok = %v", item, ok)
	}
	if item.Track.Album == nil || item.Track.Album.ID != "MPSP1" {
		t.Errorf("episode show = %+v", item.Track.Album)
	}
}

func unmarshalCard(t *testing.T, cardJSON string) twoRowItem {
	t.Helper()
	var card twoRowItem
	if err := json.Unmarshal([]byte(cardJSON), &card); err != nil {
		t.Fatal(err)
	}
	return card
}

func searchFixture(t *testing.T, fixturePath string) music.SearchResults {
	t.Helper()
	server := newFakeServer(t)
	server.respondWithFixture(t, "search", fixturePath)
	results, err := server.client().Search(t.Context(), "query")
	if err != nil {
		t.Fatal(err)
	}
	return results
}

func searchShelfTitles(results music.SearchResults) []string {
	titles := []string{}
	for _, shelf := range results.Shelves {
		titles = append(titles, shelf.Title)
	}
	return titles
}

func withoutTrackThumbnails(track music.Track) music.Track {
	track.Thumbnails = nil
	return track
}
