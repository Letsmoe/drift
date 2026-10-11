// Package oskeystore keeps Drift's Neoworks secrets under a key held by the
// operating system's keyring: Secret Service on Linux, the Keychain on macOS
// and the Credential Manager on Windows. Android passes its own keystore in.
package oskeystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "dev.letsmoe.drift"
	keySize        = 32
)

// ErrWrapped means a wrapped secret was too short, altered, or sealed under
// another key or label.
var ErrWrapped = errors.New("oskeystore: wrapped secret does not open")

// Keystore seals secrets with AES-GCM under a random key stored in the OS
// keyring as account, the label being the associated data.
type Keystore struct {
	account string
	aead    cipher.AEAD
}

// Open loads the key stored for account, creating it on first use.
func Open(account string) (*Keystore, error) {
	key, err := loadOrCreateKey(account)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	return &Keystore{account: account, aead: aead}, nil
}

func loadOrCreateKey(account string) ([]byte, error) {
	stored, err := keyring.Get(keyringService, account)
	if err == nil {
		return decodeKey(stored)
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return nil, fmt.Errorf("oskeystore: reading the keyring: %w", err)
	}
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("oskeystore: %w", err)
	}
	if err := keyring.Set(keyringService, account, base64.StdEncoding.EncodeToString(key)); err != nil {
		return nil, fmt.Errorf("oskeystore: writing the keyring: %w", err)
	}
	return key, nil
}

func decodeKey(stored string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(stored)
	if err != nil || len(key) != keySize {
		return nil, errors.New("oskeystore: the keyring holds a malformed key")
	}
	return key, nil
}

// Name identifies the backend to libneoworks.
func (keystore *Keystore) Name() string {
	return "os-keyring"
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

// Delete removes the key from the keyring; everything sealed under it is lost.
func (keystore *Keystore) Delete() error {
	err := keyring.Delete(keyringService, keystore.account)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("oskeystore: %w", err)
	}
	return nil
}
