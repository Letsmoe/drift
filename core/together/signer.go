package together

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const userIDBitsPerSymbol = 5

var errBadSignature = errors.New("together: message not signed by its sender")

// signer is the user's identity key; its public half determines the user ID.
type signer struct {
	privateKey ed25519.PrivateKey
	userID     string
}

func newSigner() (*signer, error) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("together: identity key: %w", err)
	}
	return signerOf(privateKey), nil
}

// signerFromSeed restores a signer from the hex seed the library stores.
func signerFromSeed(seedHex string) (*signer, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("together: stored identity key is invalid")
	}
	return signerOf(ed25519.NewKeyFromSeed(seed)), nil
}

func signerOf(privateKey ed25519.PrivateKey) *signer {
	publicKey, _ := privateKey.Public().(ed25519.PublicKey)
	return &signer{privateKey: privateKey, userID: userIDOf(publicKey)}
}

func (signer *signer) seedHex() string {
	return hex.EncodeToString(signer.privateKey.Seed())
}

func (signer *signer) publicKey() ed25519.PublicKey {
	publicKey, _ := signer.privateKey.Public().(ed25519.PublicKey)
	return publicKey
}

// userIDOf turns the first 60 bits of the key's hash into 12 ID symbols.
func userIDOf(publicKey ed25519.PublicKey) string {
	hash := sha256.Sum256(publicKey)
	userID := make([]byte, userIDLength)
	for index := range userID {
		userID[index] = idAlphabet[bitsAt(hash[:], index*userIDBitsPerSymbol)]
	}
	return string(userID)
}

// bitsAt reads the 5 bits starting at bit offset (most significant first).
func bitsAt(data []byte, offset int) byte {
	var value byte
	for bit := range userIDBitsPerSymbol {
		position := offset + bit
		set := data[position/8] >> (7 - position%8) & 1
		value = value<<1 | set
	}
	return value
}

// envelope carries a message with the key that signed it. The signature
// covers the topic too, so a message can't be replayed on another topic.
type envelope struct {
	Body      json.RawMessage `json:"body"`
	PublicKey []byte          `json:"publicKey"`
	Signature []byte          `json:"signature"`
}

func signedBytes(topic string, body []byte) []byte {
	return append([]byte(topic+"\n"), body...)
}

// seal signs message for topic and encrypts it with key.
func (signer *signer) seal(key *sealKey, topic string, message any) (string, error) {
	body, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("together: encode message: %w", err)
	}
	signed := envelope{
		Body:      body,
		PublicKey: signer.publicKey(),
		Signature: ed25519.Sign(signer.privateKey, signedBytes(topic, body)),
	}
	plain, err := json.Marshal(signed)
	if err != nil {
		return "", fmt.Errorf("together: encode envelope: %w", err)
	}
	return key.seal(plain)
}

// openSigned decrypts content, checks its signature and decodes the body into
// message. It returns the signer's user ID.
func openSigned(key *sealKey, topic, content string, message any) (string, error) {
	plain, err := key.open(content)
	if err != nil {
		return "", err
	}
	var signed envelope
	if err := json.Unmarshal(plain, &signed); err != nil {
		return "", fmt.Errorf("together: decode envelope: %w", err)
	}
	if len(signed.PublicKey) != ed25519.PublicKeySize ||
		!ed25519.Verify(signed.PublicKey, signedBytes(topic, signed.Body), signed.Signature) {
		return "", errBadSignature
	}
	if err := json.Unmarshal(signed.Body, message); err != nil {
		return "", fmt.Errorf("together: decode message: %w", err)
	}
	return userIDOf(signed.PublicKey), nil
}
