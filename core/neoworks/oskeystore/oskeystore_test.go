package oskeystore

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

var errNoKeyring = errors.New("the name is not activatable")

func mustOpen(t *testing.T, sources ...keySource) *Keystore {
	t.Helper()
	keystore, err := openFrom("test", sources)
	if err != nil {
		t.Fatal(err)
	}
	return keystore
}

func TestKeystoreSealsUnderAKeyItKeeps(t *testing.T) {
	keyring.MockInit()
	directory := t.TempDir()
	keystore, err := Open("test", directory)
	if err != nil {
		t.Fatal(err)
	}
	if keystore.Backend() != BackendSystem {
		t.Fatalf("backend %s, want %s", keystore.Backend(), BackendSystem)
	}
	wrapped, err := keystore.Wrap([]byte("secret"), []byte("label"))
	if err != nil {
		t.Fatal(err)
	}

	reopened, err := Open("test", directory)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := reopened.Unwrap(wrapped, []byte("label"))
	if err != nil || !bytes.Equal(secret, []byte("secret")) {
		t.Fatalf("reopened keystore unwrapped %q %v", secret, err)
	}
	if _, err := reopened.Unwrap(wrapped, []byte("other label")); !errors.Is(err, ErrWrapped) {
		t.Fatalf("another label: got %v, want %v", err, ErrWrapped)
	}
	other, err := Open("other", directory)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Unwrap(wrapped, []byte("label")); !errors.Is(err, ErrWrapped) {
		t.Fatalf("another account's key: got %v, want %v", err, ErrWrapped)
	}
}

func TestWithoutAKeyringTheKeyGoesToAPrivateFile(t *testing.T) {
	keyring.MockInitWithError(errNoKeyring)
	directory := t.TempDir()
	keystore := mustOpen(t, secretService{}, keyFile{directory: directory})
	if keystore.Backend() != BackendFile {
		t.Fatalf("backend %s, want %s", keystore.Backend(), BackendFile)
	}
	info, err := os.Stat(filepath.Join(directory, "test.key"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("key file %v %v", info, err)
	}
}

// A keyring that shows up after the key went to the file must not replace it,
// or everything sealed so far would be lost.
func TestAnExistingKeyWinsOverAKeyringThatAppearsLater(t *testing.T) {
	keyring.MockInitWithError(errNoKeyring)
	directory := t.TempDir()
	wrapped, err := mustOpen(t, secretService{}, keyFile{directory: directory}).Wrap([]byte("secret"), nil)
	if err != nil {
		t.Fatal(err)
	}

	keyring.MockInit()
	reopened := mustOpen(t, secretService{}, keyFile{directory: directory})
	if reopened.Backend() != BackendFile {
		t.Fatalf("backend %s, want the file that holds the key", reopened.Backend())
	}
	if _, err := reopened.Unwrap(wrapped, nil); err != nil {
		t.Fatalf("the earlier key was not kept: %v", err)
	}
}

func TestOpenFailsWhenNothingKeepsTheKey(t *testing.T) {
	keyring.MockInitWithError(errNoKeyring)
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := openFrom("test", []keySource{secretService{}, keyFile{directory: filepath.Join(blocker, "keys")}}); err == nil {
		t.Fatal("open must fail when no source can keep the key")
	}
}
