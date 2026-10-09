package importer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/letsmoe/drift-music/core/innertube"
	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

// fakeAccount is a signed-in YouTube Music library held in memory.
type fakeAccount struct {
	liked          []music.Track
	playlists      []music.Item
	playlistTracks map[string][]music.Track
	albums         []music.Item
	artists        []music.Item
	podcasts       []music.Item
	failPlaylist   string
	expired        bool
}

func (account *fakeAccount) LikedSongs(context.Context) ([]music.Track, error) {
	if account.expired {
		return nil, innertube.ErrSessionExpired
	}
	return account.liked, nil
}

func (account *fakeAccount) LibraryPlaylists(context.Context) ([]music.Item, error) {
	return account.playlists, nil
}

func (account *fakeAccount) PlaylistTracks(_ context.Context, playlistID string) ([]music.Track, error) {
	if playlistID == account.failPlaylist {
		return nil, errors.New("playlist is private")
	}
	return account.playlistTracks[playlistID], nil
}

func (account *fakeAccount) LibraryAlbums(context.Context) ([]music.Item, error) {
	return account.albums, nil
}

func (account *fakeAccount) LibraryArtists(context.Context) ([]music.Item, error) {
	return account.artists, nil
}

func (account *fakeAccount) LibraryPodcasts(context.Context) ([]music.Item, error) {
	return account.podcasts, nil
}

func accountSong(title, artist string) music.Track {
	track := song(title, artist)
	track.VideoID = "video-" + title
	return track
}

func newFakeAccount() *fakeAccount {
	return &fakeAccount{
		liked: []music.Track{accountSong("Newest", "Artist A"), accountSong("Older", "Artist B"), accountSong("Oldest", "Artist C")},
		playlists: []music.Item{
			{Kind: music.ItemKindPlaylist, ID: "VLLM", Title: "Liked Music"},
			{Kind: music.ItemKindPlaylist, ID: "VLPLroad", Title: "Road trip", Subtitle: "Playlist • 2 songs"},
			{Kind: music.ItemKindPlaylist, ID: "VLPLfocus", Title: "Focus", Subtitle: "Playlist • 1,200 songs"},
			{Kind: music.ItemKindPlaylist, ID: "VLSE", Title: "Episodes for later"},
		},
		playlistTracks: map[string][]music.Track{
			"PLroad":  {accountSong("Newest", "Artist A"), accountSong("Older", "Artist B")},
			"PLfocus": {accountSong("Oldest", "Artist C")},
		},
		albums:   []music.Item{{Kind: music.ItemKindAlbum, ID: "MPREbdisc", Title: "Discovery"}},
		artists:  []music.Item{{Kind: music.ItemKindArtist, ID: "UCdaft", Title: "Daft Punk"}},
		podcasts: []music.Item{{Kind: music.ItemKindPodcast, ID: "MPSPPLdark", Title: "Darknet Diaries"}},
	}
}

