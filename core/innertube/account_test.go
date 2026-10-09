package innertube

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/letsmoe/drift-music/core/music"
)

const testCookieHeader = "SID=sid-value; SAPISID=sapisid-value; __Secure-1PAPISID=first-party-value; __Secure-3PAPISID=third-party-value; HSID=hsid"

// TestSessionAuthorization checks the header against hashes computed by hand:
// sha1("1700000000 <cookie value> https://music.youtube.com").
func TestSessionAuthorization(t *testing.T) {
	session := Session{CookieHeader: testCookieHeader}
	moment := time.Unix(1700000000, 0)

	got := session.authorization(moment)

	want := "SAPISIDHASH 1700000000_11055f89efabedf355d8a9cb9845ef0187ec5fdc " +
		"SAPISID1PHASH 1700000000_243ac4b871120f72b80cbab662d8aa6e1808c763 " +
		"SAPISID3PHASH 1700000000_64ec577e30e589215cfcced206e05dcf870e2ed8"
	if got != want {
		t.Errorf("authorization = %q\nwant %q", got, want)
	}
}

func TestSessionAuthorizationFallsBackToThirdPartyCookie(t *testing.T) {
	session := Session{CookieHeader: "__Secure-3PAPISID=third-party-value"}
	got := session.authorization(time.Unix(1700000000, 0))
	if !strings.HasPrefix(got, "SAPISIDHASH 1700000000_") || !strings.Contains(got, " SAPISID3PHASH ") || strings.Contains(got, "SAPISID1PHASH") {
		t.Errorf("authorization = %q", got)
	}
}

func TestSessionValidateNeedsSAPISID(t *testing.T) {
	withoutSAPISID := Session{CookieHeader: "SID=a; HSID=b"}
	err := withoutSAPISID.Validate()
	if !errors.Is(err, ErrNoSessionCookies) {
		t.Errorf("err = %v, want ErrNoSessionCookies", err)
	}
	complete := Session{CookieHeader: testCookieHeader}
	if complete.Validate() != nil {
		t.Error("a header with SAPISID must validate")
	}
}

func TestSessionNeverPrintsCookies(t *testing.T) {
	session := Session{CookieHeader: testCookieHeader}
	for _, text := range []string{session.String(), session.GoString(), session.LogValue().String()} {
		if strings.Contains(text, "sapisid-value") {
			t.Errorf("%q leaks the cookies", text)
		}
	}
}

// accountServer answers browse requests by browse ID or continuation token.
type accountServer struct {
	*httptest.Server

	mutex    sync.Mutex
	fixtures map[string]string
	headers  []http.Header
}

func newAccountServer(t *testing.T, fixtures map[string]string) *accountServer {
	t.Helper()
	server := &accountServer{fixtures: fixtures}
	server.Server = httptest.NewServer(http.HandlerFunc(server.handle))
	t.Cleanup(server.Close)
	return server
}

func (server *accountServer) handle(writer http.ResponseWriter, request *http.Request) {
	endpoint := strings.TrimPrefix(request.URL.Path, "/youtubei/v1/")
	var body map[string]any
	_ = json.NewDecoder(request.Body).Decode(&body)

	server.mutex.Lock()
	server.headers = append(server.headers, request.Header)
	server.mutex.Unlock()

	key := endpoint
	for _, field := range []string{"browseId", "continuation"} {
		value, ok := body[field].(string)
		if ok {
			key = value
		}
	}
	if endpoint == "visitor_id" {
		_, _ = writer.Write([]byte(`{"responseContext":{"visitorData":"visitor"}}`))
		return
	}
	path, found := server.fixtures[key]
	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = writer.Write(contents)
}

func (server *accountServer) client() *AccountClient {
	return NewAccountClient(Session{CookieHeader: testCookieHeader}, WithBaseURL(server.URL))
}

func TestAccount(t *testing.T) {
	server := newAccountServer(t, map[string]string{"account/account_menu": "testdata/account_menu.json"})

	account, err := server.client().Account(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	want := music.Account{Name: "Ada Lovelace", Handle: "@ada", Email: "ada@example.com", PhotoURL: "https://yt3.ggpht.com/ada=s108"}
	if account != want {
		t.Errorf("account = %+v, want %+v", account, want)
	}
}

func TestAccountSendsSessionHeaders(t *testing.T) {
	server := newAccountServer(t, map[string]string{"account/account_menu": "testdata/account_menu.json"})
	_, err := server.client().Account(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	headers := server.headers[len(server.headers)-1]
	if headers.Get("Cookie") != testCookieHeader {
		t.Errorf("Cookie = %q", headers.Get("Cookie"))
	}
	if !strings.HasPrefix(headers.Get("Authorization"), "SAPISIDHASH ") {
		t.Errorf("Authorization = %q", headers.Get("Authorization"))
	}
	if headers.Get("X-Origin") != musicOrigin || headers.Get("Origin") != musicOrigin || headers.Get("X-Goog-AuthUser") != "0" {
		t.Errorf("origin headers = %v", headers)
	}
}

func TestAnonymousRequestsCarryNoSession(t *testing.T) {
	server := newAccountServer(t, map[string]string{"FEmusic_home": "testdata/home.json"})
	_, err := NewClient(WithBaseURL(server.URL)).Home(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, headers := range server.headers {
		if headers.Get("Cookie") != "" || headers.Get("Authorization") != "" {
			t.Errorf("anonymous request carried credentials: %v", headers)
		}
	}
}

func TestAccountSignedOutIsSessionExpired(t *testing.T) {
	server := newAccountServer(t, map[string]string{"account/account_menu": "testdata/account_menu_signed_out.json"})
	_, err := server.client().Account(t.Context())
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("err = %v, want ErrSessionExpired", err)
	}
}

func TestUnauthorizedIsSessionExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "visitor_id") {
			_, _ = writer.Write([]byte(`{"responseContext":{"visitorData":"visitor"}}`))
			return
		}
		writer.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	client := NewAccountClient(Session{CookieHeader: testCookieHeader}, WithBaseURL(server.URL))

	_, err := client.LikedSongs(t.Context())
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("err = %v, want ErrSessionExpired", err)
	}
}

