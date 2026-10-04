package browsercookies

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The cookie databases below are synthesized in temp directories; no real
// browser profile is ever read by these tests.

func createDatabase(t *testing.T, path, schema string, inserts ...[]any) {
	t.Helper()
	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.ExecContext(context.Background(), schema)
	if err != nil {
		t.Fatal(err)
	}
	for _, insert := range inserts {
		_, err = db.ExecContext(context.Background(), insert[0].(string), insert[1:]...)
		if err != nil {
			t.Fatal(err)
		}
	}
}

const (
	firefoxSchema = `CREATE TABLE moz_cookies (host TEXT, name TEXT, value TEXT, path TEXT, expiry INTEGER, isSecure INTEGER, originAttributes TEXT NOT NULL DEFAULT '')`
	firefoxInsert = `INSERT INTO moz_cookies (host, name, value, path, expiry, isSecure, originAttributes) VALUES (?, ?, ?, ?, ?, 1, ?)`
)

func firefoxRow(host, name, value string, expiry int64, origin string) []any {
	return []any{firefoxInsert, host, name, value, "/", expiry, origin}
}

func newEnvironment(t *testing.T) environment {
	t.Helper()
	return environment{goos: "linux", homeDir: t.TempDir(), tempDir: t.TempDir(), password: func(browserSpec) ([]byte, error) {
		return nil, errors.New("no keyring in tests")
	}}
}

func TestFirefoxSession(t *testing.T) {
	env := newEnvironment(t)
	future := time.Now().Add(time.Hour).Unix()
	createDatabase(t, filepath.Join(env.homeDir, ".mozilla/firefox/abc.default/cookies.sqlite"), firefoxSchema,
		firefoxRow(".youtube.com", "SAPISID", "sapisid-value", future, ""),
		firefoxRow(".youtube.com", "__Secure-3PAPISID", "third", future*1000, ""),
		firefoxRow("music.youtube.com", "YSC", "specific", future, ""),
		firefoxRow(".youtube.com", "YSC", "domain", future, ""),
		firefoxRow(".youtube.com", "OLD", "expired", time.Now().Add(-time.Hour).Unix(), ""),
		firefoxRow(".google.com", "SID", "google-only", future, ""),
		firefoxRow(".youtube.com", "CONTAINER", "other-container", future, "^userContextId=2"),
	)

	header, err := env.load("Firefox", "abc.default")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	want := "SAPISID=sapisid-value; YSC=specific; __Secure-3PAPISID=third"
	if header != want {
		t.Errorf("header = %q, want %q", header, want)
	}
	for _, leftover := range mustGlob(t, filepath.Join(env.tempDir, "otter-cookies-*")) {
		t.Errorf("temp copy %s was not deleted", leftover)
	}
}

