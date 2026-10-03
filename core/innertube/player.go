package innertube

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const playabilityOK = "OK"

// Format is one downloadable stream of a video. Formats whose URL is ciphered
// (signatureCipher) are left out, since this client does not solve ciphers.
type Format struct {
	Itag          int
	MimeType      string
	Bitrate       int
	ContentLength int64
	URL           string
}

// IsAudio reports whether the format carries audio only.
func (format Format) IsAudio() bool {
	return strings.HasPrefix(format.MimeType, "audio/")
}

// Player is the playable part of a player response.
type Player struct {
	Formats   []Format
	ExpiresIn time.Duration
}

// PlayabilityError is returned when YouTube refuses to play a video for a profile,
// e.g. "LOGIN_REQUIRED: Sign in to confirm you're not a bot".
type PlayabilityError struct {
	Profile string
	Status  string
	Reason  string
}

func (playabilityError *PlayabilityError) Error() string {
	return fmt.Sprintf("innertube player (%s): %s: %s", playabilityError.Profile, playabilityError.Status, playabilityError.Reason)
}

type playerRequest struct {
	Context        requestContext `json:"context"`
	VideoID        string         `json:"videoId"`
	ContentCheckOK bool           `json:"contentCheckOk"`
	RacyCheckOK    bool           `json:"racyCheckOk"`
}

type playerResponse struct {
	PlayabilityStatus struct {
		Status string
		Reason string
	}
	StreamingData struct {
		ExpiresInSeconds string
		AdaptiveFormats  []adaptiveFormat
	}
}

type adaptiveFormat struct {
	Itag          int
	MimeType      string
	Bitrate       int
	ContentLength string
	URL           string
}

// Player requests the streams of videoID as profile.
func (client *Client) Player(ctx context.Context, profile Profile, videoID string) (Player, error) {
	call, err := client.visitorCall(ctx, profile, "player", func(context requestContext) any {
		return playerRequest{Context: context, VideoID: videoID, ContentCheckOK: true, RacyCheckOK: true}
	})
	if err != nil {
		return Player{}, err
	}

	var response playerResponse
	if err := client.post(ctx, call, &response); err != nil {
		return Player{}, err
	}
	return parsePlayer(profile, response)
}

func parsePlayer(profile Profile, response playerResponse) (Player, error) {
	status := response.PlayabilityStatus
	if status.Status != playabilityOK {
		return Player{}, &PlayabilityError{Profile: profile.Name, Status: status.Status, Reason: status.Reason}
	}

	expiresInSeconds, _ := strconv.Atoi(response.StreamingData.ExpiresInSeconds)
	return Player{
		Formats:   parseFormats(response.StreamingData.AdaptiveFormats),
		ExpiresIn: time.Duration(expiresInSeconds) * time.Second,
	}, nil
}

func parseFormats(adaptiveFormats []adaptiveFormat) []Format {
	formats := []Format{}
	for _, adaptive := range adaptiveFormats {
		if adaptive.URL == "" {
			continue
		}
		contentLength, _ := strconv.ParseInt(adaptive.ContentLength, 10, 64)
		formats = append(formats, Format{
			Itag:          adaptive.Itag,
			MimeType:      adaptive.MimeType,
			Bitrate:       adaptive.Bitrate,
			ContentLength: contentLength,
			URL:           adaptive.URL,
		})
	}
	return formats
}
