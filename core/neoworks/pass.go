package neoworks

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	libneoworks "github.com/neoworks-dev/libneoworks-go"

	"github.com/letsmoe/drift/core/library"
)

// remoteNode is the live node chosen for a natural key.
type remoteNode struct {
	summary libneoworks.NodeSummary
	// content is set once the node was read.
	content *entity
}

// tablePass reconciles one table.
type tablePass struct {
	librarySync *LibrarySync
	mapping     collectionMapping
	userID      string
	report      *ReconcileReport

	links      map[string]library.NodeLink
	local      map[string]entity
	remote     map[string]*remoteNode
	tombstones map[string]bool
	// written holds the node of each key this pass wrote, whose update time is read back afterwards.
	written map[string]string
}

func newTablePass(librarySync *LibrarySync, mapping collectionMapping, userID string, report *ReconcileReport) *tablePass {
	return &tablePass{
		librarySync: librarySync,
		mapping:     mapping,
		userID:      userID,
		report:      report,
		remote:      map[string]*remoteNode{},
		tombstones:  map[string]bool{},
		written:     map[string]string{},
	}
}

func (pass *tablePass) run(ctx context.Context) error {
	err := pass.load(ctx)
	if err != nil {
		return err
	}
	failure := pass.reconcileKeys(ctx)
	return errors.Join(failure, pass.refreshWrittenLinks(ctx))
}

// ownRows lists the rows of this table the signed-in user owns, deleted ones
// included. Rows shared by other people are left to their owners.
func (pass *tablePass) ownRows() ([]libneoworks.NodeSummary, error) {
	summaries, err := pass.librarySync.nodes.ListNodes("", collection)
	if err != nil {
		return nil, err
	}
	var rows []libneoworks.NodeSummary
	for _, summary := range summaries {
		if summary.Kind != libneoworks.NodeKindRow || summary.OwnerID != pass.userID {
			continue
		}
		if summary.Table != nil && *summary.Table == pass.mapping.table() {
			rows = append(rows, summary)
		}
	}
	return rows, nil
}

func (pass *tablePass) load(ctx context.Context) error {
	var err error
	pass.links, err = pass.librarySync.library.NodeLinks(ctx, pass.mapping.name())
	if err != nil {
		return err
	}
	entities, err := pass.mapping.local(ctx)
	if err != nil {
		return err
	}
	pass.local = map[string]entity{}
	for _, item := range entities {
		pass.local[item.key] = item
	}
	return pass.loadRemote()
}

// loadRemote groups the live rows by natural key and keeps the newest of each.
func (pass *tablePass) loadRemote() error {
	summaries, err := pass.ownRows()
	if err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}
	keyOfNode := map[string]string{}
	for key, link := range pass.links {
		keyOfNode[link.NodeID] = key
	}
	groups := map[string][]*remoteNode{}
	for _, summary := range summaries {
		pass.classify(summary, keyOfNode, groups)
	}
	for key, group := range groups {
		pass.remote[key] = pass.keepNewest(group)
	}
	return nil
}

// classify files a row under its natural key, or notes it as deleted. A row
// that is rejected or cannot be read is left alone.
func (pass *tablePass) classify(summary libneoworks.NodeSummary, keyOfNode map[string]string, groups map[string][]*remoteNode) {
	if summary.Rejected {
		return
	}
	if summary.Deleted {
		pass.tombstones[summary.ID] = true
		return
	}
	node := &remoteNode{summary: summary}
	key, linked := keyOfNode[summary.ID]
	if !linked {
		var err error
		key, err = pass.discoverKey(node)
		if err != nil {
			pass.librarySync.logger.Warn("neoworks: skipping node", "table", pass.mapping.table(), "node", summary.ID, "error", err)
			return
		}
	}
	groups[key] = append(groups[key], node)
}

func (pass *tablePass) discoverKey(node *remoteNode) (string, error) {
	item, err := pass.read(node)
	if err != nil {
		return "", err
	}
	return item.key, nil
}