func mustGlob(t *testing.T, pattern string) []string {
	t.Helper()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

func TestDetectFirefoxWithoutSession(t *testing.T) {
	env := newEnvironment(t)
	future := time.Now().Add(time.Hour).Unix()
	createDatabase(t, filepath.Join(env.homeDir, ".mozilla/firefox/signed-in/cookies.sqlite"), firefoxSchema,
		firefoxRow(".youtube.com", "SAPISID", "value", future, ""))
	createDatabase(t, filepath.Join(env.homeDir, ".mozilla/firefox/signed-out/cookies.sqlite"), firefoxSchema,
		firefoxRow(".youtube.com", "YSC", "value", future, ""))

	sessions := env.detect()

	if len(sessions) != 2 {
		t.Fatalf("sessions = %+v", sessions)
	}
	if sessions[0].Profile != "signed-in" || !sessions[0].HasYouTubeSession {
		t.Errorf("first = %+v", sessions[0])
	}
	if sessions[1].HasYouTubeSession || sessions[1].Reason == "" {
		t.Errorf("second = %+v", sessions[1])
	}
}

func TestLoadSessionUnknownProfile(t *testing.T) {
	_, err := newEnvironment(t).load("Firefox", "missing")
	if !errors.Is(err, ErrProfileNotFound) {
		t.Errorf("err = %v", err)
	}
}

const (
	chromiumSchema = `CREATE TABLE cookies (host_key TEXT, name TEXT, value TEXT, encrypted_value BLOB, path TEXT, expires_utc INTEGER)`
	chromiumInsert = `INSERT INTO cookies (host_key, name, value, encrypted_value, path, expires_utc) VALUES (?, ?, ?, ?, ?, ?)`
)

func webkitMicros(moment time.Time) int64 {
	return moment.UnixMicro() + webkitEpochOffsetMicros
}

// encryptValue encrypts like Chromium on Linux: v10/v11 prefix, AES-128-CBC, IV of spaces.
func encryptValue(t *testing.T, prefix, password, value string, withHostHash bool) []byte {
	hostKey := ".youtube.com"
	t.Helper()
	key, err := deriveChromiumKey([]byte(password), linuxKeyIterations)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte(value)
	if withHostHash {
		hash := sha256.Sum256([]byte(hostKey))
		plaintext = append(hash[:], plaintext...)
	}
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, bytes.Repeat([]byte{' '}, aes.BlockSize)).CryptBlocks(ciphertext, plaintext)
	return append([]byte(prefix), ciphertext...)
}

func chromiumRow(host, name string, plain string, encrypted []byte, expires int64) []any {
	return []any{chromiumInsert, host, name, plain, encrypted, "/", expires}
}

func TestChromiumV10AndPlainValues(t *testing.T) {
	env := newEnvironment(t)
	future := webkitMicros(time.Now().Add(time.Hour))
	createDatabase(t, filepath.Join(env.homeDir, ".config/google-chrome/Default/Network/Cookies"), chromiumSchema,
		chromiumRow(".youtube.com", "SAPISID", "", encryptValue(t, "v10", "peanuts", "chrome-sapisid", false), future),
		chromiumRow(".youtube.com", "__Secure-3PAPISID", "", encryptValue(t, "v10", "peanuts", "hashed-third", true), future),
		chromiumRow(".youtube.com", "PLAIN", "plain-value", nil, future),
		chromiumRow(".youtube.com", "SESSION", "session-cookie", nil, 0),
		chromiumRow(".youtube.com", "STALE", "stale", nil, webkitMicros(time.Now().Add(-time.Hour))),
	)

	header, err := env.load("Chrome", "Default")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	want := "PLAIN=plain-value; SAPISID=chrome-sapisid; SESSION=session-cookie; __Secure-3PAPISID=hashed-third"
	if header != want {
		t.Errorf("header = %q, want %q", header, want)
	}
}

func TestChromiumV11UsesKeyringPassword(t *testing.T) {
	env := newEnvironment(t)
	asked := 0
	env.password = func(spec browserSpec) ([]byte, error) {
		asked++
		if spec.keyringName != "Chromium" {
			t.Errorf("keyring name = %q", spec.keyringName)
		}
		return []byte("keyring-password"), nil
	}
	future := webkitMicros(time.Now().Add(time.Hour))
	createDatabase(t, filepath.Join(env.homeDir, ".config/chromium/Profile 1/Cookies"), chromiumSchema,
		chromiumRow(".youtube.com", "SAPISID", "", encryptValue(t, "v11", "keyring-password", "v11-value", false), future),
		chromiumRow(".youtube.com", "SID", "", encryptValue(t, "v11", "keyring-password", "second", false), future),
	)

	header, err := env.load("Chromium", "Profile 1")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if header != "SAPISID=v11-value; SID=second" || asked != 1 {
		t.Errorf("header = %q, keyring asked %d times", header, asked)
	}
}

