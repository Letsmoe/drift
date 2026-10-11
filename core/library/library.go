// Package library is the local SQLite store: the user's ratings, playlists,
// saved items, play history, search history, recommendation impressions and a
// network-response cache. YouTube Music is never asked to hold any of it.
package library

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite" // pure Go driver, registers "sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const (
	impressionRetention = 90 * 24 * time.Hour
	busyTimeoutMillis   = 5000
)

// ErrNotFound is returned (wrapped) when a playlist or playlist entry does not exist.
var ErrNotFound = errors.New("library: not found")

// Library is the open database. It is safe for concurrent use.
type Library struct {
	db  *sql.DB
	now func() time.Time

	listenerMutex   sync.RWMutex
	changeListeners []func(Change)
}

// Option configures Open.
type Option func(*Library)

// WithClock replaces the wall clock, for tests.
func WithClock(now func() time.Time) Option {
	return func(library *Library) {
		library.now = now
	}
}

// Open opens the database at path, creating it and its parent directory when
// missing, applies pending migrations and purges expired cache rows and old
// impressions.
func Open(ctx context.Context, path string, options ...Option) (*Library, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return nil, fmt.Errorf("library: create directory: %w", err)
	}
	db, err := sql.Open("sqlite", dataSourceName(path))
	if err != nil {
		return nil, fmt.Errorf("library: open %s: %w", path, err)
	}
	library := &Library{db: db, now: time.Now}
	for _, option := range options {
		option(library)
	}
	err = library.initialize(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return library, nil
}

func dataSourceName(path string) string {
	query := url.Values{}
	query.Set("_txlock", "immediate")
	query.Add("_pragma", "journal_mode(WAL)")
	query.Add("_pragma", "foreign_keys(1)")
	query.Add("_pragma", "busy_timeout("+strconv.Itoa(busyTimeoutMillis)+")")
	source := url.URL{Scheme: "file", Path: path, RawQuery: query.Encode()}
	return source.String()
}

func (library *Library) initialize(ctx context.Context) error {
	err := library.migrate(ctx)
	if err != nil {
		return err
	}
	return library.purgeExpired(ctx)
}

// Close closes the database.
func (library *Library) Close() error {
	err := library.db.Close()
	if err != nil {
		return fmt.Errorf("library: close: %w", err)
	}
	return nil
}

// DB exposes the database for read queries by other core packages.
func (library *Library) DB() *sql.DB {
	return library.db
}

type migration struct {
	version int
	name    string
	sql     string
}

func (library *Library) migrate(ctx context.Context) error {
	_, err := library.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS _migration (
		version    INTEGER PRIMARY KEY,
		name       TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("library: create migration table: %w", err)
	}
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	for _, pending := range migrations {
		err = library.applyMigration(ctx, pending)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("library: read migrations: %w", err)
	}
	migrations := make([]migration, 0, len(files))
	for _, file := range files {
		loaded, err := loadMigration(file.Name())
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, loaded)
	}
	sort.Slice(migrations, func(left, right int) bool {
		return migrations[left].version < migrations[right].version
	})
	return migrations, nil
}

func loadMigration(fileName string) (migration, error) {
	versionText, name, _ := strings.Cut(strings.TrimSuffix(fileName, ".sql"), "_")
	version, err := strconv.Atoi(versionText)
	if err != nil {
		return migration{}, fmt.Errorf("library: migration %q has no numeric version: %w", fileName, err)
	}
	contents, err := migrationFiles.ReadFile("migrations/" + fileName)
	if err != nil {
		return migration{}, fmt.Errorf("library: read migration %q: %w", fileName, err)
	}
	return migration{version: version, name: name, sql: string(contents)}, nil
}

func (library *Library) applyMigration(ctx context.Context, pending migration) error {
	var applied int
	err := library.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM _migration WHERE version = ?`, pending.version).Scan(&applied)
	if err != nil {
		return fmt.Errorf("library: check migration %d: %w", pending.version, err)
	}
	if applied > 0 {
		return nil
	}
	return library.inTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, pending.sql)
		if err != nil {
			return fmt.Errorf("apply migration %d %s: %w", pending.version, pending.name, err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO _migration (version, name, applied_at) VALUES (?, ?, ?)`,
			pending.version, pending.name, library.timestamp())
		if err != nil {
			return fmt.Errorf("record migration %d: %w", pending.version, err)
		}
		return nil
	})
}

func (library *Library) purgeExpired(ctx context.Context) error {
	now := library.now()
	_, err := library.db.ExecContext(ctx, `DELETE FROM cache WHERE expires_at <= ?`, FormatTime(now))
	if err != nil {
		return fmt.Errorf("library: purge cache: %w", err)
	}
	_, err = library.db.ExecContext(ctx, `DELETE FROM impression WHERE shown_at < ?`, FormatTime(now.Add(-impressionRetention)))
	if err != nil {
		return fmt.Errorf("library: purge impressions: %w", err)
	}
	return nil
}

// inTransaction runs change in a write transaction and commits it, or rolls
// back when change fails. Errors come back wrapped with the library prefix.
func (library *Library) inTransaction(ctx context.Context, change func(tx *sql.Tx) error) error {
	tx, err := library.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("library: begin transaction: %w", err)
	}
	err = change(tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("library: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("library: commit: %w", err)
	}
	return nil
}

func (library *Library) timestamp() string {
	return FormatTime(library.now())
}

// timeLayout is fixed width and UTC, so text order equals time order.
const timeLayout = "2006-01-02T15:04:05.000Z"

// FormatTime renders a time the way every timestamp column stores it.
func FormatTime(moment time.Time) string {
	return moment.UTC().Format(timeLayout)
}

// ParseTime reads a timestamp column written by FormatTime.
func ParseTime(text string) (time.Time, error) {
	moment, err := time.Parse(timeLayout, text)
	if err != nil {
		return time.Time{}, fmt.Errorf("library: parse time %q: %w", text, err)
	}
	return moment, nil
}
