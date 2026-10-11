package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/neoworks"
	"github.com/letsmoe/drift/core/neoworks/oskeystore"
)

// recordedEvents keeps what the backend emitted.
type recordedEvents struct {
	mutex  sync.Mutex
	events []string
}

func (recorded *recordedEvents) emit(name string, _ any) {
	recorded.mutex.Lock()
	defer recorded.mutex.Unlock()
	recorded.events = append(recorded.events, name)
}

func newNeoworksTestApp(t *testing.T) (*App, *recordedEvents) {
	t.Helper()
	keyring.MockInit()
	keystore, err := oskeystore.Open("neoworks-test", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	session, err := neoworks.Open(neoworks.Config{
		OAuthURL: "https://oauth.neoworks.test", APIURL: "https://api.neoworks.test", ClientID: neoworksClientID,
		InstallName: "Drift test", Directory: t.TempDir(), Keystore: keystore,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	recorded := &recordedEvents{}
	app := &App{ctx: context.Background(), logger: discardLogger(), emit: recorded.emit, streamBaseURL: "http://127.0.0.1:49152"}
	app.neoworks = &neoworksAccount{session: session, status: NeoworksStatus{Available: true}}
	return app, recorded
}

func TestNeoworksSignInComesBackToTheBackend(t *testing.T) {
	app, recorded := newNeoworksTestApp(t)
	address, err := app.NeoworksSignIn()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host != "oauth.neoworks.test" {
		t.Fatalf("authorize URL %s %v", address, err)
	}
	if parsed.Query().Get("redirect_uri") != "http://127.0.0.1:49152"+neoworksCallbackPath {
		t.Fatalf("redirect_uri %q", parsed.Query().Get("redirect_uri"))
	}
	if !app.NeoworksStatus().SigningIn || len(recorded.events) != 1 {
		t.Fatalf("status %+v, events %v", app.NeoworksStatus(), recorded.events)
	}
}

func TestNeoworksCallbackRefusesAForgedState(t *testing.T) {
	app, _ := newNeoworksTestApp(t)
	if _, err := app.NeoworksSignIn(); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, neoworksCallbackPath+"?state=forged&code=code-1", nil)
	app.neoworksCallbackHandler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "did not finish") {
		t.Fatalf("callback answered %d: %s", recorder.Code, recorder.Body.String())
	}
	status := app.NeoworksStatus()
	if status.SigningIn || status.SignedIn || status.Error == "" {
		t.Fatalf("status after a forged callback %+v", status)
	}
}

func TestNeoworksCallbackBeforeStartupAnswersUnavailable(t *testing.T) {
	app := &App{logger: discardLogger()}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, neoworksCallbackPath+"?state=x", nil)
	app.neoworksCallbackHandler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("callback answered %d", recorder.Code)
	}
}

func TestNeoworksSyncReactsToLibraryChangesAndSignOutForgetsTheLinks(t *testing.T) {
	app, _ := newNeoworksTestApp(t)
	lib, err := library.Open(t.Context(), filepath.Join(t.TempDir(), "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lib.Close() })
	app.library = lib
	app.startNeoworksSync(app.neoworks.session)
	if app.neoworks.syncer == nil {
		t.Fatal("a library and a session must start the syncer")
	}
	t.Cleanup(app.neoworks.syncer.Stop)
	link := library.NodeLink{Collection: "ratings", Key: "abc", NodeID: "node-1", ContentHash: "h"}
	if err = lib.SaveNodeLink(t.Context(), link); err != nil {
		t.Fatal(err)
	}

	if err = app.NeoworksSignOut(); err != nil {
		t.Fatal(err)
	}

	links, err := lib.NodeLinks(t.Context(), "ratings")
	if err != nil || len(links) != 0 {
		t.Errorf("links after sign-out = %v, %v", links, err)
	}
}
