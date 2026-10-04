package recommend

import (
	"context"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
)

// profileOf returns the engine's current profile.
func profileOf(t *testing.T, environment *testEnvironment) *Profile {
	t.Helper()
	profile, err := environment.engine.Profile(context.Background())
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	return profile
}

// TestPlayWeightClassifiesPlays covers completed, partial and skipped plays.
func TestPlayWeightClassifiesPlays(t *testing.T) {
	old := time.Hour
	cases := []struct {
		name      string
		progress  int
		duration  int
		completed bool
		age       time.Duration
		weight    float64
		counts    bool
	}{
		{"completed flag", 10, 200, true, old, weightCompletedPlay, true},
		{"eighty percent", 160, 200, false, old, weightCompletedPlay, true},
		{"partial", 100, 200, false, old, weightPartialPlay, true},
		{"past thirty seconds is partial", 31, 600, false, old, weightPartialPlay, true},
		{"short track under thirty seconds but over half", 25, 40, false, old, weightPartialPlay, true},
		{"early skip", 5, 200, false, old, weightEarlySkip, true},
		{"fresh skip may still be starting", 5, 200, false, time.Minute, 0, false},
		{"unknown duration early skip", 5, 0, false, old, weightEarlySkip, true},
	}
	for _, testCase := range cases {
		weight, counts := playWeight(testCase.progress, testCase.duration, testCase.completed, testCase.age)
		if weight != testCase.weight || counts != testCase.counts {
			t.Errorf("%s: got (%v, %v), want (%v, %v)", testCase.name, weight, counts, testCase.weight, testCase.counts)
		}
	}
}

// TestProfileSpreadsWeightOverArtistsAndAlbums checks the credit split and normalization.
func TestProfileSpreadsWeightOverArtistsAndAlbums(t *testing.T) {
	environment := newTestEnvironment(t)
	duet := makeTrack("duet", "artist:lead", "album:duet")
	duet.Artists = append(duet.Artists, music.Artist{ID: "artist:guest", Name: "Guest"})
	environment.like(t, duet, 0)
	environment.like(t, makeTrack("solo", "artist:other", "album:solo"), 0)
	environment.like(t, makeTrack("solo2", "artist:other", "album:solo"), 0)

	profile := profileOf(t, environment)
	assertClose(t, "other", profile.Artists["artist:other"], 1)
	assertClose(t, "lead", profile.Artists["artist:lead"], 0.5)
	assertClose(t, "guest", profile.Artists["artist:guest"], 0.5/1.4142135623730951)
	assertClose(t, "solo album", profile.Albums["album:solo"], 1)
	assertClose(t, "duet album", profile.Albums["album:duet"], 0.5)
}

// TestProfileDecaysOldSignals checks the 30 day half-life.
func TestProfileDecaysOldSignals(t *testing.T) {
	environment := newTestEnvironment(t)
	environment.like(t, makeTrack("old", "artist:old", ""), AffinityHalfLife)
	environment.like(t, makeTrack("new", "artist:new", ""), 0)

	profile := profileOf(t, environment)
	assertClose(t, "new", profile.Artists["artist:new"], 1)
	assertClose(t, "old", profile.Artists["artist:old"], 0.5)
}

// TestProfileRepeatPlaysAddUp checks that every play row is its own signal.
func TestProfileRepeatPlaysAddUp(t *testing.T) {
	environment := newTestEnvironment(t)
	repeated := makeTrack("repeated", "artist:repeated", "")
	for range 3 {
		environment.play(t, repeated, time.Hour)
	}
	environment.play(t, makeTrack("once", "artist:once", ""), time.Hour)

	profile := profileOf(t, environment)
	assertClose(t, "once", profile.Artists["artist:once"], 1.0/3)
	if len(profile.Seeds) != 2 || profile.Seeds[0].VideoID != "repeated" {
		t.Errorf("seeds = %+v, want the repeated track first", profile.Seeds)
	}
}

