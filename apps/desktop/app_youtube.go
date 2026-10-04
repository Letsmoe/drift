package main

import (
	"fmt"

	"github.com/neoworks/otter-music/core/account"
	"github.com/neoworks/otter-music/core/browsercookies"
	"github.com/neoworks/otter-music/core/importer"
)

func (app *App) requireAccount() (*account.Service, error) {
	if app.account == nil {
		return nil, errImportUnavailable
	}
	return app.account, nil
}

// DetectBrowserSessions lists the browser profiles on this computer and whether
// each is signed in to YouTube. It opens cookie stores, so the frontend calls
// it only when the user asks.
func (app *App) DetectBrowserSessions() []browsercookies.BrowserSession {
	return browsercookies.DetectBrowserSessions()
}

// YouTubeSignInFromBrowser reads the YouTube Music session of a browser
// profile, validates it and stores it for library import.
func (app *App) YouTubeSignInFromBrowser(browser, profile string) (account.Status, error) {
	service, err := app.requireAccount()
	if err != nil {
		return account.Status{}, err
	}
	cookieHeader, err := browsercookies.LoadSession(browser, profile)
	if err != nil {
		return account.Status{}, fmt.Errorf("read the %s session: %w", browser, err)
	}
	return service.SignIn(app.ctx, cookieHeader)
}

// YouTubeSignInWithCookies signs in with pasted text: a Cookie header, a
// block of request headers or a "Copy as cURL" command.
func (app *App) YouTubeSignInWithCookies(pasted string) (account.Status, error) {
	service, err := app.requireAccount()
	if err != nil {
		return account.Status{}, err
	}
	cookieHeader, err := browsercookies.ParseCookieHeader(pasted)
	if err != nil {
		return account.Status{}, err
	}
	return service.SignIn(app.ctx, cookieHeader)
}

// YouTubeAccountStatus returns who is signed in and when the library was last
// synced, without asking YouTube.
func (app *App) YouTubeAccountStatus() (account.Status, error) {
	service, err := app.requireAccount()
	if err != nil {
		return account.Status{}, err
	}
	return service.Status(app.ctx)
}

// YouTubeSignOut deletes the stored YouTube Music session.
func (app *App) YouTubeSignOut() error {
	service, err := app.requireAccount()
	if err != nil {
		return err
	}
	return service.SignOut(app.ctx)
}

// YouTubePreview counts the signed-in account's library.
func (app *App) YouTubePreview() (importer.Summary, error) {
	runner, err := app.requireImporter()
	if err != nil {
		return importer.Summary{}, err
	}
	return runner.Preview(app.ctx, importer.Source{Kind: importer.SourceYouTubeAccount})
}

// YouTubeImport imports the selected parts of the account's library in the
// background and reports through the import-progress, import-done and
// import-error events. Running it again re-syncs.
func (app *App) YouTubeImport(options importer.Options) error {
	return app.startImport(importer.Source{Kind: importer.SourceYouTubeAccount}, options)
}
