package browsercookies

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// millisecondThreshold separates expiry values in seconds from ones in
// milliseconds; newer Firefox versions store some in milliseconds.
const millisecondThreshold = 100_000_000_000

// readFirefoxCookies reads the cookies of the YouTube hosts from a
// Firefox-family cookies.sqlite, leaving out those of container tabs.
func readFirefoxCookies(tempParent, databasePath string) ([]Cookie, error) {
	db, cleanup, err := openDatabaseCopy(tempParent, databasePath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	rows, err := db.QueryContext(context.Background(), `SELECT host, name, value, path, expiry FROM moz_cookies
		WHERE host IN (?, ?, ?) AND originAttributes = ''`, musicHosts[0], musicHosts[1], musicHosts[2])
	if err != nil {
		return nil, fmt.Errorf("query Firefox cookies: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanFirefoxRows(rows)
}

func scanFirefoxRows(rows *sql.Rows) ([]Cookie, error) {
	cookies := []Cookie{}
	for rows.Next() {
		var cookie Cookie
		var expiry int64
		err := rows.Scan(&cookie.Host, &cookie.Name, &cookie.Value, &cookie.Path, &expiry)
		if err != nil {
			return nil, fmt.Errorf("read Firefox cookie: %w", err)
		}
		cookie.Expires = firefoxExpiry(expiry)
		cookies = append(cookies, cookie)
	}
	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read Firefox cookies: %w", err)
	}
	return cookies, nil
}

func firefoxExpiry(expiry int64) time.Time {
	if expiry <= 0 {
		return time.Time{}
	}
	if expiry > millisecondThreshold {
		return time.UnixMilli(expiry)
	}
	return time.Unix(expiry, 0)
}
