package neoworks

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	libneoworks "github.com/neoworks-dev/libneoworks-go"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

// recordingSession and recordingReconciler log the order of a pass.
type recordingSession struct {
	mutex     sync.Mutex
	events    *[]string
	signedIn  bool
	syncError error
}

func (session *recordingSession) SignedIn() bool { return session.signedIn }

func (session *recordingSession) Sync(context.Context) (libneoworks.SyncReport, error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	*session.events = append(*session.events, "sync")
	return libneoworks.SyncReport{}, session.syncError
}

func (session *recordingSession) SignOut() error {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	*session.events = append(*session.events, "sign out")
	return nil
}

type recordingReconciler struct {
	session *recordingSession
	written int
	err     error
}

func (reconciler *recordingReconciler) Reconcile(context.Context) (ReconcileReport, error) {
	reconciler.session.mutex.Lock()
	defer reconciler.session.mutex.Unlock()
	*reconciler.session.events = append(*reconciler.session.events, "reconcile")
	return ReconcileReport{NodesWritten: reconciler.written}, reconciler.err
}

func (reconciler *recordingReconciler) Forget(context.Context) error {
	reconciler.session.mutex.Lock()
	defer reconciler.session.mutex.Unlock()
	*reconciler.session.events = append(*reconciler.session.events, "forget")
	return nil
}

func newRecordingSyncer(written int) (*Syncer, *recordingSession, *recordingReconciler) {
	events := &[]string{}
	session := &recordingSession{events: events, signedIn: true}
	reconciler := &recordingReconciler{session: session, written: written}
	syncer := NewSyncer(SyncerConfig{
		Session:     session,
		LibrarySync: reconciler,
		Debounce:    20 * time.Millisecond,
		Interval:    time.Hour,
	})
	return syncer, session, reconciler
}

func (session *recordingSession) snapshot() []string {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return append([]string(nil), *session.events...)
}

func waitForEvents(t *testing.T, session *recordingSession, count int) []string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if events := session.snapshot(); len(events) >= count {
			return events
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected %d events, got %v", count, session.snapshot())
	return nil
}

func equalEvents(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func TestAPassPullsReconcilesAndUploadsOnlyWhatWasWritten(t *testing.T) {
	syncer, session, reconciler := newRecordingSyncer(2)
	if _, err := syncer.SyncNow(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := session.snapshot(); !equalEvents(got, []string{"sync", "reconcile", "sync"}) {
		t.Errorf("events = %v", got)
	}

	reconciler.written = 0
	_, _ = syncer.SyncNow(t.Context())
	if got := session.snapshot(); !equalEvents(got[3:], []string{"sync", "reconcile"}) {
		t.Errorf("a pass that wrote nothing did %v", got[3:])
	}
}

func TestAFailedPullSkipsTheReconcile(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(1)
	session.syncError = errors.New("offline")
	if _, err := syncer.SyncNow(t.Context()); !errors.Is(err, session.syncError) {
		t.Fatalf("err = %v", err)
	}
	if got := session.snapshot(); !equalEvents(got, []string{"sync"}) {
		t.Errorf("events = %v", got)
	}
}

func TestAPartialReconcileFailureStillUploadsWhatItWrote(t *testing.T) {
	syncer, session, reconciler := newRecordingSyncer(1)
	reconciler.err = errors.New("one track missing")
	if _, err := syncer.SyncNow(t.Context()); !errors.Is(err, reconciler.err) {
		t.Fatalf("err = %v", err)
	}
	if got := session.snapshot(); !equalEvents(got, []string{"sync", "reconcile", "sync"}) {
		t.Errorf("events = %v", got)
	}
}

func TestSignedOutPassesDoNothing(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(1)
	session.signedIn = false
	if _, err := syncer.SyncNow(t.Context()); !errors.Is(err, ErrSignedOut) {
		t.Fatalf("err = %v", err)
	}
	if got := session.snapshot(); len(got) != 0 {
		t.Errorf("events = %v", got)
	}
}

func TestManyRequestsMakeOnePassAfterTheDebounce(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)
	waitForEvents(t, session, 2)
	before := len(session.snapshot())

	for range 20 {
		syncer.Request()
		time.Sleep(time.Millisecond)
	}

	events := waitForEvents(t, session, before+2)
	time.Sleep(100 * time.Millisecond)
	if got := session.snapshot(); len(got) != len(events) {
		t.Errorf("a burst made more than one pass: %v", got)
	}
}

func TestLibraryChangesScheduleAPassAndDownloadsStillHearThem(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	store, err := library.Open(t.Context(), t.TempDir()+"/library.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	var otherListener []library.Change
	var mutex sync.Mutex
	store.OnChange(func(change library.Change) {
		mutex.Lock()
		defer mutex.Unlock()
		otherListener = append(otherListener, change)
	})
	syncer.Watch(store)
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)
	waitForEvents(t, session, 2)
	before := len(session.snapshot())

	if err = store.Save(t.Context(), music.Item{Kind: music.ItemKindAlbum, ID: "MPREb_1", Title: "Album"}); err != nil {
		t.Fatal(err)
	}

	waitForEvents(t, session, before+2)
	mutex.Lock()
	defer mutex.Unlock()
	if len(otherListener) != 1 || otherListener[0].Kind != library.ChangeSaved {
		t.Errorf("the other listener heard %v", otherListener)
	}
}

func TestTheIntervalSyncsWithoutAnyChange(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.config.Interval = 30 * time.Millisecond
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)

	waitForEvents(t, session, 6)
}

