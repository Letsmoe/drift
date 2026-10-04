package mobile

import (
	"context"

	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/queue"
)

// queueStep is the result of QueueNext and QueueAdvance. Moved is false when
// the queue had nowhere to go (stop playback); after QueueAdvance with repeat
// one it is true while the index stays put (restart the track).
type queueStep struct {
	State queue.State `json:"state"`
	Moved bool        `json:"moved"`
}

// engineExtender feeds autoplay from the recommendation engine once it exists.
type engineExtender struct {
	core *Core
}

func (extender engineExtender) Extend(ctx context.Context, upcoming, history []music.Track) ([]music.Track, error) {
	if extender.core.engine == nil {
		return nil, nil
	}
	return extender.core.engine.Extend(ctx, upcoming, history)
}

func (core *Core) newQueue() *queue.Queue {
	return queue.New(
		queue.WithExtender(engineExtender{core: core}),
		queue.WithOnChange(core.emitQueueState),
		queue.WithLogger(core.logger),
	)
}

func (core *Core) emitQueueState(state queue.State) {
	if core.listener == nil {
		return
	}
	core.listener.OnQueueChanged(mustEncode(state))
}

// QueueState returns the current queue.State JSON.
func (core *Core) QueueState() string {
	return mustEncode(core.queue.State())
}

// QueuePlay replaces the queue with tracks (JSON array) and starts at startIndex.
func (core *Core) QueuePlay(tracksJSON string, kind string, startIndex int, playContext string) (string, error) {
	var tracks []music.Track
	err := decode(tracksJSON, &tracks)
	if err != nil {
		return "", err
	}
	return mustEncode(core.queue.Play(tracks, music.ItemKind(kind), startIndex, playContext)), nil
}

// QueuePlayNext inserts tracks after the current entry.
func (core *Core) QueuePlayNext(tracksJSON string) (string, error) {
	var tracks []music.Track
	err := decode(tracksJSON, &tracks)
	if err != nil {
		return "", err
	}
	return mustEncode(core.queue.PlayNext(tracks)), nil
}

// QueueAdd appends tracks to the queue.
func (core *Core) QueueAdd(tracksJSON string) (string, error) {
	var tracks []music.Track
	err := decode(tracksJSON, &tracks)
	if err != nil {
		return "", err
	}
	return mustEncode(core.queue.Add(tracks)), nil
}

// QueueRemove removes an entry.
func (core *Core) QueueRemove(entryID int64) string {
	return mustEncode(core.queue.Remove(entryID))
}

// QueueMove moves an entry to newIndex.
func (core *Core) QueueMove(entryID int64, newIndex int) string {
	return mustEncode(core.queue.Move(entryID, newIndex))
}

// QueueJump makes an entry current.
func (core *Core) QueueJump(entryID int64) string {
	return mustEncode(core.queue.Jump(entryID))
}

// QueueClear empties the queue.
func (core *Core) QueueClear() string {
	return mustEncode(core.queue.Clear())
}

// QueueNext is a user skip to the next entry; it returns {"state", "moved"} JSON.
func (core *Core) QueueNext() string {
	state, moved := core.queue.Next()
	return mustEncode(queueStep{State: state, Moved: moved})
}

// QueueAdvance is called when the current track ended; same JSON as QueueNext.
func (core *Core) QueueAdvance() string {
	state, moved := core.queue.Advance()
	return mustEncode(queueStep{State: state, Moved: moved})
}

// QueuePrevious moves back one entry.
func (core *Core) QueuePrevious() string {
	return mustEncode(core.queue.Previous())
}

// QueueSetShuffle turns shuffle on or off.
func (core *Core) QueueSetShuffle(enabled bool) string {
	return mustEncode(core.queue.SetShuffle(enabled))
}

// QueueSetRepeat sets the repeat mode: "off", "all" or "one".
func (core *Core) QueueSetRepeat(mode string) string {
	return mustEncode(core.queue.SetRepeat(queue.RepeatMode(mode)))
}

// QueueSetAutoplay turns autoplay on or off.
func (core *Core) QueueSetAutoplay(enabled bool) string {
	return mustEncode(core.queue.SetAutoplay(enabled))
}
