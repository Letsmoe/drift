package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/letsmoe/drift/core/music"
)

// YouTubeSession is the stored YouTube Music sign-in. The cookie header is a
// credential: it never prints itself and no error carries it.
type YouTubeSession struct {
	CookieHeader string
	Account      music.Account
	CreatedAt    time.Time
	ValidatedAt  time.Time
	// LastSyncedAt is zero until a library import finished.
	LastSyncedAt time.Time
}

// String hides the cookie header from logs.
func (session YouTubeSession) String() string { return "library.YouTubeSession(redacted)" }

// GoString hides the cookie header from %#v output.
func (session YouTubeSession) GoString() string { return session.String() }

// LogValue hides the cookie header from slog.
func (session YouTubeSession) LogValue() slog.Value { return slog.StringValue(session.String()) }

// SaveYouTubeSession stores the session of account, replacing any earlier one.
// The previous last-synced time is kept when the account did not change.
func (library *Library) SaveYouTubeSession(ctx context.Context, cookieHeader string, account music.Account) error {
	now := library.timestamp()
	_, err := library.db.ExecContext(ctx, `INSERT INTO youtube_session
		(id, cookie_header, account_name, account_handle, account_email, account_photo, created_at, validated_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			cookie_header = excluded.cookie_header, account_name = excluded.account_name,
			account_handle = excluded.account_handle, account_email = excluded.account_email,
			account_photo = excluded.account_photo, created_at = excluded.created_at,
			validated_at = excluded.validated_at,
			last_synced_at = CASE WHEN account_handle = excluded.account_handle AND account_name = excluded.account_name
				THEN last_synced_at ELSE '' END`,
		cookieHeader, account.Name, account.Handle, account.Email, account.PhotoURL, now, now)
	if err != nil {
		// The driver error never contains bound values, but say nothing about them anyway.
		return fmt.Errorf("library: save YouTube session: %w", err)
	}
	return nil
}

// YouTubeSession returns the stored session; found is false when nobody is signed in.
func (library *Library) YouTubeSession(ctx context.Context) (session YouTubeSession, found bool, err error) {
	var createdAt, validatedAt, lastSyncedAt string
	err = library.db.QueryRowContext(ctx, `SELECT cookie_header, account_name, account_handle, account_email, account_photo,
		created_at, validated_at, last_synced_at FROM youtube_session WHERE id = 1`).
		Scan(&session.CookieHeader, &session.Account.Name, &session.Account.Handle, &session.Account.Email,
			&session.Account.PhotoURL, &createdAt, &validatedAt, &lastSyncedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return YouTubeSession{}, false, nil
	}
	if err != nil {
		return YouTubeSession{}, false, fmt.Errorf("library: load YouTube session: %w", err)
	}
	err = parseSessionTimes(&session, createdAt, validatedAt, lastSyncedAt)
	if err != nil {
		return YouTubeSession{}, false, err
	}
	return session, true, nil
}

func parseSessionTimes(session *YouTubeSession, createdAt, validatedAt, lastSyncedAt string) error {
	var err error
	session.CreatedAt, err = ParseTime(createdAt)
	if err != nil {
		return err
	}
	session.ValidatedAt, err = ParseTime(validatedAt)
	if err != nil {
		return err
	}
	if lastSyncedAt == "" {
		return nil
	}
	session.LastSyncedAt, err = ParseTime(lastSyncedAt)
	return err
}

// ClearYouTubeSession signs out: it deletes the stored session. Signing out
// when nobody is signed in is not an error.
func (library *Library) ClearYouTubeSession(ctx context.Context) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM youtube_session`)
	if err != nil {
		return fmt.Errorf("library: clear YouTube session: %w", err)
	}
	return nil
}

// MarkYouTubeValidated records that the session just worked.
func (library *Library) MarkYouTubeValidated(ctx context.Context) error {
	return library.updateYouTubeTime(ctx, "validated_at")
}

// MarkYouTubeSynced records that a library import from the account finished.
func (library *Library) MarkYouTubeSynced(ctx context.Context) error {
	return library.updateYouTubeTime(ctx, "last_synced_at")
}

func (library *Library) updateYouTubeTime(ctx context.Context, column string) error {
	_, err := library.db.ExecContext(ctx, `UPDATE youtube_session SET `+column+` = ? WHERE id = 1`, library.timestamp())
	if err != nil {
		return fmt.Errorf("library: update YouTube session %s: %w", column, err)
	}
	return nil
}
