package mobile

import (
	"sync"
	"testing"

	"github.com/letsmoe/drift/core/download"
)

type recordingDownloadListener struct {
	mutex   sync.Mutex
	updates []string
}

func (listener *recordingDownloadListener) OnDownloadsChanged(updateJSON string) {
	listener.mutex.Lock()
	defer listener.mutex.Unlock()
	listener.updates = append(listener.updates, updateJSON)
}

func TestDownloadsQueueWhileBlockedAndReportThroughTheListener(t *testing.T) {
	core, _ := openTestCore(t)
	listener := &recordingDownloadListener{}
	core.SetDownloadListener(listener)
	core.SetAllowDownloads(false)

	err := core.DownloadTracks(twoTracksJSON, "song")
	if err != nil {
		t.Fatal(err)
	}

	states, err := core.DownloadStates(`["vid1","vid2","unknown"]`)
	if err != nil {
		t.Fatal(err)
	}
	statuses := decodeInto[map[string]download.Status](t, states)
	if len(statuses) != 2 || statuses["vid1"].State != download.StatePaused {
		t.Errorf("states = %s, want two paused downloads", states)
	}
	listener.mutex.Lock()
	defer listener.mutex.Unlock()
	if len(listener.updates) == 0 {
		t.Error("the download listener received nothing")
	}
}

func TestDownloadListingsAreNeverNull(t *testing.T) {
	core, _ := openTestCore(t)

	for name, list := range map[string]func() (string, error){
		"queue":       core.DownloadQueue,
		"downloaded":  core.DownloadedItems,
		"tracks":      core.DownloadedTracks,
		"collections": core.DownloadedCollections,
	} {
		encoded, err := list()
		if err != nil || encoded != "[]" {
			t.Errorf("%s = %q (err %v), want []", name, encoded, err)
		}
	}
}
