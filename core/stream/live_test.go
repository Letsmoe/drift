//go:build live

// Live tests hit real YouTube. They run in the daily canary workflow to catch
// breakage before users do: go test -tags live ./...
package stream

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/letsmoe/drift-music/core/innertube"
)

// liveDownloadBytes is past the CDN's 1 MiB window that unauthenticated clients get cut off at.
const liveDownloadBytes = 2 << 20

func TestLiveSearchResolveAndDownload(t *testing.T) {
	client := innertube.NewClient()

	tracks, err := client.SearchSongs(t.Context(), "daft punk around the world")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(tracks) == 0 {
		t.Fatal("search returned no songs; the response layout probably changed")
	}

	proxy := NewProxy(NewResolver(client), slog.Default())
	server := httptest.NewServer(proxy.Handler())
	defer server.Close()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+Path(tracks[0].VideoID), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Range", "bytes=0-")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusPartialContent && response.StatusCode != http.StatusOK {
		t.Fatalf("stream %s: status %d", tracks[0].VideoID, response.StatusCode)
	}
	expectedBytes := expectedDownload(response.ContentLength)
	downloaded, err := io.Copy(io.Discard, io.LimitReader(response.Body, expectedBytes))
	if err != nil || downloaded < expectedBytes {
		t.Fatalf("downloaded %d of %d bytes: %v", downloaded, expectedBytes, err)
	}
}

func expectedDownload(contentLength int64) int64 {
	if contentLength > 0 && contentLength < liveDownloadBytes {
		return contentLength
	}
	return liveDownloadBytes
}
