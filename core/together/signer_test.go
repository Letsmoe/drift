package together

import (
	"errors"
	"testing"
)

func TestUserIDComesFromTheKey(t *testing.T) {
	created, err := newSigner()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := signerFromSeed(created.seedHex())
	if err != nil {
		t.Fatal(err)
	}
	if restored.userID != created.userID {
		t.Fatalf("restored ID %s, want %s", restored.userID, created.userID)
	}
	if normalized, err := NormalizeUserID(FormatUserID(created.userID)); err != nil || normalized != created.userID {
		t.Fatalf("ID %q does not normalize to itself: %q, %v", created.userID, normalized, err)
	}
	other, err := newSigner()
	if err != nil {
		t.Fatal(err)
	}
	if other.userID == created.userID {
		t.Fatal("two keys share an ID")
	}
	if _, err := signerFromSeed("not hex"); err == nil {
		t.Fatal("accepted an invalid seed")
	}
}

func TestSignedMessagesNameTheirSigner(t *testing.T) {
	sender, err := newSigner()
	if err != nil {
		t.Fatal(err)
	}
	key, err := deriveRoomKey(sender.userID)
	if err != nil {
		t.Fatal(err)
	}
	content, err := sender.seal(key, "topic-a", map[string]string{"hello": "world"})
	if err != nil {
		t.Fatal(err)
	}

	var message map[string]string
	signerUserID, err := openSigned(key, "topic-a", content, &message)
	if err != nil || signerUserID != sender.userID || message["hello"] != "world" {
		t.Fatalf("openSigned = %q, %v, %v", signerUserID, message, err)
	}
	if _, err := openSigned(key, "topic-b", content, &message); !errors.Is(err, errBadSignature) {
		t.Fatalf("replay on another topic: %v", err)
	}
}
