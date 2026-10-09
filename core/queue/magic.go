package queue

import (
	"context"
	"slices"

	"github.com/letsmoe/drift/core/music"
)

const (
	// magicSpacing is how many of the user's upcoming entries come before each recommendation.
	magicSpacing = 3
	// magicLimit caps the recommendations one mix asks for.
	magicLimit = 30
	// magicSeedLimit is how many queued tracks the recommendations are based on.
	magicSeedLimit = 5
)

// Recommender supplies magic shuffle's recommendations: count tracks that fit
// the seeds, none of them in exclude.
type Recommender interface {
	Mix(ctx context.Context, seeds, exclude []music.Track, count int) ([]music.Track, error)
}

// mixRequest is one Recommender call plus the queue generation it was made for.
type mixRequest struct {
	generation int
	seeds      []music.Track
	exclude    []music.Track
	count      int
}

// SetMagicShuffle turns magic shuffle on (which turns shuffle on) or off (which
// removes the recommendations still upcoming but keeps shuffle on).
func (queue *Queue) SetMagicShuffle(enabled bool) State {
	return queue.apply(func() bool {
		if queue.magic == enabled {
			return false
		}
		if !enabled {
			queue.disableMagicLocked()
			return true
		}
		queue.magic = true
		queue.magicPending = true
		if !queue.shuffle {
			queue.shuffle = true
			queue.shuffleUpcomingLocked()
		}
		return true
	})
}

// disableMagicLocked ends magic shuffle; played recommendations stay in the history.
func (queue *Queue) disableMagicLocked() {
	queue.magic = false
	queue.magicPending = false
	upcoming := slices.DeleteFunc(slices.Clone(queue.items[queue.index+1:]), func(queued item) bool {
		return queued.entry.Magic
	})
	queue.items = append(queue.items[:queue.index+1], upcoming...)
}

// requestMixLocked asks for recommendations for entries the user just queued.
func (queue *Queue) requestMixLocked() {
	if queue.magic {
		queue.magicPending = true
	}
}

// mixRequestLocked returns a request when magic shuffle is missing
// recommendations and marks a mix as running. It returns nil otherwise.
func (queue *Queue) mixRequestLocked() *mixRequest {
	if !queue.magic || !queue.magicPending || queue.mixing || queue.recommender == nil || queue.index < 0 {
		return nil
	}
	queue.magicPending = false
	count := queue.missingMagicLocked()
	if count == 0 {
		return nil
	}
	queue.mixing = true
	return &mixRequest{
		generation: queue.generation,
		seeds:      queue.magicSeedsLocked(),
		exclude:    queue.tracksLocked(0, len(queue.items)),
		count:      count,
	}
}

// missingMagicLocked is how many more recommendations the upcoming entries take.
func (queue *Queue) missingMagicLocked() int {
	userEntries, magicEntries := 0, 0
	for _, queued := range queue.items[queue.index+1:] {
		switch {
		case queued.entry.Magic:
			magicEntries++
		case !queued.entry.Autoplay:
			userEntries++
		}
	}
	return max(0, min(userEntries/magicSpacing, magicLimit)-magicEntries)
}

// magicSeedsLocked picks up to magicSeedLimit random user-chosen tracks of the queue.
func (queue *Queue) magicSeedsLocked() []music.Track {
	var chosen []music.Track
	for _, queued := range queue.items {
		if !queued.entry.Magic && !queued.entry.Autoplay {
			chosen = append(chosen, queued.entry.Track)
		}
	}
	queue.random.Shuffle(len(chosen), func(left, right int) {
		chosen[left], chosen[right] = chosen[right], chosen[left]
	})
	return chosen[:min(len(chosen), magicSeedLimit)]
}

func (queue *Queue) startMix(request *mixRequest) {
	if request == nil {
		return
	}
	go queue.runMix(*request)
}

func (queue *Queue) runMix(request mixRequest) {
	tracks, err := queue.recommender.Mix(context.Background(), request.seeds, request.exclude, request.count)
	if err != nil {
		queue.logExtensionFailure(err)
	}
	queue.finishMix(request.generation, tracks)
}

// finishMix mixes the recommendations into the queue they were made for. When
// the queue was replaced meanwhile they are dropped; a mix requested in the
// meantime starts now.
func (queue *Queue) finishMix(generation int, tracks []music.Track) {
	queue.mutex.Lock()
	queue.mixing = false
	changed := false
	if generation == queue.generation && queue.magic {
		changed = queue.insertMagicLocked(tracks)
	}
	if changed {
		queue.renumberLocked()
	}
	state := queue.stateLocked()
	next := queue.mixRequestLocked()
	queue.mutex.Unlock()
	if changed {
		queue.notify(state)
	}
	queue.startMix(next)
}

// insertMagicLocked places one recommendation after every magicSpacing user
// entries since the last recommendation, among the upcoming entries before the
// autoplay tail. Tracks already in the queue are skipped.
func (queue *Queue) insertMagicLocked(tracks []music.Track) bool {
	pending := queue.unqueuedLocked(tracks)
	if len(pending) == 0 {
		return false
	}
	tail := queue.autoplayTailLocked()
	upcoming := queue.items[queue.index+1 : tail]
	rebuilt := make([]item, 0, len(upcoming)+len(pending))
	sinceMagic := 0
	for position, queued := range upcoming {
		rebuilt = append(rebuilt, queued)
		sinceMagic = countSinceMagic(sinceMagic, queued)
		if sinceMagic < magicSpacing || len(pending) == 0 || nextIsMagic(upcoming, position) {
			continue
		}
		rebuilt = append(rebuilt, queue.newMagicItemLocked(pending[0]))
		pending = pending[1:]
		sinceMagic = 0
	}
	queue.items = slices.Concat(queue.items[:queue.index+1], rebuilt, queue.items[tail:])
	return true
}

func countSinceMagic(sinceMagic int, queued item) int {
	if queued.entry.Magic {
		return 0
	}
	return sinceMagic + 1
}

func nextIsMagic(upcoming []item, position int) bool {
	return position+1 < len(upcoming) && upcoming[position+1].entry.Magic
}

func (queue *Queue) newMagicItemLocked(track music.Track) item {
	magic := queue.newItemLocked(track, music.ItemKindSong, false)
	magic.entry.Magic = true
	magic.order = queue.endOrderLocked()
	return magic
}

// unqueuedLocked drops tracks that are anywhere in the queue, and duplicates.
func (queue *Queue) unqueuedLocked(tracks []music.Track) []music.Track {
	known := map[string]bool{}
	for _, queued := range queue.items {
		known[queued.entry.Track.VideoID] = true
	}
	var fresh []music.Track
	for _, track := range tracks {
		if known[track.VideoID] {
			continue
		}
		known[track.VideoID] = true
		fresh = append(fresh, track)
	}
	return fresh
}
