package browsercookies

import (
	"fmt"
	"sync"
)

// keyringPassword reads the password a Chromium browser protects its cookie
// key with, from the system keyring.
type keyringPassword func(spec browserSpec) ([]byte, error)

// platformDecryptor decrypts the v10/v11 values of one browser on one OS.
// Keyring keys are requested lazily and at most once.
type platformDecryptor struct {
	goos     string
	spec     browserSpec
	password keyringPassword

	keyringOnce sync.Once
	keyringKey  []byte
	keyringErr  error
}

func newDecryptor(goos string, spec browserSpec, password keyringPassword) *platformDecryptor {
	return &platformDecryptor{goos: goos, spec: spec, password: password}
}

func (decryptor *platformDecryptor) decrypt(hostKey string, encrypted []byte) (string, error) {
	key, err := decryptor.keyFor(encryptionVersion(encrypted))
	if err != nil {
		return "", err
	}
	return decryptChromiumValue(key, encrypted, hostKey)
}

func (decryptor *platformDecryptor) keyFor(version string) ([]byte, error) {
	switch {
	case decryptor.goos == "linux" && version == "v10":
		return deriveChromiumKey([]byte(linuxPlainPassword), linuxKeyIterations)
	case decryptor.goos == "linux" && version == "v11":
		return decryptor.keyringKeyWith(linuxKeyIterations)
	case decryptor.goos == "darwin" && version == "v10":
		return decryptor.keyringKeyWith(macKeyIterations)
	default:
		return nil, fmt.Errorf("%w (%q on %s)", errUnsupportedEncryption, version, decryptor.goos)
	}
}

func (decryptor *platformDecryptor) keyringKeyWith(iterations int) ([]byte, error) {
	decryptor.keyringOnce.Do(func() {
		password, err := decryptor.password(decryptor.spec)
		if err != nil {
			decryptor.keyringErr = fmt.Errorf("read the %s Safe Storage password: %w", decryptor.spec.keyringName, err)
			return
		}
		decryptor.keyringKey, decryptor.keyringErr = deriveChromiumKey(password, iterations)
	})
	return decryptor.keyringKey, decryptor.keyringErr
}
