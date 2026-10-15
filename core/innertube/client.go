// Package innertube is a client for YouTube's internal InnerTube API.
package innertube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const defaultTimeout = 15 * time.Second

// defaultRetryDelays are the pauses before each retry of an anonymous request
// YouTube refused with a transient status. It intermittently answers 403 to
// requests that go through when repeated.
var defaultRetryDelays = []time.Duration{400 * time.Millisecond, 1200 * time.Millisecond}

// Client sends InnerTube requests. It is safe for concurrent use.
type Client struct {
	httpClient *http.Client
	// baseURL replaces every profile's host when set, e.g. with a test server.
	baseURL  string
	language string
	region   string
	// retryDelays are the pauses before each retry; one entry per retry.
	retryDelays []time.Duration

	visitorMutex sync.Mutex
	visitorData  string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithBaseURL sends all requests to baseURL instead of the profiles' hosts.
func WithBaseURL(baseURL string) Option {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

// WithLocale sets the language (hl) and region (gl) YouTube localizes results for.
func WithLocale(language, region string) Option {
	return func(client *Client) {
		client.language = language
		client.region = region
	}
}

// WithVisitorData reuses a visitor ID from an earlier session instead of requesting one.
func WithVisitorData(visitorData string) Option {
	return func(client *Client) {
		client.visitorData = visitorData
	}
}

// NewClient creates a client with English/US defaults.
func NewClient(options ...Option) *Client {
	client := &Client{
		httpClient:  &http.Client{Timeout: defaultTimeout},
		language:    "en",
		region:      "US",
		retryDelays: defaultRetryDelays,
	}
	for _, option := range options {
		option(client)
	}
	return client
}

// StatusError is returned when InnerTube answers with a non-200 status.
type StatusError struct {
	Endpoint   string
	StatusCode int
}

func (statusError *StatusError) Error() string {
	return fmt.Sprintf("innertube %s: unexpected status %d", statusError.Endpoint, statusError.StatusCode)
}

func (client *Client) requestContext(profile Profile, visitorData string) requestContext {
	return requestContext{
		Client: clientInfo{
			ClientName:    profile.Name,
			ClientVersion: profile.Version,
			UserAgent:     profile.UserAgent,
			DeviceMake:    profile.DeviceMake,
			DeviceModel:   profile.DeviceModel,
			OSName:        profile.OSName,
			OSVersion:     profile.OSVersion,
			Language:      client.language,
			Region:        client.region,
			VisitorData:   visitorData,
		},
	}
}

type apiCall struct {
	profile     Profile
	endpoint    string
	visitorData string
	payload     any
	// session is set for requests that act as the signed-in user; nil means anonymous.
	session *Session
}

// post sends call, retrying an anonymous call YouTube refused with a
// transient status. A signed-in call is never retried: there a 403 means the
// session expired.
func (client *Client) post(ctx context.Context, call apiCall, response any) error {
	err := client.postOnce(ctx, call, response)
	for _, delay := range client.retryDelays {
		if call.session != nil || !isTransient(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(delay):
		}
		err = client.postOnce(ctx, call, response)
	}
	return err
}

// isTransient reports whether err is a status YouTube often answers once and
// not on the next attempt: 403, 429 and server errors.
func isTransient(err error) bool {
	var statusError *StatusError
	if !errors.As(err, &statusError) {
		return false
	}
	code := statusError.StatusCode
	return code == http.StatusForbidden || code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}

func (client *Client) postOnce(ctx context.Context, call apiCall, response any) error {
	request, err := client.newRequest(ctx, call)
	if err != nil {
		return err
	}

	httpResponse, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("innertube %s: %w", call.endpoint, err)
	}
	// The body is fully read or abandoned on error; a failed close changes nothing.
	defer func() { _ = httpResponse.Body.Close() }()

	if call.session != nil {
		return decodeSessionResponse(call, httpResponse, response)
	}
	if httpResponse.StatusCode != http.StatusOK {
		return &StatusError{Endpoint: call.endpoint, StatusCode: httpResponse.StatusCode}
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(response); err != nil {
		return fmt.Errorf("innertube %s: decode response: %w", call.endpoint, err)
	}
	return nil
}

func (client *Client) newRequest(ctx context.Context, call apiCall) (*http.Request, error) {
	body, err := json.Marshal(call.payload)
	if err != nil {
		return nil, fmt.Errorf("innertube %s: encode request: %w", call.endpoint, err)
	}

	url := client.endpointURL(call.profile, call.endpoint)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("innertube %s: build request: %w", call.endpoint, err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", call.profile.origin())
	request.Header.Set("User-Agent", call.profile.UserAgent)
	request.Header.Set("X-YouTube-Client-Name", strconv.Itoa(call.profile.ID))
	request.Header.Set("X-YouTube-Client-Version", call.profile.Version)
	if call.visitorData != "" {
		request.Header.Set("X-Goog-Visitor-Id", call.visitorData)
	}
	if call.session != nil {
		setSessionHeaders(request, *call.session)
	}
	return request, nil
}

func (client *Client) endpointURL(profile Profile, endpoint string) string {
	baseURL := profile.origin()
	if client.baseURL != "" {
		baseURL = client.baseURL
	}
	return baseURL + "/youtubei/v1/" + endpoint + "?prettyPrint=false"
}
