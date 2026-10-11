package neoworks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// tokensLabel binds the sealed tokens to their purpose in the keystore.
var tokensLabel = []byte("drift.neoworks.tokens")

// refreshMargin is how long before expiry an access token is refreshed.
const refreshMargin = time.Minute

// tokens are a signed-in user's OAuth tokens. They are credentials and live
// on disk only sealed under the keystore.
type tokens struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	UserID       string    `json:"userId"`
}

// tokenResponse is the token endpoint's answer.
type tokenResponse struct {
	AccessToken   string          `json:"access_token"`
	RefreshToken  string          `json:"refresh_token"`
	ExpiresIn     int             `json:"expires_in"`
	NeoworksGrant json.RawMessage `json:"neoworks_grant"`
}

func loadTokens(config Config) (*tokens, error) {
	sealed, err := os.ReadFile(filepath.Join(config.Directory, "tokens.sealed"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("neoworks: %w", err)
	}
	plain, err := config.Keystore.Unwrap(sealed, tokensLabel)
	if err != nil {
		return nil, fmt.Errorf("neoworks: unsealing tokens: %w", err)
	}
	var loaded tokens
	if err := json.Unmarshal(plain, &loaded); err != nil {
		return nil, fmt.Errorf("neoworks: tokens: %w", err)
	}
	return &loaded, nil
}

func saveTokens(config Config, saved *tokens) error {
	plain, err := json.Marshal(saved)
	if err != nil {
		return fmt.Errorf("neoworks: tokens: %w", err)
	}
	sealed, err := config.Keystore.Wrap(plain, tokensLabel)
	if err != nil {
		return fmt.Errorf("neoworks: sealing tokens: %w", err)
	}
	return writePrivateFile(filepath.Join(config.Directory, "tokens.sealed"), sealed)
}

// tokensFrom turns a token response into tokens; the user ID is the access
// token's subject.
func tokensFrom(response tokenResponse, now time.Time) (*tokens, error) {
	userID, err := subjectOf(response.AccessToken)
	if err != nil {
		return nil, err
	}
	return &tokens{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    now.Add(time.Duration(response.ExpiresIn) * time.Second),
		UserID:       userID,
	}, nil
}

// subjectOf reads the sub claim of a JWT without verifying it: the token came
// straight from the token endpoint over TLS, and the API verifies it.
func subjectOf(accessToken string) (string, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return "", errors.New("neoworks: the access token is not a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("neoworks: access token payload: %w", err)
	}
	var claims struct {
		Subject string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Subject == "" {
		return "", errors.New("neoworks: the access token names no subject")
	}
	return claims.Subject, nil
}

// accessToken returns a valid access token, refreshing it when it is about to
// expire. A refresh the server refuses signs the session out.
func (session *Session) accessToken(ctx context.Context) (string, error) {
	session.mutex.Lock()
	current := session.tokens
	session.mutex.Unlock()
	if current == nil {
		return "", ErrSignedOut
	}
	if time.Until(current.ExpiresAt) > refreshMargin {
		return current.AccessToken, nil
	}
	refreshed, err := session.refresh(ctx, current)
	if errors.Is(err, errTokenRefused) {
		return "", errors.Join(ErrSignedOut, session.SignOut())
	}
	if err != nil {
		return "", err
	}
	return refreshed.AccessToken, nil
}

// errTokenRefused means the token endpoint refused the code or refresh token.
var errTokenRefused = errors.New("neoworks: the token endpoint refused the grant")

func (session *Session) refresh(ctx context.Context, current *tokens) (*tokens, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {current.RefreshToken},
		"client_id":     {session.config.ClientID},
	}
	response, err := session.postToken(ctx, form)
	if err != nil {
		return nil, err
	}
	refreshed, err := tokensFrom(response, time.Now())
	if err != nil {
		return nil, err
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = current.RefreshToken
	}
	if err := saveTokens(session.config, refreshed); err != nil {
		return nil, err
	}
	session.mutex.Lock()
	session.tokens = refreshed
	session.mutex.Unlock()
	return refreshed, nil
}

// postToken calls the token endpoint. A 400 or 401 is errTokenRefused; other
// failures leave the tokens alone.
func (session *Session) postToken(ctx context.Context, form url.Values) (tokenResponse, error) {
	var answer tokenResponse
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, session.config.OAuthURL+"/oauth/token",
		strings.NewReader(form.Encode()))
	if err != nil {
		return answer, fmt.Errorf("neoworks: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := session.config.HTTPClient.Do(request)
	if err != nil {
		return answer, fmt.Errorf("neoworks: token endpoint: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return answer, fmt.Errorf("neoworks: token endpoint: %w", err)
	}
	if response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusUnauthorized {
		return answer, fmt.Errorf("%w: %s", errTokenRefused, body)
	}
	if response.StatusCode != http.StatusOK {
		return answer, fmt.Errorf("neoworks: token endpoint answered %d: %s", response.StatusCode, body)
	}
	if err := json.Unmarshal(body, &answer); err != nil {
		return answer, fmt.Errorf("neoworks: token response: %w", err)
	}
	return answer, nil
}
