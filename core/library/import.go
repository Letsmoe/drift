package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/letsmoe/drift/core/music"
)

// ImportMatch is the remembered result of matching one external item to
// YouTube Music. An empty TargetID means the item was searched and not found.
type ImportMatch struct {
	Source    string
	SourceURI string
	Kind      music.ItemKind
	TargetID  string
	Score     float64
	// Payload is the matched music.Track or music.Item as JSON, so applying a
	// stored match needs no network.
	Payload   string
	MatchedAt time.Time
}

// ImportMatch returns the stored match of an external item; found is false
// when the item was never matched.
func (library *Library) ImportMatch(ctx context.Context, source, sourceURI string) (match ImportMatch, found bool, err error) {
	var kind, matchedAt string
	err = library.db.QueryRowContext(ctx, `SELECT kind, target_id, score, payload, matched_at FROM import_match
		WHERE source = ? AND source_uri = ?`, source, sourceURI).
		Scan(&kind, &match.TargetID, &match.Score, &match.Payload, &matchedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ImportMatch{}, false, nil
	}
	if err != nil {
		return ImportMatch{}, false, fmt.Errorf("library: load import match %s: %w", sourceURI, err)
	}
	match.Source = source
	match.SourceURI = sourceURI
	match.Kind = music.ItemKind(kind)
	match.MatchedAt, err = ParseTime(matchedAt)
	if err != nil {
		return ImportMatch{}, false, err
	}
	return match, true, nil
}

// StoreImportMatch remembers a match, replacing an earlier one for the same item.
func (library *Library) StoreImportMatch(ctx context.Context, match ImportMatch) error {
	_, err := library.db.ExecContext(ctx, `INSERT INTO import_match (source, source_uri, kind, target_id, score, payload, matched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (source, source_uri) DO UPDATE SET
			kind = excluded.kind, target_id = excluded.target_id, score = excluded.score,
			payload = excluded.payload, matched_at = excluded.matched_at`,
		match.Source, match.SourceURI, string(match.Kind), match.TargetID, match.Score, match.Payload, library.timestamp())
	if err != nil {
		return fmt.Errorf("library: store import match %s: %w", match.SourceURI, err)
	}
	return nil
}

// ImportedPlaylist describes an external playlist to write into the library.
type ImportedPlaylist struct {
	Source      string
	SourceKey   string
	Name        string
	Description string
	Tracks      []music.Track
}

// ImportPlaylist creates the local playlist for an external one, or, when an
// earlier import made one that still exists, renames it and replaces its
// entries. It returns the local playlist ID.
func (library *Library) ImportPlaylist(ctx context.Context, imported ImportedPlaylist) (string, error) {
	var playlistID string
	err := library.inTransaction(ctx, func(tx *sql.Tx) error {
		existingID, err := linkedPlaylistID(ctx, tx, imported.Source, imported.SourceKey)
		if err != nil {
			return err
		}
		playlistID, err = library.writeImportedPlaylist(ctx, tx, existingID, imported)
		return err
	})
	if err != nil {
		return "", err
	}
	library.emitChange(Change{Kind: ChangePlaylist, ID: playlistID})
	return playlistID, nil
}

func linkedPlaylistID(ctx context.Context, tx *sql.Tx, source, sourceKey string) (string, error) {
	var playlistID string
	err := tx.QueryRowContext(ctx, `SELECT playlist_id FROM import_playlist WHERE source = ? AND source_key = ?`, source, sourceKey).
		Scan(&playlistID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find imported playlist %s: %w", sourceKey, err)
	}
	return playlistID, nil
}

func (library *Library) writeImportedPlaylist(ctx context.Context, tx *sql.Tx, existingID string, imported ImportedPlaylist) (string, error) {
	if existingID != "" {
		updated, err := library.replaceImportedPlaylist(ctx, tx, existingID, imported)
		if err != nil || updated {
			return existingID, err
		}
	}
	return library.createImportedPlaylist(ctx, tx, imported)
}

// replaceImportedPlaylist reports false when the linked playlist is gone.
func (library *Library) replaceImportedPlaylist(ctx context.Context, tx *sql.Tx, id string, imported ImportedPlaylist) (bool, error) {
	result, err := tx.ExecContext(ctx, `UPDATE playlist SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		imported.Name, imported.Description, library.timestamp(), id)
	if err != nil {
		return false, fmt.Errorf("update imported playlist %s: %w", id, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return false, nil
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM playlist_entry WHERE playlist_id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("clear imported playlist %s: %w", id, err)
	}
	return true, library.appendEntries(ctx, tx, id, imported.Tracks)
}

func (library *Library) createImportedPlaylist(ctx context.Context, tx *sql.Tx, imported ImportedPlaylist) (string, error) {
	id, err := newPlaylistID()
	if err != nil {
		return "", err
	}
	now := library.timestamp()
	_, err = tx.ExecContext(ctx, `INSERT INTO playlist (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, imported.Name, imported.Description, now, now)
	if err != nil {
		return "", fmt.Errorf("create imported playlist: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO import_playlist (source, source_key, playlist_id) VALUES (?, ?, ?)
		ON CONFLICT (source, source_key) DO UPDATE SET playlist_id = excluded.playlist_id`,
		imported.Source, imported.SourceKey, id)
	if err != nil {
		return "", fmt.Errorf("link imported playlist %s: %w", imported.SourceKey, err)
	}
	return id, library.appendEntries(ctx, tx, id, imported.Tracks)
}