func TestLikedSongsFollowsContinuations(t *testing.T) {
	server := newAccountServer(t, map[string]string{
		"VLLM":       "testdata/liked_songs.json",
		"liked-next": "testdata/liked_songs_continuation.json",
	})

	tracks, err := server.client().LikedSongs(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	var videoIDs []string
	for _, track := range tracks {
		videoIDs = append(videoIDs, track.VideoID)
	}
	if strings.Join(videoIDs, ",") != "vid1,vid2,vid3" {
		t.Fatalf("video IDs = %v", videoIDs)
	}
	if tracks[0].Title != "One More Time" || tracks[0].Artists[0].Name != "Daft Punk" || tracks[0].Album.Name != "Discovery" || tracks[0].DurationSeconds != 320 {
		t.Errorf("first track = %+v", tracks[0])
	}
}

func TestPrivatePlaylistHeader(t *testing.T) {
	var response pageResponse
	contents, err := os.ReadFile("testdata/liked_songs.json")
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(contents, &response)
	if err != nil {
		t.Fatal(err)
	}
	header := response.header()
	if header == nil || joinText(header.Title.Runs) != "Liked Music" {
		t.Errorf("header = %+v", header)
	}
}

func itemIDs(items []music.Item) string {
	ids := []string{}
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return strings.Join(ids, ",")
}

func TestLibraryPlaylists(t *testing.T) {
	server := newAccountServer(t, map[string]string{
		"FEmusic_liked_playlists": "testdata/library_playlists.json",
		"playlists-next":          "testdata/library_playlists_continuation.json",
	})

	items, err := server.client().LibraryPlaylists(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	// The create-playlist button is skipped and the repeated playlist is kept once.
	if got := itemIDs(items); got != "VLLM,VLPLroad,VLSE,VLPLfocus" {
		t.Errorf("ids = %s", got)
	}
	if items[1].Title != "Road trip" || items[1].Subtitle != "Playlist • 2 songs" {
		t.Errorf("playlist = %+v", items[1])
	}
}

func TestLibraryAlbums(t *testing.T) {
	server := newAccountServer(t, map[string]string{"FEmusic_liked_albums": "testdata/library_albums.json"})
	items, err := server.client().LibraryAlbums(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := itemIDs(items); got != "MPREbdisc,MPREbokc" {
		t.Errorf("ids = %s", got)
	}
}

func TestLibraryArtists(t *testing.T) {
	server := newAccountServer(t, map[string]string{
		"FEmusic_library_corpus_artists": "testdata/library_artists.json",
		"artists-next":                   "testdata/library_artists_continuation.json",
	})
	items, err := server.client().LibraryArtists(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := itemIDs(items); got != "UCdaft,UCradio,UCbonobo" {
		t.Errorf("ids = %s", got)
	}
	if items[0].Kind != music.ItemKindArtist || items[0].Title != "Daft Punk" || items[0].Subtitle != "3.2M subscribers" {
		t.Errorf("artist = %+v", items[0])
	}
}

func TestLibraryPodcasts(t *testing.T) {
	server := newAccountServer(t, map[string]string{"FEmusic_library_non_music_audio_list": "testdata/library_podcasts.json"})
	items, err := server.client().LibraryPodcasts(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got := itemIDs(items); got != "MPSPPLdark" {
		t.Errorf("ids = %s", got)
	}
}

func TestEmptyLibraryIsNotAnError(t *testing.T) {
	empty := t.TempDir() + "/empty.json"
	err := os.WriteFile(empty, []byte(`{"contents":{"singleColumnBrowseResultsRenderer":{"tabs":[{"tabRenderer":{"content":{"sectionListRenderer":{}}}}]}}}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	server := newAccountServer(t, map[string]string{"FEmusic_liked_albums": empty})

	items, err := server.client().LibraryAlbums(t.Context())
	if err != nil || len(items) != 0 {
		t.Errorf("items = %v, err = %v", items, err)
	}
}
