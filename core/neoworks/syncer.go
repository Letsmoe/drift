package neoworks

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	libneoworks "github.com/neoworks-dev/libneoworks-go"

	"github.com/letsmoe/drift/core/library"
)

const (
	// DefaultDebounce is how long the library must stay quiet before a change is synced.
	DefaultDebounce = time.Second
	// DefaultInterval is how often a signed-in install syncs without any change.
	// The change feed announces remote changes, so this is only the fallback for
	// when it is not connected.
	DefaultInterval = 15 * time.Minute
)

// syncSession is the part of Session a Syncer drives.
type syncSession interface {
	SignedIn() bool
	Sync(ctx context.Context) (libneoworks.SyncReport, error)
	SignOut() error
}

// reconciler maps the library to nodes and back.
type reconciler interface {
	Reconcile(ctx context.Context) (ReconcileReport, error)
	Forget(ctx context.Context) error
}

// SyncState is what an observer hears about a pass.
type SyncState struct {
	// Syncing is true when the pass starts and false when it ended.
	Syncing bool
	// Err is the pass's failure; set when Syncing is false.
	Err error
	// Report says what the pass wrote; set when Syncing is false.
	Report ReconcileReport
}

// SyncerConfig holds what a Syncer needs; Debounce and Interval default to the package defaults.
type SyncerConfig struct {
	Session     syncSession
	LibrarySync reconciler
	Debounce    time.Duration
	Interval    time.Duration
	// OnState hears every pass start and end, from the goroutine that runs it.
	OnState func(SyncState)
	Logger  *slog.Logger
}

// Syncer keeps the library and Neoworks in step while signed in: a pass pulls,
// reconciles the library with the nodes and uploads what that wrote. Passes
// run one at a time, after local changes once the library was quiet for the
// debounce time, after remote changes the change feed announces, and on a
// fixed interval.
type Syncer struct {
	config   SyncerConfig
	requests chan struct{}

	passMutex sync.Mutex
	running   sync.WaitGroup
	cancel    context.CancelFunc

	sequenceMutex sync.Mutex
	// announcedSequence is the newest remote sequence the change feed named.
	announcedSequence int64
	// pulledSequence is the newest announced sequence a pass started after,
	// which therefore pulled it.
	pulledSequence int64
}

// NewSyncer builds a Syncer; Start runs it.
func NewSyncer(config SyncerConfig) *Syncer {
	if config.Debounce <= 0 {
		config.Debounce = DefaultDebounce
	}
	if config.Interval <= 0 {
		config.Interval = DefaultInterval
	}
	if config.Logger == nil {
		config.Logger = slog.New(slog.DiscardHandler)
	}
	return &Syncer{config: config, requests: make(chan struct{}, 1)}
}

// Watch syncs after every change to the library.
func (syncer *Syncer) Watch(store *library.Library) {
	store.OnChange(func(library.Change) { syncer.Request() })
}

// Request schedules a pass after the debounce time; requests made before it
// passes push it back.
func (syncer *Syncer) Request() {
	select {
	case syncer.requests <- struct{}{}:
	default:
	}
}

// NoteRemoteChange schedules a pass when sequence is newer than what the last
// pass pulled. The sequences are the change feed's event ids.
func (syncer *Syncer) NoteRemoteChange(sequence int64) {
	syncer.sequenceMutex.Lock()
	newer := sequence > syncer.pulledSequence
	syncer.announcedSequence = max(syncer.announcedSequence, sequence)
	syncer.sequenceMutex.Unlock()
	if newer {
		syncer.Request()
	}
}

// NoteResync schedules a pass because the change feed cannot say what changed.
// Sequences may start over after it, so none counts as pulled any more.
func (syncer *Syncer) NoteResync() {
	syncer.forgetSequences()
	syncer.Request()
}

func (syncer *Syncer) forgetSequences() {
	syncer.sequenceMutex.Lock()
	defer syncer.sequenceMutex.Unlock()
	syncer.announcedSequence = 0
	syncer.pulledSequence = 0
}

