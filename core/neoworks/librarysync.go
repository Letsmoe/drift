package neoworks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	libneoworks "github.com/neoworks-dev/libneoworks-go"

	"github.com/letsmoe/drift/core/library"
)

// Nodes is the part of the libneoworks client the library mapping uses.
// *libneoworks.Client implements it.
type Nodes interface {
	CreateNode(node libneoworks.NewNode) (string, error)
	UpdateNode(update libneoworks.NodeUpdate) error
	DeleteNode(nodeID string, recursive bool) error
	ReadNode(nodeID string) (json.RawMessage, error)
	ListNodes(parentID, collection string) ([]libneoworks.NodeSummary, error)
}

// LibrarySyncConfig holds what LibrarySync needs.
type LibrarySyncConfig struct {
	Library *library.Library
	Nodes   Nodes
	// UserID returns the signed-in user's ID, empty when signed out.
	UserID func() string
	Logger *slog.Logger
}

// LibrarySync maps the library to the nodes of the @drift collections and
// back. It reads and writes the local store only; Session.Sync moves the
// nodes to and from the server.
type LibrarySync struct {
	library  *library.Library
	nodes    Nodes
	userID   func() string
	logger   *slog.Logger
	mappings []collectionMapping
}

// NewLibrarySync builds the mapping for every @drift collection.
func NewLibrarySync(config LibrarySyncConfig) *LibrarySync {
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &LibrarySync{
		library: config.Library,
		nodes:   config.Nodes,
		userID:  config.UserID,
		logger:  logger,
		mappings: []collectionMapping{
			trackMapping{library: config.Library},
			ratingMapping{library: config.Library},
			savedMapping{library: config.Library},
			playlistMapping{library: config.Library},
		},
	}
}

// ReconcileReport says what one reconcile did.
type ReconcileReport struct {
	// NodesWritten counts nodes created, updated or deleted; they still have to be synced.
	NodesWritten int
	// LibraryWritten counts library entities created, changed or removed from nodes.
	LibraryWritten int
	// WithoutRoot lists the collections skipped because the account holds no root of them yet.
	WithoutRoot []string
}

// Forget drops the links to nodes, for when the synced copy is wiped.
func (librarySync *LibrarySync) Forget(ctx context.Context) error {
	return librarySync.library.ClearNodeLinks(ctx)
}

// Reconcile applies what other devices wrote into the library and writes local
// changes as nodes, collection by collection, tracks first. Descriptors must
// be registered and the nodes pulled before; a collection that fails does not
// stop the others.
func (librarySync *LibrarySync) Reconcile(ctx context.Context) (ReconcileReport, error) {
	var report ReconcileReport
	userID := librarySync.userID()
	if userID == "" {
		return report, ErrSignedOut
	}
	quietCtx := library.WithoutNotifications(ctx)
	var failures []error
	for _, mapping := range librarySync.mappings {
		pass := newCollectionPass(librarySync, mapping, userID, &report)
		err := pass.run(quietCtx)
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", collectionPath(mapping.name()), err))
		}
	}
	return report, errors.Join(failures...)
}

// entity is one library entity as node content.
type entity struct {
	key string
	// canonical is the typed content as JSON, so equal content is equal bytes.
	canonical []byte
	updatedAt time.Time
	// model is the mapping's typed content.
	model any
}

func (item entity) hash() string {
	sum := sha256.Sum256(item.canonical)
	return hex.EncodeToString(sum[:])
}

// content is the canonical JSON as the map libneoworks takes.
func (item entity) content() (map[string]any, error) {
	var content map[string]any
	err := json.Unmarshal(item.canonical, &content)
	if err != nil {
		return nil, fmt.Errorf("decode content of %s: %w", item.key, err)
	}
	return content, nil
}

// collectionMapping converts one collection between the library and node content.
type collectionMapping interface {
	// name is the collection name without scope.
	name() string
	// local lists the library's entities of this collection.
	local(ctx context.Context) ([]entity, error)
	// decode reads node content. The entity's key is the one the mapping derives for a node it has not met.
	decode(nodeID string, raw []byte) (entity, error)
	// apply writes an entity that came from a node into the library.
	apply(ctx context.Context, remote entity) error
	// remove deletes the library entity of a key whose node was deleted elsewhere.
	remove(ctx context.Context, key string) error
	// optionalFields are the content fields an update must clear when an entity no longer has them.
	optionalFields() []string
	// keepsNodes says the nodes outlive their library entities: a node is never deleted for a
	// missing entity, and a deleted node comes back while the entity exists.
	keepsNodes() bool
	// ownsKey says only this install writes the entity: node content is never applied.
	ownsKey(key string) bool
}

func canonicalJSON(model any) ([]byte, error) {
	encoded, err := json.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("encode content: %w", err)
	}
	return encoded, nil
}

// decodeModel parses raw node content into a typed model; fields the model lacks are dropped.
func decodeModel[Model any](raw []byte) (Model, error) {
	var model Model
	err := json.Unmarshal(raw, &model)
	if err != nil {
		return model, fmt.Errorf("decode content: %w", err)
	}
	return model, nil
}

// parseNodeTime reads the time of a node summary or a timestamp field; zero when unreadable.
func parseNodeTime(text string) time.Time {
	moment, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}
	}
	return moment
}

// normalizeTime renders a timestamp field the way the library writes it, so
// content from a node and from the library compare equal.
func normalizeTime(text string) (string, error) {
	moment, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return "", fmt.Errorf("timestamp %q: %w", text, err)
	}
	return library.FormatTime(moment), nil
}
