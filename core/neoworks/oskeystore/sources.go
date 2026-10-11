package oskeystore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

// errKeyNotFound means a source works but holds no key for the account yet.
var errKeyNotFound = errors.New("oskeystore: no key stored")

// keySource is one place the key can live. load answers errKeyNotFound when
// the source works but has no key, and any other error when it does not work.
type keySource interface {
	name() string
	load(account string) ([]byte, error)
	store(account string, key []byte) error
}

// secretService is the freedesktop Secret Service, the macOS Keychain or the
// Windows Credential Manager, whichever go-keyring talks to on this system.
type secretService struct{}

func (secretService) name() string { return BackendSystem }

func (secretService) load(account string) ([]byte, error) {
	stored, err := keyring.Get(keyringService, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, errKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("oskeystore: system keyring: %w", err)
	}
	return decodeKey(stored)
}

func (secretService) store(account string, key []byte) error {
	if err := keyring.Set(keyringService, account, encodeKey(key)); err != nil {
		return fmt.Errorf("oskeystore: system keyring: %w", err)
	}
	return nil
}

// keyFile keeps the key in a file only the user can read: anyone who can read
// the user's files can read the key. It is the last resort.
type keyFile struct {
	directory string
}

func (keyFile) name() string { return BackendFile }

func (source keyFile) path(account string) string {
	return filepath.Join(source.directory, account+".key")
}

func (source keyFile) load(account string) ([]byte, error) {
	stored, err := os.ReadFile(source.path(account))
	if errors.Is(err, os.ErrNotExist) {
		return nil, errKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("oskeystore: key file: %w", err)
	}
	return decodeKey(string(stored))
}

func (source keyFile) store(account string, key []byte) error {
	if err := os.MkdirAll(source.directory, 0o700); err != nil {
		return fmt.Errorf("oskeystore: key file: %w", err)
	}
	err := os.WriteFile(source.path(account), []byte(encodeKey(key)), 0o600)
	if err != nil {
		return fmt.Errorf("oskeystore: key file: %w", err)
	}
	return nil
}
