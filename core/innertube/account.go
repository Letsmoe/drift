package innertube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/letsmoe/drift/core/music"
)

// AccountClient reads the signed-in user's YouTube Music library. It is the
// only client that sends the session's cookies; catalog browsing, search and
// playback stay on the anonymous Client.
type AccountClient struct {
	client  *Client
	session Session
}

// NewAccountClient returns a client acting as the user of session.
func NewAccountClient(session Session, options ...Option) *AccountClient {
	return &AccountClient{client: NewClient(options...), session: session}
}

// setSessionHeaders makes a request look like the signed-in web player.
func setSessionHeaders(request *http.Request, session Session) {
	request.Header.Set("Cookie", session.CookieHeader)
	request.Header.Set("Authorization", session.authorization(time.Now()))
	request.Header.Set("X-Origin", musicOrigin)
	request.Header.Set("Origin", musicOrigin)
	request.Header.Set("X-Goog-AuthUser", "0")
}

// signedInState is the part of every response telling whether YouTube saw a
// signed-in user (serviceTrackingParams carry logged_in = 0 or 1).
type signedInState struct {
	ResponseContext struct {
		ServiceTrackingParams []struct {
			Params []struct {
				Key   string
				Value string
			}
		}
	}
}

func (state signedInState) signedOut() bool {
	for _, service := range state.ResponseContext.ServiceTrackingParams {
		for _, param := range service.Params {
			if param.Key == "logged_in" && param.Value == "0" {
				return true
			}
		}
	}
	return false
}

// decodeSessionResponse decodes the answer of an authenticated call and maps
// every sign that YouTube saw no user to ErrSessionExpired.
func decodeSessionResponse(call apiCall, httpResponse *http.Response, response any) error {
	if httpResponse.StatusCode == http.StatusUnauthorized || httpResponse.StatusCode == http.StatusForbidden {
		return ErrSessionExpired
	}
	if httpResponse.StatusCode != http.StatusOK {
		return &StatusError{Endpoint: call.endpoint, StatusCode: httpResponse.StatusCode}
	}
	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return fmt.Errorf("innertube %s: read response: %w", call.endpoint, err)
	}
	var state signedInState
	if json.Unmarshal(body, &state) == nil && state.signedOut() {
		return ErrSessionExpired
	}
	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("innertube %s: decode response: %w", call.endpoint, err)
	}
	return nil
}

// sessionCall builds an authenticated apiCall with the visitor ID attached.
func (account *AccountClient) sessionCall(ctx context.Context, endpoint string, buildPayload func(requestContext) any) (apiCall, error) {
	call, err := account.client.visitorCall(ctx, WebRemix, endpoint, buildPayload)
	if err != nil {
		return apiCall{}, err
	}
	call.session = &account.session
	return call, nil
}

func (account *AccountClient) browse(ctx context.Context, request browseRequest, response any) error {
	call, err := account.sessionCall(ctx, "browse", func(context requestContext) any {
		request.Context = context
		return request
	})
	if err != nil {
		return err
	}
	return account.client.post(ctx, call, response)
}

type accountMenuRequest struct {
	Context requestContext `json:"context"`
}

type accountMenuResponse struct {
	Actions []struct {
		OpenPopupAction struct {
			Popup struct {
				MultiPageMenuRenderer struct {
					Header struct {
						ActiveAccountHeaderRenderer accountHeader
					}
				}
			}
		}
	}
}

type accountHeader struct {
	AccountName struct {
		Runs []run
	}
	ChannelHandle struct {
		Runs []run
	}
	Email struct {
		Runs []run
	}
	AccountPhoto struct {
		Thumbnails []thumbnail
	}
}

// Account returns who the session belongs to. It is the cheapest way to check
// that the session still works: a signed-out answer has no account header.
func (account *AccountClient) Account(ctx context.Context) (music.Account, error) {
	call, err := account.sessionCall(ctx, "account/account_menu", func(context requestContext) any {
		return accountMenuRequest{Context: context}
	})
	if err != nil {
		return music.Account{}, err
	}
	var response accountMenuResponse
	err = account.client.post(ctx, call, &response)
	if err != nil {
		return music.Account{}, err
	}
	return parseAccount(response)
}

func parseAccount(response accountMenuResponse) (music.Account, error) {
	if len(response.Actions) == 0 {
		return music.Account{}, ErrSessionExpired
	}
	header := response.Actions[0].OpenPopupAction.Popup.MultiPageMenuRenderer.Header.ActiveAccountHeaderRenderer
	name := joinText(header.AccountName.Runs)
	if name == "" {
		return music.Account{}, ErrSessionExpired
	}
	return music.Account{
		Name:     name,
		Handle:   joinText(header.ChannelHandle.Runs),
		Email:    joinText(header.Email.Runs),
		PhotoURL: largestThumbnailURL(header.AccountPhoto.Thumbnails),
	}, nil
}

func largestThumbnailURL(thumbnails []thumbnail) string {
	best := thumbnail{}
	for _, candidate := range thumbnails {
		if candidate.URL != "" && candidate.Width >= best.Width {
			best = candidate
		}
	}
	return best.URL
}
