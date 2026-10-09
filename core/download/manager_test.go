package download

import (
	"bytes"
	"os"
	"testing"

	"github.com/letsmoe/drift-music/core/music"
)

func TestDownloadsTrackAndExposesLocalFiles(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()

	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "abc", StateDone)

	audio, found := manager.AudioFile("abc")
	if !found {
		t.Fatal("no local audio for a finished download")
	}
	contents, err := os.ReadFile(audio.Path)
	if err != nil || !bytes.Equal(contents, audioPayload) {
		t.Errorf("stored audio differs from the stream (err %v)", err)
	}
	if audio.MimeType != `audio/webm; codecs="opus"` {
		t.Errorf("mime type = %q", audio.MimeType)
	}
	cover, found := manager.CoverFile("abc")
	if !found || cover.MimeType != "image/jpeg" {
		t.Errorf("cover = %+v, found %v", cover, found)
	}
	assertDownloadedList(t, manager)
}

func assertDownloadedList(t *testing.T, manager *Manager) {
	t.Helper()
	tracks, err := manager.DownloadedTracks(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 || tracks[0].Thumbnails[0].URL != "http://local/cover/abc" {
		t.Errorf("downloaded tracks = %+v, want one with the local cover", tracks)
	}
	usage, err := manager.StorageUsage(t.Context())
	if err != nil || usage.Count != 1 || usage.Bytes != int64(len(audioPayload)) {
		t.Errorf("usage = %+v (err %v)", usage, err)
	}
}

func TestReResolvesWhenCDNRejectsTheURL(t *testing.T) {
	fixture := newFixture(t)
	fixture.cdn.rejectGeneration = 0
	manager := fixture.startManager()

	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "abc", StateDone)

	if fixture.resolver.invalidations() != 1 {
		t.Errorf("invalidations = %d, want 1", fixture.resolver.invalidations())
	}
	ranges := fixture.cdn.recordedRanges()
	if len(ranges) == 0 || ranges[0] != "bytes=0-" {
		t.Errorf("ranges = %v, want the first request to ask bytes=0-", ranges)
	}
}

func TestResumesAnInterruptedTransferFromThePartialFile(t *testing.T) {
	fixture := newFixture(t)
	fixture.cdn.breakFirst = true
	fixture.cdn.breakAfter = 20000
	manager := fixture.startManager()

	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "abc", StateDone)

	audio, _ := manager.AudioFile("abc")
	contents, err := os.ReadFile(audio.Path)
	if err != nil || !bytes.Equal(contents, audioPayload) {
		t.Fatalf("resumed file differs from the stream (err %v)", err)
	}
	ranges := fixture.cdn.recordedRanges()
	if len(ranges) != 2 || ranges[0] != "bytes=0-" || ranges[1] == "bytes=0-" {
		t.Errorf("ranges = %v, want a full request then a resumed one", ranges)
	}
}

func TestMarksATrackFailedAfterTheRetriesRunOut(t *testing.T) {
	fixture := newFixture(t)
	fixture.cdn.rejectAll = true
	manager := fixture.startManager(WithMaxAttempts(2))

	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "abc", StateFailed)
	statuses, _ := manager.States(t.Context(), []string{"abc"})
	if statuses["abc"].Error == "" {
		t.Error("a failed download should say why")
	}
}

func TestQueueSurvivesARestart(t *testing.T) {
	fixture := newFixture(t)
	first := fixture.startManager()
	first.SetAllowDownloads(false)
	err := first.DownloadTracks(t.Context(), []music.Track{fixture.track("abc"), fixture.track("def")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	if state := stateOf(t, first, "abc"); state != StatePaused {
		t.Fatalf("state while blocked = %q, want paused", state)
	}
	first.Close()

	second := fixture.startManager()

	waitForState(t, second, "abc", StateDone)
	waitForState(t, second, "def", StateDone)
}

func TestPauseAndResume(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	manager.Pause()

	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	if state := stateOf(t, manager, "abc"); state != StatePaused {
		t.Fatalf("state = %q, want paused", state)
	}
	summary, err := manager.Summary(t.Context())
	if err != nil || !summary.Paused || summary.Queued != 1 {
		t.Fatalf("summary = %+v (err %v)", summary, err)
	}

	manager.Resume()

	waitForState(t, manager, "abc", StateDone)
}

func TestCancelForgetsAnUnfinishedDownloadOnly(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	manager.Pause()
	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}

	err = manager.Cancel(t.Context(), "abc")
	if err != nil {
		t.Fatal(err)
	}

	if state := stateOf(t, manager, "abc"); state != "" {
		t.Errorf("state after cancel = %q, want none", state)
	}
}

func TestRejectsUnsafeVideoIDs(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()

	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("../escape")}, music.ItemKindSong)

	if err == nil {
		t.Fatal("a video ID with path separators must be rejected")
	}
}

func TestDeleteAllRemovesFilesAndRows(t *testing.T) {
	fixture := newFixture(t)
	manager := fixture.startManager()
	err := manager.DownloadTracks(t.Context(), []music.Track{fixture.track("abc")}, music.ItemKindSong)
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, "abc", StateDone)

	err = manager.DeleteAll(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(fixture.directory)
	if len(entries) != 0 {
		t.Errorf("files left behind: %v", entries)
	}
	usage, _ := manager.StorageUsage(t.Context())
	if usage.Count != 0 {
		t.Errorf("usage = %+v", usage)
	}
}
