package neoworks

import (
	"slices"
	"testing"
	"time"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

// play records a play of the track at the given moment on the install.
func play(t *testing.T, install *testInstall, at time.Time, videoID string) int64 {
	t.Helper()
	*install.clock = at
	playID, err := install.library.StartPlay(t.Context(), syncedTestTrack(videoID), music.ItemKindSong, "album:MPREb_1")
	if err != nil {
		t.Fatal(err)
	}
	return playID
}

func logsByDay(store *fakeNodeStore) map[string]map[string]any {
	logs := map[string]map[string]any{}
	for _, node := range store.liveItems("history") {
		key := node["deviceId"].(string) + "|" + node["day"].(string)
		logs[key] = node
	}
	return logs
}

func TestPlaysAndSearchesBecomeOneLogPerLocalDay(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store, "laptop")
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Skip("no time zone database")
	}
	laptop.inLocation(berlin)
	lateEvening := time.Date(2026, 10, 2, 21, 30, 0, 0, time.UTC)
	playID := play(t, laptop, lateEvening, "abc")
	_ = laptop.library.UpdatePlay(t.Context(), playID, 200, true)
	play(t, laptop, lateEvening.Add(time.Hour), "def")
	*laptop.clock = lateEvening.Add(30 * time.Minute)
	_ = laptop.library.RecordSearch(t.Context(), "daft punk")

	laptop.reconcile(t)

	logs := logsByDay(store)
	if len(logs) != 2 {
		t.Fatalf("logs = %v, want one for 2026-10-02 and one for 2026-10-03 in Berlin", logs)
	}
	first := logs["laptop|2026-10-02"]
	plays, _ := first["plays"].([]any)
	if len(plays) != 1 || plays[0].(map[string]any)["completed"] != true || plays[0].(map[string]any)["context"] != "album:MPREb_1" {
		t.Errorf("first day plays = %v", first["plays"])
	}
	if first["deviceName"] != "Drift on laptop" {
		t.Errorf("device name = %v", first["deviceName"])
	}
	second := logs["laptop|2026-10-03"]
	searches, _ := second["searches"].([]any)
	if len(searches) != 1 || searches[0].(map[string]any)["query"] != "daft punk" {
		t.Errorf("second day searches = %v", second["searches"])
	}
}

func TestANewPlayUpdatesTodaysLogInsteadOfAddingANode(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store, "laptop")
	noon := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	play(t, laptop, noon, "abc")
	laptop.reconcile(t)
	play(t, laptop, noon.Add(time.Hour), "def")

	laptop.reconcile(t)

	logs := logsByDay(store)
	plays, _ := logs["laptop|2026-10-02"]["plays"].([]any)
	if len(logs) != 1 || len(plays) != 2 {
		t.Errorf("logs = %v", logs)
	}
	if report := laptop.reconcile(t); report.NodesWritten != 0 {
		t.Errorf("an unchanged log was written again: %+v", report)
	}
}

func TestAnotherDevicesLogAddsPlaysAndSearchesWithoutReturningAsOwnHistory(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store, "laptop")
	phone := newTestInstall(t, store, "phone")
	noon := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	play(t, laptop, noon, "abc")
	_ = laptop.library.RecordSearch(t.Context(), "Daft Punk")
	laptop.reconcile(t)
	play(t, phone, noon.Add(time.Hour), "def")

	phone.reconcile(t)
	laptop.reconcile(t)

	phonePlays, _ := phone.library.RecentTracks(t.Context(), 10)
	var recent []string
	for _, track := range phonePlays {
		recent = append(recent, track.VideoID)
	}
	if !slices.Equal(recent, []string{"def", "abc"}) {
		t.Errorf("phone recent tracks = %v", recent)
	}
	searches, _ := phone.library.RecentSearches(t.Context(), 10)
	if !slices.Equal(searches, []string{"Daft Punk"}) {
		t.Errorf("phone searches = %v", searches)
	}
	if len(logsByDay(store)) != 2 {
		t.Errorf("each device writes its own log, got %v", logsByDay(store))
	}
	own, _ := phone.library.OwnPlays(t.Context())
	if len(own) != 1 || own[0].VideoID != "def" {
		t.Errorf("the laptop's play must not become the phone's own: %v", own)
	}
	for _, install := range []*testInstall{laptop, phone} {
		if report := install.reconcile(t); report.NodesWritten != 0 || report.LibraryWritten != 0 {
			t.Errorf("must settle, wrote %+v", report)
		}
	}
	laptopPlays, _ := laptop.library.OwnPlays(t.Context())
	if len(laptopPlays) != 1 {
		t.Errorf("laptop own plays = %v", laptopPlays)
	}
}

