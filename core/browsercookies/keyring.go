package browsercookies

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

const keyringCommandTimeout = 30 * time.Second

// systemKeyringPassword reads the "<name> Safe Storage" password of a
// Chromium browser: from the macOS Keychain, or on Linux from the Secret
// Service (GNOME Keyring, KeePassXC, ...) with KWallet as fallback.
func systemKeyringPassword(spec browserSpec) ([]byte, error) {
	switch runtime.GOOS {
	case "darwin":
		return macKeychainPassword(spec)
	case "linux":
		return linuxKeyringPassword(spec)
	default:
		return nil, errors.New("no keyring support on this OS")
	}
}

func linuxKeyringPassword(spec browserSpec) ([]byte, error) {
	label := spec.keyringName + " Safe Storage"
	password, secretErr := secretServicePassword(label)
	if secretErr == nil {
		return password, nil
	}
	password, walletErr := kwalletPassword(spec.keyringName)
	if walletErr == nil {
		return password, nil
	}
	return nil, fmt.Errorf("secret service: %w; kwallet: %w", secretErr, walletErr)
}

func macKeychainPassword(spec browserSpec) ([]byte, error) {
	label := spec.keyringName + " Safe Storage"
	return runKeyringCommand("security", "find-generic-password", "-w", "-a", spec.keyringName, "-s", label)
}

// kwalletPassword uses kwallet-query, as yt-dlp does, on the default wallet.
func kwalletPassword(keyringName string) ([]byte, error) {
	password, err := runKeyringCommand("kwallet-query", "--read-password", keyringName+" Safe Storage",
		"--folder", keyringName+" Keys", "kdewallet")
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(string(password), "Failed to read") {
		return nil, errors.New("entry not found in KWallet")
	}
	return password, nil
}

func runKeyringCommand(name string, arguments ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), keyringCommandTimeout)
	defer cancel()
	var stdout bytes.Buffer
	command := exec.CommandContext(ctx, name, arguments...)
	command.Stdout = &stdout
	err := command.Run()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	password := bytes.TrimRight(stdout.Bytes(), "\r\n")
	if len(password) == 0 {
		return nil, fmt.Errorf("%s returned no password", name)
	}
	return password, nil
}

// memoizePassword asks the keyring once per browser, also when it fails, so a
// browser with several profiles prompts at most once.
func memoizePassword(source keyringPassword) keyringPassword {
	type answer struct {
		password []byte
		err      error
	}
	var mutex sync.Mutex
	answers := map[string]answer{}
	return func(spec browserSpec) ([]byte, error) {
		mutex.Lock()
		defer mutex.Unlock()
		cached, found := answers[spec.keyringName]
		if !found {
			cached.password, cached.err = source(spec)
			answers[spec.keyringName] = cached
		}
		return cached.password, cached.err
	}
}
