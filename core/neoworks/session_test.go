package neoworks

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	libneoworks "github.com/neoworks-dev/libneoworks-go"
	"github.com/zalando/go-keyring"

	"github.com/letsmoe/drift/core/neoworks/oskeystore"
)

// fakeNeoworks answers the OAuth token endpoint and the API's identity lookup.
type fakeNeoworks struct {
	server *httptest.Server

	mutex            sync.Mutex
	tokenForms       []url.Values
	tokenStatus      int
	identityRequests []*http.Request
	apiAuthorization []string
}

func newFakeNeoworks(t *testing.T) *fakeNeoworks {
	t.Helper()
	fake := &fakeNeoworks{tokenStatus: http.StatusOK}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth/token", fake.token)
	mux.HandleFunc("GET /api/v1/keys/identity", fake.identity)
	mux.HandleFunc("GET /oauth/userinfo", fake.userinfo)
	mux.HandleFunc("/api/", fake.api)
	fake.server = httptest.NewServer(mux)
	t.Cleanup(fake.server.Close)
	return fake
}

func (fake *fakeNeoworks) token(response http.ResponseWriter, request *http.Request) {
	_ = request.ParseForm()
	fake.mutex.Lock()
	fake.tokenForms = append(fake.tokenForms, request.PostForm)
	status := fake.tokenStatus
	fake.mutex.Unlock()
	if status != http.StatusOK {
		http.Error(response, `{"error":"invalid_grant"}`, status)
		return
	}
	answer := map[string]any{
		"access_token": testAccessToken, "refresh_token": "refresh-2", "expires_in": 900,
		"neoworks_grant": map[string]any{"installId": "x", "certificate": "x", "certificateSignature": "x", "grants": []any{}},
	}
	_ = json.NewEncoder(response).Encode(answer)
}

func (fake *fakeNeoworks) identity(response http.ResponseWriter, request *http.Request) {
	fake.mutex.Lock()
	fake.identityRequests = append(fake.identityRequests, request)
	fake.mutex.Unlock()
	_, _ = response.Write([]byte(`{"userId":"user-1","encPub":"AA","signPub":"AA","selfSig":"AA","version":1}`))
}

func (fake *fakeNeoworks) userinfo(response http.ResponseWriter, request *http.Request) {
	if request.Header.Get("Authorization") != "Bearer "+testAccessToken {
		http.Error(response, "unauthorized", http.StatusUnauthorized)
		return
	}
	_, _ = response.Write([]byte(`{"sub":"user-1","name":"Ada Lovelace","email":"ada@example.com"}`))
}

func (fake *fakeNeoworks) api(response http.ResponseWriter, request *http.Request) {
	fake.mutex.Lock()
	fake.apiAuthorization = append(fake.apiAuthorization, request.Header.Get("Authorization"))
	fake.mutex.Unlock()
	_, _ = response.Write([]byte("{}"))
}

// testAccessToken is an unsigned JWT naming user-1; the client only reads its sub.
var testAccessToken = "e30." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user-1"}`)) + ".sig"

