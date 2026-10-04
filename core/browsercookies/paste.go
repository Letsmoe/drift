package browsercookies

import (
	"errors"
	"regexp"
	"strings"
)

// ErrNoCookies means pasted text holds no Cookie header.
var ErrNoCookies = errors.New("browsercookies: no cookies found in the pasted text")

// ErrNoSession means the cookies hold no signed-in YouTube session.
var ErrNoSession = errors.New("browsercookies: the cookies hold no YouTube session (no SAPISID cookie); copy them from a signed-in music.youtube.com tab")

var (
	// "Cookie: a=b" as a header line, also inside -H '...' of a cURL command.
	headerLinePattern = regexp.MustCompile(`(?im)(?:^|[\s'"])cookie:[ \t]*([^'"\r\n]+)`)
	// curl -b 'a=b' / --cookie 'a=b'
	curlCookiePattern = regexp.MustCompile(`(?:^|\s)(?:-b|--cookie)\s+(?:'([^']*)'|"([^"]*)")`)
)

// ParseCookieHeader extracts a Cookie header value from pasted text: the bare
// header value, a "Cookie: ..." line, a block of request headers or a "Copy as
// cURL" command. It fails when the result holds no YouTube session cookie.
func ParseCookieHeader(pasted string) (string, error) {
	header := extractCookieHeader(strings.TrimSpace(pasted))
	if header == "" {
		return "", ErrNoCookies
	}
	if !headerHasSession(header) {
		return "", ErrNoSession
	}
	return header, nil
}

func extractCookieHeader(text string) string {
	if text == "" {
		return ""
	}
	match := curlCookiePattern.FindStringSubmatch(text)
	if match != nil {
		return strings.TrimSpace(match[1] + match[2])
	}
	match = headerLinePattern.FindStringSubmatch(text)
	if match != nil {
		return strings.TrimSpace(match[1])
	}
	if looksLikeBareHeader(text) {
		return text
	}
	return cookieAfterNameLine(text)
}

// looksLikeBareHeader accepts a single line of name=value pairs.
func looksLikeBareHeader(text string) bool {
	return !strings.ContainsAny(text, "\r\n") && strings.Contains(text, "=") && !strings.Contains(text, " -H ")
}

// cookieAfterNameLine handles DevTools tables copied as "cookie" on one line
// and the value on the next.
func cookieAfterNameLine(text string) string {
	lines := strings.Split(text, "\n")
	for index := 0; index+1 < len(lines); index++ {
		if strings.EqualFold(strings.TrimSpace(lines[index]), "cookie") {
			return strings.TrimSpace(lines[index+1])
		}
	}
	return ""
}

func headerHasSession(header string) bool {
	for _, pair := range strings.Split(header, ";") {
		name, value, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found || value == "" {
			continue
		}
		for _, sessionName := range sessionCookieNames {
			if name == sessionName {
				return true
			}
		}
	}
	return false
}
