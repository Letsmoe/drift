-- "Don't recommend" can also name a podcast show (its MPSP... browse ID).
CREATE TABLE recommendation_block_new (
    kind       TEXT NOT NULL CHECK (kind IN ('track', 'artist', 'podcast')),
    id         TEXT NOT NULL,
    name       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    PRIMARY KEY (kind, id)
);

INSERT INTO recommendation_block_new (kind, id, name, created_at)
    SELECT kind, id, name, created_at FROM recommendation_block;

DROP TABLE recommendation_block;

ALTER TABLE recommendation_block_new RENAME TO recommendation_block;
