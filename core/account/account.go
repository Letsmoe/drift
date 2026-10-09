// Package account manages the optional YouTube Music sign-in. The session
// (the user's cookies) is stored in the local library and used only to read
// the user's own library for import and sync; catalog browsing, search,
// recommendations and playback stay anonymous.
package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/letsmoe/drift/core/importer"
	"github.com/letsmoe/drift/core/innertube"
	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

// ErrNotSignedIn means a call needs a YouTube Music session and none is stored.
var ErrNotSignedIn = errors.New("account: not signed in to YouTube Music")

// Client is what the service needs from an authenticated YouTube Music client.
// *innertube.AccountClient implements it.
type Client interface {
	importer.Account
	Account(ctx context.Context) (music.Account, error)
}

// ClientFactory builds the client for a session.
type ClientFactory func(session innertube.Session) Client

// Status says who is signed in. Times are RFC 3339; empty when unknown.
type Status struct {
	SignedIn     bool          `json:"signedIn"`
	Account      music.Account `json:"account"`
	ValidatedAt  string        `json:"validatedAt"`
	LastSyncedAt string        `json:"lastSyncedAt"`
}

// Service signs in and out and hands out authenticated library readers.
type Service struct {
	library   *library.Library
	newClient ClientFactory
}

// Option configures New.
type Option func(*Service)

// WithClientFactory replaces the InnerTube client factory, for tests.
func WithClientFactory(factory ClientFactory) Option {
	return func(service *Service) {
		service.newClient = factory
	}
}

// New returns a service storing its session in lib.
func New(lib *library.Library, options ...Option) *Service {
	service := &Service{library: lib}
	service.newClient = func(session innertube.Session) Client {
		return innertube.NewAccountClient(session)
	}
	for _, option := range options {
		option(service)
	}
	return service
}

// SignIn validates a cookie header against YouTube Music and stores it. The
// cookies are stored only when they work.
func (service *Service) SignIn(ctx context.Context, cookieHeader string) (Status, error) {
	session := innertube.Session{CookieHeader: cookieHeader}
	err := session.Validate()
	if err != nil {
		return Status{}, err
	}
	account, err := service.newClient(session).Account(ctx)
	if err != nil {
		return Status{}, fmt.Errorf("account: sign in: %w", err)
	}
	err = service.library.SaveYouTubeSession(ctx, cookieHeader, account)
	if err != nil {
		return Status{}, err
	}
	return service.Status(ctx)
}

// Status reads the stored sign-in without asking YouTube.
func (service *Service) Status(ctx context.Context) (Status, error) {
	stored, found, err := service.library.YouTubeSession(ctx)
	if err != nil || !found {
		return Status{}, err
	}
	status := Status{
		SignedIn:    true,
		Account:     stored.Account,
		ValidatedAt: library.FormatTime(stored.ValidatedAt),
	}
	if !stored.LastSyncedAt.IsZero() {
		status.LastSyncedAt = library.FormatTime(stored.LastSyncedAt)
	}
	return status, nil
}

// SignOut deletes the stored session.
func (service *Service) SignOut(ctx context.Context) error {
	return service.library.ClearYouTubeSession(ctx)
}

// Reader returns the library reader of the stored session, or ErrNotSignedIn.
func (service *Service) Reader(ctx context.Context) (importer.Account, error) {
	stored, found, err := service.library.YouTubeSession(ctx)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotSignedIn
	}
	return service.newClient(innertube.Session{CookieHeader: stored.CookieHeader}), nil
}
