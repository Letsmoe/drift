package queue

// Adopt replaces the queue with state from another queue, as Listen Together
// does with the queue of whoever changed it last. Entry IDs are kept, so a
// player sees an unchanged current entry as the same entry and keeps playing.
// Pending extensions and mixes are dropped; adopting never starts a mix, but
// may start an autoplay extension when the adopted queue runs low.
func (queue *Queue) Adopt(state State) State {
	queue.mutex.Lock()
	queue.generation++
	queue.magicPending = false
	queue.items = itemsOf(state.Entries)
	queue.index = adoptedIndex(state.Index, len(queue.items))
	queue.shuffle = state.Shuffle
	queue.magic = state.MagicShuffle
	queue.repeat = adoptedRepeat(state.Repeat)
	queue.autoplay = state.Autoplay
	queue.context = state.Context
	queue.nextEntryID = max(queue.nextEntryID, highestEntryID(state.Entries))
	adopted := queue.stateLocked()
	request := queue.extensionRequestLocked()
	queue.mutex.Unlock()
	queue.notify(adopted)
	queue.startExtension(request)
	return adopted
}

// itemsOf keeps the given play order as the insertion order shuffle off restores.
func itemsOf(entries []Entry) []item {
	items := make([]item, len(entries))
	for position, entry := range entries {
		items[position] = item{entry: entry, order: int64(position)}
	}
	return items
}

func adoptedIndex(index, length int) int {
	if length == 0 {
		return -1
	}
	return min(max(index, 0), length-1)
}

func adoptedRepeat(mode RepeatMode) RepeatMode {
	if mode == RepeatAll || mode == RepeatOne {
		return mode
	}
	return RepeatOff
}

func highestEntryID(entries []Entry) int64 {
	var highest int64
	for _, entry := range entries {
		highest = max(highest, entry.ID)
	}
	return highest
}
