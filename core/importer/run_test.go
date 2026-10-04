package importer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
)

// fakeCatalog answers song searches from a title table and counts calls.
type fakeCatalog struct {
	calls atomic.Int32
	fail  bool
}

func (catalog *fakeCatalog) SearchSongs(_ context.Context, query string) ([]music.Track, error) {
	catalog.calls.Add(1)
	if catalog.fail {
		return nil, errors.New("offline")
	}
	songs := map[string]music.Track{
		"hold up beyonce":          song("Hold Up", "Beyoncé"),
		"midnight city m83":        song("Midnight City", "M83"),
		"redbone childish gambino": song("Redbone", "Childish Gambino"),
	}
	candidates := []music.Track{}
	for key, track := range songs {
		if foldText(query) == key {
			track.VideoID = "video-" + key
			candidates = append(candidates, track)
		}
	}
	return candidates, nil
}

func (catalog *fakeCatalog) Search(_ context.Context, _ string) (music.SearchResults, error) {
	catalog.calls.Add(1)
	if catalog.fail {
		return music.SearchResults{}, errors.New("offline")
	}
	items := []music.Item{
		{Kind: music.ItemKindAlbum, ID: "MPREb_ok", Title: "OK Computer", Subtitle: "Album • Radiohead • 1997"},
		{Kind: music.ItemKindArtist, ID: "UCbonobo", Title: "Bonobo"},
		{Kind: music.ItemKindPodcast, ID: "MPSPPLdark", Title: "Darknet Diaries", Subtitle: "Podcast"},
	}
	return music.SearchResults{Shelves: []music.Shelf{{Title: "Results", Items: items}}}, nil
}