// read loads and decodes the node's content once.
func (pass *tablePass) read(node *remoteNode) (entity, error) {
	if node.content != nil {
		return *node.content, nil
	}
	raw, err := pass.librarySync.nodes.ReadNode(node.summary.ID)
	if err != nil {
		return entity{}, fmt.Errorf("read node %s: %w", node.summary.ID, err)
	}
	item, err := pass.mapping.decode(node.summary.ID, raw)
	if err != nil {
		return entity{}, fmt.Errorf("node %s: %w", node.summary.ID, err)
	}
	node.content = &item
	return item, nil
}

// keepNewest returns the newest node of a group and deletes the others: two
// devices that created the same entity offline left duplicates.
func (pass *tablePass) keepNewest(group []*remoteNode) *remoteNode {
	newest := group[0]
	for _, candidate := range group[1:] {
		if newerNode(candidate.summary, newest.summary) {
			newest = candidate
		}
	}
	for _, candidate := range group {
		if candidate != newest {
			pass.deleteDuplicate(candidate)
		}
	}
	return newest
}

func newerNode(candidate, current libneoworks.NodeSummary) bool {
	candidateTime := parseNodeTime(candidate.UpdatedAt)
	currentTime := parseNodeTime(current.UpdatedAt)
	if candidateTime.Equal(currentTime) {
		return candidate.ID < current.ID
	}
	return candidateTime.After(currentTime)
}

func (pass *tablePass) deleteDuplicate(duplicate *remoteNode) {
	_, err := pass.librarySync.nodes.DeleteNode(duplicate.summary.ID)
	if err != nil {
		pass.librarySync.logger.Warn("neoworks: deleting duplicate node", "table", pass.mapping.table(), "node", duplicate.summary.ID, "error", err)
		return
	}
	pass.report.NodesWritten++
}

func (pass *tablePass) reconcileKeys(ctx context.Context) error {
	keys := map[string]bool{}
	for key := range pass.local {
		keys[key] = true
	}
	for key := range pass.remote {
		keys[key] = true
	}
	for key := range pass.links {
		keys[key] = true
	}
	var failures []error
	for _, key := range slices.Sorted(maps.Keys(keys)) {
		err := pass.reconcileKey(ctx, key)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", key, err))
		}
	}
	return errors.Join(failures...)
}

func (pass *tablePass) reconcileKey(ctx context.Context, key string) error {
	state, err := pass.situationOf(key)
	if err != nil {
		return err
	}
	return pass.execute(ctx, key, decide(state))
}

func (pass *tablePass) situationOf(key string) (situation, error) {
	state := situation{owned: pass.mapping.ownsKey(key), keepsNodes: pass.mapping.keepsNodes()}
	pass.describeLocal(key, &state)
	node, hasNode := pass.remote[key]
	link, linked := pass.links[key]
	state.hasNode = hasNode
	state.linked = linked && (!hasNode || link.NodeID == node.summary.ID)
	if linked {
		state.linkHash = link.ContentHash
		state.tombstone = pass.tombstones[link.NodeID]
	}
	if !hasNode {
		return state, nil
	}
	state.remoteUpdatedAt = parseNodeTime(node.summary.UpdatedAt)
	state.nodeMoved = state.linked && node.summary.UpdatedAt != link.NodeUpdatedAt
	state.remoteHash = state.linkHash
	if !state.linked || state.nodeMoved {
		return pass.describeRemote(node, state)
	}
	return state, nil
}

func (pass *tablePass) describeLocal(key string, state *situation) {
	item, found := pass.local[key]
	if !found {
		return
	}
	state.hasLocal = true
	state.localHash = item.hash()
	state.localUpdatedAt = item.updatedAt
}

func (pass *tablePass) describeRemote(node *remoteNode, state situation) (situation, error) {
	item, err := pass.read(node)
	if err != nil {
		return state, err
	}
	state.remoteHash = item.hash()
	if !item.updatedAt.IsZero() {
		state.remoteUpdatedAt = item.updatedAt
	}
	return state, nil
}
