package mobile

import (
	"path/filepath"
	"testing"

	"github.com/letsmoe/drift/core/importer"
)

type recordingImportListener struct {
	done chan string
}

func (listener *recordingImportListener) OnImportProgress(string) {}

func (listener *recordingImportListener) OnImportDone(resultJSON, errorMessage string) {
	listener.done <- resultJSON + "|" + errorMessage
}

func TestSpotifyPreviewExport(t *testing.T) {
	core, _ := openTestCore(t)
	encoded, err := core.SpotifyPreviewExport(filepath.Join("..", "spotify", "testdata", "export"))
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	summary := decodeInto[importer.Summary](t, encoded)
	if summary.Liked != 3 || len(summary.Playlists) != 2 {
		t.Errorf("summary = %+v", summary)
	}
}

func TestSpotifyImportReportsFailure(t *testing.T) {
	core, _ := openTestCore(t)
	listener := &recordingImportListener{done: make(chan string, 1)}
	err := core.SpotifyImport("export", filepath.Join(t.TempDir(), "missing.zip"), `{"liked":true}`, listener)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	outcome := <-listener.done
	if outcome[0] != '|' {
		t.Errorf("outcome = %q, want an error message and no result", outcome)
	}
}

func TestYouTubeAccountStartsSignedOut(t *testing.T) {
	core, _ := openTestCore(t)

	encoded, err := core.YouTubeAccount()
	if err != nil {
		t.Fatalf("account: %v", err)
	}
	if status := decodeInto[map[string]any](t, encoded); status["signedIn"] != false {
		t.Errorf("status = %v", status)
	}
	_, err = core.YouTubePreview()
	if err == nil {
		t.Error("previewing without a session must fail")
	}
	if core.YouTubeSignOut() != nil {
		t.Error("signing out while signed out is not an error")
	}
}

func TestYouTubeSignInRejectsCookiesWithoutSession(t *testing.T) {
	core, _ := openTestCore(t)
	_, err := core.YouTubeSignIn("YSC=nothing")
	if err == nil {
		t.Error("cookies without SAPISID must be rejected")
	}
}

func TestYouTubeImportWithoutSessionReportsFailure(t *testing.T) {
	core, _ := openTestCore(t)
	listener := &recordingImportListener{done: make(chan string, 1)}
	err := core.YouTubeImport(`{"liked":true}`, listener)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if outcome := <-listener.done; outcome[0] != '|' {
		t.Errorf("outcome = %q, want an error and no result", outcome)
	}
}