func testConfig(t *testing.T, fake *fakeNeoworks, directory string) Config {
	t.Helper()
	keystore, err := oskeystore.Open("neoworks-test", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return Config{
		OAuthURL: fake.server.URL, APIURL: fake.server.URL, ClientID: "drift",
		InstallName: "Drift test", Directory: directory, Keystore: keystore,
	}
}

func openTestSession(t *testing.T, config Config) *Session {
	t.Helper()
	session, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	return session
}

func TestOpenKeepsTheInstallAcrossRuns(t *testing.T) {
	keyring.MockInit()
	config := testConfig(t, newFakeNeoworks(t), t.TempDir())
	first, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	installID := first.description.InstallID
	first.Close()

	second := openTestSession(t, config)
	if second.description.InstallID != installID {
		t.Fatalf("install changed from %s to %s", installID, second.description.InstallID)
	}
	info, err := os.Stat(filepath.Join(config.Directory, "install.sealed"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("sealed install: %v %v", info, err)
	}
	if second.SignedIn() {
		t.Fatal("a fresh install is not signed in")
	}
}

func TestBeginSignInAsksForDriftsCollectionsWithPKCE(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	session := openTestSession(t, testConfig(t, fake, t.TempDir()))
	address, err := session.BeginSignIn("http://127.0.0.1:49152/neoworks/callback")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(address)
	if err != nil || parsed.Path != "/oauth/authorize" {
		t.Fatalf("authorize URL %s %v", address, err)
	}
	query := parsed.Query()
	digest := sha256.Sum256([]byte(session.pending.verifier))
	wants := map[string]string{
		"client_id": "drift", "redirect_uri": "http://127.0.0.1:49152/neoworks/callback",
		"code_challenge_method": "S256", "code_challenge": base64.RawURLEncoding.EncodeToString(digest[:]),
		"install_id": session.description.InstallID, "install_sign_pub": session.description.InstallSignPub,
		"install_name": "Drift test", "state": session.pending.state,
	}
	for name, want := range wants {
		if query.Get(name) != want {
			t.Errorf("%s = %q, want %q", name, query.Get(name), want)
		}
	}
	if !strings.Contains(query.Get("scope"), "@drift/history:read") || !strings.Contains(query.Get("scope"), "@drift/blocks:write") {
		t.Errorf("scope %q misses Drift's collections", query.Get("scope"))
	}
}

func TestCompleteSignInRefusesAnotherState(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	session := openTestSession(t, testConfig(t, fake, t.TempDir()))
	if _, err := session.BeginSignIn("http://127.0.0.1/neoworks/callback"); err != nil {
		t.Fatal(err)
	}
	err := session.CompleteSignIn(context.Background(), url.Values{"state": {"forged"}, "code": {"code-1"}})
	if !errors.Is(err, ErrSignInState) {
		t.Fatalf("got %v, want %v", err, ErrSignInState)
	}
	if len(fake.tokenForms) != 0 {
		t.Fatal("a forged callback must not reach the token endpoint")
	}
}

// The fake grant is not signed by the owner, so libneoworks must refuse it and
// the session must stay signed out.
func TestCompleteSignInExchangesTheCodeAndRefusesAnUnsignedGrant(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	config := testConfig(t, fake, t.TempDir())
	session := openTestSession(t, config)
	if _, err := session.BeginSignIn("http://127.0.0.1/neoworks/callback"); err != nil {
		t.Fatal(err)
	}
	verifier := session.pending.verifier
	callback := url.Values{"state": {session.pending.state}, "code": {"code-1"}}

	if err := session.CompleteSignIn(context.Background(), callback); err == nil {
		t.Fatal("an unsigned grant must be refused")
	}
	form := fake.tokenForms[0]
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "code-1" ||
		form.Get("code_verifier") != verifier || form.Get("redirect_uri") != "http://127.0.0.1/neoworks/callback" {
		t.Errorf("token request %v", form)
	}
	identity := fake.identityRequests[0]
	if identity.URL.Query().Get("userId") != "user-1" || identity.Header.Get("Authorization") != "Bearer "+testAccessToken {
		t.Errorf("identity lookup %s %s", identity.URL, identity.Header.Get("Authorization"))
	}
	if session.SignedIn() {
		t.Fatal("a refused grant must leave the session signed out")
	}
	if _, err := os.Stat(filepath.Join(config.Directory, "tokens.sealed")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("tokens of a refused grant were kept: %v", err)
	}
}

func signedInSession(t *testing.T, fake *fakeNeoworks, expiresIn time.Duration) (*Session, Config) {
	t.Helper()
	config := testConfig(t, fake, t.TempDir())
	stored := &tokens{AccessToken: testAccessToken, RefreshToken: "refresh-1", ExpiresAt: time.Now().Add(expiresIn), UserID: "user-1"}
	if err := os.MkdirAll(config.Directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveTokens(config, stored); err != nil {
		t.Fatal(err)
	}
	return openTestSession(t, config), config
}

func TestAccessTokenIsRefreshedBeforeItExpires(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	session, config := signedInSession(t, fake, 30*time.Second)
	if _, err := session.accessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(fake.tokenForms) != 1 || fake.tokenForms[0].Get("refresh_token") != "refresh-1" {
		t.Fatalf("refresh requests %v", fake.tokenForms)
	}
	reloaded, err := loadTokens(config)
	if err != nil || reloaded.RefreshToken != "refresh-2" || time.Until(reloaded.ExpiresAt) < 10*time.Minute {
		t.Fatalf("refreshed tokens were not kept: %+v %v", reloaded, err)
	}
}

func TestARefusedRefreshSignsOut(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	fake.tokenStatus = http.StatusBadRequest
	session, config := signedInSession(t, fake, 0)
	if _, err := session.accessToken(context.Background()); !errors.Is(err, ErrSignedOut) {
		t.Fatalf("got %v, want %v", err, ErrSignedOut)
	}
	if session.SignedIn() {
		t.Fatal("still signed in")
	}
	if _, err := os.Stat(filepath.Join(config.Directory, "tokens.sealed")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("tokens kept after a refused refresh: %v", err)
	}
}

func TestTransportSendsTheTokenOnlyToTheAPI(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	session, _ := signedInSession(t, fake, time.Hour)
	ctx := context.Background()
	if _, err := (transport{session: session}).Do(ctx, libneoworks.Request{Method: "GET", URL: "/api/v1/nodes/pull", IsAPI: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := (transport{session: session}).Do(ctx, libneoworks.Request{Method: "GET", URL: fake.server.URL + "/api/blob"}); err != nil {
		t.Fatal(err)
	}
	if fake.apiAuthorization[0] != "Bearer "+testAccessToken || fake.apiAuthorization[1] != "" {
		t.Fatalf("authorization headers %q", fake.apiAuthorization)
	}
}

func TestSyncNeedsASignedInUser(t *testing.T) {
	keyring.MockInit()
	session := openTestSession(t, testConfig(t, newFakeNeoworks(t), t.TempDir()))
	if _, err := session.Sync(context.Background()); !errors.Is(err, ErrSignedOut) {
		t.Fatalf("got %v, want %v", err, ErrSignedOut)
	}
}

func TestProfileNamesTheSignedInUser(t *testing.T) {
	keyring.MockInit()
	session, _ := signedInSession(t, newFakeNeoworks(t), time.Hour)
	profile, err := session.Profile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := Profile{UserID: "user-1", Name: "Ada Lovelace", Email: "ada@example.com"}
	if profile != want {
		t.Fatalf("profile %+v, want %+v", profile, want)
	}
}

func TestACancelledSignInRefusesItsCallback(t *testing.T) {
	keyring.MockInit()
	fake := newFakeNeoworks(t)
	session := openTestSession(t, testConfig(t, fake, t.TempDir()))
	if _, err := session.BeginSignIn("http://127.0.0.1/neoworks/callback"); err != nil {
		t.Fatal(err)
	}
	state := session.pending.state
	session.CancelSignIn()
	err := session.CompleteSignIn(context.Background(), url.Values{"state": {state}, "code": {"code-1"}})
	if !errors.Is(err, ErrSignInState) {
		t.Fatalf("got %v, want %v", err, ErrSignInState)
	}
}
