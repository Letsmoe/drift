package neoworks

import (
	"errors"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	libneoworks "github.com/neoworks-dev/libneoworks-go"
	"github.com/zalando/go-keyring"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

func syncedTestTrack(videoID string) music.Track {
	base := "https://lh3.googleusercontent.com/" + videoID
	return music.Track{
		VideoID:         videoID,
		Title:           "Title " + videoID,
		Artists:         []music.Artist{{ID: "UC1", Name: "Artist"}, {Name: "Unlinked"}},
		Album:           &music.Album{ID: "MPREb_1", Name: "Album"},
		DurationSeconds: 215,
		Thumbnails: []music.Thumbnail{
			{URL: base + "=w60-h60-l90-rj", Width: 60, Height: 60},
			{URL: base + "=w226-h226-l90-rj", Width: 226, Height: 226},
			{URL: base + "=w1080-h1080-l90-rj", Width: 1080, Height: 1080},
		},
	}
}

func TestThumbnailsKeepOneCoverAndRebuildTheUsualSizes(t *testing.T) {
	base := "https://lh3.googleusercontent.com/abc"
	url := syncedThumbnailURL(syncedTestTrack("abc").Thumbnails)
	if url != "https://lh3.googleusercontent.com/abc=w544-h544-l90-rj" {
		t.Fatalf("synced url = %s", url)
	}
	rebuilt := rebuildThumbnails(url)
	wantWidths := []int{60, 120, 226, 544}
	if len(rebuilt) != len(wantWidths) {
		t.Fatalf("rebuilt %+v", rebuilt)
	}
	for index, width := range wantWidths {
		want := music.Thumbnail{URL: base + "=w" + strconv.Itoa(width) + "-h" + strconv.Itoa(width) + "-l90-rj", Width: width, Height: width}
		if rebuilt[index] != want {
			t.Errorf("thumbnail %d = %+v, want %+v", index, rebuilt[index], want)
		}
	}
	if syncedThumbnailURL(rebuilt) != url {
		t.Errorf("the rebuilt list does not map back to %s", url)
	}
}

func TestThumbnailsKeepTheAspectRatioAndSizelessURLs(t *testing.T) {
	rebuilt := rebuildThumbnails("https://lh3.googleusercontent.com/x=w400-h225")
	if rebuilt[0].URL != "https://lh3.googleusercontent.com/x=w60-h34" || rebuilt[3].Height != 306 {
		t.Errorf("rebuilt %+v", rebuilt)
	}
	plain := "https://i.ytimg.com/vi/id/hqdefault.jpg"
	got := rebuildThumbnails(plain)
	if len(got) != 1 || got[0].URL != plain {
		t.Errorf("sizeless url became %+v", got)
	}
	if syncedThumbnailURL(nil) != "" || len(rebuildThumbnails("")) != 0 {
		t.Error("no cover must stay no cover")
	}
}

func TestARatingBecomesATrackNodeAndARatingNode(t *testing.T) {
	store := newFakeNodeStore()
	install := newTestInstall(t, store)
	ctx := t.Context()
	if err := install.library.SetRating(ctx, syncedTestTrack("abc"), music.ItemKindSong, library.RatingLike); err != nil {
		t.Fatal(err)
	}

	report := install.reconcile(t)
	if report.NodesWritten != 2 {
		t.Fatalf("nodes written = %d, want a track and a rating", report.NodesWritten)
	}
	tracks := store.liveItems("tracks")
	if len(tracks) != 1 || tracks[0]["videoId"] != "abc" || tracks[0]["thumbnailUrl"] != "https://lh3.googleusercontent.com/abc=w544-h544-l90-rj" {
		t.Errorf("track nodes = %v", tracks)
	}
	ratings := store.liveItems("ratings")
	if len(ratings) != 1 || ratings[0]["opinion"] != "like" || ratings[0]["videoId"] != "abc" {
		t.Errorf("rating nodes = %v", ratings)
	}
	links, err := install.library.NodeLinks(ctx, "ratings")
	if err != nil || links["abc"].NodeID == "" || links["abc"].NodeUpdatedAt == "" {
		t.Errorf("rating link = %+v, %v", links, err)
	}
}

func TestAReconcileWithoutChangesWritesNothing(t *testing.T) {
	store := newFakeNodeStore()
	install := newTestInstall(t, store)
	if err := install.library.SetRating(t.Context(), syncedTestTrack("abc"), music.ItemKindSong, library.RatingLike); err != nil {
		t.Fatal(err)
	}
	install.reconcile(t)

	report := install.reconcile(t)
	if report.NodesWritten != 0 || report.LibraryWritten != 0 {
		t.Errorf("second reconcile wrote %+v", report)
	}
}

func TestChangingAndRemovingARatingUpdatesAndDeletesItsNode(t *testing.T) {
	store := newFakeNodeStore()
	install := newTestInstall(t, store)
	ctx := t.Context()
	track := syncedTestTrack("abc")
	_ = install.library.SetRating(ctx, track, music.ItemKindSong, library.RatingLike)
	install.reconcile(t)

	install.advance(time.Minute)
	_ = install.library.SetRating(ctx, track, music.ItemKindSong, library.RatingDislike)
	install.reconcile(t)
	ratings := store.liveItems("ratings")
	if len(ratings) != 1 || ratings[0]["opinion"] != "dislike" {
		t.Fatalf("after dislike: %v", ratings)
	}

	_ = install.library.SetRating(ctx, track, music.ItemKindSong, library.RatingNone)
	install.reconcile(t)
	if ratings := store.liveItems("ratings"); len(ratings) != 0 {
		t.Errorf("after unrating: %v", ratings)
	}
	if links, _ := install.library.NodeLinks(ctx, "ratings"); len(links) != 0 {
		t.Errorf("links after unrating: %v", links)
	}
	if tracks := store.liveItems("tracks"); len(tracks) != 1 {
		t.Errorf("the track node must stay, got %v", tracks)
	}
}

func TestAnotherInstallReceivesTheRatingAndTheTrackWithoutAskingYouTube(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	ctx := t.Context()
	_ = laptop.library.SetRating(ctx, syncedTestTrack("abc"), music.ItemKindSong, library.RatingLike)
	laptop.reconcile(t)

	report := phone.reconcile(t)
	if report.LibraryWritten != 2 || report.NodesWritten != 0 {
		t.Fatalf("phone reconcile = %+v", report)
	}
	liked, err := phone.library.LikedTracks(ctx)
	if err != nil || len(liked) != 1 {
		t.Fatalf("liked = %v, %v", liked, err)
	}
	got := liked[0]
	if got.Title != "Title abc" || got.DurationSeconds != 215 || got.Album == nil || got.Album.ID != "MPREb_1" {
		t.Errorf("track = %+v", got)
	}
	if len(got.Artists) != 2 || got.Artists[0].ID != "UC1" || got.Artists[1].ID != "" {
		t.Errorf("artists = %+v", got.Artists)
	}
	if len(got.Thumbnails) != 4 || got.Thumbnails[3].Width != 544 {
		t.Errorf("thumbnails = %+v", got.Thumbnails)
	}
	if again := phone.reconcile(t); again.NodesWritten != 0 || again.LibraryWritten != 0 {
		t.Errorf("phone must settle, wrote %+v", again)
	}
}

func TestAnUnlikeOnOneInstallRemovesTheLikeOnTheOther(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	ctx := t.Context()
	track := syncedTestTrack("abc")
	_ = laptop.library.SetRating(ctx, track, music.ItemKindSong, library.RatingLike)
	laptop.reconcile(t)
	phone.reconcile(t)

	_ = laptop.library.SetRating(ctx, track, music.ItemKindSong, library.RatingNone)
	laptop.reconcile(t)
	phone.reconcile(t)

	ratings, _ := phone.library.Ratings(ctx, []string{"abc"})
	if len(ratings) != 0 {
		t.Errorf("phone still rates the track: %v", ratings)
	}
}

func TestTwoOfflineLikesMergeIntoTheNewestAndTheDuplicateNodeGoes(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store)
	phone := newTestInstall(t, store)
	ctx := t.Context()
	track := syncedTestTrack("abc")
	laptop.advance(time.Hour)
	_ = laptop.library.SetRating(ctx, track, music.ItemKindSong, library.RatingDislike)
	laptop.reconcile(t)
	phone.nodes.goOffline()
	phone.advance(2 * time.Hour)
	_ = phone.library.SetRating(ctx, track, music.ItemKindSong, library.RatingLike)
	phone.reconcile(t)
	phone.nodes.goOnline()

	if len(store.liveItems("ratings")) != 2 || len(store.liveItems("tracks")) != 2 {
		t.Fatalf("setup must leave duplicates, got %d ratings and %d tracks", len(store.liveItems("ratings")), len(store.liveItems("tracks")))
	}
	phone.reconcile(t)
	laptop.reconcile(t)

	ratings := store.liveItems("ratings")
	if len(ratings) != 1 || ratings[0]["opinion"] != "like" {
		t.Fatalf("rating nodes = %v, want the phone's newer like only", ratings)
	}
	if len(store.liveItems("tracks")) != 1 {
		t.Errorf("track nodes = %v", store.liveItems("tracks"))
	}
	rated, _ := laptop.library.Ratings(ctx, []string{"abc"})
	if rated["abc"] != library.RatingLike {
		t.Errorf("laptop rating = %v, want like", rated["abc"])
	}
	if report := laptop.reconcile(t); report.NodesWritten != 0 || report.LibraryWritten != 0 {
		t.Errorf("laptop must settle, wrote %+v", report)
	}
}

