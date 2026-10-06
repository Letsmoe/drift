-- Tracks and artists the user asked never to be recommended. Unlike a dislike
-- (track_rating -1) this is not a taste signal: it only filters the output.
CREATE TABLE recommendation_block (
    kind       TEXT NOT NULL CHECK (kind IN ('track', 'artist')),
    id         TEXT NOT NULL,
    name       TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    PRIMARY KEY (kind, id)
);
