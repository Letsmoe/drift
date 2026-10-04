package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/neoworks/otter-music/core/music"
)

// executor is what *sql.DB and *sql.Tx share.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type rowScanner interface {
	Scan(dest ...any) error
}

// trackColumns is the column list scanTrack expects, for the track table
// aliased as t.
const trackColumns = `t.video_id, t.title, t.artists, t.album_id, t.album_name, t.duration_seconds, t.thumbnails`

const upsertTrackSQL = `INSERT INTO track
	(video_id, kind, title, artists, album_id, album_name, duration_seconds, thumbnails, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT (video_id) DO UPDATE SET
		kind = CASE WHEN excluded.kind = 'episode' THEN 'episode' ELSE track.kind END,
		title = excluded.title,
		artists = excluded.artists,
		album_id = COALESCE(excluded.album_id, track.album_id),
		album_name = COALESCE(excluded.album_name, track.album_name),
		duration_seconds = excluded.duration_seconds,
		thumbnails = excluded.thumbnails,
		updated_at = excluded.updated_at`

// SaveTracks caches track metadata. A track stored as an episode stays one.
func (library *Library) SaveTracks(ctx context.Context, tracks []music.Track, kind music.ItemKind) error {
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		return library.upsertTracks(ctx, tx, tracks, kind)
	})
}

func (library *Library) upsertTracks(ctx context.Context, db executor, tracks []music.Track, kind music.ItemKind) error {
	for _, track := range tracks {
		err := library.upsertTrack(ctx, db, track, kind)
		if err != nil {
			return err
		}
	}
	return nil
}

func (library *Library) upsertTrack(ctx context.Context, db executor, track music.Track, kind music.ItemKind) error {
	if track.VideoID == "" {
		return nil
	}
	artists, err := marshalList(track.Artists)
	if err != nil {
		return err
	}
	thumbnails, err := marshalList(track.Thumbnails)
	if err != nil {
		return err
	}
	var albumID, albumName *string
	if track.Album != nil {
		albumID = &track.Album.ID
		albumName = &track.Album.Name
	}
	_, err = db.ExecContext(ctx, upsertTrackSQL, track.VideoID, storedKind(kind), track.Title, artists,
		albumID, albumName, track.DurationSeconds, thumbnails, library.timestamp())
	if err != nil {
		return fmt.Errorf("save track %s: %w", track.VideoID, err)
	}
	return nil
}

func storedKind(kind music.ItemKind) string {
	if kind == music.ItemKindEpisode {
		return string(music.ItemKindEpisode)
	}
	return string(music.ItemKindSong)
}

// marshalList encodes a slice as a JSON array, never "null".
func marshalList[T any](list []T) (string, error) {
	if list == nil {
		return "[]", nil
	}
	encoded, err := json.Marshal(list)
	if err != nil {
		return "", fmt.Errorf("encode json: %w", err)
	}
	return string(encoded), nil
}

// scanTrack reads trackColumns; leading receives any columns selected before them.
func scanTrack(row rowScanner, leading ...any) (music.Track, error) {
	var track music.Track
	var artists, thumbnails string
	var albumID, albumName sql.NullString
	destinations := slices.Concat(leading, []any{&track.VideoID, &track.Title, &artists, &albumID, &albumName, &track.DurationSeconds, &thumbnails})
	err := row.Scan(destinations...)
	if err != nil {
		return music.Track{}, fmt.Errorf("scan track: %w", err)
	}
	err = json.Unmarshal([]byte(artists), &track.Artists)
	if err != nil {
		return music.Track{}, fmt.Errorf("decode artists of %s: %w", track.VideoID, err)
	}
	err = json.Unmarshal([]byte(thumbnails), &track.Thumbnails)
	if err != nil {
		return music.Track{}, fmt.Errorf("decode thumbnails of %s: %w", track.VideoID, err)
	}
	if albumID.Valid {
		track.Album = &music.Album{ID: albumID.String, Name: albumName.String}
	}
	return track, nil
}

func scanTracks(rows *sql.Rows) ([]music.Track, error) {
	defer func() { _ = rows.Close() }()
	tracks := []music.Track{}
	for rows.Next() {
		track, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read tracks: %w", err)
	}
	return tracks, nil
}

// Tracks returns the cached metadata of the given videos; unknown IDs are absent.
func (library *Library) Tracks(ctx context.Context, videoIDs []string) (map[string]music.Track, error) {
	found := map[string]music.Track{}
	if len(videoIDs) == 0 {
		return found, nil
	}
	rows, err := library.db.QueryContext(ctx,
		`SELECT `+trackColumns+` FROM track t WHERE t.video_id IN (`+placeholders(len(videoIDs))+`)`,
		stringArguments(videoIDs)...)
	if err != nil {
		return nil, fmt.Errorf("library: query tracks: %w", err)
	}
	tracks, err := scanTracks(rows)
	if err != nil {
		return nil, fmt.Errorf("library: %w", err)
	}
	for _, track := range tracks {
		found[track.VideoID] = track
	}
	return found, nil
}

func placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func stringArguments(values []string) []any {
	arguments := make([]any, len(values))
	for index, value := range values {
		arguments[index] = value
	}
	return arguments
}
