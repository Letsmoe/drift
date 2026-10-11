package neoworks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	libneoworks "github.com/neoworks-dev/libneoworks-go"
)

// transport performs libneoworks' requests: API paths on the API origin with
// the user's access token, anything else as given.
type transport struct {
	session *Session
}

func (transport transport) Do(ctx context.Context, request libneoworks.Request) (libneoworks.Response, error) {
	var answer libneoworks.Response
	address := request.URL
	if request.IsAPI {
		address = transport.session.config.APIURL + request.URL
	}
	var body io.Reader
	if len(request.Body) > 0 {
		body = bytes.NewReader(request.Body)
	}
	outgoing, err := http.NewRequestWithContext(ctx, request.Method, address, body)
	if err != nil {
		return answer, fmt.Errorf("neoworks: %w", err)
	}
	if request.ContentType != "" {
		outgoing.Header.Set("Content-Type", request.ContentType)
	}
	if request.IsAPI {
		accessToken, err := transport.session.accessToken(ctx)
		if err != nil {
			return answer, err
		}
		outgoing.Header.Set("Authorization", "Bearer "+accessToken)
	}
	response, err := transport.session.config.HTTPClient.Do(outgoing)
	if err != nil {
		return answer, fmt.Errorf("neoworks: %s %s: %w", request.Method, request.URL, err)
	}
	defer func() { _ = response.Body.Close() }()
	answer.Status = response.StatusCode
	answer.Body, err = io.ReadAll(response.Body)
	if err != nil {
		return answer, fmt.Errorf("neoworks: %s %s: %w", request.Method, request.URL, err)
	}
	return answer, nil
}
