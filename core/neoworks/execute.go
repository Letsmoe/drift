package neoworks

import (
	"context"
	"fmt"

	libneoworks "github.com/neoworks-dev/libneoworks-go"

	"github.com/letsmoe/drift/core/library"
)

func (pass *collectionPass) execute(ctx context.Context, key string, chosen action) error {
	switch chosen {
	case actionCreateNode:
		return pass.createNode(ctx, key)
	case actionUpdateNode:
		return pass.updateNode(ctx, key)
	case actionDeleteNode:
		return pass.deleteNode(ctx, key)
	case actionApplyRemote:
		return pass.applyRemote(ctx, key)
	case actionRemoveLocal:
		return pass.removeLocal(ctx, key)
	case actionLinkOnly:
		return pass.linkRemote(ctx, key)
	case actionDropLink:
		return pass.librarySync.library.DeleteNodeLink(ctx, pass.mapping.name(), key)
	}
	return nil
}

func (pass *collectionPass) createNode(ctx context.Context, key string) error {
	item := pass.local[key]
	content, err := item.content()
	if err != nil {
		return err
	}
	nodeID, err := pass.librarySync.nodes.CreateNode(libneoworks.NewNode{
		ParentID:   pass.rootID,
		Collection: pass.path(),
		Kind:       libneoworks.NodeKindItem,
		Content:    content,
	})
	if err != nil {
		return fmt.Errorf("create node: %w", err)
	}
	pass.report.NodesWritten++
	pass.written[key] = nodeID
	return pass.saveLink(ctx, key, nodeID, item.hash(), "")
}

func (pass *collectionPass) updateNode(ctx context.Context, key string) error {
	item := pass.local[key]
	node := pass.remote[key]
	changes, err := pass.changesFor(item)
	if err != nil {
		return err
	}
	err = pass.librarySync.nodes.UpdateNode(libneoworks.NodeUpdate{NodeID: node.summary.ID, Changes: changes})
	if err != nil {
		return fmt.Errorf("update node %s: %w", node.summary.ID, err)
	}
	pass.report.NodesWritten++
	pass.written[key] = node.summary.ID
	return pass.saveLink(ctx, key, node.summary.ID, item.hash(), "")
}

// changesFor is the whole content, with the optional fields the entity lacks cleared.
func (pass *collectionPass) changesFor(item entity) (map[string]any, error) {
	changes, err := item.content()
	if err != nil {
		return nil, err
	}
	for _, field := range pass.mapping.optionalFields() {
		if _, present := changes[field]; !present {
			changes[field] = nil
		}
	}
	return changes, nil
}

func (pass *collectionPass) deleteNode(ctx context.Context, key string) error {
	node := pass.remote[key]
	err := pass.librarySync.nodes.DeleteNode(node.summary.ID, false)
	if err != nil {
		return fmt.Errorf("delete node %s: %w", node.summary.ID, err)
	}
	pass.report.NodesWritten++
	return pass.librarySync.library.DeleteNodeLink(ctx, pass.mapping.name(), key)
}

func (pass *collectionPass) applyRemote(ctx context.Context, key string) error {
	node := pass.remote[key]
	item, err := pass.read(node)
	if err != nil {
		return err
	}
	item.key = key
	err = pass.mapping.apply(ctx, item)
	if err != nil {
		return fmt.Errorf("apply node %s: %w", node.summary.ID, err)
	}
	pass.report.LibraryWritten++
	return pass.saveLink(ctx, key, node.summary.ID, item.hash(), node.summary.UpdatedAt)
}

func (pass *collectionPass) removeLocal(ctx context.Context, key string) error {
	err := pass.mapping.remove(ctx, key)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}
	pass.report.LibraryWritten++
	return pass.librarySync.library.DeleteNodeLink(ctx, pass.mapping.name(), key)
}

func (pass *collectionPass) linkRemote(ctx context.Context, key string) error {
	node := pass.remote[key]
	item, err := pass.read(node)
	if err != nil {
		return err
	}
	return pass.saveLink(ctx, key, node.summary.ID, item.hash(), node.summary.UpdatedAt)
}

func (pass *collectionPass) saveLink(ctx context.Context, key, nodeID, hash, nodeUpdatedAt string) error {
	return pass.librarySync.library.SaveNodeLink(ctx, library.NodeLink{
		Collection:    pass.mapping.name(),
		Key:           key,
		NodeID:        nodeID,
		ContentHash:   hash,
		NodeUpdatedAt: nodeUpdatedAt,
	})
}

// refreshWrittenLinks records the update time the store gave the nodes this
// pass wrote, so the next pass does not mistake them for remote edits.
func (pass *collectionPass) refreshWrittenLinks(ctx context.Context) error {
	if len(pass.written) == 0 {
		return nil
	}
	summaries, err := pass.librarySync.nodes.ListNodes(pass.rootID, pass.path())
	if err != nil {
		return fmt.Errorf("list written nodes: %w", err)
	}
	updatedAt := map[string]string{}
	for _, summary := range summaries {
		updatedAt[summary.ID] = summary.UpdatedAt
	}
	links, err := pass.librarySync.library.NodeLinks(ctx, pass.mapping.name())
	if err != nil {
		return err
	}
	for key, nodeID := range pass.written {
		link, found := links[key]
		if !found || link.NodeID != nodeID {
			continue
		}
		link.NodeUpdatedAt = updatedAt[nodeID]
		err = pass.librarySync.library.SaveNodeLink(ctx, link)
		if err != nil {
			return err
		}
	}
	return nil
}
