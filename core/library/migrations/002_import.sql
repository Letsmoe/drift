-- Matches of external items (e.g. Spotify tracks) to YouTube Music, so a
-- re-import does not search again. An empty target_id records "no match".
CREATE TABLE import_match (
    source     TEXT NOT NULL,
    source_uri TEXT NOT NULL,
    kind       TEXT NOT NULL,
    target_id  TEXT NOT NULL DEFAULT '',
    score      REAL NOT NULL DEFAULT 0,
    payload    TEXT NOT NULL DEFAULT '',
    matched_at TEXT NOT NULL,
    PRIMARY KEY (source, source_uri)
);

-- The local playlist an external playlist was imported into, so a re-import
-- replaces its entries instead of creating a duplicate.
CREATE TABLE import_playlist (
    source      TEXT NOT NULL,
    source_key  TEXT NOT NULL,
    playlist_id TEXT NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    PRIMARY KEY (source, source_key)
);
