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
	request := searchRequest{
		Context: client.requestContext(),
		Query:   query,
		Params:  songsFilterParams,
	}

	var response searchResponse
	if err := client.post(ctx, "search", request, &response); err != nil {
		return nil, err
	}
	return parseSongs(response.listItems()), nil
}
