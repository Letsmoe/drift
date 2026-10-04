package recommend

import (
	"context"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

// seedListening gives the user one completed play of a track by artist:seed
// and wires the catalog around it: a radio, an artist page with a related
// artist, a search and a home feed.
func seedListening(t *testing.T, environment *testEnvironment) {
	t.Helper()
	seed := makeTrack("seed", "artist:seed", "album:seed")
	environment.play(t, seed, time.Hour)
	catalog := environment.catalog
	catalog.radios["seed"] = []music.Track{seed, makeTrack("radio1", "artist:r1", "album:r1"), makeTrack("radio2", "artist:r2", "album:r2")}
	catalog.artists["artist:seed"] = music.ArtistPage{
		ID: "artist:seed", Name: "Seed Artist",
		TopSongs: []music.Track{makeTrack("top1", "artist:seed", "album:t1"), makeTrack("top2", "artist:seed", "album:t2")},
		Shelves: []music.Shelf{
			{Title: "Albums", Items: []music.Item{{Kind: music.ItemKindAlbum, ID: "album:x"}}},
			{Title: "Fans might also like", Items: []music.Item{
				{Kind: music.ItemKindArtist, ID: "artist:seed", Title: "Seed Artist"},
				{Kind: music.ItemKindArtist, ID: "artist:new", Title: "New Artist"},
			}},
		},
	}
	catalog.artists["artist:new"] = music.ArtistPage{ID: "artist:new", Name: "New Artist", TopSongs: []music.Track{makeTrack("new1", "artist:new", "album:n1")}}
	catalog.home = music.HomePage{Shelves: []music.Shelf{{Title: "Quick picks", Items: []music.Item{songItem(makeTrack("home1", "artist:h", "album:h"))}}}}
}

// TestFeedCombinesEverySource checks that each source contributes with its reason.
func TestFeedCombinesEverySource(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	environment.catalog.searches["lofi beats"] = []music.Track{makeTrack("found1", "artist:f", "album:f")}
	err := environment.library.RecordSearch(context.Background(), "lofi beats")
	if err != nil {
		t.Fatalf("record search: %v", err)
	}
	loved := makeTrack("loved", "artist:l", "album:l")
	environment.like(t, loved, 40*24*time.Hour)

	items := environment.feed(t)
	wants := map[string]struct{ source, reason string }{
		"radio1": {SourceRadio, "Because you listened to Title seed"},
		"top1":   {SourceArtist, "More from Seed Artist"},
		"new1":   {SourceExplore, "Explore New Artist"},
		"found1": {SourceSearch, "From your search “lofi beats”"},
		"loved":  {SourceRediscover, "Rediscover"},
		"home1":  {SourceYouTube, "Popular on YouTube Music"},
	}
	for videoID, want := range wants {
		item := itemByVideoID(items, videoID)
		if item == nil {
			t.Errorf("%s missing from the feed", videoID)
			continue
		}
		if item.Source != want.source || item.Reason.Text != want.reason {
			t.Errorf("%s: source %q reason %q, want %q %q", videoID, item.Source, item.Reason.Text, want.source, want.reason)
		}
	}
}

// TestFeedRanksRadioAboveHomeAndSkipsSeed checks the source prior and the seed exclusion.
func TestFeedRanksRadioAboveHomeAndSkipsSeed(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	items := environment.feed(t)
	radio := itemByVideoID(items, "radio1")
	home := itemByVideoID(items, "home1")
	if radio == nil || home == nil || radio.Score <= home.Score {
		t.Fatalf("radio must outscore home: %+v vs %+v", radio, home)
	}
	seedItem := itemByVideoID(items, "seed")
	if seedItem != nil {
		t.Errorf("the seed itself must not come back from its own radio: %+v", seedItem)
	}
}

// TestFeedExcludesDislikedTracks checks that thumbs-down tracks never appear.
func TestFeedExcludesDislikedTracks(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	environment.dislike(t, makeTrack("radio1", "artist:r1", "album:r1"), 400*24*time.Hour)

	items := environment.feed(t)
	if itemByVideoID(items, "radio1") != nil {
		t.Error("a disliked track must be excluded, however old the dislike")
	}
	if itemByVideoID(items, "radio2") == nil {
		t.Error("other radio tracks must stay")
	}
}

// TestFeedSinksTracksOfDislikedArtists checks that negative artist affinity lowers the score.
func TestFeedSinksTracksOfDislikedArtists(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	environment.dislike(t, makeTrack("bad", "artist:r1", "album:bad"), time.Hour)

	items := environment.feed(t)
	sunk, neutral := itemByVideoID(items, "radio1"), itemByVideoID(items, "radio2")
	if sunk == nil || neutral == nil || sunk.Score >= neutral.Score {
		t.Errorf("a disliked artist's track must rank below a neutral one: %+v vs %+v", sunk, neutral)
	}
}

// TestFeedSurvivesFailingSources checks that failing catalog calls are skipped.
func TestFeedSurvivesFailingSources(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	environment.catalog.failing["radio"] = true
	environment.catalog.failing["home"] = true

	items := environment.feed(t)
	if itemByVideoID(items, "top1") == nil {
		t.Error("artist tracks must still be served")
	}
	if itemByVideoID(items, "radio1") != nil || itemByVideoID(items, "home1") != nil {
		t.Error("failed sources must contribute nothing")
	}
}

// TestFeedWithEverySourceDownIsEmptyNotAnError checks the offline case.
func TestFeedWithEverySourceDownIsEmptyNotAnError(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	for _, method := range []string{"radio", "artist", "search", "home"} {
		environment.catalog.failing[method] = true
	}
	if items := environment.feed(t); len(items) != 0 {
		t.Errorf("got %d items with every source down", len(items))
	}
}

// TestColdStartServesYouTubeHome checks the empty-profile feed.
func TestColdStartServesYouTubeHome(t *testing.T) {
	environment := newTestEnvironment(t)
	environment.catalog.home = music.HomePage{Shelves: []music.Shelf{
		{Title: "Quick picks", Items: []music.Item{
			songItem(makeTrack("home1", "artist:a", "album:a")),
			{Kind: music.ItemKindAlbum, ID: "album:b", Title: "An album"},
			songItem(makeTrack("home2", "artist:b", "album:b")),
		}},
	}}

	if !profileOf(t, environment).Empty() {
		t.Fatal("a fresh library must have an empty profile")
	}
	items := environment.feed(t)
	if len(items) != 2 {
		t.Fatalf("got %d items, want the 2 home songs", len(items))
	}
	for _, item := range items {
		if item.Source != SourceYouTube {
			t.Errorf("cold start item from %q, want %q", item.Source, SourceYouTube)
		}
	}
}

// TestFeedMemoizesRankingAndCachesCatalogResponses checks both cache layers.
func TestFeedMemoizesRankingAndCachesCatalogResponses(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	ctx := context.Background()
	environment.feed(t)
	radioCalls := environment.catalog.callCount("radio")

	_, err := environment.engine.Feed(ctx, 10, 0, false)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	environment.clock.set(testNow.Add(11 * time.Minute))
	_, err = environment.engine.Feed(ctx, 10, 24, false)
	if err != nil {
		t.Fatalf("feed page: %v", err)
	}
	if environment.engine.cache.builtAt != testNow {
		t.Error("later pages must reuse the 11 minute old ranking")
	}
	_, err = environment.engine.Feed(ctx, 10, 0, false)
	if err != nil {
		t.Fatalf("feed first page: %v", err)
	}
	if !environment.engine.cache.builtAt.After(testNow) {
		t.Error("the first page must re-rank after 10 minutes")
	}
	if environment.catalog.callCount("radio") != radioCalls {
		t.Error("the re-rank must read the radio from the library cache")
	}

	environment.clock.set(testNow.Add(25 * time.Hour))
	environment.feed(t)
	if environment.catalog.callCount("radio") == radioCalls {
		t.Error("catalog responses must expire after a day")
	}
}

// TestRowsGroupByReason checks titled rows, thin and YouTube-only rows dropped.
func TestRowsGroupByReason(t *testing.T) {
	environment := newTestEnvironment(t)
	seedListening(t, environment)
	environment.catalog.radios["seed"] = append(environment.catalog.radios["seed"], makeTrack("radio3", "artist:r3", "album:r3"))
	environment.catalog.artists["artist:seed"] = music.ArtistPage{ID: "artist:seed", Name: "Seed Artist", TopSongs: []music.Track{
		makeTrack("top1", "artist:seed", "album:t1"), makeTrack("top2", "artist:seed", "album:t2"), makeTrack("top3", "artist:seed", "album:t3"),
	}}

	rows, err := environment.engine.Rows(context.Background(), 10, 10)
	if err != nil {
		t.Fatalf("rows: %v", err)
	}
	titles := map[string]int{}
	for _, row := range rows {
		titles[row.Title] = len(row.Items)
	}
	if titles["Because you listened to Title seed"] != 3 || titles["More from Seed Artist"] != 3 {
		t.Errorf("rows = %v", titles)
	}
	if _, found := titles["Popular on YouTube Music"]; found {
		t.Error("YouTube-only rows must be dropped")
	}
	if len(rows) > 1 && rowStrength(rows[0]) < rowStrength(rows[1]) {
		t.Error("rows must be ordered strongest first")
	}
}

// TestGroupRowsTrimsAndDropsThinRows checks the limits.
func TestGroupRowsTrimsAndDropsThinRows(t *testing.T) {
	scored := []Item{}
	for _, videoID := range []string{"a1", "a2", "a3", "a4"} {
		scored = append(scored, Item{Track: makeTrack(videoID, "", ""), Reason: rediscoverReason(), Score: 1})
	}
	scored = append(scored, Item{Track: makeTrack("b1", "", ""), Reason: searchReason("thin"), Score: 2})

	rows := groupRows(scored, 5, 2)
	if len(rows) != 1 || len(rows[0].Items) != 2 {
		t.Fatalf("rows = %+v, want one row trimmed to 2 items", rows)
	}
}

// TestPenaltiesSinkRecentPlaysAndIgnoredShows checks the played and fatigue penalties.
func TestPenaltiesSinkRecentPlaysAndIgnoredShows(t *testing.T) {
	state := &snapshot{
		now:        testNow,
		lastPlayed: map[string]time.Time{"today": testNow.Add(-2 * time.Hour), "week": testNow.Add(-3 * 24 * time.Hour)},
		ignored:    map[string]ignoredShows{"shown": {count: 4}, "once": {count: 1}},
	}
	assertClose(t, "played today", penalties("today", state), penaltyPlayedToday)
	assertClose(t, "played this week", penalties("week", state), penaltyPlayedThisWeek)
	assertClose(t, "shown four times", penalties("shown", state), 3*penaltyFatiguePerShow)
	assertClose(t, "shown once is free", penalties("once", state), 0)
	assertClose(t, "unknown track", penalties("new", state), 0)
}
