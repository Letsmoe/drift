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
}

// neoworksAccount is the Neoworks session and what the menu knows about it.
type neoworksAccount struct {
	session *neoworks.Session
	mutex   sync.Mutex
	status  NeoworksStatus
}

// openNeoworks opens the Neoworks session next to the library. Failure leaves
// sign-in unavailable and Drift works as before.
func (app *App) openNeoworks() {
	app.neoworks = &neoworksAccount{}
	if app.dataDirectory == "" {
		return
	}
	config, err := neoworksConfig(app.dataDirectory)
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
	app.neoworks.status = NeoworksStatus{Available: true, SignedIn: session.SignedIn()}
	if session.SignedIn() {
		go app.refreshNeoworksProfile()
	}
}

func neoworksConfig(dataDirectory string) (neoworks.Config, error) {
	keystore, err := oskeystore.Open("neoworks")
	if err != nil {
		return neoworks.Config{}, err
	}
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
		Directory:   filepath.Join(dataDirectory, "neoworks"),
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
	if app.neoworks != nil && app.neoworks.session != nil {
		app.neoworks.session.Close()
	}
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
	err = session.SignOut()
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		*status = NeoworksStatus{Available: true}
	})
	return err
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
	session := app.neoworks.session
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		status.Syncing = true
	})
	report, err := session.Sync(app.ctx)
	app.updateNeoworksStatus(func(status *NeoworksStatus) {
		status.Syncing = false
		status.SignedIn = session.SignedIn()
		status.Error = syncFailure(report.Failure, err)
		if status.Error == "" {
			now := time.Now()
			status.LastSyncedAt = &now
		}
	})
	if err != nil || report.Failure != "" {
		app.logger.Warn("neoworks sync", "error", err, "failure", report.Failure)
	}
}

func syncFailure(failure string, err error) string {
	if err != nil {
		return err.Error()
	}
	return failure
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
