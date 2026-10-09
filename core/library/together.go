package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TogetherIdentity is who this user is in Listen Together. SigningKey is a
// credential: it never prints itself.
type TogetherIdentity struct {
	UserID string
	// SigningKey is the hex Ed25519 seed UserID is derived from.
	SigningKey    string
	Name          string
	SharePresence bool
}

// String hides the signing key from logs.
func (identity TogetherIdentity) String() string {
	return fmt.Sprintf("library.TogetherIdentity{UserID: %s, Name: %q}", identity.UserID, identity.Name)
}

// GoString hides the signing key from %#v output.
func (identity TogetherIdentity) GoString() string { return identity.String() }

// TogetherFriend is someone the user listens with.
type TogetherFriend struct {
	UserID string
	Name   string
	// Trusted friends join the user's sessions without being asked for.
	Trusted bool
	AddedAt time.Time
}

// TogetherIdentity returns the stored identity; found is false before the first save.
func (library *Library) TogetherIdentity(ctx context.Context) (identity TogetherIdentity, found bool, err error) {
	err = library.db.QueryRowContext(ctx, `SELECT user_id, signing_key, name, share_presence
		FROM together_identity WHERE id = 1`).
		Scan(&identity.UserID, &identity.SigningKey, &identity.Name, &identity.SharePresence)
	if errors.Is(err, sql.ErrNoRows) {
		return TogetherIdentity{}, false, nil
	}
	if err != nil {
		return TogetherIdentity{}, false, fmt.Errorf("library: load together identity: %w", err)
	}
	return identity, true, nil
}

// SaveTogetherIdentity stores the identity, replacing the earlier one.
func (library *Library) SaveTogetherIdentity(ctx context.Context, identity TogetherIdentity) error {
	if identity.UserID == "" || identity.SigningKey == "" {
		return errors.New("library: save together identity: empty user ID or key")
	}
	_, err := library.db.ExecContext(ctx, `INSERT INTO together_identity
		(id, user_id, signing_key, name, share_presence, created_at) VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET user_id = excluded.user_id, signing_key = excluded.signing_key,
			name = excluded.name, share_presence = excluded.share_presence`,
		identity.UserID, identity.SigningKey, identity.Name, identity.SharePresence, library.timestamp())
	if err != nil {
		// The driver error never contains bound values, but say nothing about them anyway.
		return fmt.Errorf("library: save together identity: %w", err)
	}
	return nil
}

// SaveTogetherFriend adds a friend or updates an existing one's name and trust.
func (library *Library) SaveTogetherFriend(ctx context.Context, friend TogetherFriend) error {
	if friend.UserID == "" {
		return errors.New("library: save friend: empty user ID")
	}
	_, err := library.db.ExecContext(ctx, `INSERT INTO together_friend (user_id, name, trusted, added_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET name = excluded.name, trusted = excluded.trusted`,
		friend.UserID, friend.Name, friend.Trusted, library.timestamp())
	if err != nil {
		return fmt.Errorf("library: save friend %s: %w", friend.UserID, err)
	}
	return nil
}

// RemoveTogetherFriend deletes a friend; removing an unknown one is not an error.
func (library *Library) RemoveTogetherFriend(ctx context.Context, userID string) error {
	_, err := library.db.ExecContext(ctx, `DELETE FROM together_friend WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("library: remove friend %s: %w", userID, err)
	}
	return nil
}

// TogetherFriends lists every friend by name.
func (library *Library) TogetherFriends(ctx context.Context) ([]TogetherFriend, error) {
	rows, err := library.db.QueryContext(ctx, `SELECT user_id, name, trusted, added_at FROM together_friend
		ORDER BY name COLLATE NOCASE, user_id`)
	if err != nil {
		return nil, fmt.Errorf("library: query friends: %w", err)
	}
	defer func() { _ = rows.Close() }()

	friends := []TogetherFriend{}
	for rows.Next() {
		friend, err := scanFriend(rows.Scan)
		if err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("library: read friends: %w", err)
	}
	return friends, nil
}

func scanFriend(scan func(...any) error) (TogetherFriend, error) {
	var friend TogetherFriend
	var addedAt string
	if err := scan(&friend.UserID, &friend.Name, &friend.Trusted, &addedAt); err != nil {
		return TogetherFriend{}, fmt.Errorf("library: scan friend: %w", err)
	}
	parsed, err := ParseTime(addedAt)
	if err != nil {
		return TogetherFriend{}, fmt.Errorf("library: friend time: %w", err)
	}
	friend.AddedAt = parsed
	return friend, nil
}
