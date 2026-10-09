package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/letsmoe/drift/core/innertube"
)

// fakePlayers answers Player calls per profile name and counts them.
type fakePlayers struct {
	players map[string]innertube.Player
	errors  map[string]error
	calls   []string
}

func (players *fakePlayers) Player(_ context.Context, profile innertube.Profile, _ string) (innertube.Player, error) {
	players.calls = append(players.calls, profile.Name)
	if err, failed := players.errors[profile.Name]; failed {
		return innertube.Player{}, err
	}
	return players.players[profile.Name], nil
}

var (
	opus    = innertube.Format{Itag: 251, MimeType: `audio/webm; codecs="opus"`, Bitrate: 141465, URL: "https://cdn/251"}
	opusLow = innertube.Format{Itag: 250, MimeType: `audio/webm; codecs="opus"`, Bitrate: 71940, URL: "https://cdn/250"}
	aac     = innertube.Format{Itag: 140, MimeType: `audio/mp4; codecs="mp4a.40.2"`, Bitrate: 130993, URL: "https://cdn/140"}
	video   = innertube.Format{Itag: 137, MimeType: `video/mp4; codecs="avc1"`, Bitrate: 999999, URL: "https://cdn/137"}
)

var (
	primaryProfile  = innertube.Profile{Name: "PRIMARY"}
	fallbackProfile = innertube.Profile{Name: "FALLBACK"}
)

func TestResolvePrefersOpusWithHighestBitrate(t *testing.T) {
	players := &fakePlayers{players: map[string]innertube.Player{
		"PRIMARY": {Formats: []innertube.Format{video, aac, opusLow, opus}, ExpiresIn: time.Hour},
	}}
	resolver := NewResolver(players, WithProfiles(primaryProfile))

	stream, err := resolver.Resolve(t.Context(), "video")
	if err != nil {
		t.Fatal(err)
	}
	if stream.URL != opus.URL || stream.Profile != "PRIMARY" {
		t.Errorf("stream = %+v, want itag 251 from PRIMARY", stream)
	}
}

func TestResolveHonorsMimePreference(t *testing.T) {
	players := &fakePlayers{players: map[string]innertube.Player{
		"PRIMARY": {Formats: []innertube.Format{opus, aac}, ExpiresIn: time.Hour},
	}}
	resolver := NewResolver(players, WithProfiles(primaryProfile), WithMimePreference("audio/mp4"))

	stream, err := resolver.Resolve(t.Context(), "video")
	if err != nil {
		t.Fatal(err)
	}
	if stream.URL != aac.URL {
		t.Errorf("URL = %s, want AAC", stream.URL)
	}
}

func TestResolveFallsBackToNextProfile(t *testing.T) {
	players := &fakePlayers{
		errors: map[string]error{"PRIMARY": &innertube.PlayabilityError{Status: "LOGIN_REQUIRED"}},
		players: map[string]innertube.Player{
			"FALLBACK": {Formats: []innertube.Format{aac}, ExpiresIn: time.Hour},
		},
	}
	resolver := NewResolver(players, WithProfiles(primaryProfile, fallbackProfile))

	stream, err := resolver.Resolve(t.Context(), "video")
	if err != nil {
		t.Fatal(err)
	}
	if stream.Profile != "FALLBACK" {
		t.Errorf("profile = %s, want FALLBACK", stream.Profile)
	}
}

func TestResolveReportsEveryProfileFailure(t *testing.T) {
	players := &fakePlayers{
		errors: map[string]error{"PRIMARY": &innertube.PlayabilityError{Status: "LOGIN_REQUIRED"}},
		players: map[string]innertube.Player{
			"FALLBACK": {Formats: []innertube.Format{video}, ExpiresIn: time.Hour},
		},
	}
	resolver := NewResolver(players, WithProfiles(primaryProfile, fallbackProfile))

	_, err := resolver.Resolve(t.Context(), "video")

	var playabilityError *innertube.PlayabilityError
	if !errors.As(err, &playabilityError) || !errors.Is(err, ErrNoAudioFormat) {
		t.Errorf("err = %v, want both profile failures", err)
	}
}

func TestResolveCachesUntilExpiry(t *testing.T) {
	players := &fakePlayers{players: map[string]innertube.Player{
		"PRIMARY": {Formats: []innertube.Format{opus}, ExpiresIn: time.Hour},
	}}
	resolver := NewResolver(players, WithProfiles(primaryProfile))
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	resolver.now = func() time.Time { return now }

	resolveOrFail(t, resolver)
	resolveOrFail(t, resolver)
	if len(players.calls) != 1 {
		t.Fatalf("player called %d times while fresh, want 1", len(players.calls))
	}

	now = now.Add(time.Hour - expiryMargin + time.Second)
	resolveOrFail(t, resolver)
	if len(players.calls) != 2 {
		t.Errorf("player called %d times after expiry, want 2", len(players.calls))
	}
}

func TestInvalidateForcesNewResolve(t *testing.T) {
	players := &fakePlayers{players: map[string]innertube.Player{
		"PRIMARY": {Formats: []innertube.Format{opus}, ExpiresIn: time.Hour},
	}}
	resolver := NewResolver(players, WithProfiles(primaryProfile))

	resolveOrFail(t, resolver)
	resolver.Invalidate("video")
	resolveOrFail(t, resolver)

	if len(players.calls) != 2 {
		t.Errorf("player called %d times, want 2", len(players.calls))
	}
}

func resolveOrFail(t *testing.T, resolver *Resolver) {
	t.Helper()
	if _, err := resolver.Resolve(t.Context(), "video"); err != nil {
		t.Fatal(err)
	}
}
