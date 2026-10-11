-- Maps each synced library entity to its Neoworks node. content_hash is the
-- hash of what was last written to or read from the node, and node_updated_at
-- is the node's update time at that moment; a difference from the library or
-- from the node says which side changed since.
CREATE TABLE neoworks_link (
    collection      TEXT NOT NULL,
    natural_key     TEXT NOT NULL,
    node_id         TEXT NOT NULL,
    content_hash    TEXT NOT NULL,
    node_updated_at TEXT NOT NULL,
    PRIMARY KEY (collection, natural_key)
);

CREATE UNIQUE INDEX neoworks_link_node ON neoworks_link (node_id);