func newTestImporter(t *testing.T, catalog Catalog) (*Importer, *library.Library) {
	t.Helper()
	lib, err := library.Open(context.Background(), filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatalf("open library: %v", err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(lib, catalog, logger, WithRequestDelay(0)), lib
}

var exportSource = Source{Kind: SourceExport, Location: filepath.Join("..", "spotify", "testdata", "export")}

func allOptions(playlistKeys ...string) Options {
	return Options{Liked: true, Playlists: playlistKeys, Albums: true, Artists: true, Shows: true}
}

func TestPreview(t *testing.T) {
	importer, _ := newTestImporter(t, &fakeCatalog{})
	summary, err := importer.Preview(context.Background(), exportSource)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if summary.Liked != 3 || summary.Albums != 1 || summary.Artists != 1 || summary.Shows != 1 || len(summary.Playlists) != 2 {
		t.Errorf("summary = %+v", summary)
	}
	if summary.Playlists[0].TrackCount != 2 || summary.TrackLimit != 0 {
		t.Errorf("playlist summary = %+v, limit %d", summary.Playlists[0], summary.TrackLimit)
	}
}

func TestRunImportsEverything(t *testing.T) {
	catalog := &fakeCatalog{}
	importer, lib := newTestImporter(t, catalog)
	summary, _ := importer.Preview(context.Background(), exportSource)

	var last Progress
	result, err := importer.Run(context.Background(), exportSource, allOptions(summary.Playlists[0].Key), func(progress Progress) { last = progress })
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.Liked.Matched != 1 || result.Liked.Unmatched != 2 {
		t.Errorf("liked = %+v, want 1 matched (Hold Up) and 2 unmatched", result.Liked)
	}
	if result.Albums.Matched != 1 || result.Artists.Matched != 1 || result.Shows.Matched != 1 {
		t.Errorf("saved = %+v %+v %+v", result.Albums, result.Artists, result.Shows)
	}
	if len(result.Playlists) != 1 || result.Playlists[0].Matched != 2 {
		t.Fatalf("playlists = %+v", result.Playlists)
	}
	if last.Phase != PhaseSaving || last.Done != last.Total {
		t.Errorf("last progress = %+v", last)
	}
	assertLibrary(t, lib, result.Playlists[0].PlaylistID)
}

func assertLibrary(t *testing.T, lib *library.Library, playlistID string) {
	t.Helper()
	liked, err := lib.LikedTracks(context.Background())
	if err != nil || len(liked) != 1 || liked[0].Title != "Hold Up" {
		t.Errorf("liked = %+v, %v", liked, err)
	}
	playlist, err := lib.Playlist(context.Background(), playlistID)
	if err != nil {
		t.Fatalf("playlist: %v", err)
	}
	if playlist.Name != "Road Trip" || playlist.Description != "Windows down" || len(playlist.Entries) != 2 ||
		playlist.Entries[0].Track.Title != "Midnight City" || playlist.Entries[1].Track.Title != "Redbone" {
		t.Errorf("playlist = %+v", playlist)
	}
	saved, err := lib.SavedItems(context.Background())
	if err != nil || len(saved) != 3 {
		t.Errorf("saved = %+v, %v", saved, err)
	}
}

func TestRunTwiceIsIdempotentAndReusesMatches(t *testing.T) {
	catalog := &fakeCatalog{}
	importer, lib := newTestImporter(t, catalog)
	summary, _ := importer.Preview(context.Background(), exportSource)
	options := allOptions(summary.Playlists[0].Key)

	first, err := importer.Run(context.Background(), exportSource, options, nil)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	callsAfterFirst := catalog.calls.Load()
	second, err := importer.Run(context.Background(), exportSource, options, nil)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if catalog.calls.Load() != callsAfterFirst {
		t.Errorf("second run searched again: %d calls, was %d", catalog.calls.Load(), callsAfterFirst)
	}
	if second.Playlists[0].PlaylistID != first.Playlists[0].PlaylistID {
		t.Error("second run created another playlist")
	}
	playlists, _ := lib.Playlists(context.Background())
	if len(playlists) != 1 || playlists[0].TrackCount != 2 {
		t.Errorf("playlists after re-import = %+v", playlists)
	}
	liked, _ := lib.LikedTracks(context.Background())
	if len(liked) != 1 {
		t.Errorf("liked after re-import = %d", len(liked))
	}
}

func TestRetryUnmatchedSearchesAgain(t *testing.T) {
	catalog := &fakeCatalog{}
	importer, _ := newTestImporter(t, catalog)
	options := Options{Liked: true}
	_, err := importer.Run(context.Background(), exportSource, options, nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	calls := catalog.calls.Load()
	options.RetryUnmatched = true
	_, err = importer.Run(context.Background(), exportSource, options, nil)
	if err != nil {
		t.Fatalf("retry run: %v", err)
	}
	if catalog.calls.Load() != calls+2 {
		t.Errorf("retry searched %d times, want 2 (the unmatched tracks)", catalog.calls.Load()-calls)
	}
}

func TestRunStopsWhenCatalogKeepsFailing(t *testing.T) {
	importer, lib := newTestImporter(t, &fakeCatalog{fail: true})
	importer.requestDelay = 0
	_, err := importer.Run(context.Background(), exportSource, allOptions(), nil)
	if !errors.Is(err, ErrCatalogUnavailable) {
		t.Fatalf("err = %v", err)
	}
	liked, _ := lib.LikedTracks(context.Background())
	if len(liked) != 0 {
		t.Error("a failed run wrote likes")
	}
}

func TestRunCancelled(t *testing.T) {
	importer, _ := newTestImporter(t, &fakeCatalog{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := importer.Run(ctx, exportSource, allOptions(), nil)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestRunRejectsConcurrentImport(t *testing.T) {
	importer, _ := newTestImporter(t, &fakeCatalog{})
	importer.running.Lock()
	defer importer.running.Unlock()
	_, err := importer.Run(context.Background(), exportSource, allOptions(), nil)
	if !errors.Is(err, ErrBusy) {
		t.Errorf("err = %v, want ErrBusy", err)
	}
}

func TestUnknownSource(t *testing.T) {
	importer, _ := newTestImporter(t, &fakeCatalog{})
	_, err := importer.Preview(context.Background(), Source{Kind: "ftp"})
	if !errors.Is(err, ErrUnknownSource) {
		t.Errorf("err = %v", err)
	}
}

func TestStartReportsDoneAndCancel(t *testing.T) {
	importer, _ := newTestImporter(t, &fakeCatalog{})
	done := make(chan error, 1)
	err := importer.Start(context.Background(), exportSource, Options{Liked: true}, nil, func(_ Result, err error) { done <- err })
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if err = <-done; err != nil {
		t.Errorf("done error = %v", err)
	}
	importer.Cancel()
}
