package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
)

// forwardedHeaders are copied from the CDN response so the player can seek.
var forwardedHeaders = []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"}

// fullRange is sent upstream when the player asks for the whole file: without a
// Range header the CDN throttles to roughly playback speed (~30 KiB/s instead of
// several MiB/s), which leaves the player buffering for seconds before it starts.
const fullRange = "bytes=0-"

// Proxy serves GET /stream/{videoId} by forwarding range requests to the CDN.
// Players get a stable URL; expiry and re-resolving stay inside the core.
type Proxy struct {
	resolver   *Resolver
	httpClient *http.Client
	logger     *slog.Logger
	local      atomic.Pointer[LocalFiles]
}

// NewProxy creates a proxy for resolver's streams.
func NewProxy(resolver *Resolver, logger *slog.Logger) *Proxy {
	return &Proxy{
		resolver: resolver,
		// No timeout: a response body lives as long as the track plays.
		httpClient: &http.Client{},
		logger:     logger,
	}
}

// Path returns the proxy path that plays videoID.
func Path(videoID string) string {
	return "/stream/" + url.PathEscape(videoID)
}

// Handler returns the proxy's routes.
func (proxy *Proxy) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /stream/{videoId}", proxy.serveStream)
	mux.HandleFunc("GET /cover/{videoId}", proxy.serveCover)
	return mux
}

func (proxy *Proxy) serveStream(writer http.ResponseWriter, request *http.Request) {
	videoID := request.PathValue("videoId")
	if proxy.serveLocalAudio(writer, request, videoID) {
		return
	}
	clientRange := request.Header.Get("Range")

	upstream, err := proxy.openUpstream(request.Context(), videoID, upstreamRange(clientRange))
	if err != nil {
		proxy.logger.Error("stream failed", "videoId", videoID, "error", err)
		http.Error(writer, "stream unavailable", http.StatusBadGateway)
		return
	}
	defer func() { _ = upstream.Body.Close() }()

	copyHeaders(writer.Header(), upstream.Header)
	writer.Header().Set("Cache-Control", "no-store")
	status := upstream.StatusCode
	if clientRange == "" && status == http.StatusPartialContent {
		// The player never asked for a range, so it expects a plain 200 for fullRange.
		writer.Header().Del("Content-Range")
		status = http.StatusOK
	}
	writer.WriteHeader(status)
	proxy.logger.Debug("stream", "videoId", videoID, "range", clientRange, "status", status)

	// Copy errors are the listener going away (skip, pause, quit); nothing to recover.
	_, _ = io.Copy(writer, upstream.Body)
}

// openUpstream fetches the stream, re-resolving once if the CDN rejects a cached URL.
func (proxy *Proxy) openUpstream(ctx context.Context, videoID, rangeHeader string) (*http.Response, error) {
	upstream, err := proxy.fetch(ctx, videoID, rangeHeader)
	if !errors.Is(err, errURLRejected) {
		return upstream, err
	}

	proxy.resolver.Invalidate(videoID)
	return proxy.fetch(ctx, videoID, rangeHeader)
}

var errURLRejected = errors.New("CDN rejected the stream URL")

func (proxy *Proxy) fetch(ctx context.Context, videoID, rangeHeader string) (*http.Response, error) {
	stream, err := proxy.resolver.Resolve(ctx, videoID)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, stream.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("build CDN request: %w", err)
	}
	request.Header.Set("Range", rangeHeader)

	response, err := proxy.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("CDN request: %w", err)
	}
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusGone {
		_ = response.Body.Close()
		return nil, fmt.Errorf("%w: status %d", errURLRejected, response.StatusCode)
	}
	return response, nil
}

func upstreamRange(clientRange string) string {
	if clientRange == "" {
		return fullRange
	}
	return clientRange
}

func copyHeaders(destination, source http.Header) {
	for _, name := range forwardedHeaders {
		if value := source.Get(name); value != "" {
			destination.Set(name, value)
		}
	}
}
