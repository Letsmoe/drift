package queue

import (
	"slices"

	"github.com/letsmoe/drift-music/core/music"
)

// PlayNext inserts tracks right after the current entry. Into an empty queue
// they become the queue and the first one is current. They are queued as songs.
func (queue *Queue) PlayNext(tracks []music.Track) State {
	return queue.apply(func() bool {
		queue.requestMixLocked()
		return queue.insertLocked(queue.index+1, tracks, false, queue.frontOrderLocked(len(tracks)))
	})
}

// Add appends tracks after the user-added entries but before any autoplay
// entries, which always stay at the tail. Into an empty queue they become the
// queue and the first one is current. They are queued as songs.
func (queue *Queue) Add(tracks []music.Track) State {
	return queue.apply(func() bool {
		queue.requestMixLocked()
		return queue.insertLocked(queue.autoplayTailLocked(), tracks, false, queue.endOrderLocked())
	})
}

func (queue *Queue) autoplayTailLocked() int {
	for position := queue.index + 1; position < len(queue.items); position++ {
		if queue.items[position].entry.Autoplay {
			return position
		}
	}
	return len(queue.items)
}

func (queue *Queue) insertLocked(position int, tracks []music.Track, autoplay bool, firstOrder int64) bool {
	if len(tracks) == 0 {
		return false
	}
	inserted := make([]item, len(tracks))
	for offset, track := range tracks {
		inserted[offset] = queue.newItemLocked(track, music.ItemKindSong, autoplay)
		inserted[offset].order = firstOrder + int64(offset)
	}
	queue.items = slices.Insert(queue.items, position, inserted...)
	if queue.index < 0 {
		queue.index = 0
	}
	return true
}

// frontOrderLocked is an order key that sorts count entries before every
// upcoming entry. Order only matters to shuffle off, which restores it.
func (queue *Queue) frontOrderLocked(count int) int64 {
	if queue.index+1 >= len(queue.items) {
		return queue.endOrderLocked()
	}
	lowest := queue.items[queue.index+1].order
	for _, queued := range queue.items[queue.index+1:] {
		lowest = min(lowest, queued.order)
	}
	return lowest - int64(count)
}

// endOrderLocked is an order key that sorts after every entry.
func (queue *Queue) endOrderLocked() int64 {
	var highest int64 = -1
	for _, queued := range queue.items {
		highest = max(highest, queued.order)
	}
	return highest + 1
}

// Remove deletes an entry. Removing the current entry makes the next one current.
func (queue *Queue) Remove(entryID int64) State {
	return queue.apply(func() bool {
		position := queue.positionOfLocked(entryID)
		if position < 0 {
			return false
		}
		queue.items = slices.Delete(queue.items, position, position+1)
		if position < queue.index {
			queue.index--
		}
		queue.index = min(queue.index, len(queue.items)-1)
		return true
	})
}

// Move puts an upcoming entry at newIndex (clamped to the upcoming range).
// History and the current entry cannot be moved. A moved entry counts as
// chosen by the user, so it is no longer an autoplay or magic entry.
func (queue *Queue) Move(entryID int64, newIndex int) State {
	return queue.apply(func() bool {
		position := queue.positionOfLocked(entryID)
		if position <= queue.index {
			return false
		}
		moved := queue.items[position]
		moved.entry.Autoplay = false
		moved.entry.Magic = false
		queue.items = slices.Delete(queue.items, position, position+1)
		target := min(max(newIndex, queue.index+1), len(queue.items))
		queue.items = slices.Insert(queue.items, target, moved)
		return true
	})
}

// Clear removes everything but the current entry and cancels pending autoplay extensions.
func (queue *Queue) Clear() State {
	return queue.apply(func() bool {
		queue.generation++
		if queue.index < 0 {
			return false
		}
		queue.items = []item{queue.items[queue.index]}
		queue.index = 0
		return true
	})
}
