package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// CacheGet decodes the cached value for key into into. Expired or missing
// entries report found as false.
func (library *Library) CacheGet(ctx context.Context, key string, into any) (bool, error) {
	var value string
	err := library.db.QueryRowContext(ctx, `SELECT value FROM cache WHERE key = ? AND expires_at > ?`, key, library.timestamp()).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("library: read cache %q: %w", key, err)
	}
	err = json.Unmarshal([]byte(value), into)
	if err != nil {
		return false, fmt.Errorf("library: decode cache %q: %w", key, err)
	}
	return true, nil
}

// CachePut stores value under key for ttl. Expired rows are purged on Open.
func (library *Library) CachePut(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("library: encode cache %q: %w", key, err)
	}
	_, err = library.db.ExecContext(ctx, `INSERT INTO cache (key, value, expires_at) VALUES (?, ?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value, expires_at = excluded.expires_at`,
		key, string(encoded), FormatTime(library.now().Add(ttl)))
	if err != nil {
		return fmt.Errorf("library: write cache %q: %w", key, err)
	}
	return nil
}
