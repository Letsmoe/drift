// Package neoworks signs Drift in to Neoworks and syncs the library through it,
// end to end encrypted, using libneoworks. Signing in is optional; without it
// nothing here talks to the network.
package neoworks

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	libneoworks "github.com/neoworks-dev/libneoworks-go"
)

// ErrSignedOut means the action needs a signed-in session.
var ErrSignedOut = errors.New("neoworks: not signed in")

// Config says where Neoworks is and where the session keeps its files.
type Config struct {
	// OAuthURL and APIURL are origins, e.g. https://oauth.neoworks.dev.
	OAuthURL string
	APIURL   string
	ClientID string
	// InstallName is shown on the consent screen, e.g. "Drift on laptop".
	InstallName string
	// Directory holds the sealed install and tokens and the encrypted store.
	Directory string
	Keystore  libneoworks.Keystore
	// HTTPClient defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// Session is this install's connection to a Neoworks account: its keys, the
// encrypted local store, the tokens of the signed-in user and the sync loop.
type Session struct {
	config Config
	client *libneoworks.Client
	// description is the install's ID and public keys, sent at sign-in.
	description libneoworks.InstallDescription

	mutex                 sync.Mutex
	tokens                *tokens
	pending               *pendingSignIn
	descriptorsRegistered bool
}

// Open loads the install sealed in config.Directory, creating it on first
// use, and the tokens of a previous sign-in.
func Open(config Config) (*Session, error) {
	if config.HTTPClient == nil {
		config.HTTPClient = http.DefaultClient
	}
	if err := os.MkdirAll(config.Directory, 0o700); err != nil {
		return nil, fmt.Errorf("neoworks: %w", err)
	}
	install, err := loadOrCreateInstall(config)
	if err != nil {
		return nil, err
	}
	defer install.Close()
	description, err := install.Describe()
	if err != nil {
		return nil, fmt.Errorf("neoworks: describing the install: %w", err)
	}
	client, err := libneoworks.OpenClient(install, filepath.Join(config.Directory, "store.db"))
	if err != nil {
		return nil, fmt.Errorf("neoworks: opening the store: %w", err)
	}
	session := &Session{config: config, client: client, description: description}
	session.tokens, err = loadTokens(config)
	if err != nil {
		client.Close()
		return nil, err
	}
	return session, nil
}

func loadOrCreateInstall(config Config) (*libneoworks.Install, error) {
	path := filepath.Join(config.Directory, "install.sealed")
	sealed, err := os.ReadFile(path)
	if err == nil {
		install, err := libneoworks.OpenInstall(sealed, config.Keystore)
		if err != nil {
			return nil, fmt.Errorf("neoworks: unsealing the install: %w", err)
		}
		return install, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("neoworks: %w", err)
	}
	install, err := libneoworks.NewInstall()
	if err != nil {
		return nil, fmt.Errorf("neoworks: creating the install: %w", err)
	}
	sealed, err = install.Seal(config.Keystore)
	if err == nil {
		err = writePrivateFile(path, sealed)
	}
	if err != nil {
		install.Close()
		return nil, err
	}
	return install, nil
}

func writePrivateFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("neoworks: %w", err)
	}
	return nil
}

// SignedIn reports whether a user is signed in on this install.
func (session *Session) SignedIn() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	return session.tokens != nil
}

// UserID is the signed-in user's Neoworks ID, empty when signed out.
func (session *Session) UserID() string {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	if session.tokens == nil {
		return ""
	}
	return session.tokens.UserID
}

// InstallID identifies this install; it names the device in the synced history.
func (session *Session) InstallID() string {
	return session.description.InstallID
}

// Nodes is the encrypted node store of this session. Its collections can only
// be used once a Sync registered the descriptors.
func (session *Session) Nodes() Nodes {
	return session.client
}

// SignOut forgets the tokens and wipes the local store. The install keeps its
// keys, so signing in again needs a new consent but no new install.
func (session *Session) SignOut() error {
	session.mutex.Lock()
	session.tokens = nil
	session.pending = nil
	session.descriptorsRegistered = false
	session.mutex.Unlock()
	if err := session.client.Wipe(); err != nil {
		return fmt.Errorf("neoworks: wiping the store: %w", err)
	}
	err := os.Remove(filepath.Join(session.config.Directory, "tokens.sealed"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("neoworks: %w", err)
	}
	return nil
}

// Close frees the session. It must not be used afterwards.
func (session *Session) Close() {
	session.client.Close()
}
