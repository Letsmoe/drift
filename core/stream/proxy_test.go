package stream

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/innertube"
)

const audioBytes = "0123456789"

// fakeCDN serves audioBytes for range requests; the first rejectFirst requests get 403.
type fakeCDN struct {
	*httptest.Server
	rejectFirst  int32
	requestCount atomic.Int32
	lastRange    atomic.Value
}

func newFakeCDN(t *testing.T, rejectFirst int32) *fakeCDN {
	t.Helper()
	cdn := &fakeCDN{rejectFirst: rejectFirst}
	cdn.Server = httptest.NewServer(http.HandlerFunc(cdn.serve))
	t.Cleanup(cdn.Close)
	return cdn
}

func (cdn *fakeCDN) serve(writer http.ResponseWriter, request *http.Request) {
	rangeHeader := request.Header.Get("Range")
	cdn.lastRange.Store(rangeHeader)
	if cdn.requestCount.Add(1) <= cdn.rejectFirst {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	first, last := parseRange(rangeHeader)
	writer.Header().Set("Content-Type", "audio/webm")
	writer.Header().Set("Content-Length", strconv.Itoa(last-first+1))
	writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", first, last, len(audioBytes)))
	writer.Header().Set("X-Internal", "not forwarded")
	writer.WriteHeader(http.StatusPartialContent)
	_, _ = writer.Write([]byte(audioBytes[first : last+1]))
}

// parseRange reads "bytes=first-" or "bytes=first-last"; anything else means the whole file.
func parseRange(rangeHeader string) (int, int) {
	first, last := 0, len(audioBytes)-1
	if _, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &first, &last); err != nil {
		return first, len(audioBytes) - 1
	}
	return first, last
}

func newTestProxy(cdnURL string) (*Proxy, *fakePlayers) {
	players := &fakePlayers{players: map[string]innertube.Player{
		"PRIMARY": {
			Formats:   []innertube.Format{{MimeType: "audio/webm", URL: cdnURL}},
			ExpiresIn: time.Hour,
		},
	}}
	resolver := NewResolver(players, WithProfiles(primaryProfile))
	return NewProxy(resolver, slog.New(slog.DiscardHandler)), players
}

// proxyResponse is a fully read proxy response.
type proxyResponse struct {
	statusCode int
	header     http.Header
	body       string
}

func requestRange(t *testing.T, proxy *Proxy, rangeHeader string) proxyResponse {
	t.Helper()
	return fetchProxy(t, proxy, Path("video"), rangeHeader)
}

func requestPath(t *testing.T, proxy *Proxy, path string) proxyResponse {
	t.Helper()
	return fetchProxy(t, proxy, path, "")
}

func fetchProxy(t *testing.T, proxy *Proxy, path, rangeHeader string) proxyResponse {
	t.Helper()
	server := httptest.NewServer(proxy.Handler())
	t.Cleanup(server.Close)

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rangeHeader != "" {
		request.Header.Set("Range", rangeHeader)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return proxyResponse{statusCode: response.StatusCode, header: response.Header, body: string(body)}
}

func TestProxyForwardsRangeRequests(t *testing.T) {
	cdn := newFakeCDN(t, 0)
	proxy, _ := newTestProxy(cdn.URL)

	response := requestRange(t, proxy, "bytes=2-5")

	if cdn.lastRange.Load() != "bytes=2-5" {
		t.Errorf("upstream Range = %q, want bytes=2-5", cdn.lastRange.Load())
	}
	if response.statusCode != http.StatusPartialContent || response.body != "2345" {
		t.Errorf("got %d %q, want 206 \"2345\"", response.statusCode, response.body)
	}
	if response.header.Get("Content-Range") != "bytes 2-5/10" {
		t.Errorf("Content-Range = %q", response.header.Get("Content-Range"))
	}
	if response.header.Get("X-Internal") != "" {
		t.Error("forwarded a header outside the allow list")
	}
}

func TestProxyReResolvesRejectedURL(t *testing.T) {
	cdn := newFakeCDN(t, 1)
	proxy, players := newTestProxy(cdn.URL)

	response := requestRange(t, proxy, "bytes=2-5")

	if response.statusCode != http.StatusPartialContent {
		t.Errorf("status = %d, want 206", response.statusCode)
	}
	if cdn.requestCount.Load() != 2 || len(players.calls) != 2 {
		t.Errorf("CDN requests = %d, resolves = %d, want 2 and 2", cdn.requestCount.Load(), len(players.calls))
	}
}

func TestProxyReportsUnresolvableStream(t *testing.T) {
	cdn := newFakeCDN(t, 2)
	proxy, _ := newTestProxy(cdn.URL)

	response := requestRange(t, proxy, "bytes=2-5")

	if response.statusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", response.statusCode)
	}
}

func TestProxyRequestsFullRangeWhenPlayerSendsNone(t *testing.T) {
	cdn := newFakeCDN(t, 0)
	proxy, _ := newTestProxy(cdn.URL)

	response := requestRange(t, proxy, "")

	if cdn.lastRange.Load() != fullRange {
		t.Errorf("upstream Range = %q, want %q", cdn.lastRange.Load(), fullRange)
	}
	if response.statusCode != http.StatusOK || response.body != audioBytes {
		t.Errorf("got %d %q, want 200 with the whole file", response.statusCode, response.body)
	}
	if response.header.Get("Content-Range") != "" {
		t.Errorf("Content-Range = %q on a 200", response.header.Get("Content-Range"))
	}
	if response.header.Get("Content-Length") != strconv.Itoa(len(audioBytes)) {
		t.Errorf("Content-Length = %q", response.header.Get("Content-Length"))
	}
}
