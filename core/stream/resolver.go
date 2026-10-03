// Package stream turns video IDs into playable audio: it resolves stream URLs
// through a chain of InnerTube profiles and serves them through a local proxy.
package stream

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/neoworks/otter-music/core/innertube"
)

// expiryMargin renews URLs before YouTube's expiry so a long track never starts on a dying URL.
const expiryMargin = 10 * time.Minute

// ErrNoAudioFormat means a profile returned streams, but none with plain-URL audio.
var ErrNoAudioFormat = errors.New("no audio format with a plain URL")

// Stream is a resolved, directly fetchable audio stream.
type Stream struct {
	VideoID       string
	URL           string
	MimeType      string
	Bitrate       int
	ContentLength int64
	ExpiresAt     time.Time
	// Profile is the InnerTube client that produced the URL.
	Profile string
}

type playerSource interface {
	Player(ctx context.Context, profile innertube.Profile, videoID string) (innertube.Player, error)
}

// Resolver finds the best audio stream for a video. It is safe for concurrent use.
type Resolver struct {
	players        playerSource
	profiles       []innertube.Profile
	mimePreference []string
	now            func() time.Time

	cacheMutex sync.Mutex
	cache      map[string]Stream
}

// ResolverOption configures a Resolver.
type ResolverOption func(*Resolver)

// WithProfiles sets the InnerTube profiles tried in order until one yields a stream.
func WithProfiles(profiles ...innertube.Profile) ResolverOption {
	return func(resolver *Resolver) {
		resolver.profiles = profiles
	}
}

// WithMimePreference sets the container preference, e.g. "audio/mp4" first for
// AVPlayer, which cannot play WebM. Within a container the highest bitrate wins.
func WithMimePreference(mimeTypes ...string) ResolverOption {
	return func(resolver *Resolver) {
		resolver.mimePreference = mimeTypes
	}
}

// NewResolver creates a resolver that uses VisionOS and prefers Opus (WebM) over AAC (MP4).
func NewResolver(players playerSource, options ...ResolverOption) *Resolver {
	resolver := &Resolver{
		players:        players,
		profiles:       []innertube.Profile{innertube.VisionOS},
		mimePreference: []string{"audio/webm", "audio/mp4"},
		now:            time.Now,
		cache:          map[string]Stream{},
	}
	for _, option := range options {
		option(resolver)
	}
	return resolver
}

// Resolve returns a cached stream for videoID while it is fresh, otherwise asks
// each profile in order and returns the first stream found.
func (resolver *Resolver) Resolve(ctx context.Context, videoID string) (Stream, error) {
	if cached, found := resolver.cached(videoID); found {
		return cached, nil
	}

	var failures []error
	for _, profile := range resolver.profiles {
		stream, err := resolver.resolveWith(ctx, profile, videoID)
		if err == nil {
			resolver.store(stream)
			return stream, nil
		}
		failures = append(failures, err)
	}
	return Stream{}, fmt.Errorf("resolve %s: %w", videoID, errors.Join(failures...))
}

// Invalidate drops the cached stream, e.g. after the CDN rejected its URL.
func (resolver *Resolver) Invalidate(videoID string) {
	resolver.cacheMutex.Lock()
	defer resolver.cacheMutex.Unlock()
	delete(resolver.cache, videoID)
}

func (resolver *Resolver) resolveWith(ctx context.Context, profile innertube.Profile, videoID string) (Stream, error) {
	player, err := resolver.players.Player(ctx, profile, videoID)
	if err != nil {
		return Stream{}, err
	}

	format, found := resolver.bestAudioFormat(player.Formats)
	if !found {
		return Stream{}, fmt.Errorf("%s: %w", profile.Name, ErrNoAudioFormat)
	}

	return Stream{
		VideoID:       videoID,
		URL:           format.URL,
		MimeType:      format.MimeType,
		Bitrate:       format.Bitrate,
		ContentLength: format.ContentLength,
		ExpiresAt:     resolver.now().Add(player.ExpiresIn - expiryMargin),
		Profile:       profile.Name,
	}, nil
}

func (resolver *Resolver) bestAudioFormat(formats []innertube.Format) (innertube.Format, bool) {
	for _, mimeType := range resolver.mimePreference {
		format, found := highestBitrate(formats, mimeType)
		if found {
			return format, true
		}
	}
	return innertube.Format{}, false
}

func highestBitrate(formats []innertube.Format, mimeTypePrefix string) (innertube.Format, bool) {
	var best innertube.Format
	found := false
	for _, format := range formats {
		if !strings.HasPrefix(format.MimeType, mimeTypePrefix) {
			continue
		}
		if !found || format.Bitrate > best.Bitrate {
			best = format
			found = true
		}
	}
	return best, found
}

func (resolver *Resolver) cached(videoID string) (Stream, bool) {
	resolver.cacheMutex.Lock()
	defer resolver.cacheMutex.Unlock()

	stream, found := resolver.cache[videoID]
	if !found || resolver.now().After(stream.ExpiresAt) {
		return Stream{}, false
	}
	return stream, true
}

func (resolver *Resolver) store(stream Stream) {
	resolver.cacheMutex.Lock()
	defer resolver.cacheMutex.Unlock()
	resolver.cache[stream.VideoID] = stream
}