// TestProfileKeepsDislikedArtistsNegative checks that distaste stays below zero.
func TestProfileKeepsDislikedArtistsNegative(t *testing.T) {
	environment := newTestEnvironment(t)
	environment.like(t, makeTrack("loved", "artist:loved", ""), 0)
	hated := makeTrack("hated", "artist:hated", "")
	environment.dislike(t, hated, 0)

	profile := profileOf(t, environment)
	if profile.Artists["artist:hated"] >= 0 {
		t.Errorf("hated artist affinity = %v, want negative", profile.Artists["artist:hated"])
	}
	if !profile.Disliked["hated"] {
		t.Error("the disliked track must be in Disliked")
	}
}

// TestProfileBoostsFollowedArtists checks that a follow raises an artist.
func TestProfileBoostsFollowedArtists(t *testing.T) {
	environment := newTestEnvironment(t)
	environment.like(t, makeTrack("liked", "artist:liked", ""), 0)
	err := environment.library.Save(context.Background(), music.Item{Kind: music.ItemKindArtist, ID: "artist:followed", Title: "Followed"})
	if err != nil {
		t.Fatalf("follow: %v", err)
	}

	profile := profileOf(t, environment)
	assertClose(t, "followed", profile.Artists["artist:followed"], 1)
	assertClose(t, "liked", profile.Artists["artist:liked"], weightLike/followedArtistBoost)
	if profile.ArtistNames["artist:followed"] != "Followed" {
		t.Errorf("followed artist name = %q", profile.ArtistNames["artist:followed"])
	}
}

// TestProfileCountsPlaylistAddsAndClicks checks the weaker positive signals.
func TestProfileCountsPlaylistAddsAndClicks(t *testing.T) {
	environment := newTestEnvironment(t)
	ctx := context.Background()
	added := makeTrack("added", "artist:added", "")
	playlist, err := environment.library.CreatePlaylist(ctx, "Mine")
	if err != nil {
		t.Fatalf("create playlist: %v", err)
	}
	err = environment.library.AddToPlaylist(ctx, playlist.ID, []music.Track{added})
	if err != nil {
		t.Fatalf("add to playlist: %v", err)
	}
	err = environment.library.SaveTracks(ctx, []music.Track{makeTrack("clicked", "artist:clicked", "")}, music.ItemKindSong)
	if err != nil {
		t.Fatalf("save track: %v", err)
	}
	err = environment.library.RecordImpressions(ctx, []library.Impression{{VideoID: "clicked", Source: SourceRadio, Surface: "home", Clicked: true}})
	if err != nil {
		t.Fatalf("record click: %v", err)
	}

	profile := profileOf(t, environment)
	assertClose(t, "clicked", profile.Artists["artist:clicked"], weightClick/weightPlaylistAdd)
	assertClose(t, "added", profile.Artists["artist:added"], 1)
}

// TestIgnoredImpressionsCountOnlyUntilClickedOrPlayed checks the negative impression signal and its cap.
func TestIgnoredImpressionsCountOnlyUntilClickedOrPlayed(t *testing.T) {
	environment := newTestEnvironment(t)
	ctx := context.Background()
	for _, track := range []music.Track{makeTrack("ignored", "artist:a", ""), makeTrack("played", "artist:b", "")} {
		err := environment.library.SaveTracks(ctx, []music.Track{track}, music.ItemKindSong)
		if err != nil {
			t.Fatalf("save track: %v", err)
		}
	}
	environment.ago(time.Hour, func() {
		shown := []library.Impression{}
		for range 5 {
			shown = append(shown, library.Impression{VideoID: "ignored"}, library.Impression{VideoID: "played"})
		}
		err := environment.library.RecordImpressions(ctx, shown)
		if err != nil {
			t.Fatalf("record impressions: %v", err)
		}
	})
	environment.play(t, makeTrack("played", "artist:b", ""), 30*time.Minute)

	ignored, err := environment.engine.ignoredImpressions(ctx, testNow)
	if err != nil {
		t.Fatalf("ignored impressions: %v", err)
	}
	if ignored["ignored"].count != 5 {
		t.Errorf("ignored count = %d, want 5", ignored["ignored"].count)
	}
	if _, found := ignored["played"]; found {
		t.Error("a played track's impressions must not count as ignored")
	}
	signals := ignoredImpressionSignals(ignored)
	assertClose(t, "capped weight", signals[0].weight, weightIgnoredImpression*ignoredImpressionCap)
}
