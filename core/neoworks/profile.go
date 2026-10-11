package neoworks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Profile is who is signed in, as the OAuth server's userinfo says.
type Profile struct {
	UserID string `json:"sub"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// Profile asks the OAuth server who is signed in.
func (session *Session) Profile(ctx context.Context) (Profile, error) {
	var profile Profile
	accessToken, err := session.accessToken(ctx)
	if err != nil {
		return profile, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, session.config.OAuthURL+"/oauth/userinfo", nil)
	if err != nil {
		return profile, fmt.Errorf("neoworks: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response, err := session.config.HTTPClient.Do(request)
	if err != nil {
		return profile, fmt.Errorf("neoworks: userinfo: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return profile, fmt.Errorf("neoworks: userinfo: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return profile, fmt.Errorf("neoworks: userinfo answered %d: %s", response.StatusCode, body)
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		return profile, fmt.Errorf("neoworks: userinfo: %w", err)
	}
	return profile, nil
}
