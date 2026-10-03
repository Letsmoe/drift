package innertube

import (
	"context"
	"errors"
)

// errNoVisitorData means InnerTube answered visitor_id without a visitor ID.
var errNoVisitorData = errors.New("innertube visitor_id: response has no visitorData")

type visitorRequest struct {
	Context requestContext `json:"context"`
}

type visitorResponse struct {
	ResponseContext struct {
		VisitorData string
	}
}

// VisitorData returns the visitor ID sent with every request, requesting one on
// first use. Without it the player endpoint answers "Sign in to confirm you're not a bot".
func (client *Client) VisitorData(ctx context.Context) (string, error) {
	client.visitorMutex.Lock()
	defer client.visitorMutex.Unlock()

	if client.visitorData != "" {
		return client.visitorData, nil
	}

	visitorData, err := client.fetchVisitorData(ctx)
	if err != nil {
		return "", err
	}
	client.visitorData = visitorData
	return visitorData, nil
}

func (client *Client) fetchVisitorData(ctx context.Context) (string, error) {
	call := apiCall{
		profile:  WebRemix,
		endpoint: "visitor_id",
		payload:  visitorRequest{Context: client.requestContext(WebRemix, "")},
	}

	var response visitorResponse
	if err := client.post(ctx, call, &response); err != nil {
		return "", err
	}
	if response.ResponseContext.VisitorData == "" {
		return "", errNoVisitorData
	}
	return response.ResponseContext.VisitorData, nil
}

// visitorCall builds an apiCall for profile with the visitor ID attached.
func (client *Client) visitorCall(ctx context.Context, profile Profile, endpoint string, buildPayload func(requestContext) any) (apiCall, error) {
	visitorData, err := client.VisitorData(ctx)
	if err != nil {
		return apiCall{}, err
	}
	return apiCall{
		profile:     profile,
		endpoint:    endpoint,
		visitorData: visitorData,
		payload:     buildPayload(client.requestContext(profile, visitorData)),
	}, nil
}
