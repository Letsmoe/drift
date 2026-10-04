-- The signed-in YouTube Music session used only to read the user's library
-- for import and sync. At most one row. cookie_header is a credential.
CREATE TABLE youtube_session (
    id               INTEGER PRIMARY KEY CHECK (id = 1),
    cookie_header    TEXT NOT NULL,
    account_name     TEXT NOT NULL DEFAULT '',
    account_handle   TEXT NOT NULL DEFAULT '',
    account_email    TEXT NOT NULL DEFAULT '',
    account_photo    TEXT NOT NULL DEFAULT '',
    created_at       TEXT NOT NULL,
    validated_at     TEXT NOT NULL,
    last_synced_at   TEXT NOT NULL DEFAULT ''
);
