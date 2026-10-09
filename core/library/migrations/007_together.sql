-- Listen Together: this user's identity. user_id is derived from the public half
-- of signing_key (an Ed25519 seed, hex), so nobody else can claim it. One row.
CREATE TABLE together_identity (
    id             INTEGER PRIMARY KEY CHECK (id = 1),
    user_id        TEXT NOT NULL,
    signing_key    TEXT NOT NULL,
    name           TEXT NOT NULL,
    -- Whether friends see what this user is playing.
    share_presence INTEGER NOT NULL DEFAULT 1,
    created_at     TEXT NOT NULL
);

-- People the user listens with, by their fixed ID. Trusted friends join the
-- user's sessions without being asked for.
CREATE TABLE together_friend (
    user_id  TEXT PRIMARY KEY,
    name     TEXT NOT NULL DEFAULT '',
    trusted  INTEGER NOT NULL DEFAULT 0,
    added_at TEXT NOT NULL
);
