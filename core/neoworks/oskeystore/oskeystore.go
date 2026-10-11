// Package oskeystore keeps Drift's Neoworks secrets under a key the operating
// system holds: the Secret Service on Linux, the Keychain on macOS, the
// Credential Manager on Windows, or an existing KWallet wallet. Without any of
// them the key lives in a file only the user can read. Android passes its own
// keystore in.
package oskeystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// Backends a key can live in.
const (
	BackendSystem  = "system-keyring"
	BackendKWallet = "kwallet"
	BackendFile    = "file"
)

const (
	keyringService = "dev.letsmoe.drift"
	keySize        = 32
)

// ErrWrapped means a wrapped secret was too short, altered, or sealed under
// another key or label.
var ErrWrapped = errors.New("oskeystore: wrapped secret does not open")

// Keystore seals secrets with AES-GCM under a random key, the label being the
// associated data.
type Keystore struct {
	account string
	source  keySource
	aead    cipher.AEAD
}

// Open loads the key stored for account, creating it on first use. The key
// file, the last resort, goes in fallbackDirectory.
func Open(account, fallbackDirectory string) (*Keystore, error) {
	sources := append([]keySource{secretService{}}, platformSources()...)
	sources = append(sources, keyFile{directory: fallbackDirectory})
	return openFrom(account, sources)
}

// openFrom uses the first source that already holds a key, so a keyring that
// appears later never orphans the key in use; without one it stores a new key
// in the first source that takes it.
func openFrom(account string, sources []keySource) (*Keystore, error) {
	var failures []error
	var working []keySource
	for _, source := range sources {
		key, err := source.load(account)
		if err == nil {
			return newKeystore(account, source, key)
		}
		if errors.Is(err, errKeyNotFound) {
			working = append(working, source)
			continue
		}
		failures = append(failures, err)
	}
	return createKey(account, working, failures)
}

func createKey(account string, sources []keySource, failures []error) (*Keystore, error) {
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	for _, source := range sources {
		err := source.store(account, key)
		if err == nil {
			return newKeystore(account, source, key)
		}
		failures = append(failures, err)
	}
	return nil, errors.Join(append([]error{errors.New("oskeystore: nowhere to keep the key")}, failures...)...)
}

func newKeystore(account string, source keySource, key []byte) (*Keystore, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	return &Keystore{account: account, source: source, aead: aead}, nil
}

func encodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func decodeKey(stored string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(stored)
	if err != nil || len(key) != keySize {
		return nil, errors.New("oskeystore: the stored key is malformed")
	}
	return key, nil
}

// Backend says where the key lives: BackendSystem, BackendKWallet or BackendFile.
func (keystore *Keystore) Backend() string {
	return keystore.source.name()
}

// Name identifies the backend to libneoworks.
func (keystore *Keystore) Name() string {
	return keystore.Backend()
}

// Wrap seals secret, bound to label.
func (keystore *Keystore) Wrap(secret, label []byte) ([]byte, error) {
	nonce := make([]byte, keystore.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	return keystore.aead.Seal(nonce, nonce, secret, label), nil
}

// Unwrap opens what Wrap sealed under the same label.
func (keystore *Keystore) Unwrap(wrapped, label []byte) ([]byte, error) {
	nonceSize := keystore.aead.NonceSize()
	if len(wrapped) < nonceSize {
		return nil, ErrWrapped
	}
	secret, err := keystore.aead.Open(nil, wrapped[:nonceSize], wrapped[nonceSize:], label)
	if err != nil {
		return nil, ErrWrapped
	}
	return secret, nil
}
