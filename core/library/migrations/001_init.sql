CREATE TABLE track (
    video_id         TEXT PRIMARY KEY,
    kind             TEXT NOT NULL DEFAULT 'song',
    title            TEXT NOT NULL,
    artists          TEXT NOT NULL DEFAULT '[]',
    album_id         TEXT,
    album_name       TEXT,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    thumbnails       TEXT NOT NULL DEFAULT '[]',
    updated_at       TEXT NOT NULL
);

CREATE TABLE track_rating (
    video_id   TEXT PRIMARY KEY REFERENCES track (video_id) ON DELETE CASCADE,
    rating     INTEGER NOT NULL CHECK (rating IN (1, -1)),
    updated_at TEXT NOT NULL
);

CREATE INDEX track_rating_updated_at ON track_rating (rating, updated_at);

CREATE TABLE playlist (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE playlist_entry (
    id          INTEGER PRIMARY KEY,
    playlist_id TEXT NOT NULL REFERENCES playlist (id) ON DELETE CASCADE,
    video_id    TEXT NOT NULL REFERENCES track (video_id),
    position    INTEGER NOT NULL,
    added_at    TEXT NOT NULL
);

CREATE INDEX playlist_entry_position ON playlist_entry (playlist_id, position);

CREATE TABLE saved_item (
    kind       TEXT NOT NULL,
    id         TEXT NOT NULL,
    title      TEXT NOT NULL,
    subtitle   TEXT NOT NULL DEFAULT '',
    thumbnails TEXT NOT NULL DEFAULT '[]',
    saved_at   TEXT NOT NULL,
    PRIMARY KEY (kind, id)
);

CREATE TABLE play (
    id               INTEGER PRIMARY KEY,
    video_id         TEXT NOT NULL REFERENCES track (video_id),
    context          TEXT NOT NULL DEFAULT '',
    started_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    progress_seconds INTEGER NOT NULL DEFAULT 0,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    completed        INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX play_video_started ON play (video_id, started_at);
CREATE INDEX play_started ON play (started_at);

CREATE TABLE search_history (
    id               INTEGER PRIMARY KEY,
    query            TEXT NOT NULL,
    normalized_query TEXT NOT NULL,
    created_at       TEXT NOT NULL
);

CREATE INDEX search_history_created ON search_history (created_at);
CREATE INDEX search_history_normalized ON search_history (normalized_query);

CREATE TABLE impression (
    id       INTEGER PRIMARY KEY,
    video_id TEXT NOT NULL,
    source   TEXT NOT NULL DEFAULT '',
    surface  TEXT NOT NULL DEFAULT '',
    position INTEGER,
    clicked  INTEGER NOT NULL DEFAULT 0,
    shown_at TEXT NOT NULL
);

CREATE INDEX impression_video_shown ON impression (video_id, shown_at);
CREATE INDEX impression_shown ON impression (shown_at);

CREATE TABLE cache (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    expires_at TEXT NOT NULL
);

CREATE INDEX cache_expires ON cache (expires_at);
