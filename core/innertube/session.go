package innertube

import (
	"crypto/sha1" //nolint:gosec // SAPISIDHASH is defined by Google as SHA-1; it is not used for security here.
	"encoding/hex"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

// musicOrigin is the origin SAPISIDHASH values are bound to.
const musicOrigin = "https://music.youtube.com"

// ErrSessionExpired means YouTube Music treated an authenticated request as
// signed out: the cookies are expired, revoked or incomplete.
var ErrSessionExpired = errors.New("innertube: the YouTube Music session has expired; sign in again")

// ErrNoSessionCookies means a cookie header holds none of the cookies a
// signed-in YouTube Music session needs.
var ErrNoSessionCookies = errors.New("innertube: the cookies hold no YouTube Music session (no SAPISID cookie)")

// Session is a signed-in YouTube Music web session: the Cookie header of
// music.youtube.com. It is a credential, so it never prints itself.
type Session struct {
	CookieHeader string
}

// String hides the cookies from logs and error messages.
func (session Session) String() string { return "innertube.Session(redacted)" }

// GoString hides the cookies from %#v output.
func (session Session) GoString() string { return session.String() }

// LogValue hides the cookies from slog.
func (session Session) LogValue() slog.Value { return slog.StringValue(session.String()) }

// Cookie names that carry the signed-in identity.
const (
	cookieSAPISID          = "SAPISID"
	cookieSecure1PAPISID   = "__Secure-1PAPISID"
	cookieSecure3PAPISID   = "__Secure-3PAPISID"
	hashSchemeSAPISID      = "SAPISIDHASH"
	hashSchemeSAPISID1P    = "SAPISID1PHASH"
	hashSchemeSAPISID3P    = "SAPISID3PHASH"
	cookiePairSeparator    = ";"
	cookieNameValueDivider = "="
)

// cookies returns the cookie header's name to value pairs.
func (session Session) cookies() map[string]string {
	cookies := map[string]string{}
	for _, pair := range strings.Split(session.CookieHeader, cookiePairSeparator) {
		name, value, found := strings.Cut(strings.TrimSpace(pair), cookieNameValueDivider)
		if !found || name == "" {
			continue
		}
		cookies[name] = value
	}
	return cookies
}

// Validate reports ErrNoSessionCookies when the header lacks the cookie the
// authorization header is derived from.
func (session Session) Validate() error {
	cookies := session.cookies()
	if sapisid(cookies) == "" {
		return ErrNoSessionCookies
	}
	return nil
}

// sapisid prefers SAPISID and falls back to the third-party twin, which holds
// the same value.
func sapisid(cookies map[string]string) string {
	if cookies[cookieSAPISID] != "" {
		return cookies[cookieSAPISID]
	}
	return cookies[cookieSecure3PAPISID]
}

// authorization builds the Authorization header: one "<scheme> <time>_<hash>"
// entry per SAPISID-family cookie the session holds.
func (session Session) authorization(now time.Time) string {
	cookies := session.cookies()
	timestamp := now.Unix()
	entries := []string{}
	entries = appendHashEntry(entries, hashSchemeSAPISID, sapisid(cookies), timestamp)
	entries = appendHashEntry(entries, hashSchemeSAPISID1P, cookies[cookieSecure1PAPISID], timestamp)
	entries = appendHashEntry(entries, hashSchemeSAPISID3P, cookies[cookieSecure3PAPISID], timestamp)
	return strings.Join(entries, " ")
}

func appendHashEntry(entries []string, scheme, cookieValue string, timestamp int64) []string {
	if cookieValue == "" {
		return entries
	}
	return append(entries, scheme+" "+sapisidHash(cookieValue, timestamp))
}

// sapisidHash is "<timestamp>_<sha1 of "<timestamp> <cookie> <origin>">".
func sapisidHash(cookieValue string, timestamp int64) string {
	timestampText := strconv.FormatInt(timestamp, 10)
	digest := sha1.Sum([]byte(timestampText + " " + cookieValue + " " + musicOrigin)) //nolint:gosec // see import
	return timestampText + "_" + hex.EncodeToString(digest[:])
}