func TestACollectionWithoutARootIsSkipped(t *testing.T) {
	store := newFakeNodeStore()
	for nodeID, node := range store.nodes {
		if node.summary.Collection == "@drift/ratings" {
			delete(store.nodes, nodeID)
		}
	}
	install := newTestInstall(t, store)
	_ = install.library.SetRating(t.Context(), syncedTestTrack("def"), music.ItemKindSong, library.RatingLike)

	report := install.reconcile(t)
	if len(report.WithoutRoot) != 1 || report.WithoutRoot[0] != "ratings" {
		t.Errorf("without root = %v", report.WithoutRoot)
	}
	if len(store.liveItems("tracks")) != 1 {
		t.Error("the other collections must still sync")
	}
}

func TestReconcileNeedsASignedInUser(t *testing.T) {
	store := newFakeNodeStore()
	install := newTestInstall(t, store)
	install.sync.userID = func() string { return "" }
	if _, err := install.sync.Reconcile(t.Context()); !errors.Is(err, ErrSignedOut) {
		t.Errorf("err = %v, want ErrSignedOut", err)
	}
}

var _ Nodes = (*libneoworks.Client)(nil)

func TestAFreshStoreHoldsNoRootsSoNothingIsMapped(t *testing.T) {
	keyring.MockInit()
	session, _ := signedInSession(t, newFakeNeoworks(t), time.Hour)
	store := openTestLibrary(t)
	_ = store.SetRating(t.Context(), syncedTestTrack("abc"), music.ItemKindSong, library.RatingLike)
	librarySync := NewLibrarySync(LibrarySyncConfig{Library: store, Nodes: session.Nodes(), UserID: session.UserID})

	report, err := librarySync.Reconcile(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.WithoutRoot) != len(librarySync.mappings) || report.NodesWritten != 0 {
		t.Errorf("report = %+v", report)
	}
}

func openTestLibrary(t *testing.T) *library.Library {
	t.Helper()
	store, err := library.Open(t.Context(), filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}
