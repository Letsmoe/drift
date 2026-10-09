package together

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

// 32 symbols without the look-alikes I, O, 0 and 1; a byte modulo 32 is unbiased.
const idAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// 12 symbols are 60 bits of the public key hash: too many to guess through relay
// traffic or to find a second key for, few enough to read out.
const userIDLength = 12

const userIDGroupLength = 4

// Key derivation salts; changing one makes old and new versions unable to meet.
// They differ from otter's, so the two apps never answer each other.
const (
	roomSalt     = "drift-room-v1"
	inboxSalt    = "drift-inbox-v1"
	presenceSalt = "drift-presence-v1"
)

// ErrInvalidUserID is returned for text that is not a user ID.
var ErrInvalidUserID = errors.New("that is not a valid Listen Together ID")

// NormalizeUserID accepts what people type or paste ("k7qd-9mza-x2pt", "K7QD 9MZA X2PT").
func NormalizeUserID(input string) (string, error) {
	var userID strings.Builder
	for _, character := range strings.ToUpper(strings.TrimSpace(input)) {
		if character == '-' || character == ' ' {
			continue
		}
		if !strings.ContainsRune(idAlphabet, character) {
			return "", ErrInvalidUserID
		}
		userID.WriteRune(character)
	}
	if userID.Len() != userIDLength {
		return "", ErrInvalidUserID
	}
	return userID.String(), nil
}

// FormatUserID groups an ID for reading: "K7QD-9MZA-X2PT".
func FormatUserID(userID string) string {
	var groups []string
	for start := 0; start < len(userID); start += userIDGroupLength {
		end := min(start+userIDGroupLength, len(userID))
		groups = append(groups, userID[start:end])
	}
	return strings.Join(groups, "-")
}

// sealKey encrypts signaling for one room or inbox. Relays only see the topic
// and ciphertext; the topic is derived from the key, not from the ID directly.
type sealKey struct {
	aead  cipher.AEAD
	topic string
}

func deriveRoomKey(hostUserID string) (*sealKey, error) {
	return deriveKey(hostUserID, roomSalt)
}

func deriveInboxKey(userID string) (*sealKey, error) {
	return deriveKey(userID, inboxSalt)
}

// derivePresenceKey seals what a user is playing; everyone who knows the ID can read it.
func derivePresenceKey(userID string) (*sealKey, error) {
	return deriveKey(userID, presenceSalt)
}

func deriveKey(userID, salt string) (*sealKey, error) {
	// scrypt makes guessing IDs from relay traffic expensive.
	secret, err := scrypt.Key([]byte(userID), []byte(salt), 1<<15, 8, 1, 64)
	if err != nil {
		return nil, fmt.Errorf("together: derive key: %w", err)
	}
	block, err := aes.NewCipher(secret[:32])
	if err != nil {
		return nil, fmt.Errorf("together: cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("together: cipher: %w", err)
	}
	topicHash := sha256.Sum256(secret[32:])
	return &sealKey{aead: aead, topic: hex.EncodeToString(topicHash[:16])}, nil
}

func (key *sealKey) seal(plain []byte) (string, error) {
	nonce := make([]byte, key.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("together: nonce: %w", err)
	}
	sealed := key.aead.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (key *sealKey) open(content string) ([]byte, error) {
	sealed, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("together: sealed message: %w", err)
	}
	nonceSize := key.aead.NonceSize()
	if len(sealed) < nonceSize {
		return nil, errors.New("together: sealed message too short")
	}
	plain, err := key.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("together: open sealed message: %w", err)
	}
	return plain, nil
}
