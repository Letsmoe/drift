package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/letsmoe/drift/core/neoworks"
	"github.com/letsmoe/drift/core/neoworks/oskeystore"
)

const (
	neoworksCallbackPath = "/neoworks/callback"
	neoworksClientID     = "drift"
	// neoworksDefaultDomain is where Drift signs in unless DRIFT_NEOWORKS_DOMAIN
	// names another stack, e.g. neoworks.localhost:8443 for local development.
	neoworksDefaultDomain = "neoworks.dev"
)

var errNeoworksUnavailable = errors.New("neoworks sign-in is unavailable: the library or the system keyring could not be opened")

// NeoworksStatus is what the account menu shows.
type NeoworksStatus struct {
	// Available is false when the session could not be opened; sign-in is then hidden.
	Available bool `json:"available"`
	SignedIn  bool `json:"signedIn"`
	// SigningIn is true while the browser shows the Neoworks sign-in.
	SigningIn bool   `json:"signingIn"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Syncing   bool   `json:"syncing"`
	// LastSyncedAt is when the last sync finished without a failure.
	LastSyncedAt *time.Time `json:"lastSyncedAt"`
	// Error is the last sign-in or sync failure, empty after a success.
	Error string `json:"error"`
	// KeyInFile is true when no system keyring answered and the key that seals
	// the tokens and install keys lives in a file in Drift's data directory.
	KeyInFile bool `json:"keyInFile"`
}

// neoworksAccount is the Neoworks session and what the menu knows about it.
type neoworksAccount struct {
	session *neoworks.Session
	// syncer keeps the library in step with Neoworks; nil when the library could not be opened.
	syncer    *neoworks.Syncer
	keyInFile bool
	mutex     sync.Mutex
	status    NeoworksStatus
}

// signedOutStatus is the status of an available session nobody signed in to.
func (account *neoworksAccount) signedOutStatus() NeoworksStatus {
	return NeoworksStatus{Available: true, KeyInFile: account.keyInFile}
}

// openNeoworks opens the Neoworks session next to the library. Failure leaves
// sign-in unavailable and Drift works as before.
func (app *App) openNeoworks() {
	app.neoworks = &neoworksAccount{}
	if app.dataDirectory == "" {
		return
	}
	directory := filepath.Join(app.dataDirectory, "neoworks")
	keystore, err := oskeystore.Open("neoworks", directory)
	if err != nil {
		app.logger.Error("neoworks keystore", "error", err)
		return
	}
	if keystore.Backend() == oskeystore.BackendFile {
		app.logger.Warn("no system keyring answered; the Neoworks key is kept in a file", "directory", directory)
	}
	config, err := neoworksConfig(directory, keystore)
	if err != nil {
		app.logger.Error("neoworks config", "error", err)
		return
	}
	session, err := neoworks.Open(config)
	if err != nil {
		app.logger.Error("neoworks session failed to open", "error", err)
		return
	}
	app.neoworks.session = session
	app.neoworks.keyInFile = keystore.Backend() == oskeystore.BackendFile
	app.neoworks.status = app.neoworks.signedOutStatus()
	app.neoworks.status.SignedIn = session.SignedIn()
	app.startNeoworksSync(session)
	if session.SignedIn() {
		go app.refreshNeoworksProfile()
	}
}

// startNeoworksSync keeps the library in step with Neoworks from now on: after
// every change to it, on an interval and once at startup, while signed in.
func (app *App) startNeoworksSync(session *neoworks.Session) {
	if app.library == nil {
		return
	}
	librarySync := neoworks.NewLibrarySync(neoworks.LibrarySyncConfig{
		Library: app.library,
		Nodes:   session.Nodes(),
		UserID:  session.UserID,
		Device:  neoworks.Device{ID: session.InstallID(), Name: installName()},
		Logger:  app.logger,
	})
	syncer := neoworks.NewSyncer(neoworks.SyncerConfig{
		Session:     session,
		LibrarySync: librarySync,
		OnState:     app.observeNeoworksSync,
		Logger:      app.logger,
	})
	syncer.Watch(app.library)
	syncer.Start(app.ctx)
	app.neoworks.syncer = syncer
}

func neoworksConfig(directory string, keystore *oskeystore.Keystore) (neoworks.Config, error) {
	httpClient, err := neoworksHTTPClient()
	if err != nil {
		return neoworks.Config{}, err
	}
	domain := os.Getenv("DRIFT_NEOWORKS_DOMAIN")
	if domain == "" {
		domain = neoworksDefaultDomain
	}
	return neoworks.Config{
		OAuthURL:    "https://oauth." + domain,
		APIURL:      "https://api." + domain,
		ClientID:    neoworksClientID,
		InstallName: installName(),
		Directory:   directory,
		Keystore:    keystore,
		HTTPClient:  httpClient,
	}, nil
}

// neoworksHTTPClient trusts the system roots plus the PEM file named by
// DRIFT_NEOWORKS_CA, which a local stack's own CA needs.
func neoworksHTTPClient() (*http.Client, error) {
	authorityFile := os.Getenv("DRIFT_NEOWORKS_CA")
	if authorityFile == "" {
		return &http.Client{Timeout: time.Minute}, nil
	}
	pem, err := os.ReadFile(authorityFile)
	if err != nil {
		return nil, fmt.Errorf("read DRIFT_NEOWORKS_CA: %w", err)
	}
	roots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("system certificates: %w", err)
	}
	if !roots.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("DRIFT_NEOWORKS_CA %s holds no certificate", authorityFile)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	return &http.Client{Timeout: time.Minute, Transport: transport}, nil
}

func installName() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "Drift"
	}
	return "Drift on " + hostname
}

func (app *App) closeNeoworks() {
	if app.neoworks == nil || app.neoworks.session == nil {
		return
	}
	if app.neoworks.syncer != nil {
		app.neoworks.syncer.Stop()
	}
	app.neoworks.session.Close()
}

func (app *App) requireNeoworks() (*neoworks.Session, error) {
	if app.neoworks == nil || app.neoworks.session == nil {
		return nil, errNeoworksUnavailable
	}
	return app.neoworks.session, nil
}

// updateNeoworksStatus changes the status and tells the frontend.
func (app *App) updateNeoworksStatus(change func(status *NeoworksStatus)) {
	app.neoworks.mutex.Lock()
	change(&app.neoworks.status)
	status := app.neoworks.status
	app.neoworks.mutex.Unlock()
	app.emit("neoworks", status)
}

// NeoworksStatus returns whether a Neoworks account is signed in and how sync went.
func (app *App) NeoworksStatus() NeoworksStatus {
	if app.neoworks == nil {
		return NeoworksStatus{}
	}
	app.neoworks.mutex.Lock()
	defer app.neoworks.mutex.Unlock()
	return app.neoworks.status
}

// NeoworksSignIn starts signing in and returns the URL the frontend opens in
// the browser; the browser comes back to the backend's callback route.
func (app *App) NeoworksSignIn() (string, error) {
	session, err := app.requireNeoworks()
	if err != nil {
		return "", err
	}
	address, err := session.BeginSignIn(app.streamBaseURL + neoworksCallbackPath)
	if err != nil {
		return "", err
	}
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		status.SigningIn = true
		status.Error = ""
	})
	return address, nil
}

// NeoworksCancelSignIn abandons a sign-in; a late callback is then refused.
func (app *App) NeoworksCancelSignIn() {
	session, err := app.requireNeoworks()
	if err != nil {
		return
	}
	session.CancelSignIn()
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		status.SigningIn = false
	})
}

// NeoworksSignOut signs out and wipes the synced copy; the library stays.
func (app *App) NeoworksSignOut() error {
	session, err := app.requireNeoworks()
	if err != nil {
		return err
	}
	err = app.signOutNeoworks(session)
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		*status = app.neoworks.signedOutStatus()
	})
	return err
}

// signOutNeoworks also forgets which nodes hold which library entities, so a
// later sign-in does not mistake the wiped nodes for deletions.
func (app *App) signOutNeoworks(session *neoworks.Session) error {
	if app.neoworks.syncer == nil {
		return session.SignOut()
	}
	return app.neoworks.syncer.SignOut(app.ctx)
}

// NeoworksSync syncs now; the outcome arrives as a neoworks event.
func (app *App) NeoworksSync() error {
	if _, err := app.requireNeoworks(); err != nil {
		return err
	}
	go app.syncNeoworks()
	return nil
}

func (app *App) syncNeoworks() {
	if app.neoworks.syncer != nil {
		_, _ = app.neoworks.syncer.SyncNow(app.ctx)
		return
	}
	session := app.neoworks.session
	app.observeNeoworksSync(neoworks.SyncState{Syncing: true})
	_, err := session.Sync(app.ctx)
	app.observeNeoworksSync(neoworks.SyncState{Err: err})
	if err != nil {
		app.logger.Warn("neoworks sync", "error", err)
	}
}

// observeNeoworksSync shows a sync pass in the account menu.
func (app *App) observeNeoworksSync(state neoworks.SyncState) {
	session := app.neoworks.session
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		status.Syncing = state.Syncing
		if state.Syncing {
			return
		}
		status.SignedIn = session.SignedIn()
		status.Error = ""
		if state.Err != nil {
			status.Error = state.Err.Error()
			return
		}
		now := time.Now()
		status.LastSyncedAt = &now
	})
}

func (app *App) refreshNeoworksProfile() {
	ctx, cancel := context.WithTimeout(app.ctx, 30*time.Second)
	defer cancel()
	profile, err := app.neoworks.session.Profile(ctx)
	if err != nil {
		app.logger.Warn("neoworks profile", "error", err)
		return
	}
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		status.Name = profile.Name
		status.Email = profile.Email
		status.SignedIn = true
	})
}

// neoworksCallbackHandler finishes a sign-in when the browser comes back from
// Neoworks, then fetches the profile and syncs.
func (app *App) neoworksCallbackHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		session, err := app.requireNeoworks()
		if err != nil {
			writer.WriteHeader(http.StatusServiceUnavailable)
			_ = callbackPage.Execute(writer, callbackPageData{Failed: true, Detail: err.Error()})
			return
		}
		err = session.CompleteSignIn(request.Context(), request.URL.Query())
		app.updateNeoworksStatus(func(status *NeoworksStatus) {
			status.SigningIn = false
			status.SignedIn = err == nil
			status.Error = ""
			if err != nil {
				status.Error = err.Error()
			}
		})
		if err != nil {
			app.logger.Warn("neoworks sign-in", "error", err)
			writer.WriteHeader(http.StatusBadRequest)
			_ = callbackPage.Execute(writer, callbackPageData{Failed: true, Detail: err.Error()})
			return
		}
		_ = callbackPage.Execute(writer, callbackPageData{})
		go app.refreshNeoworksProfile()
		go app.syncNeoworks()
	})
}

type callbackPageData struct {
	Failed bool
	Detail string
}

var callbackPage = template.Must(template.New("callback").Parse(`<!doctype html>
<html lang="en">
<meta charset="utf-8">
<title>Drift</title>
<style>
body { font: 16px system-ui, sans-serif; background: #111; color: #eee; display: grid; place-items: center; min-height: 100vh; margin: 0; }
main { max-width: 28rem; text-align: center; }
p { color: #aaa; }
</style>
<main>
{{if .Failed}}
<h1>Sign-in did not finish</h1>
<p>Drift could not finish signing in to Neoworks. Go back to Drift and try again.</p>
<p><small>{{.Detail}}</small></p>
{{else}}
<h1>Signed in to Neoworks</h1>
<p>You can close this tab and go back to Drift.</p>
{{end}}
</main>
</html>
`))
