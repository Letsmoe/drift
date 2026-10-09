package mobile

import (
	"github.com/letsmoe/drift/core/account"
	"github.com/letsmoe/drift/core/importer"
)

func (core *Core) requireAccount() (*account.Service, error) {
	if core.account == nil {
		return nil, errAccountUnavailable
	}
	return core.account, nil
}

// YouTubeSignIn validates the Cookie header of a signed-in music.youtube.com
// session (Android passes its WebView's cookies) and stores it for library
// import. It returns account.Status JSON. The cookies are used only to read
// the user's library; they never reach catalog, search or playback requests.
func (core *Core) YouTubeSignIn(cookieHeader string) (string, error) {
	service, err := core.requireAccount()
	if err != nil {
		return "", err
	}
	status, err := service.SignIn(core.ctx, cookieHeader)
	return encodeResult(status, err)
}

// YouTubeAccount returns the stored sign-in as account.Status JSON; signedIn
// is false when nobody is signed in. It makes no network request.
func (core *Core) YouTubeAccount() (string, error) {
	service, err := core.requireAccount()
	if err != nil {
		return "", err
	}
	status, err := service.Status(core.ctx)
	return encodeResult(status, err)
}

// YouTubeSignOut deletes the stored session.
func (core *Core) YouTubeSignOut() error {
	service, err := core.requireAccount()
	if err != nil {
		return err
	}
	return service.SignOut(core.ctx)
}

// YouTubePreview counts the signed-in account's library as importer.Summary JSON.
func (core *Core) YouTubePreview() (string, error) {
	return core.previewSource(importer.Source{Kind: importer.SourceYouTubeAccount})
}

// YouTubeImport imports the selected parts (optionsJSON is an importer.Options
// object) of the account's library in the background; run it again to re-sync.
// Progress and the end arrive on listener like SpotifyImport; cancel it with
// SpotifyCancelImport.
func (core *Core) YouTubeImport(optionsJSON string, listener ImportListener) error {
	return core.startImport(importer.Source{Kind: importer.SourceYouTubeAccount}, optionsJSON, listener)
}