func TestStopWaitsForARunningPass(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.Start(t.Context())
	waitForEvents(t, session, 2)
	syncer.Stop()
	count := len(session.snapshot())
	syncer.Request()
	time.Sleep(100 * time.Millisecond)
	if got := len(session.snapshot()); got != count {
		t.Errorf("a stopped syncer ran %d more events", got-count)
	}
}

func TestSigningOutForgetsTheLinksBeforeWipingTheStore(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	if err := syncer.SignOut(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := session.snapshot(); !equalEvents(got, []string{"forget", "sign out"}) {
		t.Errorf("events = %v", got)
	}
}

func TestTheObserverHearsEachPass(t *testing.T) {
	syncer, _, _ := newRecordingSyncer(3)
	var states []SyncState
	syncer.config.OnState = func(state SyncState) { states = append(states, state) }
	_, _ = syncer.SyncNow(t.Context())
	if len(states) != 2 || !states[0].Syncing || states[1].Syncing || states[1].Report.NodesWritten != 3 {
		t.Errorf("states = %+v", states)
	}
}

func TestTheDefaultIntervalIsTheFallbackForAnUnconnectedFeed(t *testing.T) {
	if DefaultInterval != 15*time.Minute {
		t.Errorf("DefaultInterval = %v", DefaultInterval)
	}
}

func TestAnAnnouncedRemoteChangeSchedulesAPass(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)
	waitForEvents(t, session, 2)
	before := len(session.snapshot())

	syncer.NoteRemoteChange(5)

	waitForEvents(t, session, before+2)
}

func TestARemoteChangeAlreadyPulledSchedulesNothing(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)
	waitForEvents(t, session, 2)
	syncer.NoteRemoteChange(5)
	waitForEvents(t, session, 4)
	time.Sleep(100 * time.Millisecond)
	before := len(session.snapshot())

	syncer.NoteRemoteChange(5)
	syncer.NoteRemoteChange(3)
	time.Sleep(150 * time.Millisecond)

	if got := len(session.snapshot()); got != before {
		t.Errorf("old sequences made %d more events", got-before)
	}
}

func TestANewerRemoteChangeThanThePullSchedulesAnotherPass(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)
	waitForEvents(t, session, 2)
	syncer.NoteRemoteChange(5)
	waitForEvents(t, session, 4)
	time.Sleep(100 * time.Millisecond)
	before := len(session.snapshot())

	syncer.NoteRemoteChange(6)

	waitForEvents(t, session, before+2)
}

func TestAChangeAnnouncedDuringAPassIsNotCountedAsPulled(t *testing.T) {
	syncer, _, _ := newRecordingSyncer(0)
	if _, err := syncer.SyncNow(t.Context()); err != nil {
		t.Fatal(err)
	}
	syncer.NoteRemoteChange(8)
	syncer.sequenceMutex.Lock()
	pulledBeforeThePass := syncer.pulledSequence
	syncer.sequenceMutex.Unlock()

	if pulledBeforeThePass != 0 {
		t.Fatalf("announcing must not mark anything as pulled, pulled = %d", pulledBeforeThePass)
	}
	if _, err := syncer.SyncNow(t.Context()); err != nil {
		t.Fatal(err)
	}
	syncer.sequenceMutex.Lock()
	defer syncer.sequenceMutex.Unlock()
	if syncer.pulledSequence != 8 {
		t.Errorf("a pass started after the announcement pulls it, pulled = %d", syncer.pulledSequence)
	}
}

func TestAFailedPullDoesNotCountAnnouncedChangesAsPulled(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	session.syncError = errors.New("offline")
	syncer.NoteRemoteChange(8)

	_, _ = syncer.SyncNow(t.Context())

	syncer.sequenceMutex.Lock()
	defer syncer.sequenceMutex.Unlock()
	if syncer.pulledSequence != 0 {
		t.Errorf("pulled = %d after a pull that failed", syncer.pulledSequence)
	}
}

func TestAResyncMakesTheNextChangeCountEvenWithALowerSequence(t *testing.T) {
	syncer, session, _ := newRecordingSyncer(0)
	syncer.Start(t.Context())
	t.Cleanup(syncer.Stop)
	waitForEvents(t, session, 2)
	syncer.NoteRemoteChange(50)
	waitForEvents(t, session, 4)
	time.Sleep(100 * time.Millisecond)

	syncer.NoteResync()
	waitForEvents(t, session, 6)
	time.Sleep(100 * time.Millisecond)
	before := len(session.snapshot())
	syncer.NoteRemoteChange(2)

	waitForEvents(t, session, before+2)
}

func TestSigningOutForgetsTheSequences(t *testing.T) {
	syncer, _, _ := newRecordingSyncer(0)
	syncer.NoteRemoteChange(9)
	if _, err := syncer.SyncNow(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := syncer.SignOut(t.Context()); err != nil {
		t.Fatal(err)
	}

	syncer.sequenceMutex.Lock()
	defer syncer.sequenceMutex.Unlock()
	if syncer.pulledSequence != 0 || syncer.announcedSequence != 0 {
		t.Errorf("sequences survived the sign-out: pulled %d, announced %d", syncer.pulledSequence, syncer.announcedSequence)
	}
}
