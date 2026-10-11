package library

import "context"

// ChangeKind says which part of the library a Change touched.
type ChangeKind string

// Change kinds.
const (
	// ChangeLikes is any change to the ratings, likes and dislikes.
	ChangeLikes ChangeKind = "likes"
	// ChangePlaylist is any change to a local playlist; Change.ID is its ID.
	ChangePlaylist ChangeKind = "playlist"
	// ChangeSaved is any change to the saved albums, artists, podcasts and playlists.
	ChangeSaved ChangeKind = "saved"
	// ChangeBlocks is any change to the recommendation blocks.
	ChangeBlocks ChangeKind = "blocks"
	// ChangeHistory is a new play, a finished play or any change to the search history.
	ChangeHistory ChangeKind = "history"
)

// Change describes a committed edit to the library.
type Change struct {
	Kind ChangeKind
	ID   string
}

type quietKey struct{}

// WithoutNotifications returns a context under which library methods make
// their edits without reporting them, for writes that apply a synced copy.
func WithoutNotifications(ctx context.Context) context.Context {
	return context.WithValue(ctx, quietKey{}, true)
}

// OnChange adds a listener called after each committed change, on the goroutine
// that made it; it must not block. Listeners are never removed.
func (library *Library) OnChange(listener func(Change)) {
	library.listenerMutex.Lock()
	defer library.listenerMutex.Unlock()
	library.changeListeners = append(library.changeListeners, listener)
}

// notifyChange passes err through, emitting change only when it is nil.
func (library *Library) notifyChange(ctx context.Context, err error, change Change) error {
	if err != nil {
		return err
	}
	library.emitChange(ctx, change)
	return nil
}

func (library *Library) emitChange(ctx context.Context, change Change) {
	if quiet, _ := ctx.Value(quietKey{}).(bool); quiet {
		return
	}
	library.listenerMutex.RLock()
	listeners := library.changeListeners
	library.listenerMutex.RUnlock()
	for _, listener := range listeners {
		listener(change)
	}
}