func TestChromiumLockedKeyringIsReportedPerProfile(t *testing.T) {
	env := newEnvironment(t)
	future := webkitMicros(time.Now().Add(time.Hour))
	createDatabase(t, filepath.Join(env.homeDir, ".config/google-chrome/Default/Cookies"), chromiumSchema,
		chromiumRow(".youtube.com", "SAPISID", "", encryptValue(t, "v11", "x", "value", false), future))
	createDatabase(t, filepath.Join(env.homeDir, ".mozilla/firefox/ff/cookies.sqlite"), firefoxSchema,
		firefoxRow(".youtube.com", "SAPISID", "value", time.Now().Add(time.Hour).Unix(), ""))

	sessions := env.detect()

	if len(sessions) != 2 {
		t.Fatalf("sessions = %+v", sessions)
	}
	for _, session := range sessions {
		switch session.Browser {
		case "Chrome":
			if session.HasYouTubeSession || !strings.Contains(session.Reason, "Safe Storage") {
				t.Errorf("Chrome = %+v", session)
			}
		case "Firefox":
			if !session.HasYouTubeSession {
				t.Errorf("Firefox = %+v", session)
			}
		}
	}
}

func TestChromiumWrongKeyIsAnError(t *testing.T) {
	env := newEnvironment(t)
	future := webkitMicros(time.Now().Add(time.Hour))
	createDatabase(t, filepath.Join(env.homeDir, ".config/google-chrome/Default/Cookies"), chromiumSchema,
		chromiumRow(".youtube.com", "SAPISID", "", encryptValue(t, "v10", "not-peanuts", "value", false), future))

	_, err := env.load("Chrome", "Default")
	if err == nil {
		t.Error("a value encrypted with another key must not load")
	}
}

func TestWindowsChromiumIsSkipped(t *testing.T) {
	env := newEnvironment(t)
	env.goos = "windows"
	env.localAppDataDir = t.TempDir()
	createDatabase(t, filepath.Join(env.localAppDataDir, `Google/Chrome/User Data/Default/Network/Cookies`), chromiumSchema)

	sessions := env.detect()

	if len(sessions) != 1 || sessions[0].HasYouTubeSession || !strings.Contains(sessions[0].Reason, "app-bound") {
		t.Errorf("sessions = %+v", sessions)
	}
}

func TestDecryptPlaintextShorterThanHostHash(t *testing.T) {
	key, _ := deriveChromiumKey([]byte(linuxPlainPassword), linuxKeyIterations)
	value, err := decryptChromiumValue(key, encryptValue(t, "v10", linuxPlainPassword, "short", false), ".youtube.com")
	if err != nil || value != "short" {
		t.Errorf("value = %q, err = %v", value, err)
	}
}

func TestParseCookieHeader(t *testing.T) {
	const cookies = "SID=a; SAPISID=b; HSID=c"
	tests := map[string]string{
		"bare header":          cookies,
		"header line":          "Cookie: " + cookies,
		"headers block":        "GET /youtubei HTTP/2\nHost: music.youtube.com\ncookie: " + cookies + "\nuser-agent: x\n",
		"curl header":          "curl 'https://music.youtube.com/youtubei/v1/browse' \\\n  -H 'accept: */*' \\\n  -H 'cookie: " + cookies + "' \\\n  -H 'origin: https://music.youtube.com'",
		"curl cookie flag":     "curl 'https://music.youtube.com' -b '" + cookies + "' -H 'x: y'",
		"name line then value": "cookie\n" + cookies + "\nreferer",
	}
	for name, pasted := range tests {
		got, err := ParseCookieHeader(pasted)
		if err != nil || got != cookies {
			t.Errorf("%s: got %q, err %v", name, got, err)
		}
	}
}

func TestParseCookieHeaderErrors(t *testing.T) {
	_, err := ParseCookieHeader("   ")
	if !errors.Is(err, ErrNoCookies) {
		t.Errorf("empty: err = %v", err)
	}
	_, err = ParseCookieHeader("Cookie: YSC=a; VISITOR_INFO1_LIVE=b")
	if !errors.Is(err, ErrNoSession) {
		t.Errorf("no session cookie: err = %v", err)
	}
	_, err = ParseCookieHeader("just some words")
	if !errors.Is(err, ErrNoCookies) {
		t.Errorf("prose: err = %v", err)
	}
}
