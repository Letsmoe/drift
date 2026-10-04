-- Offline downloads: one row per downloaded (or queued) video, plus the
-- collections (album, playlist, local playlist, liked songs) the user asked to
-- have downloaded. A file is kept while it is explicit or any collection lists it.
CREATE TABLE download (
    video_id      TEXT PRIMARY KEY,
    track         TEXT NOT NULL,
    kind          TEXT NOT NULL DEFAULT 'song',
    state         TEXT NOT NULL CHECK (state IN ('queued', 'downloading', 'done', 'failed')),
    explicit      INTEGER NOT NULL DEFAULT 0,
    file_path     TEXT NOT NULL DEFAULT '',
    cover_path    TEXT NOT NULL DEFAULT '',
    format        TEXT NOT NULL DEFAULT '',
    size_bytes    INTEGER NOT NULL DEFAULT 0,
    error         TEXT NOT NULL DEFAULT '',
    queued_at     TEXT NOT NULL,
    downloaded_at TEXT NOT NULL DEFAULT ''
);

CREATE INDEX download_state_queued ON download (state, queued_at);

CREATE TABLE download_collection (
    kind       TEXT NOT NULL,
    id         TEXT NOT NULL,
    title      TEXT NOT NULL DEFAULT '',
    keep       INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    synced_at  TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (kind, id)
);

CREATE TABLE download_collection_track (
    collection_kind TEXT NOT NULL,
    collection_id   TEXT NOT NULL,
    video_id        TEXT NOT NULL REFERENCES download (video_id) ON DELETE CASCADE,
    position        INTEGER NOT NULL,
    PRIMARY KEY (collection_kind, collection_id, video_id),
    FOREIGN KEY (collection_kind, collection_id) REFERENCES download_collection (kind, id) ON DELETE CASCADE
);

CREATE INDEX download_collection_track_video ON download_collection_track (video_id);
