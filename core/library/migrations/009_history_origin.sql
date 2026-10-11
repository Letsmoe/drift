-- Plays and searches that came from another install's synced history carry that
-- install's ID; the ones made here keep the empty origin and are the only ones
-- written to this install's day logs.
ALTER TABLE play ADD COLUMN origin TEXT NOT NULL DEFAULT '';

ALTER TABLE search_history ADD COLUMN origin TEXT NOT NULL DEFAULT '';

CREATE INDEX play_origin_started ON play (origin, started_at);

CREATE INDEX search_history_origin_created ON search_history (origin, created_at);
