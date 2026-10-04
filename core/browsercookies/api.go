package browsercookies

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"
)

// BrowserSession is a browser profile and what its cookie store holds.
type BrowserSession struct {
	Browser string `json:"browser"`
	Profile string `json:"profile"`
	// HasYouTubeSession says the profile is signed in to YouTube.
	HasYouTubeSession bool `json:"hasYouTubeSession"`
	// Reason explains why a profile has no usable session: not signed in,
	// locked keyring, unsupported encryption, ...
	Reason string `json:"reason"`
}

// ErrProfileNotFound means no profile matches the browser and profile name.
var ErrProfileNotFound = errors.New("browsercookies: no such browser profile")

// DetectBrowserSessions inspects the cookie stores of the browsers installed
// for this user. It reads cookie databases (copies of them) and may ask the
// system keyring for a key, so call it only on the user's explicit action.
func DetectBrowserSessions() []BrowserSession {
	return defaultEnvironment().detect()
}

// LoadSession returns the Cookie header music.youtube.com would receive from
// the named browser profile.
func LoadSession(browser, profileName string) (string, error) {
	return defaultEnvironment().load(browser, profileName)
}

func defaultEnvironment() environment {
	homeDir, _ := os.UserHomeDir()
	return environment{
		goos:            runtime.GOOS,
		homeDir:         homeDir,
		appDataDir:      os.Getenv("APPDATA"),
		localAppDataDir: os.Getenv("LOCALAPPDATA"),
		tempDir:         os.TempDir(),
		password:        systemKeyringPassword,
	}
}

func (env environment) detect() []BrowserSession {
	env.password = memoizePassword(env.password)
	sessions := []BrowserSession{}
	for _, candidate := range env.profiles() {
		session := BrowserSession{Browser: candidate.spec.name, Profile: candidate.name}
		cookies, err := env.readCookies(candidate)
		switch {
		case err != nil:
			session.Reason = err.Error()
		case !hasSession(musicCookies(cookies, time.Now())):
			session.Reason = "not signed in to YouTube Music"
		default:
			session.HasYouTubeSession = true
		}
		sessions = append(sessions, session)
	}
	return sessions
}

func (env environment) load(browser, profileName string) (string, error) {
	for _, candidate := range env.profiles() {
		if candidate.spec.name != browser || candidate.name != profileName {
			continue
		}
		return env.loadProfile(candidate)
	}
	return "", fmt.Errorf("%w: %s %s", ErrProfileNotFound, browser, profileName)
}

func (env environment) loadProfile(candidate profile) (string, error) {
	cookies, err := env.readCookies(candidate)
	if err != nil {
		return "", err
	}
	selected := musicCookies(cookies, time.Now())
	if !hasSession(selected) {
		return "", ErrNoSession
	}
	return cookieHeader(selected), nil
}

// errWindowsChromium explains why Chromium browsers on Windows are skipped.
var errWindowsChromium = errors.New("chromium browsers on Windows encrypt cookies with app-bound keys that other programs cannot read; use Firefox or paste the cookies")

func (env environment) readCookies(candidate profile) ([]Cookie, error) {
	if candidate.spec.family == familyFirefox {
		return readFirefoxCookies(env.tempDir, candidate.databasePath)
	}
	if env.goos == "windows" {
		return nil, errWindowsChromium
	}
	spec := candidate.spec
	if env.goos == "darwin" && macKeyringNames[spec.name] != "" {
		spec.keyringName = macKeyringNames[spec.name]
	}
	return readChromiumCookies(env.tempDir, candidate.databasePath, newDecryptor(env.goos, spec, env.password))
}
