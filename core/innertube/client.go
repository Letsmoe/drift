// Package innertube is a client for YouTube Music's internal InnerTube API.
package innertube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://music.youtube.com/youtubei/v1"
	musicOrigin    = "https://music.youtube.com"

	webRemixClientName = "WEB_REMIX"
	// InnerTube accepts any recent version; bump it when YouTube starts rejecting requests.
	webRemixClientVersion = "1.20260930.01.00"
	webRemixUserAgent     = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"

	defaultTimeout = 15 * time.Second
)

// Client sends requests to InnerTube as the YouTube Music web client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	language   string
	region     string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

// WithBaseURL points the client at another InnerTube host, e.g. a test server.
func WithBaseURL(baseURL string) Option {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

// WithLocale sets the language (hl) and region (gl) YouTube Music localizes results for.
func WithLocale(language, region string) Option {
	return func(client *Client) {
		client.language = language
		client.region = region
	}
}

// NewClient creates a client with English/US defaults.
func NewClient(options ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    defaultBaseURL,
		language:   "en",
		region:     "US",
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

type requestContext struct {
	Client clientInfo `json:"client"`
}

type clientInfo struct {
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
	Language      string `json:"hl"`
	Region        string `json:"gl"`
}

func (client *Client) requestContext() requestContext {
	return requestContext{
		Client: clientInfo{
			ClientName:    webRemixClientName,
			ClientVersion: webRemixClientVersion,
			Language:      client.language,
			Region:        client.region,
		},
	}
}

func (client *Client) post(ctx context.Context, endpoint string, payload, response any) error {
	request, err := client.newRequest(ctx, endpoint, payload)
	if err != nil {
		return err
	}

	httpResponse, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("innertube %s: %w", endpoint, err)
	}
	// The body is fully read or abandoned on error; a failed close changes nothing.
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode != http.StatusOK {
		return &StatusError{Endpoint: endpoint, StatusCode: httpResponse.StatusCode}
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(response); err != nil {
		return fmt.Errorf("innertube %s: decode response: %w", endpoint, err)
	}
	return nil
}

func (client *Client) newRequest(ctx context.Context, endpoint string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("innertube %s: encode request: %w", endpoint, err)
	}

	url := client.baseURL + "/" + endpoint + "?prettyPrint=false"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("innertube %s: build request: %w", endpoint, err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", musicOrigin)
	request.Header.Set("User-Agent", webRemixUserAgent)
	return request, nil
}
