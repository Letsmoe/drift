package browsercookies

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1" //nolint:gosec // Chromium derives its cookie key with PBKDF2-HMAC-SHA1.
	"crypto/sha256"
	"errors"
	"fmt"
)

const (
	linuxPlainPassword   = "peanuts"
	keySalt              = "saltysalt"
	linuxKeyIterations   = 1
	macKeyIterations     = 1003
	chromiumKeyLength    = 16
	encryptionPrefixSize = 3
	hostHashSize         = sha256.Size
)

var errUnsupportedEncryption = errors.New("unsupported cookie encryption")

// deriveChromiumKey is Chromium's PBKDF2-HMAC-SHA1 over the keyring password.
func deriveChromiumKey(password []byte, iterations int) ([]byte, error) {
	key, err := pbkdf2.Key(sha1.New, string(password), []byte(keySalt), iterations, chromiumKeyLength) //nolint:gosec // see import
	if err != nil {
		return nil, fmt.Errorf("derive cookie key: %w", err)
	}
	return key, nil
}

// decryptChromiumValue decrypts a v10/v11 value: AES-128-CBC with an IV of 16
// spaces and PKCS#7 padding. Newer Chromium prepends SHA-256(host) to the
// plaintext, which is dropped when it matches.
func decryptChromiumValue(key, encrypted []byte, hostKey string) (string, error) {
	if len(encrypted) <= encryptionPrefixSize {
		return "", errUnsupportedEncryption
	}
	ciphertext := encrypted[encryptionPrefixSize:]
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("cookie ciphertext has an invalid length")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("cookie cipher: %w", err)
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, bytes.Repeat([]byte{' '}, aes.BlockSize)).CryptBlocks(plaintext, ciphertext)
	plaintext, err = removePadding(plaintext)
	if err != nil {
		return "", err
	}
	return string(dropHostHash(plaintext, hostKey)), nil
}

func removePadding(plaintext []byte) ([]byte, error) {
	padding := int(plaintext[len(plaintext)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(plaintext) {
		return nil, errors.New("cookie decryption failed: wrong key")
	}
	return plaintext[:len(plaintext)-padding], nil
}

func dropHostHash(plaintext []byte, hostKey string) []byte {
	if len(plaintext) < hostHashSize {
		return plaintext
	}
	hash := sha256.Sum256([]byte(hostKey))
	if bytes.Equal(plaintext[:hostHashSize], hash[:]) {
		return plaintext[hostHashSize:]
	}
	return plaintext
}

// encryptionVersion returns "v10" or "v11" for an encrypted value.
func encryptionVersion(encrypted []byte) string {
	if len(encrypted) < encryptionPrefixSize {
		return ""
	}
	return string(encrypted[:encryptionPrefixSize])
}
