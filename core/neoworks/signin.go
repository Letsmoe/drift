package neoworks

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrSignInState means a callback does not belong to the sign-in in progress.
var ErrSignInState = errors.New("neoworks: the sign-in callback does not match the sign-in in progress")

// pendingSignIn is the authorization request waiting for its callback.
type pendingSignIn struct {
	state       string
	verifier    string
	redirectURI string
}

// BeginSignIn starts an authorization request and returns the URL to open in
// the browser. The callback goes to redirectURI, a loopback address of the
// app; pass its query to CompleteSignIn. A new call replaces the pending one.
func (session *Session) BeginSignIn(redirectURI string) (string, error) {
	state, err := randomToken()
	if err != nil {
		return "", err
	}
	verifier, err := randomToken()
	if err != nil {
		return "", err
	}
	challenge := sha256.Sum256([]byte(verifier))
	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {session.config.ClientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {strings.Join(scopes(), " ")},
		"state":                 {state},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
		"install_id":            {session.description.InstallID},
		"install_enc_pub":       {session.description.InstallEncPub},
		"install_sign_pub":      {session.description.InstallSignPub},
		"install_name":          {session.config.InstallName},
	}
	session.mutex.Lock()
	session.pending = &pendingSignIn{state: state, verifier: verifier, redirectURI: redirectURI}
	session.mutex.Unlock()
	return session.config.OAuthURL + "/oauth/authorize?" + query.Encode(), nil
}

// CompleteSignIn takes the callback's query: it exchanges the code, takes in
// the grant the user consented to and keeps the tokens.
func (session *Session) CompleteSignIn(ctx context.Context, callback url.Values) error {
	pending, err := session.takePending(callback.Get("state"))
	if err != nil {
		return err
	}
	if callback.Get("error") != "" {
		return fmt.Errorf("neoworks: sign-in failed: %s %s", callback.Get("error"), callback.Get("error_description"))
	}
	response, err := session.postToken(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {callback.Get("code")},
		"redirect_uri":  {pending.redirectURI},
		"client_id":     {session.config.ClientID},
		"code_verifier": {pending.verifier},
	})
	if err != nil {
		return err
	}
	signedIn, err := tokensFrom(response, time.Now())
	if err != nil {
		return err
	}
	if len(response.NeoworksGrant) == 0 {
		return errors.New("neoworks: the token response carries no grant")
	}
	if err := session.acceptGrant(ctx, response.NeoworksGrant, signedIn); err != nil {
		return err
	}
	if err := saveTokens(session.config, signedIn); err != nil {
		return err
	}
	session.mutex.Lock()
	session.tokens = signedIn
	session.descriptorsRegistered = false
	session.mutex.Unlock()
	return nil
}

// takePending removes and returns the pending sign-in when state matches it.
func (session *Session) takePending(state string) (*pendingSignIn, error) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	pending := session.pending
	if pending == nil || subtle.ConstantTimeCompare([]byte(pending.state), []byte(state)) != 1 {
		return nil, ErrSignInState
	}
	session.pending = nil
	return pending, nil
}

// acceptGrant verifies the grant against the owner's public identity and
// stores the node keys it carries.
func (session *Session) acceptGrant(ctx context.Context, grant json.RawMessage, owner *tokens) error {
	identity, err := session.ownerIdentity(ctx, owner)
	if err != nil {
		return err
	}
	if err := session.client.AcceptGrant(grant, identity); err != nil {
		return fmt.Errorf("neoworks: accepting the grant: %w", err)
	}
	return nil
}

func (session *Session) ownerIdentity(ctx context.Context, owner *tokens) (json.RawMessage, error) {
	address := session.config.APIURL + "/api/v1/keys/identity?" + url.Values{"userId": {owner.UserID}}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, fmt.Errorf("neoworks: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+owner.AccessToken)
	response, err := session.config.HTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("neoworks: identity lookup: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("neoworks: identity lookup: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("neoworks: identity lookup answered %d: %s", response.StatusCode, body)
	}
	return body, nil
}

func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("neoworks: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
