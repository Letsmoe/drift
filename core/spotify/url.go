package spotify

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Kinds of public Spotify page that can be imported.
const (
	KindPlaylist = "playlist"
	KindAlbum    = "album"
)

// ErrUnsupportedURL means a link is not a Spotify playlist or album.
var ErrUnsupportedURL = errors.New("spotify: not a playlist or album link")

// Reference points at one public playlist or album.
type Reference struct {
	Kind string
	ID   string
}

// URI renders the reference as a spotify: URI.
func (reference Reference) URI() string {
	return "spotify:" + reference.Kind + ":" + reference.ID
}

// ParseURL understands open.spotify.com links (also /embed/ and /intl-xx/
// ones) and spotify: URIs of playlists and albums.
func ParseURL(text string) (Reference, error) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "spotify:") {
		return parseSegments(strings.Split(text, ":")[1:])
	}
	parsed, err := url.Parse(text)
	if err != nil || parsed.Hostname() != "open.spotify.com" {
		return Reference{}, fmt.Errorf("%w: %q", ErrUnsupportedURL, text)
	}
	return parseSegments(strings.Split(strings.Trim(parsed.Path, "/"), "/"))
}

// parseSegments finds "<kind>, <id>" in segments such as
// [intl-de, playlist, ID] or [user, someone, playlist, ID].
func parseSegments(segments []string) (Reference, error) {
	for index := 0; index+1 < len(segments); index++ {
		kind := segments[index]
		if (kind == KindPlaylist || kind == KindAlbum) && isSpotifyID(segments[index+1]) {
			return Reference{Kind: kind, ID: segments[index+1]}, nil
		}
	}
	return Reference{}, fmt.Errorf("%w: %q", ErrUnsupportedURL, strings.Join(segments, "/"))
}

func isSpotifyID(text string) bool {
	if text == "" {
		return false
	}
	for _, character := range text {
		isLetter := character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
		isDigit := character >= '0' && character <= '9'
		if !isLetter && !isDigit {
			return false
		}
	}
	return true
}
