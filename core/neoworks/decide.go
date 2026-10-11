package neoworks

import "time"

// action is what one entity needs so that its library row and its node agree.
type action int

const (
	actionNothing action = iota
	// actionLinkOnly records the node's current state; library and node already agree.
	actionLinkOnly
	actionCreateNode
	actionUpdateNode
	actionDeleteNode
	actionApplyRemote
	actionRemoveLocal
	actionDropLink
)

// situation is everything decide looks at for one natural key.
type situation struct {
	hasLocal       bool
	localHash      string
	localUpdatedAt time.Time

	hasNode bool
	// remoteUpdatedAt is when the node's content was last changed: the update time
	// in the content when it has one, else the node's own update time.
	remoteUpdatedAt time.Time
	// remoteHash is the node's content hash; for a linked node whose update time
	// did not move since the last sync it is the link's hash.
	remoteHash string
	// nodeMoved says the node's update time differs from the one the link recorded.
	nodeMoved bool

	linked   bool
	linkHash string
	// tombstone says the linked node is known to be deleted.
	tombstone bool

	owned      bool
	keepsNodes bool
}

// decide applies the merge rules. A side that changed since the link wins over
// one that did not; when both changed, or nothing was linked yet and the
// contents differ, the newer updatedAt wins. A deletion on one side removes
// the other only when that other did not change since the link. Entities this
// install owns are never overwritten from a node, and nodes that outlive their
// entities are never deleted.
func decide(state situation) action {
	if !state.hasNode {
		return decideWithoutNode(state)
	}
	if !state.linked {
		return decideUnlinked(state)
	}
	return decideLinked(state)
}

func decideWithoutNode(state situation) action {
	if !state.linked {
		return createIfLocal(state)
	}
	if !state.tombstone {
		return actionNothing
	}
	if !state.hasLocal {
		return actionDropLink
	}
	if state.owned || state.keepsNodes || state.localHash != state.linkHash {
		return actionCreateNode
	}
	return actionRemoveLocal
}

func createIfLocal(state situation) action {
	if state.hasLocal {
		return actionCreateNode
	}
	return actionNothing
}

func decideUnlinked(state situation) action {
	if !state.hasLocal {
		return applyUnlessOwned(state)
	}
	if state.localHash == state.remoteHash {
		return actionLinkOnly
	}
	if state.owned || state.localUpdatedAt.After(state.remoteUpdatedAt) {
		return actionUpdateNode
	}
	return actionApplyRemote
}

func applyUnlessOwned(state situation) action {
	if state.owned {
		return actionNothing
	}
	return actionApplyRemote
}

func decideLinked(state situation) action {
	remoteChanged := !state.owned && state.remoteHash != state.linkHash
	if !state.hasLocal {
		return decideLocalDeletion(state, remoteChanged)
	}
	localChanged := state.localHash != state.linkHash
	switch {
	case remoteChanged && localChanged:
		return newerSideWins(state)
	case remoteChanged:
		return actionApplyRemote
	case localChanged:
		return actionUpdateNode
	case state.nodeMoved:
		return actionLinkOnly
	}
	return actionNothing
}

func decideLocalDeletion(state situation, remoteChanged bool) action {
	if state.owned || state.keepsNodes {
		return actionNothing
	}
	if remoteChanged {
		return actionApplyRemote
	}
	return actionDeleteNode
}

func newerSideWins(state situation) action {
	if state.localUpdatedAt.After(state.remoteUpdatedAt) {
		return actionUpdateNode
	}
	return actionApplyRemote
}
