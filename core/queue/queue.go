// Package queue is the playback queue: what plays now, what plays next, what
// played before, plus shuffle, repeat and autoplay. It has no UI; clients only
// render the State it returns.
package queue

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

// autoplayThreshold is how few upcoming entries trigger an autoplay extension.
const autoplayThreshold = 3

// historyLimit caps how much played history the Extender is given.
const historyLimit = 20

// RepeatMode says what happens when playback reaches the end of the queue or a track.
type RepeatMode string

// Repeat modes.
const (
	RepeatOff RepeatMode = "off"
	RepeatAll RepeatMode = "all"
	RepeatOne RepeatMode = "one"
)

// Entry is one queue slot. ID is unique per insertion, so the same track can
// be queued twice.
type Entry struct {
	ID    int64          `json:"id"`
	Track music.Track    `json:"track"`
	Kind  music.ItemKind `json:"kind"`
	// Autoplay marks entries added by the Extender rather than the user.
	Autoplay bool `json:"autoplay"`
}

// State is a snapshot of the queue.
type State struct {
	// Entries are in play order: played history, the current entry, upcoming.
	Entries []Entry `json:"entries"`
	// Index is the current entry, or -1 when the queue is empty.
	Index    int        `json:"index"`
	Shuffle  bool       `json:"shuffle"`
	Repeat   RepeatMode `json:"repeat"`
	Autoplay bool       `json:"autoplay"`
	// Context says where playback was started, e.g. "album:MPREb_x" or "radio".
	Context string `json:"context"`
}

// Extender supplies more tracks when the queue runs low. history ends with the
// current track (oldest first, at most 20); upcoming is what is still queued.
type Extender interface {
	Extend(ctx context.Context, upcoming []music.Track, history []music.Track) ([]music.Track, error)
}

// Option configures New.
type Option func(*Queue)

// WithExtender sets the source of autoplay tracks.
func WithExtender(extender Extender) Option {
	return func(queue *Queue) {
		queue.extender = extender
	}
}

// WithOnChange registers a callback that receives every new State, including
// those caused by autoplay extensions. It is called without the queue locked.
func WithOnChange(onChange func(State)) Option {
	return func(queue *Queue) {
		queue.onChange = onChange
	}
}

// WithLogger sets where Extender errors are logged.
func WithLogger(logger *slog.Logger) Option {
	return func(queue *Queue) {
		queue.logger = logger
	}
}

// WithRandom sets the randomness used for shuffling, for deterministic tests.
func WithRandom(random *rand.Rand) Option {
	return func(queue *Queue) {
		queue.random = random
	}
}

// item is an entry plus the position it had in insertion order, which shuffle
// off restores.
type item struct {
	entry Entry
	order int64
}

// Queue is safe for concurrent use.
type Queue struct {
	mutex    sync.Mutex
	items    []item
	index    int
	shuffle  bool
	repeat   RepeatMode
	autoplay bool
	context  string

	nextEntryID int64
	// generation changes when the queue is replaced or cleared, so extension
	// results computed for the old queue can be recognized and dropped.
	generation int
	extending  bool

	extender Extender
	onChange func(State)
	logger   *slog.Logger
	random   *rand.Rand
}

// New returns an empty queue with autoplay on and repeat off.
func New(options ...Option) *Queue {
	queue := &Queue{
		index:    -1,
		repeat:   RepeatOff,
		autoplay: true,
		logger:   slog.Default(),
		random:   rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0)),
	}
	for _, option := range options {
		option(queue)
	}
	return queue
}

// State returns a snapshot of the queue.
func (queue *Queue) State() State {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	return queue.stateLocked()
}

func (queue *Queue) stateLocked() State {
	entries := make([]Entry, len(queue.items))
	for position, queued := range queue.items {
		entries[position] = queued.entry
	}
	return State{
		Entries:  entries,
		Index:    queue.index,
		Shuffle:  queue.shuffle,
		Repeat:   queue.repeat,
		Autoplay: queue.autoplay,
		Context:  queue.context,
	}
}

// apply runs change under the lock. When change reports a modification, the
// queue is renumbered, onChange fires and autoplay may start extending.
func (queue *Queue) apply(change func() bool) State {
	queue.mutex.Lock()
	changed := change()
	if !changed {
		state := queue.stateLocked()
		queue.mutex.Unlock()
		return state
	}
	queue.renumberLocked()
	state := queue.stateLocked()
	request := queue.extensionRequestLocked()
	queue.mutex.Unlock()
	queue.notify(state)
	queue.startExtension(request)
	return state
}

func (queue *Queue) notify(state State) {
	if queue.onChange != nil {
		queue.onChange(state)
	}
}

// renumberLocked keeps insertion order equal to play order while shuffle is
// off; while it is on, the recorded order is what shuffle off restores.
func (queue *Queue) renumberLocked() {
	if queue.shuffle {
		return
	}
	for position := range queue.items {
		queue.items[position].order = int64(position)
	}
}

func (queue *Queue) newItemLocked(track music.Track, kind music.ItemKind, autoplay bool) item {
	queue.nextEntryID++
	return item{entry: Entry{ID: queue.nextEntryID, Track: track, Kind: kind, Autoplay: autoplay}}
}

func (queue *Queue) positionOfLocked(entryID int64) int {
	for position, queued := range queue.items {
		if queued.entry.ID == entryID {
			return position
		}
	}
	return -1
}
