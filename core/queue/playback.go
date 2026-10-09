package queue

import (
	"slices"

	"github.com/letsmoe/drift-music/core/music"
)

// Play replaces the queue with tracks and starts at startIndex. With shuffle
// on, the start track plays first and the rest follow shuffled.
func (queue *Queue) Play(tracks []music.Track, kind music.ItemKind, startIndex int, playContext string) State {
	return queue.apply(func() bool {
		queue.generation++
		queue.context = playContext
		queue.items = nil
		queue.index = -1
		for position, track := range tracks {
			queued := queue.newItemLocked(track, kind, false)
			queued.order = int64(position)
			queue.items = append(queue.items, queued)
		}
		if len(queue.items) == 0 {
			return true
		}
		queue.index = min(max(startIndex, 0), len(queue.items)-1)
		if queue.shuffle {
			queue.moveCurrentToFrontAndShuffleRestLocked()
		}
		queue.requestMixLocked()
		return true
	})
}

func (queue *Queue) moveCurrentToFrontAndShuffleRestLocked() {
	current := queue.items[queue.index]
	rest := slices.Delete(slices.Clone(queue.items), queue.index, queue.index+1)
	queue.random.Shuffle(len(rest), func(left, right int) {
		rest[left], rest[right] = rest[right], rest[left]
	})
	queue.items = append([]item{current}, rest...)
	queue.index = 0
}

// Next is a user skip: it always advances. At the end it wraps with RepeatAll;
// otherwise the queue stays put and ok is false, so the UI knows to stop.
func (queue *Queue) Next() (state State, ok bool) {
	state = queue.apply(func() bool {
		ok = queue.stepForwardLocked()
		return ok
	})
	return state, ok
}

// Advance is called when a track finished. RepeatOne replays the same entry
// (ok is true and the index is unchanged, so the UI must restart it);
// otherwise it behaves like Next.
func (queue *Queue) Advance() (state State, ok bool) {
	state = queue.apply(func() bool {
		if queue.repeat == RepeatOne && queue.index >= 0 {
			ok = true
			return false
		}
		ok = queue.stepForwardLocked()
		return ok
	})
	return state, ok
}

func (queue *Queue) stepForwardLocked() bool {
	if queue.index+1 < len(queue.items) {
		queue.index++
		return true
	}
	if queue.repeat == RepeatAll && len(queue.items) > 0 {
		queue.index = 0
		return true
	}
	return false
}

// Previous moves back one entry. At the start it wraps with RepeatAll and
// otherwise stays.
func (queue *Queue) Previous() State {
	return queue.apply(func() bool {
		if queue.index > 0 {
			queue.index--
			return true
		}
		if queue.repeat == RepeatAll && len(queue.items) > 1 {
			queue.index = len(queue.items) - 1
			return true
		}
		return false
	})
}

// Jump makes the given entry current.
func (queue *Queue) Jump(entryID int64) State {
	return queue.apply(func() bool {
		position := queue.positionOfLocked(entryID)
		if position < 0 || position == queue.index {
			return false
		}
		queue.index = position
		return true
	})
}

// SetShuffle shuffles the upcoming entries (the current one stays) or, when
// turned off, puts them back in insertion order. Turning it off also ends
// magic shuffle and removes its upcoming recommendations.
func (queue *Queue) SetShuffle(enabled bool) State {
	return queue.apply(func() bool {
		if queue.shuffle == enabled {
			return false
		}
		queue.shuffle = enabled
		if enabled {
			queue.shuffleUpcomingLocked()
			return true
		}
		if queue.magic {
			queue.disableMagicLocked()
		}
		queue.restoreUpcomingOrderLocked()
		return true
	})
}

func (queue *Queue) shuffleUpcomingLocked() {
	upcoming := queue.items[queue.index+1:]
	queue.random.Shuffle(len(upcoming), func(left, right int) {
		upcoming[left], upcoming[right] = upcoming[right], upcoming[left]
	})
}

// restoreUpcomingOrderLocked sorts by insertion order with autoplay entries last.
func (queue *Queue) restoreUpcomingOrderLocked() {
	upcoming := queue.items[queue.index+1:]
	slices.SortStableFunc(upcoming, func(left, right item) int {
		if left.entry.Autoplay != right.entry.Autoplay {
			if left.entry.Autoplay {
				return 1
			}
			return -1
		}
		return int(left.order - right.order)
	})
}

// SetRepeat sets the repeat mode.
func (queue *Queue) SetRepeat(mode RepeatMode) State {
	return queue.apply(func() bool {
		if queue.repeat == mode {
			return false
		}
		queue.repeat = mode
		return true
	})
}

// SetAutoplay turns autoplay (queue extension by the Extender) on or off.
func (queue *Queue) SetAutoplay(enabled bool) State {
	return queue.apply(func() bool {
		if queue.autoplay == enabled {
			return false
		}
		queue.autoplay = enabled
		return true
	})
}
