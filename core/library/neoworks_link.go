package library

import (
	"context"
	"database/sql"
	"fmt"
)

// NodeLink ties a library entity to the Neoworks node that holds its synced copy.
type NodeLink struct {
	// Collection is the collection name without scope, e.g. "ratings".
	Collection string
	// Key is the entity's natural key in that collection, e.g. a video ID.
	Key    string
	NodeID string
	// ContentHash identifies the content last written to or read from the node.
	ContentHash string
	// NodeUpdatedAt is the node's update time as the store reported it then.
	NodeUpdatedAt string
}

// NodeLinks returns the links of one collection by natural key.
func (library *Library) NodeLinks(ctx context.Context, collection string) (map[string]NodeLink, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT natural_key, node_id, content_hash, node_updated_at
		FROM neoworks_link WHERE collection = ?`, collection)
	if err != nil {
		return nil, fmt.Errorf("library: query node links: %w", err)
	}
	defer func() { _ = rows.Close() }()
	links := map[string]NodeLink{}
	for rows.Next() {
		link := NodeLink{Collection: collection}
		err = rows.Scan(&link.Key, &link.NodeID, &link.ContentHash, &link.NodeUpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("library: scan node link: %w", err)
		}
		links[link.Key] = link
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("library: read node links: %w", err)
	}
	return links, nil
}

// SaveNodeLink stores a link, replacing the one for the same entity or the same node.
func (library *Library) SaveNodeLink(ctx context.Context, link NodeLink) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `DELETE FROM neoworks_link WHERE node_id = ?`, link.NodeID)
		if err != nil {
			return fmt.Errorf("replace node link of %s: %w", link.NodeID, err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO neoworks_link (collection, natural_key, node_id, content_hash, node_updated_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (collection, natural_key) DO UPDATE SET
				node_id = excluded.node_id, content_hash = excluded.content_hash, node_updated_at = excluded.node_updated_at`,
			link.Collection, link.Key, link.NodeID, link.ContentHash, link.NodeUpdatedAt)
		if err != nil {
			return fmt.Errorf("save node link %s %s: %w", link.Collection, link.Key, err)
		}
		return nil
	})
}

// DeleteNodeLink forgets one link; forgetting one that does not exist is not an error.
func (library *Library) DeleteNodeLink(ctx context.Context, collection, key string) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM neoworks_link WHERE collection = ? AND natural_key = ?`, collection, key)
	if err != nil {
		return fmt.Errorf("library: delete node link %s %s: %w", collection, key, err)
	}
	return nil
}

// ClearNodeLinks forgets every link, for when the synced copy is gone.
func (library *Library) ClearNodeLinks(ctx context.Context) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM neoworks_link`)
	if err != nil {
		return fmt.Errorf("library: clear node links: %w", err)
	}
	return nil
}
