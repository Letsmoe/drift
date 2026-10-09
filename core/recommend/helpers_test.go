package recommend

import (
	"context"
	"errors"
	"math"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/letsmoe/drift-music/core/library"
	"github.com/letsmoe/drift-music/core/music"
)

// testNow is the fixed clock every engine test starts at.
var testNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// testClock is a clock tests move by hand; the library and engine share it.
type testClock struct {
	mutex   sync.Mutex
	current time.Time
}

// now returns the current test time.
func (clock *testClock) now() time.Time {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	return clock.current
}

// set moves the clock.
func (clock *testClock) set(moment time.Time) {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	clock.current = moment
}

// fakeCatalog answers from fixed data, counts calls and can fail per method.
type fakeCatalog struct {
	mutex    sync.Mutex
	radios   map[string][]music.Track
	artists  map[string]music.ArtistPage
	searches map[string][]music.Track
	home     music.HomePage
	failing  map[string]bool
	calls    map[string]int
}

// newFakeCatalog returns an empty catalog.
func newFakeCatalog() *fakeCatalog {
	return &fakeCatalog{
		radios:   map[string][]music.Track{},
		artists:  map[string]music.ArtistPage{},
		searches: map[string][]music.Track{},
		failing:  map[string]bool{},
		calls:    map[string]int{},
	}
}

// record counts a call and reports the configured failure of the method.
func (catalog *fakeCatalog) record(method string) error {
	catalog.mutex.Lock()
	defer catalog.mutex.Unlock()
	catalog.calls[method]++
	if catalog.failing[method] {
		return errors.New("catalog down")
	}
	return nil
}

// callCount returns how often a method was called.
func (catalog *fakeCatalog) callCount(method string) int {
	catalog.mutex.Lock()
	defer catalog.mutex.Unlock()
	return catalog.calls[method]
}

// Radio implements Catalog.
func (catalog *fakeCatalog) Radio(_ context.Context, videoID string) ([]music.Track, error) {
	return catalog.radios[videoID], catalog.record("radio")
}

// Artist implements Catalog.
func (catalog *fakeCatalog) Artist(_ context.Context, browseID string) (music.ArtistPage, error) {
	return catalog.artists[browseID], catalog.record("artist")
}

// SearchSongs implements Catalog.
func (catalog *fakeCatalog) SearchSongs(_ context.Context, query string) ([]music.Track, error) {
	return catalog.searches[query], catalog.record("search")
}

// Home implements Catalog.
func (catalog *fakeCatalog) Home(_ context.Context) (music.HomePage, error) {
	return catalog.home, catalog.record("home")
}

// testEnvironment bundles an engine with the library and catalog behind it.
type testEnvironment struct {
	engine  *Engine
	library *library.Library
	catalog *fakeCatalog
	clock   *testClock
}

// newTestEnvironment opens a fresh library and an engine on a fixed clock.
func newTestEnvironment(t *testing.T) *testEnvironment {
	t.Helper()
	clock := &testClock{current: testNow}
	lib, err := library.Open(context.Background(), filepath.Join(t.TempDir(), "library.db"), library.WithClock(clock.now))
	if err != nil {
		t.Fatalf("open library: %v", err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	catalog := newFakeCatalog()
	engine := New(lib, catalog, WithClock(clock.now))
	return &testEnvironment{engine: engine, library: lib, catalog: catalog, clock: clock}
}

// ago runs action with the clock moved age into the past, so rows it writes
// carry that timestamp.
func (environment *testEnvironment) ago(age time.Duration, action func()) {
	environment.clock.set(testNow.Add(-age))
	action()
	environment.clock.set(testNow)
}

// makeTrack builds a 200 second track by one artist ("artist:<artistID>") on
// an album ("album:<albumID>"); empty IDs leave the credit or album out.
func makeTrack(videoID, artistID, albumIdentifier string) music.Track {
	track := music.Track{VideoID: videoID, Title: "Title " + videoID, DurationSeconds: 200, Thumbnails: []music.Thumbnail{}}
	if artistID != "" {
		track.Artists = []music.Artist{{ID: artistID, Name: "Name " + artistID}}
	}
	if albumIdentifier != "" {
		track.Album = &music.Album{ID: albumIdentifier, Name: "Album " + albumIdentifier}
	}
	return track
}

// like rates a track thumbs-up, age ago.
func (environment *testEnvironment) like(t *testing.T, track music.Track, age time.Duration) {
	t.Helper()
	environment.rate(t, track, library.RatingLike, age)
}

// dislike rates a track thumbs-down, age ago.
func (environment *testEnvironment) dislike(t *testing.T, track music.Track, age time.Duration) {
	t.Helper()
	environment.rate(t, track, library.RatingDislike, age)
}

// rate stores a rating, age ago.
func (environment *testEnvironment) rate(t *testing.T, track music.Track, rating library.Rating, age time.Duration) {
	t.Helper()
	environment.ago(age, func() {
		err := environment.library.SetRating(context.Background(), track, music.ItemKindSong, rating)
		if err != nil {
			t.Fatalf("rate %s: %v", track.VideoID, err)
		}
	})
}

// play stores one completed play, age ago.
func (environment *testEnvironment) play(t *testing.T, track music.Track, age time.Duration) {
	t.Helper()
	environment.ago(age, func() {
		ctx := context.Background()
		playID, err := environment.library.StartPlay(ctx, track, music.ItemKindSong, "queue")
		if err != nil {
			t.Fatalf("start play of %s: %v", track.VideoID, err)
		}
		err = environment.library.UpdatePlay(ctx, playID, track.DurationSeconds, true)
		if err != nil {
			t.Fatalf("update play of %s: %v", track.VideoID, err)
		}
	})
}

// feed returns the first page of a freshly ranked feed.
func (environment *testEnvironment) feed(t *testing.T) []Item {
	t.Helper()
	items, err := environment.engine.Feed(context.Background(), 100, 0, true)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	return items
}

// itemByVideoID finds a track in the items; nil when absent.
func itemByVideoID(items []Item, videoID string) *Item {
	for index := range items {
		if items[index].Track.VideoID == videoID {
			return &items[index]
		}
	}
	return nil
}

// assertClose fails when got differs from want by more than a rounding error.
func assertClose(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

// songItem wraps a track as a home shelf song item.
func songItem(track music.Track) music.Item {
	return music.Item{Kind: music.ItemKindSong, ID: track.VideoID, Title: track.Title, Track: &track}
}
