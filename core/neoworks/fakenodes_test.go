package neoworks

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	libneoworks "github.com/neoworks-dev/libneoworks-go"

	"github.com/letsmoe/drift/core/library"
)

const testUserID = "user-1"

// fakeNodeStore stands in for the decrypted node store that every install of
// one account shares once synced. Granting real node keys needs a signed
// grant the Go binding cannot produce, so the mapping is tested against this.
type fakeNodeStore struct {
	mutex  sync.Mutex
	clock  time.Time
	nextID int
	nodes  map[string]*fakeNode
}

type fakeNode struct {
	summary libneoworks.NodeSummary
	content map[string]any
}

func newFakeNodeStore() *fakeNodeStore {
	store := &fakeNodeStore{clock: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), nodes: map[string]*fakeNode{}}
	for _, name := range Collections {
		store.add(libneoworks.NodeSummary{OwnerID: testUserID, Collection: collectionPath(name), Kind: libneoworks.NodeKindRoot}, nil)
	}
	return store
}

func (store *fakeNodeStore) add(summary libneoworks.NodeSummary, content map[string]any) string {
	store.nextID++
	summary.ID = fmt.Sprintf("node-%03d", store.nextID)
	summary.UpdatedAt = store.tick()
	store.nodes[summary.ID] = &fakeNode{summary: summary, content: content}
	return summary.ID
}

func (store *fakeNodeStore) tick() string {
	store.clock = store.clock.Add(time.Second)
	return library.FormatTime(store.clock)
}

// liveItems returns the contents of the live items of a collection.
func (store *fakeNodeStore) liveItems(name string) []map[string]any {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	var items []map[string]any
	for _, node := range store.nodes {
		live := node.summary.Kind == libneoworks.NodeKindItem && !node.summary.Deleted
		if live && node.summary.Collection == collectionPath(name) {
			items = append(items, node.content)
		}
	}
	return items
}

// fakeNodes is one install's view of the store; hidden nodes are ones that
// install has not synced yet.
type fakeNodes struct {
	store  *fakeNodeStore
	hidden map[string]bool
}

func newFakeNodes(store *fakeNodeStore) *fakeNodes {
	return &fakeNodes{store: store, hidden: map[string]bool{}}
}

// goOffline hides the nodes that exist now until goOnline.
func (view *fakeNodes) goOffline() {
	view.store.mutex.Lock()
	defer view.store.mutex.Unlock()
	for nodeID, node := range view.store.nodes {
		view.hidden[nodeID] = node.summary.Kind == libneoworks.NodeKindItem
	}
}

func (view *fakeNodes) goOnline() {
	view.hidden = map[string]bool{}
}

func (view *fakeNodes) CreateNode(node libneoworks.NewNode) (string, error) {
	view.store.mutex.Lock()
	defer view.store.mutex.Unlock()
	content, err := roundTrip(node.Content)
	if err != nil {
		return "", err
	}
	summary := libneoworks.NodeSummary{OwnerID: testUserID, Collection: node.Collection, Kind: node.Kind, ParentID: &node.ParentID}
	return view.store.add(summary, content), nil
}

func (view *fakeNodes) UpdateNode(update libneoworks.NodeUpdate) error {
	view.store.mutex.Lock()
	defer view.store.mutex.Unlock()
	node, found := view.store.nodes[update.NodeID]
	if !found {
		return libneoworks.StatusNotFound
	}
	changes, err := roundTrip(update.Changes)
	if err != nil {
		return err
	}
	for field, value := range changes {
		if value == nil {
			delete(node.content, field)
			continue
		}
		node.content[field] = value
	}
	node.summary.UpdatedAt = view.store.tick()
	return nil
}

func (view *fakeNodes) DeleteNode(nodeID string, _ bool) error {
	view.store.mutex.Lock()
	defer view.store.mutex.Unlock()
	node, found := view.store.nodes[nodeID]
	if !found {
		return libneoworks.StatusNotFound
	}
	node.summary.Deleted = true
	node.summary.UpdatedAt = view.store.tick()
	return nil
}

func (view *fakeNodes) ReadNode(nodeID string) (json.RawMessage, error) {
	view.store.mutex.Lock()
	defer view.store.mutex.Unlock()
	node, found := view.store.nodes[nodeID]
	if !found || node.summary.Deleted {
		return nil, libneoworks.StatusNotFound
	}
	encoded, err := json.Marshal(node.content)
	if err != nil {
		return nil, fmt.Errorf("fake node: %w", err)
	}
	return encoded, nil
}

func (view *fakeNodes) ListNodes(parentID, collection string) ([]libneoworks.NodeSummary, error) {
	view.store.mutex.Lock()
	defer view.store.mutex.Unlock()
	var summaries []libneoworks.NodeSummary
	for nodeID, node := range view.store.nodes {
		if view.hidden[nodeID] || node.summary.Collection != collection {
			continue
		}
		if parentID != "" && (node.summary.ParentID == nil || *node.summary.ParentID != parentID) {
			continue
		}
		summaries = append(summaries, node.summary)
	}
	slices.SortFunc(summaries, func(first, second libneoworks.NodeSummary) int {
		if first.ID < second.ID {
			return -1
		}
		return 1
	})
	return summaries, nil
}

func roundTrip(content map[string]any) (map[string]any, error) {
	encoded, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("fake content: %w", err)
	}
	var decoded map[string]any
	err = json.Unmarshal(encoded, &decoded)
	if err != nil {
		return nil, fmt.Errorf("fake content: %w", err)
	}
	return decoded, nil
}

// testInstall is one install of the account: its library and its mapping.
type testInstall struct {
	library *library.Library
	nodes   *fakeNodes
	sync    *LibrarySync
	clock   *time.Time
}

func newTestInstall(t *testing.T, store *fakeNodeStore) *testInstall {
	t.Helper()
	install := &testInstall{nodes: newFakeNodes(store), clock: new(time.Time)}
	*install.clock = time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)
	var err error
	install.library, err = library.Open(context.Background(), filepath.Join(t.TempDir(), "library.db"),
		library.WithClock(func() time.Time { return *install.clock }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = install.library.Close() })
	install.sync = NewLibrarySync(LibrarySyncConfig{
		Library: install.library,
		Nodes:   install.nodes,
		UserID:  func() string { return testUserID },
	})
	return install
}

func (install *testInstall) advance(duration time.Duration) {
	*install.clock = install.clock.Add(duration)
}

func (install *testInstall) reconcile(t *testing.T) ReconcileReport {
	t.Helper()
	report, err := install.sync.Reconcile(context.Background())
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	return report
}
