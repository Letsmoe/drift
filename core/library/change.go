package library

// ChangeKind says which part of the library a Change touched.
type ChangeKind string

// Change kinds.
const (
	// ChangeLikes is any change to the liked songs.
	ChangeLikes ChangeKind = "likes"
	// ChangePlaylist is any change to a local playlist; Change.ID is its ID.
	ChangePlaylist ChangeKind = "playlist"
)

// Change describes a committed edit to likes or a local playlist.
type Change struct {
	Kind ChangeKind
	ID   string
}

// OnChange registers the listener called after each committed change, on the
// goroutine that made it; it must not block. A new listener replaces the old.
func (library *Library) OnChange(listener func(Change)) {
	library.listenerMutex.Lock()
	defer library.listenerMutex.Unlock()
	library.changeListener = listener
}

// notifyChange passes err through, emitting change only when it is nil.
func (library *Library) notifyChange(err error, change Change) error {
	if err != nil {
		return err
	}
	library.emitChange(change)
	return nil
}

func (library *Library) emitChange(change Change) {
	library.listenerMutex.RLock()
	listener := library.changeListener
	library.listenerMutex.RUnlock()
	if listener != nil {
		listener(change)
	}
}
