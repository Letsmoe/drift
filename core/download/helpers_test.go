package download

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
	"github.com/letsmoe/drift/core/stream"
)

const waitTimeout = 5 * time.Second

// audioPayload is the fake audio of every video, big enough to be sent in several chunks.
var audioPayload = bytes.Repeat([]byte("0123456789abcdef"), 4096)

// fakeCDN serves audioPayload for /audio/{id} with Range support and /cover.jpg.
type fakeCDN struct {
	*httptest.Server
	// rejectGeneration makes requests carrying that URL generation answer 403.
	rejectGeneration int
	// rejectAll makes every audio request answer 403.
	rejectAll bool
	// breakFirst cuts the first response after breakAfter bytes.
	breakFirst bool
	breakAfter int

	mutex  sync.Mutex
	ranges []string
	broken atomic.Bool
}

func newFakeCDN(t *testing.T) *fakeCDN {
	t.Helper()
	cdn := &fakeCDN{rejectGeneration: -1}
	cdn.Server = httptest.NewServer(http.HandlerFunc(cdn.serve))
	t.Cleanup(cdn.Close)
	return cdn
}

func (cdn *fakeCDN) serve(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/cover") {
		writer.Header().Set("Content-Type", "image/jpeg")
		_, _ = writer.Write([]byte("cover-bytes"))
		return
	}
	cdn.recordRange(request.Header.Get("Range"))
	generation, _ := strconv.Atoi(request.URL.Query().Get("g"))
	if cdn.rejectAll || generation == cdn.rejectGeneration {
		writer.WriteHeader(http.StatusForbidden)
		return
	}
	var first int
	_, _ = fmt.Sscanf(request.Header.Get("Range"), "bytes=%d-", &first)
	body := audioPayload[first:]
	writer.Header().Set("Content-Type", "audio/webm")
	writer.Header().Set("Content-Length", strconv.Itoa(len(body)))
	writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", first, len(audioPayload)-1, len(audioPayload)))
	writer.WriteHeader(http.StatusPartialContent)
	if cdn.breakFirst && cdn.broken.CompareAndSwap(false, true) {
		_, _ = writer.Write(body[:cdn.breakAfter])
		writer.(http.Flusher).Flush()
		panic(http.ErrAbortHandler)
	}
	_, _ = writer.Write(body)
}

func (cdn *fakeCDN) recordRange(rangeHeader string) {
	cdn.mutex.Lock()
	defer cdn.mutex.Unlock()
	cdn.ranges = append(cdn.ranges, rangeHeader)
}

func (cdn *fakeCDN) recordedRanges() []string {
	cdn.mutex.Lock()
	defer cdn.mutex.Unlock()
	return append([]string{}, cdn.ranges...)
}

// fakeResolver points every video at the fake CDN; Invalidate moves to a new URL generation.
type fakeResolver struct {
	cdnURL string

	mutex       sync.Mutex
	generation  int
	invalidated int
}

func (resolver *fakeResolver) Resolve(_ context.Context, videoID string) (stream.Stream, error) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	return stream.Stream{
		VideoID:  videoID,
		URL:      fmt.Sprintf("%s/audio/%s?g=%d", resolver.cdnURL, videoID, resolver.generation),
		MimeType: `audio/webm; codecs="opus"`,
	}, nil
}

func (resolver *fakeResolver) Invalidate(string) {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	resolver.generation++
	resolver.invalidated++
}

func (resolver *fakeResolver) invalidations() int {
	resolver.mutex.Lock()
	defer resolver.mutex.Unlock()
	return resolver.invalidated
}

// fixture is a manager over a temporary library and downloads directory.
type fixture struct {
	t         *testing.T
	library   *library.Library
	cdn       *fakeCDN
	resolver  *fakeResolver
	directory string
	database  string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()
	database := filepath.Join(root, "library.db")
	lib, err := library.Open(t.Context(), database)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	cdn := newFakeCDN(t)
	return &fixture{
		t:         t,
		library:   lib,
		cdn:       cdn,
		resolver:  &fakeResolver{cdnURL: cdn.URL},
		directory: filepath.Join(root, "downloads"),
		database:  database,
	}
}

// startManager builds and starts a manager; it is closed when the test ends.
func (fixture *fixture) startManager(options ...Option) *Manager {
	fixture.t.Helper()
	defaults := []Option{
		WithBackoff(func(int) time.Duration { return time.Millisecond }),
		WithCoverURL(func(videoID string) string { return "http://local/cover/" + videoID }),
		WithSyncInterval(time.Hour),
	}
	manager, err := New(fixture.library, fixture.resolver, fixture.directory, append(defaults, options...)...)
	if err != nil {
		fixture.t.Fatal(err)
	}
	err = manager.Start(fixture.t.Context())
	if err != nil {
		fixture.t.Fatal(err)
	}
	fixture.t.Cleanup(manager.Close)
	return manager
}

func (fixture *fixture) track(videoID string) music.Track {
	return music.Track{
		VideoID:    videoID,
		Title:      "Title " + videoID,
		Artists:    []music.Artist{{ID: "artist", Name: "Artist"}},
		Thumbnails: []music.Thumbnail{{URL: fixture.cdn.URL + "/cover.jpg", Width: 100, Height: 100}},
	}
}

func waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

func waitForState(t *testing.T, manager *Manager, videoID string, want State) {
	t.Helper()
	waitFor(t, videoID+" to be "+string(want), func() bool {
		return stateOf(t, manager, videoID) == want
	})
}

func stateOf(t *testing.T, manager *Manager, videoID string) State {
	t.Helper()
	statuses, err := manager.States(t.Context(), []string{videoID})
	if err != nil {
		t.Fatal(err)
	}
	return statuses[videoID].State
}
