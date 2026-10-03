package stream

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/innertube"
)

const audioBytes = "0123456789"

// newFakeCDN serves audioBytes for range requests; the first rejectFirst requests get 403.
func newFakeCDN(t *testing.T, rejectFirst int32) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if requestCount.Add(1) <= rejectFirst {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		if request.Header.Get("Range") != "bytes=2-5" {
			t.Errorf("Range = %q, want bytes=2-5", request.Header.Get("Range"))
		}
		writer.Header().Set("Content-Type", "audio/webm")
		writer.Header().Set("Content-Range", "bytes 2-5/10")
		writer.Header().Set("X-Internal", "not forwarded")
		writer.WriteHeader(http.StatusPartialContent)
		_, _ = writer.Write([]byte(audioBytes[2:6]))
	}))
	t.Cleanup(server.Close)
	return server, &requestCount
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

func requestRange(t *testing.T, proxy *Proxy) proxyResponse {
	t.Helper()
	server := httptest.NewServer(proxy.Handler())
	t.Cleanup(server.Close)

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL+Path("video"), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Range", "bytes=2-5")
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
	cdn, _ := newFakeCDN(t, 0)
	proxy, _ := newTestProxy(cdn.URL)

	response := requestRange(t, proxy)

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
	cdn, cdnRequests := newFakeCDN(t, 1)
	proxy, players := newTestProxy(cdn.URL)

	response := requestRange(t, proxy)

	if response.statusCode != http.StatusPartialContent {
		t.Errorf("status = %d, want 206", response.statusCode)
	}
	if cdnRequests.Load() != 2 || len(players.calls) != 2 {
		t.Errorf("CDN requests = %d, resolves = %d, want 2 and 2", cdnRequests.Load(), len(players.calls))
	}
}

func TestProxyReportsUnresolvableStream(t *testing.T) {
	cdn, _ := newFakeCDN(t, 2)
	proxy, _ := newTestProxy(cdn.URL)

	response := requestRange(t, proxy)

	if response.statusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", response.statusCode)
	}
}
