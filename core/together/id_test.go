package together

import (
	"errors"
	"testing"
)

func TestNormalizeUserIDAcceptsTypedForms(t *testing.T) {
	for _, input := range []string{"k7qd-9mza-x2pt", " K7QD 9MZA X2PT ", "K7QD9MZAX2PT"} {
		normalized, err := NormalizeUserID(input)
		if err != nil || normalized != "K7QD9MZAX2PT" {
			t.Errorf("NormalizeUserID(%q) = %q, %v", input, normalized, err)
		}
	}
	for _, input := range []string{"", "K7QD9MZA", "K7QD9MZAX2P0", "K7QD9MZAX2PTX"} {
		if _, err := NormalizeUserID(input); !errors.Is(err, ErrInvalidUserID) {
			t.Errorf("NormalizeUserID(%q) error = %v", input, err)
		}
	}
}

func TestFormatUserIDGroupsByFour(t *testing.T) {
	if formatted := FormatUserID("K7QD9MZAX2PT"); formatted != "K7QD-9MZA-X2PT" {
		t.Fatalf("FormatUserID = %q", formatted)
	}
}

func TestRoomAndInboxKeysDiffer(t *testing.T) {
	roomKey, err := deriveRoomKey("K7QD9MZAX2PT")
	if err != nil {
		t.Fatal(err)
	}
	inboxKey, err := deriveInboxKey("K7QD9MZAX2PT")
	if err != nil {
		t.Fatal(err)
	}
	if roomKey.topic == inboxKey.topic {
		t.Fatal("room and inbox share a topic")
	}
	sealed, err := roomKey.seal([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inboxKey.open(sealed); err == nil {
		t.Fatal("inbox key opened a room message")
	}
	plain, err := roomKey.open(sealed)
	if err != nil || string(plain) != "hello" {
		t.Fatalf("open = %q, %v", plain, err)
	}
}
