package account

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/letsmoe/drift/core/importer"
	"github.com/letsmoe/drift/core/innertube"
	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

type libraryReader = importer.Account

type stubClient struct {
	libraryReader
	profile music.Account
	err     error
}

func (client stubClient) Account(context.Context) (music.Account, error) {
	return client.profile, client.err
}

func newService(t *testing.T, client stubClient) *Service {
	t.Helper()
	lib, err := library.Open(context.Background(), filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	return New(lib, WithClientFactory(func(innertube.Session) Client { return client }))
}

const validCookies = "SID=a; SAPISID=b"

func TestSignInStoresValidatedSession(t *testing.T) {
	service := newService(t, stubClient{profile: music.Account{Name: "Ada", Handle: "@ada"}})
	ctx := context.Background()

	status, err := service.SignIn(ctx, validCookies)
	if err != nil {
		t.Fatalf("sign in: %v", err)
	}
	if !status.SignedIn || status.Account.Name != "Ada" || status.LastSyncedAt != "" {
		t.Errorf("status = %+v", status)
	}
	reader, err := service.Reader(ctx)
	if err != nil || reader == nil {
		t.Errorf("reader = %v, err = %v", reader, err)
	}
}

func TestSignInRejectsInvalidCookies(t *testing.T) {
	service := newService(t, stubClient{err: innertube.ErrSessionExpired})
	ctx := context.Background()

	_, err := service.SignIn(ctx, "SID=a")
	if !errors.Is(err, innertube.ErrNoSessionCookies) {
		t.Errorf("incomplete cookies: err = %v", err)
	}
	_, err = service.SignIn(ctx, validCookies)
	if !errors.Is(err, innertube.ErrSessionExpired) {
		t.Errorf("expired cookies: err = %v", err)
	}
	if strings.Contains(err.Error(), "SAPISID") {
		t.Errorf("error leaks cookies: %v", err)
	}
	status, _ := service.Status(ctx)
	if status.SignedIn {
		t.Error("a failed sign-in must store nothing")
	}
}

func TestSignOutAndReader(t *testing.T) {
	service := newService(t, stubClient{profile: music.Account{Name: "Ada"}})
	ctx := context.Background()
	_, err := service.Reader(ctx)
	if !errors.Is(err, ErrNotSignedIn) {
		t.Errorf("reader without session: err = %v", err)
	}
	_, err = service.SignIn(ctx, validCookies)
	if err != nil {
		t.Fatal(err)
	}

	err = service.SignOut(ctx)
	if err != nil {
		t.Fatal(err)
	}

	status, _ := service.Status(ctx)
	if status.SignedIn {
		t.Error("still signed in after sign-out")
	}
}