func (syncer *Syncer) announced() int64 {
	syncer.sequenceMutex.Lock()
	defer syncer.sequenceMutex.Unlock()
	return syncer.announcedSequence
}

// markPulled records that a pull ran after the feed had announced sequence.
func (syncer *Syncer) markPulled(sequence int64) {
	syncer.sequenceMutex.Lock()
	defer syncer.sequenceMutex.Unlock()
	syncer.pulledSequence = max(syncer.pulledSequence, sequence)
}

// Start runs the automatic passes until ctx ends or Stop is called, beginning
// with one now. It must be called once.
func (syncer *Syncer) Start(ctx context.Context) {
	ctx, syncer.cancel = context.WithCancel(ctx)
	syncer.running.Add(1)
	go func() {
		defer syncer.running.Done()
		syncer.loop(ctx)
	}()
	syncer.Request()
}

// Stop ends the automatic passes and waits for every running pass; the
// session must not be closed before.
func (syncer *Syncer) Stop() {
	if syncer.cancel != nil {
		syncer.cancel()
	}
	syncer.running.Wait()
}

func (syncer *Syncer) loop(ctx context.Context) {
	ticker := time.NewTicker(syncer.config.Interval)
	defer ticker.Stop()
	var due <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-syncer.requests:
			due = time.After(syncer.config.Debounce)
		case <-due:
			due = nil
			syncer.automaticPass(ctx)
		case <-ticker.C:
			syncer.automaticPass(ctx)
		}
	}
}

func (syncer *Syncer) automaticPass(ctx context.Context) {
	_, err := syncer.SyncNow(ctx)
	if err != nil && !errors.Is(err, ErrSignedOut) && !errors.Is(err, context.Canceled) {
		syncer.config.Logger.Warn("neoworks: sync", "error", err)
	}
}

// SyncNow runs one pass and waits for it, behind a pass that is running.
func (syncer *Syncer) SyncNow(ctx context.Context) (ReconcileReport, error) {
	syncer.running.Add(1)
	defer syncer.running.Done()
	syncer.passMutex.Lock()
	defer syncer.passMutex.Unlock()
	if !syncer.config.Session.SignedIn() {
		return ReconcileReport{}, ErrSignedOut
	}
	syncer.observe(SyncState{Syncing: true})
	report, err := syncer.pass(ctx, syncer.announced())
	syncer.observe(SyncState{Err: err, Report: report})
	return report, err
}

func (syncer *Syncer) observe(state SyncState) {
	if syncer.config.OnState != nil {
		syncer.config.OnState(state)
	}
}

// pass pulls, reconciles and uploads. The upload only runs when the
// reconcile wrote nodes, and it runs even when part of the reconcile failed.
// announced is the newest sequence the change feed named before the pass; the
// pull covers it.
func (syncer *Syncer) pass(ctx context.Context, announced int64) (ReconcileReport, error) {
	_, err := syncer.config.Session.Sync(ctx)
	if err != nil {
		return ReconcileReport{}, err
	}
	syncer.markPulled(announced)
	report, reconcileErr := syncer.config.LibrarySync.Reconcile(ctx)
	if report.NodesWritten == 0 {
		return report, reconcileErr
	}
	_, uploadErr := syncer.config.Session.Sync(ctx)
	return report, errors.Join(reconcileErr, uploadErr)
}

// SignOut forgets which nodes hold which library entities, then signs out and
// wipes the synced copy, after a running pass. The library itself stays.
func (syncer *Syncer) SignOut(ctx context.Context) error {
	syncer.passMutex.Lock()
	defer syncer.passMutex.Unlock()
	err := syncer.config.LibrarySync.Forget(ctx)
	if err != nil {
		return err
	}
	syncer.forgetSequences()
	return syncer.config.Session.SignOut()
}

var (
	_ syncSession = (*Session)(nil)
	_ reconciler  = (*LibrarySync)(nil)
)
