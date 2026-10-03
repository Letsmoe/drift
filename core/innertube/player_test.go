package innertube

import (
	"errors"
	"testing"
	"time"
)

func TestPlayer(t *testing.T) {
	server := newFakeServer(t)
	server.respondWithFixture(t, "player", "testdata/player_visionos.json")

	player, err := server.client().Player(t.Context(), VisionOS, "UM6Cts99XdA")
	if err != nil {
		t.Fatal(err)
	}

	request := server.lastRequest(t, "player")
	if request.body["videoId"] != "UM6Cts99XdA" {
		t.Errorf("videoId = %v", request.body["videoId"])
	}
	if request.contextClient()["clientName"] != VisionOS.Name {
		t.Errorf("clientName = %v", request.contextClient()["clientName"])
	}
	if request.headers.Get("X-YouTube-Client-Name") != "101" {
		t.Errorf("X-YouTube-Client-Name = %q", request.headers.Get("X-YouTube-Client-Name"))
	}

	if player.ExpiresIn != 21540*time.Second {
		t.Errorf("ExpiresIn = %v", player.ExpiresIn)
	}
	// The ciphered itag 250 is dropped; video itag 137 stays (callers filter by IsAudio).
	itags := []int{}
	for _, format := range player.Formats {
		itags = append(itags, format.Itag)
	}
	if len(itags) != 3 || itags[0] != 137 || itags[1] != 140 || itags[2] != 251 {
		t.Errorf("itags = %v, want [137 140 251]", itags)
	}
	if !player.Formats[2].IsAudio() || player.Formats[2].ContentLength != 2558160 {
		t.Errorf("format 251 = %+v", player.Formats[2])
	}
}

func TestPlayerPlayabilityError(t *testing.T) {
	server := newFakeServer(t)
	server.responses["player"] = []byte(`{"playabilityStatus":{"status":"LOGIN_REQUIRED","reason":"Sign in to confirm you're not a bot"}}`)

	_, err := server.client().Player(t.Context(), VisionOS, "UM6Cts99XdA")

	var playabilityError *PlayabilityError
	if !errors.As(err, &playabilityError) || playabilityError.Status != "LOGIN_REQUIRED" {
		t.Errorf("err = %v, want PlayabilityError LOGIN_REQUIRED", err)
	}
}
