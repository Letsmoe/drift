package oskeystore

import (
	"bytes"
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeystoreSealsUnderAKeyKeptInTheKeyring(t *testing.T) {
	keyring.MockInit()
	keystore, err := Open("test")
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := keystore.Wrap([]byte("secret"), []byte("label"))
	if err != nil {
		t.Fatal(err)
	}

	reopened, err := Open("test")
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

	other, err := Open("other")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Unwrap(wrapped, []byte("label")); !errors.Is(err, ErrWrapped) {
		t.Fatalf("another account's key: got %v, want %v", err, ErrWrapped)
	}
}

func TestDeleteForgetsTheKey(t *testing.T) {
	keyring.MockInit()
	keystore, err := Open("test")
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := keystore.Wrap([]byte("secret"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := keystore.Delete(); err != nil {
		t.Fatal(err)
	}
	fresh, err := Open("test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.Unwrap(wrapped, nil); !errors.Is(err, ErrWrapped) {
		t.Fatalf("after Delete: got %v, want %v", err, ErrWrapped)
	}
}