func TestAnOwnLogIsRestoredWhenItsNodeIsEditedElsewhere(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store, "laptop")
	play(t, laptop, time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC), "abc")
	laptop.reconcile(t)
	for _, node := range store.nodes {
		if node.summary.Collection == "@drift/history" && node.summary.Kind == "item" {
			node.content["plays"] = []any{}
			node.summary.UpdatedAt = store.tick()
		}
	}

	laptop.reconcile(t)

	own, _ := laptop.library.OwnPlays(t.Context())
	if len(own) != 1 {
		t.Errorf("own plays = %v", own)
	}
	plays, _ := logsByDay(store)["laptop|2026-10-02"]["plays"].([]any)
	if len(plays) != 1 {
		t.Errorf("the node was not made to match the log again: %v", logsByDay(store))
	}
}

func TestBlocksSyncAndLiftEverywhere(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store, "laptop")
	phone := newTestInstall(t, store, "phone")
	ctx := t.Context()
	_ = laptop.library.BlockRecommendation(ctx, library.BlockArtist, "UC1", "Artist")
	_ = laptop.library.BlockRecommendation(ctx, library.BlockPodcast, "UC1", "Show")
	_ = laptop.library.SaveTracks(ctx, []music.Track{syncedTestTrack("abc")}, music.ItemKindSong)
	_ = laptop.library.BlockRecommendation(ctx, library.BlockTrack, "abc", "Title abc")
	laptop.reconcile(t)

	nodes := store.liveItems("blocks")
	if len(nodes) != 3 {
		t.Fatalf("block nodes = %v", nodes)
	}
	if len(store.liveItems("tracks")) != 1 {
		t.Errorf("a blocked track needs its @drift/tracks item, got %v", store.liveItems("tracks"))
	}
	phone.reconcile(t)
	blocks, _ := phone.library.RecommendationBlocks(ctx)
	if len(blocks) != 3 {
		t.Fatalf("phone blocks = %v", blocks)
	}
	if again := phone.reconcile(t); again.NodesWritten != 0 || again.LibraryWritten != 0 {
		t.Errorf("phone must settle, wrote %+v", again)
	}

	_ = laptop.library.UnblockRecommendation(ctx, library.BlockArtist, "UC1")
	laptop.reconcile(t)
	phone.reconcile(t)
	blocks, _ = phone.library.RecommendationBlocks(ctx)
	if len(blocks) != 2 {
		t.Errorf("phone blocks after unblock = %v", blocks)
	}
}

func TestAHistoryLogWaitsForTracksThatHaveNotArrived(t *testing.T) {
	store := newFakeNodeStore()
	laptop := newTestInstall(t, store, "laptop")
	phone := newTestInstall(t, store, "phone")
	play(t, laptop, time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC), "abc")
	laptop.reconcile(t)
	for _, node := range store.nodes {
		if node.summary.Collection == "@drift/tracks" {
			node.summary.Deleted = true
		}
	}

	if _, err := phone.sync.Reconcile(t.Context()); err == nil {
		t.Fatal("a log whose track is missing must report it")
	}
	if tracks, _ := phone.library.RecentTracks(t.Context(), 10); len(tracks) != 0 {
		t.Errorf("phone tracks = %v", tracks)
	}
}
