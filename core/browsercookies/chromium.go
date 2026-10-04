package browsercookies

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// webkitEpochOffsetMicros is the distance between Chromium's cookie epoch
// (1601-01-01) and the Unix epoch, in microseconds.
const webkitEpochOffsetMicros = 11_644_473_600 * 1_000_000

// valueDecryptor turns an encrypted Chromium cookie value into plaintext.
type valueDecryptor interface {
	decrypt(hostKey string, encrypted []byte) (string, error)
}

// readChromiumCookies reads the cookies of the YouTube hosts from a Chromium
// "Cookies" database, decrypting values with decryptor.
func readChromiumCookies(tempParent, databasePath string, decryptor valueDecryptor) ([]Cookie, error) {
	db, cleanup, err := openDatabaseCopy(tempParent, databasePath)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	rows, err := db.QueryContext(context.Background(), `SELECT host_key, name, value, encrypted_value, path, expires_utc FROM cookies
		WHERE host_key IN (?, ?, ?)`, musicHosts[0], musicHosts[1], musicHosts[2])
	if err != nil {
		return nil, fmt.Errorf("query Chromium cookies: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanChromiumRows(rows, decryptor)
}

func scanChromiumRows(rows *sql.Rows, decryptor valueDecryptor) ([]Cookie, error) {
	cookies := []Cookie{}
	for rows.Next() {
		var cookie Cookie
		var encrypted []byte
		var expiresMicros int64
		err := rows.Scan(&cookie.Host, &cookie.Name, &cookie.Value, &encrypted, &cookie.Path, &expiresMicros)
		if err != nil {
			return nil, fmt.Errorf("read Chromium cookie: %w", err)
		}
		cookie.Expires = chromiumExpiry(expiresMicros)
		if cookie.Value == "" && len(encrypted) > 0 {
			cookie.Value, err = decryptor.decrypt(cookie.Host, encrypted)
			if err != nil {
				return nil, fmt.Errorf("cookie %s: %w", cookie.Name, err)
			}
		}
		cookies = append(cookies, cookie)
	}
	err := rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read Chromium cookies: %w", err)
	}
	return cookies, nil
}

func chromiumExpiry(expiresMicros int64) time.Time {
	if expiresMicros <= 0 {
		return time.Time{}
	}
	return time.UnixMicro(expiresMicros - webkitEpochOffsetMicros)
}
