package main

import (
	"context"

	"github.com/letsmoe/drift/core/music"
	"github.com/letsmoe/drift/core/queue"
)

// queueEvent is the event carrying every new queue.State.
const queueEvent = "queue"

// QueueStep is the result of QueueNext and QueueAdvance. Moved is false when
// the queue had nowhere to go (stop playback); after QueueAdvance with repeat
// one it is true while the index stays put (restart the track).
type QueueStep struct {
	State queue.State `json:"state"`
	Moved bool        `json:"moved"`
}

// engineExtender feeds autoplay and magic shuffle from the recommendation engine once it exists.
type engineExtender struct {
	app *App
}

func (extender engineExtender) Extend(ctx context.Context, upcoming, history []music.Track) ([]music.Track, error) {
	// In someone else's Listen Together room the host's autoplay extends the shared queue.
	if extender.app.engine == nil || extender.app.following() {
		return nil, nil
	}
	return extender.app.engine.Extend(ctx, upcoming, history)
}

func (extender engineExtender) Mix(ctx context.Context, seeds, exclude []music.Track, count int) ([]music.Track, error) {
	if extender.app.engine == nil {
		return nil, nil
	}
	return extender.app.engine.Mix(ctx, seeds, exclude, count)
}

func (app *App) newQueue() *queue.Queue {
	return queue.New(
		queue.WithExtender(engineExtender{app: app}),
		queue.WithRecommender(engineExtender{app: app}),
		queue.WithOnChange(app.emitQueueState),
		queue.WithLogger(app.logger),
	)
}

func (app *App) emitQueueState(state queue.State) {
	app.emit(queueEvent, state)
	if app.together != nil {
		app.together.QueueChanged(state)
	}
}

// QueueState returns the current queue.
func (app *App) QueueState() queue.State {
	return app.queue.State()
}

// QueuePlay replaces the queue with tracks and starts at startIndex.
func (app *App) QueuePlay(tracks []music.Track, kind string, startIndex int, playContext string) queue.State {
	return app.queue.Play(tracks, music.ItemKind(kind), startIndex, playContext)
}

// QueuePlayNext inserts tracks after the current entry.
func (app *App) QueuePlayNext(tracks []music.Track) queue.State {
	return app.queue.PlayNext(tracks)
}

// QueueAdd appends tracks to the queue.
func (app *App) QueueAdd(tracks []music.Track) queue.State {
	return app.queue.Add(tracks)
}

// QueueRemove removes an entry.
func (app *App) QueueRemove(entryID int64) queue.State {
	return app.queue.Remove(entryID)
}

// QueueMove moves an entry to newIndex.
func (app *App) QueueMove(entryID int64, newIndex int) queue.State {
	return app.queue.Move(entryID, newIndex)
}

// QueueJump makes an entry current.
func (app *App) QueueJump(entryID int64) queue.State {
	return app.queue.Jump(entryID)
}

// QueueClear empties the queue.
func (app *App) QueueClear() queue.State {
	return app.queue.Clear()
}

// QueueNext is a user skip to the next entry.
func (app *App) QueueNext() QueueStep {
	state, moved := app.queue.Next()
	return QueueStep{State: state, Moved: moved}
}

// QueueAdvance is called when the current track ended.
func (app *App) QueueAdvance() QueueStep {
	state, moved := app.queue.Advance()
	return QueueStep{State: state, Moved: moved}
}

// QueuePrevious moves back one entry.
func (app *App) QueuePrevious() queue.State {
	return app.queue.Previous()
}

// QueueSetShuffle turns shuffle on or off.
func (app *App) QueueSetShuffle(enabled bool) queue.State {
	return app.queue.SetShuffle(enabled)
}

// QueueSetMagicShuffle turns magic shuffle (shuffle plus mixed-in recommendations) on or off.
func (app *App) QueueSetMagicShuffle(enabled bool) queue.State {
	return app.queue.SetMagicShuffle(enabled)
}

// QueueSetRepeat sets the repeat mode: "off", "all" or "one".
func (app *App) QueueSetRepeat(mode string) queue.State {
	return app.queue.SetRepeat(queue.RepeatMode(mode))
}

// QueueSetAutoplay turns autoplay on or off.
func (app *App) QueueSetAutoplay(enabled bool) queue.State {
	return app.queue.SetAutoplay(enabled)
}
