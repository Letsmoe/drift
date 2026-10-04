// Package browsercookies reads the YouTube Music session of a browser on this
// machine, like yt-dlp's --cookies-from-browser. It is UI-free and must only
// run on the user's explicit action: it opens browser cookie stores, and for
// Chromium browsers asks the system keyring for the key that protects them.
// Cookie values are never logged.
package browsercookies

import (
	"sort"
	"strings"
	"time"
)

// Cookie is one stored browser cookie.
type Cookie struct {
	Host  string
	Name  string
	Value string
	Path  string
	// Expires is zero for session cookies.
	Expires time.Time
}

// musicHosts are the host values of cookies a browser sends to
// music.youtube.com: its own host and the .youtube.com domain.
var musicHosts = []string{"music.youtube.com", ".music.youtube.com", ".youtube.com"}

// sessionCookieNames identify a signed-in session; either one is enough.
var sessionCookieNames = []string{"SAPISID", "__Secure-3PAPISID"}

func (cookie Cookie) expired(now time.Time) bool {
	return !cookie.Expires.IsZero() && !cookie.Expires.After(now)
}

// sentToMusic reports whether a browser sends the cookie to https://music.youtube.com/.
func (cookie Cookie) sentToMusic() bool {
	hostMatches := false
	for _, host := range musicHosts {
		if cookie.Host == host {
			hostMatches = true
		}
	}
	return hostMatches && (cookie.Path == "/" || cookie.Path == "")
}

// musicCookies keeps the cookies music.youtube.com would receive. When a name
// exists for both the specific and the domain host, the specific one wins, as
// in a browser.
func musicCookies(cookies []Cookie, now time.Time) []Cookie {
	byName := map[string]Cookie{}
	for _, cookie := range cookies {
		if !cookie.sentToMusic() || cookie.expired(now) || cookie.Name == "" {
			continue
		}
		existing, found := byName[cookie.Name]
		if !found || len(cookie.Host) > len(existing.Host) {
			byName[cookie.Name] = cookie
		}
	}
	selected := make([]Cookie, 0, len(byName))
	for _, cookie := range byName {
		selected = append(selected, cookie)
	}
	sort.Slice(selected, func(left, right int) bool { return selected[left].Name < selected[right].Name })
	return selected
}

func hasSession(cookies []Cookie) bool {
	for _, cookie := range cookies {
		for _, name := range sessionCookieNames {
			if cookie.Name == name && cookie.Value != "" {
				return true
			}
		}
	}
	return false
}

// cookieHeader renders cookies as a Cookie request header value.
func cookieHeader(cookies []Cookie) string {
	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(pairs, "; ")
}