func newAccountImporter(t *testing.T, account Account) (*Importer, *library.Library) {
	t.Helper()
	lib, err := library.Open(context.Background(), filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatalf("open library: %v", err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	provider := func(context.Context) (Account, error) { return account, nil }
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(lib, &fakeCatalog{}, logger, WithAccount(provider)), lib
}

var accountSource = Source{Kind: SourceYouTubeAccount}

func TestAccountPreview(t *testing.T) {
	importer, _ := newAccountImporter(t, newFakeAccount())

	summary, err := importer.Preview(context.Background(), accountSource)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}

	if summary.Liked != 3 || summary.Albums != 1 || summary.Artists != 1 || summary.Shows != 1 {
		t.Errorf("summary = %+v", summary)
	}
	want := []PlaylistSummary{{Key: "PLroad", Name: "Road trip", TrackCount: 2}, {Key: "PLfocus", Name: "Focus", TrackCount: 1200}}
	if len(summary.Playlists) != 2 || summary.Playlists[0] != want[0] || summary.Playlists[1] != want[1] {
		t.Errorf("playlists = %+v, want %+v", summary.Playlists, want)
	}
}

func TestAccountRunImportsEverything(t *testing.T) {
	importer, lib := newAccountImporter(t, newFakeAccount())
	ctx := context.Background()

	result, err := importer.Run(ctx, accountSource, allOptions("PLroad", "PLfocus"), nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if result.Liked.Matched != 3 || result.Albums.Matched != 1 || result.Artists.Matched != 1 || result.Shows.Matched != 1 || len(result.Playlists) != 2 {
		t.Errorf("result = %+v", result)
	}
	liked, err := lib.LikedTracks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// The library lists the most recent like first; YouTube's newest like stays on top.
	if len(liked) != 3 || liked[0].Title != "Newest" || liked[2].Title != "Oldest" {
		t.Errorf("liked order = %v", titlesOf(liked))
	}
	saved, err := lib.SavedItems(ctx)
	if err != nil || len(saved) != 3 {
		t.Fatalf("saved = %v, err = %v", saved, err)
	}
	playlist, err := lib.Playlist(ctx, result.Playlists[0].PlaylistID)
	if err != nil {
		t.Fatalf("playlist: %v", err)
	}
	if playlist.Name != "Road trip" || len(playlist.Entries) != 2 {
		t.Errorf("playlist = %+v", playlist)
	}
}

func titlesOf(tracks []music.Track) []string {
	titles := []string{}
	for _, track := range tracks {
		titles = append(titles, track.Title)
	}
	return titles
}

func TestAccountRunIsRepeatable(t *testing.T) {
	importer, lib := newAccountImporter(t, newFakeAccount())
	ctx := context.Background()
	options := allOptions("PLroad")

	_, err := importer.Run(ctx, accountSource, options, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := importer.Run(ctx, accountSource, options, nil)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	liked, _ := lib.LikedTracks(ctx)
	if len(liked) != 3 || liked[0].Title != "Newest" {
		t.Errorf("liked after re-sync = %v", titlesOf(liked))
	}
	playlists, _ := lib.Playlists(ctx)
	if len(playlists) != 1 || playlists[0].TrackCount != 2 {
		t.Errorf("playlists after re-sync = %+v", playlists)
	}
	if second.Playlists[0].PlaylistID != playlists[0].ID {
		t.Error("re-sync must replace the mirrored playlist, not create another")
	}
}

func TestAccountRunSelectsOnlyChosenParts(t *testing.T) {
	importer, lib := newAccountImporter(t, newFakeAccount())
	ctx := context.Background()

	result, err := importer.Run(ctx, accountSource, Options{Artists: true}, nil)
	if err != nil {
		t.Fatal(err)
	}

	liked, _ := lib.LikedTracks(ctx)
	if len(liked) != 0 || result.Liked.Matched != 0 || result.Artists.Matched != 1 {
		t.Errorf("liked = %v, result = %+v", titlesOf(liked), result)
	}
}

func TestAccountRunSkipsUnreadablePlaylist(t *testing.T) {
	account := newFakeAccount()
	account.failPlaylist = "PLfocus"
	importer, _ := newAccountImporter(t, account)

	result, err := importer.Run(context.Background(), accountSource, Options{Playlists: []string{"PLroad", "PLfocus"}}, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(result.Playlists) != 1 || len(result.Unmatched) != 1 || result.Unmatched[0].Kind != "playlist" || result.Unmatched[0].Title != "Focus" {
		t.Errorf("result = %+v", result)
	}
}

func TestAccountRunStopsWhenSessionExpired(t *testing.T) {
	account := newFakeAccount()
	account.expired = true
	importer, _ := newAccountImporter(t, account)

	_, err := importer.Run(context.Background(), accountSource, allOptions(), nil)
	if !errors.Is(err, innertube.ErrSessionExpired) {
		t.Errorf("err = %v, want ErrSessionExpired", err)
	}
}

func TestAccountRunMarksSynced(t *testing.T) {
	importer, lib := newAccountImporter(t, newFakeAccount())
	ctx := context.Background()
	err := lib.SaveYouTubeSession(ctx, "SAPISID=x", music.Account{Name: "Ada"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = importer.Run(ctx, accountSource, Options{Liked: true}, nil)
	if err != nil {
		t.Fatal(err)
	}

	session, _, _ := lib.YouTubeSession(ctx)
	if session.LastSyncedAt.IsZero() {
		t.Error("a finished import must record the sync time")
	}
}

func TestAccountWithoutProvider(t *testing.T) {
	importer, _ := newTestImporter(t, &fakeCatalog{})
	_, err := importer.Preview(context.Background(), accountSource)
	if !errors.Is(err, ErrAccountUnavailable) {
		t.Errorf("err = %v, want ErrAccountUnavailable", err)
	}
}
