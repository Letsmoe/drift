package library

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

// testClock is a clock tests advance by hand.
type testClock struct {
	current time.Time
}

func (clock *testClock) now() time.Time {
	return clock.current
}

func (clock *testClock) advance(duration time.Duration) {
	clock.current = clock.current.Add(duration)
}

func openTestLibrary(t *testing.T) (*Library, *testClock) {
	t.Helper()
	clock := &testClock{current: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)}
	library, err := Open(context.Background(), filepath.Join(t.TempDir(), "nested", "library.db"), WithClock(clock.now))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = library.Close() })
	return library, clock
}

func testTrack(videoID string) music.Track {
	return music.Track{
		VideoID:         videoID,
		Title:           "Title " + videoID,
		Artists:         []music.Artist{{ID: "UC1", Name: "Artist"}},
		Album:           &music.Album{ID: "MPREb_1", Name: "Album"},
		DurationSeconds: 200,
		Thumbnails:      []music.Thumbnail{{URL: "https://img/" + videoID, Width: 60, Height: 60}},
	}
}

func videoIDsOf(tracks []music.Track) []string {
	videoIDs := []string{}
	for _, track := range tracks {
		videoIDs = append(videoIDs, track.VideoID)
	}
	return videoIDs
}

func assertVideoIDs(t *testing.T, got []music.Track, want ...string) {
	t.Helper()
	gotIDs := videoIDsOf(got)
	if len(gotIDs) != len(want) {
		t.Fatalf("got %v, want %v", gotIDs, want)
	}
	for index := range want {
		if gotIDs[index] != want[index] {
			t.Fatalf("got %v, want %v", gotIDs, want)
		}
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "library.db")
	first, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err = first.SaveTracks(ctx, []music.Track{testTrack("a")}, music.ItemKindSong); err != nil {
		t.Fatal(err)
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = second.Close() }()
	var applied int
	if err = second.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM _migration`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied != 6 {
		t.Fatalf("applied migrations = %d, want 6", applied)
	}
	tracks, err := second.Tracks(ctx, []string{"a"})
	if err != nil || len(tracks) != 1 {
		t.Fatalf("tracks after reopen = %v, %v", tracks, err)
	}
}

func TestTrackRoundTrip(t *testing.T) {
	ctx := context.Background()
	library, _ := openTestLibrary(t)
	bare := music.Track{VideoID: "bare", Title: "Bare"}
	if err := library.SaveTracks(ctx, []music.Track{testTrack("a"), bare}, music.ItemKindSong); err != nil {
		t.Fatal(err)
	}
	tracks, err := library.Tracks(ctx, []string{"a", "bare", "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 2 {
		t.Fatalf("got %d tracks, want 2", len(tracks))
	}
	got := tracks["a"]
	if got.Title != "Title a" || got.Album.Name != "Album" || got.Artists[0].ID != "UC1" || got.Thumbnails[0].Width != 60 {
		t.Fatalf("track did not round trip: %+v", got)
	}
	if tracks["bare"].Album != nil {
		t.Fatalf("bare track gained an album")
	}
}

func TestRatingsRoundTrip(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	for _, videoID := range []string{"a", "b", "c"} {
		clock.advance(time.Second)
		if err := library.SetRating(ctx, testTrack(videoID), music.ItemKindSong, RatingLike); err != nil {
			t.Fatal(err)
		}
	}
	clock.advance(time.Second)
	if err := library.SetRating(ctx, testTrack("d"), music.ItemKindSong, RatingDislike); err != nil {
		t.Fatal(err)
	}
	clock.advance(time.Second)
	if err := library.SetRating(ctx, testTrack("a"), music.ItemKindSong, RatingLike); err != nil {
		t.Fatal(err)
	}
	if err := library.SetRating(ctx, testTrack("b"), music.ItemKindSong, RatingNone); err != nil {
		t.Fatal(err)
	}
	ratings, err := library.Ratings(ctx, []string{"a", "b", "c", "d"})
	if err != nil {
		t.Fatal(err)
	}
	if ratings["a"] != RatingLike || ratings["d"] != RatingDislike {
		t.Fatalf("ratings = %v", ratings)
	}
	if _, rated := ratings["b"]; rated {
		t.Fatalf("removed rating still present")
	}
	liked, err := library.LikedTracks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertVideoIDs(t, liked, "a", "c")
}

func playlistVideoIDs(t *testing.T, library *Library, id string) []string {
	t.Helper()
	playlist, err := library.Playlist(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	videoIDs := []string{}
	for _, entry := range playlist.Entries {
		videoIDs = append(videoIDs, entry.Track.VideoID)
	}
	return videoIDs
}

func assertDensePositions(t *testing.T, library *Library, id string, count int) {
	t.Helper()
	rows, err := library.DB().QueryContext(context.Background(),
		`SELECT position FROM playlist_entry WHERE playlist_id = ? ORDER BY position`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	expected := 0
	for rows.Next() {
		var position int
		if err = rows.Scan(&position); err != nil {
			t.Fatal(err)
		}
		if position != expected {
			t.Fatalf("position %d, want %d", position, expected)
		}
		expected++
	}
	if expected != count {
		t.Fatalf("%d entries, want %d", expected, count)
	}
}

func TestPlaylistEntriesStayDense(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	playlist, err := library.CreatePlaylist(ctx, "Mix")
	if err != nil {
		t.Fatal(err)
	}
	clock.advance(time.Minute)
	tracks := []music.Track{testTrack("a"), testTrack("b"), testTrack("c"), testTrack("d")}
	if err = library.AddToPlaylist(ctx, playlist.ID, tracks); err != nil {
		t.Fatal(err)
	}
	loaded, err := library.Playlist(ctx, playlist.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.UpdatedAt.After(loaded.CreatedAt) {
		t.Fatalf("add did not bump updated_at")
	}
	if err = library.MovePlaylistEntry(ctx, playlist.ID, loaded.Entries[3].ID, 0); err != nil {
		t.Fatal(err)
	}
	if got := playlistVideoIDs(t, library, playlist.ID); len(got) != 4 || got[0] != "d" || got[3] != "c" {
		t.Fatalf("after move to front: %v", got)
	}
	if err = library.RemoveFromPlaylist(ctx, playlist.ID, loaded.Entries[0].ID); err != nil {
		t.Fatal(err)
	}
	if got := playlistVideoIDs(t, library, playlist.ID); len(got) != 3 || got[0] != "d" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("after remove: %v", got)
	}
	if err = library.MovePlaylistEntry(ctx, playlist.ID, loaded.Entries[1].ID, 99); err != nil {
		t.Fatal(err)
	}
	if got := playlistVideoIDs(t, library, playlist.ID); got[2] != "b" {
		t.Fatalf("after move to end: %v", got)
	}
	assertDensePositions(t, library, playlist.ID, 3)
	if err = library.AddToPlaylist(ctx, playlist.ID, []music.Track{testTrack("a")}); err != nil {
		t.Fatal(err)
	}
	assertDensePositions(t, library, playlist.ID, 4)
}

func TestPlaylistSummariesAndErrors(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	older, _ := library.CreatePlaylist(ctx, "Older")
	clock.advance(time.Minute)
	newer, _ := library.CreatePlaylist(ctx, "Newer")
	clock.advance(time.Minute)
	if err := library.AddToPlaylist(ctx, older.ID, []music.Track{testTrack("a"), testTrack("b")}); err != nil {
		t.Fatal(err)
	}
	summaries, err := library.Playlists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 || summaries[0].ID != older.ID || summaries[0].TrackCount != 2 {
		t.Fatalf("summaries = %+v", summaries)
	}
	if summaries[0].Thumbnails[0].URL != "https://img/a" || len(summaries[1].Thumbnails) != 0 {
		t.Fatalf("covers = %+v", summaries)
	}
	if err = library.DeletePlaylist(ctx, newer.ID); err != nil {
		t.Fatal(err)
	}
	_, err = library.Playlist(ctx, newer.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted playlist err = %v", err)
	}
	if err = library.AddToPlaylist(ctx, "nope", nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("add to unknown err = %v", err)
	}
	if err = library.RenamePlaylist(ctx, "nope", "x", ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("rename unknown err = %v", err)
	}
	if err = library.RemoveFromPlaylist(ctx, older.ID, 9999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("remove unknown entry err = %v", err)
	}
}

func TestSavedItems(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	album := music.Item{Kind: music.ItemKindAlbum, ID: "MPREb_1", Title: "Album", Thumbnails: []music.Thumbnail{{URL: "u"}}}
	artist := music.Item{Kind: music.ItemKindArtist, ID: "UC1", Title: "Artist"}
	if err := library.Save(ctx, album); err != nil {
		t.Fatal(err)
	}
	clock.advance(time.Second)
	if err := library.Save(ctx, artist); err != nil {
		t.Fatal(err)
	}
	saved, err := library.IsSaved(ctx, music.ItemKindAlbum, "MPREb_1")
	if err != nil || !saved {
		t.Fatalf("IsSaved = %v, %v", saved, err)
	}
	items, err := library.SavedItems(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Item.ID != "UC1" || items[1].Item.Thumbnails[0].URL != "u" {
		t.Fatalf("items = %+v", items)
	}
	if err = library.Unsave(ctx, music.ItemKindAlbum, "MPREb_1"); err != nil {
		t.Fatal(err)
	}
	saved, _ = library.IsSaved(ctx, music.ItemKindAlbum, "MPREb_1")
	if saved {
		t.Fatalf("still saved after Unsave")
	}
}

func TestPlayProgress(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	playID, err := library.StartPlay(ctx, testTrack("a"), music.ItemKindSong, "radio")
	if err != nil {
		t.Fatal(err)
	}
	updates := []struct {
		progress  int
		completed bool
	}{{120, true}, {40, false}}
	for _, update := range updates {
		if err = library.UpdatePlay(ctx, playID, update.progress, update.completed); err != nil {
			t.Fatal(err)
		}
	}
	var progress, duration int
	var completed bool
	err = library.DB().QueryRowContext(ctx, `SELECT progress_seconds, duration_seconds, completed FROM play WHERE id = ?`, playID).
		Scan(&progress, &duration, &completed)
	if err != nil {
		t.Fatal(err)
	}
	if progress != 120 || duration != 200 || !completed {
		t.Fatalf("progress=%d duration=%d completed=%v", progress, duration, completed)
	}
	if err = library.UpdatePlay(ctx, 9999, 1, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown play err = %v", err)
	}
	clock.advance(time.Minute)
	for _, videoID := range []string{"b", "a"} {
		clock.advance(time.Minute)
		if _, err = library.StartPlay(ctx, testTrack(videoID), music.ItemKindSong, "queue"); err != nil {
			t.Fatal(err)
		}
	}
	recent, err := library.RecentTracks(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	assertVideoIDs(t, recent, "a", "b")
}

func TestResumePosition(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	episode := testTrack("ep")
	assertResume := func(videoID string, wantSeconds int, wantOK bool) {
		t.Helper()
		seconds, ok, err := library.ResumePosition(ctx, videoID)
		if err != nil || seconds != wantSeconds || ok != wantOK {
			t.Fatalf("ResumePosition(%s) = %d, %v, %v; want %d, %v", videoID, seconds, ok, err, wantSeconds, wantOK)
		}
	}
	assertResume("ep", 0, false)

	playID, _ := library.StartPlay(ctx, episode, music.ItemKindEpisode, "podcast:x")
	_ = library.UpdatePlay(ctx, playID, 20, false)
	assertResume("ep", 20, false)
	_ = library.UpdatePlay(ctx, playID, 600, false)
	assertResume("ep", 600, true)

	clock.advance(time.Minute)
	finishedID, _ := library.StartPlay(ctx, episode, music.ItemKindEpisode, "podcast:x")
	_ = library.UpdatePlay(ctx, finishedID, 3000, true)
	assertResume("ep", 3000, false)

	songID, _ := library.StartPlay(ctx, testTrack("song"), music.ItemKindSong, "queue")
	_ = library.UpdatePlay(ctx, songID, 100, false)
	assertResume("song", 0, false)
}

func TestRecentSearches(t *testing.T) {
	ctx := context.Background()
	library, _ := openTestLibrary(t)
	for _, query := range []string{"Daft Punk", "  ", "radiohead", "daft   punk ", "Boards of Canada"} {
		if err := library.RecordSearch(ctx, query); err != nil {
			t.Fatal(err)
		}
	}
	searches, err := library.RecentSearches(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Boards of Canada", "daft   punk", "radiohead"}
	if len(searches) != len(want) {
		t.Fatalf("searches = %q, want %q", searches, want)
	}
	for index := range want {
		if searches[index] != want[index] {
			t.Fatalf("searches = %q, want %q", searches, want)
		}
	}
	limited, _ := library.RecentSearches(ctx, 1)
	if len(limited) != 1 {
		t.Fatalf("limit ignored: %q", limited)
	}
	if err = library.ClearSearches(ctx); err != nil {
		t.Fatal(err)
	}
	searches, _ = library.RecentSearches(ctx, 10)
	if len(searches) != 0 {
		t.Fatalf("searches after clear = %q", searches)
	}
}

func TestImpressionsPurgedAfterRetention(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	position := 2
	err := library.RecordImpressions(ctx, []Impression{
		{VideoID: "a", Source: "radio", Surface: "home", Position: &position},
		{VideoID: "b", Source: "radio", Surface: "home", Clicked: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	var nullPositions int
	err = library.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM impression WHERE position IS NULL`).Scan(&nullPositions)
	if err != nil || nullPositions != 1 {
		t.Fatalf("null positions = %d, %v", nullPositions, err)
	}
	if err = library.purgeExpired(ctx); err != nil {
		t.Fatal(err)
	}
	clock.advance(91 * 24 * time.Hour)
	if err = library.purgeExpired(ctx); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err = library.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM impression`).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("remaining impressions = %d, %v", remaining, err)
	}
}

func TestCacheExpiry(t *testing.T) {
	ctx := context.Background()
	library, clock := openTestLibrary(t)
	type payload struct{ Name string }
	if err := library.CachePut(ctx, "k", payload{Name: "x"}, time.Hour); err != nil {
		t.Fatal(err)
	}
	var got payload
	found, err := library.CacheGet(ctx, "k", &got)
	if err != nil || !found || got.Name != "x" {
		t.Fatalf("fresh cache = %v, %v, %+v", found, err, got)
	}
	clock.advance(time.Hour)
	found, err = library.CacheGet(ctx, "k", &got)
	if err != nil || found {
		t.Fatalf("expired cache found = %v, %v", found, err)
	}
	if err = library.purgeExpired(ctx); err != nil {
		t.Fatal(err)
	}
	var rows int
	if err = library.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM cache`).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("cache rows after purge = %d, %v", rows, err)
	}
}

