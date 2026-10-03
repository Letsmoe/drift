package innertube

import (
	"context"

	"github.com/neoworks/otter-music/core/music"
)

// songsFilterParams restricts search results to the "Songs" shelf.
const songsFilterParams = "EgWKAQIIAWoMEA4QChADEAQQCRAF"

type searchRequest struct {
	Context requestContext `json:"context"`
	Query   string         `json:"query"`
	Params  string         `json:"params,omitempty"`
}

// SearchSongs returns the songs YouTube Music finds for query, in ranking order.
func (client *Client) SearchSongs(ctx context.Context, query string) ([]music.Track, error) {
	call, err := client.visitorCall(ctx, WebRemix, "search", func(context requestContext) any {
		return searchRequest{Context: context, Query: query, Params: songsFilterParams}
	})
	if err != nil {
		return nil, err
	}

	var response searchResponse
	if err := client.post(ctx, call, &response); err != nil {
		return nil, err
	}
	return parseSongs(response.listItems()), nil
}