func TestTimeFormatSortsAndRoundTrips(t *testing.T) {
	early := time.Date(2026, 1, 2, 3, 4, 5, 6_000_000, time.UTC)
	late := early.Add(time.Millisecond)
	if FormatTime(early) >= FormatTime(late) {
		t.Fatalf("format does not sort")
	}
	parsed, err := ParseTime(FormatTime(early))
	if err != nil || !parsed.Equal(early) {
		t.Fatalf("round trip = %v, %v", parsed, err)
	}
}

func TestYouTubeSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	lib, _ := openTestLibrary(t)

	_, found, err := lib.YouTubeSession(ctx)
	if err != nil || found {
		t.Fatalf("empty library: found = %v, err = %v", found, err)
	}

	account := music.Account{Name: "Ada", Handle: "@ada", PhotoURL: "https://example.com/ada"}
	err = lib.SaveYouTubeSession(ctx, "SAPISID=secret", account)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	err = lib.MarkYouTubeSynced(ctx)
	if err != nil {
		t.Fatalf("mark synced: %v", err)
	}
	session, found, err := lib.YouTubeSession(ctx)
	if err != nil || !found {
		t.Fatalf("load: found = %v, err = %v", found, err)
	}
	if session.CookieHeader != "SAPISID=secret" || session.Account != account || session.LastSyncedAt.IsZero() {
		t.Errorf("session = %+v", session.Account)
	}
	if strings.Contains(fmt.Sprintf("%v %+v %#v", session, session, session), "secret") {
		t.Error("formatting a session leaks the cookies")
	}

	err = lib.SaveYouTubeSession(ctx, "SAPISID=newer", account)
	if err != nil {
		t.Fatalf("save again: %v", err)
	}
	session, _, _ = lib.YouTubeSession(ctx)
	if session.LastSyncedAt.IsZero() {
		t.Error("signing in again as the same account must keep the last-synced time")
	}
	err = lib.SaveYouTubeSession(ctx, "SAPISID=other", music.Account{Name: "Grace", Handle: "@grace"})
	if err != nil {
		t.Fatalf("save other account: %v", err)
	}
	session, _, _ = lib.YouTubeSession(ctx)
	if !session.LastSyncedAt.IsZero() {
		t.Error("a different account must start without a last-synced time")
	}

	err = lib.ClearYouTubeSession(ctx)
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	_, found, _ = lib.YouTubeSession(ctx)
	if found {
		t.Error("session still stored after sign-out")
	}
}

func TestOnChangeReportsLikesAndPlaylistEdits(t *testing.T) {
	ctx := t.Context()
	lib, err := Open(ctx, filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lib.Close() }()
	var changes []Change
	lib.OnChange(func(change Change) { changes = append(changes, change) })

	err = lib.SetRating(ctx, music.Track{VideoID: "a", Title: "A"}, music.ItemKindSong, RatingLike)
	if err != nil {
		t.Fatal(err)
	}
	playlist, err := lib.CreatePlaylist(ctx, "Mix")
	if err != nil {
		t.Fatal(err)
	}
	err = lib.AddToPlaylist(ctx, playlist.ID, []music.Track{{VideoID: "a", Title: "A"}})
	if err != nil {
		t.Fatal(err)
	}
	err = lib.DeletePlaylist(ctx, playlist.ID)
	if err != nil {
		t.Fatal(err)
	}

	want := []Change{
		{Kind: ChangeLikes},
		{Kind: ChangePlaylist, ID: playlist.ID},
		{Kind: ChangePlaylist, ID: playlist.ID},
	}
	if !slices.Equal(changes, want) {
		t.Errorf("changes = %v, want %v", changes, want)
	}
}
