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

// Search returns the results YouTube Music finds for query as shelves: a top
// result, then songs, albums, artists, playlists, videos, podcasts and episodes.
func (client *Client) Search(ctx context.Context, query string) (music.SearchResults, error) {
	call, err := client.visitorCall(ctx, WebRemix, "search", func(context requestContext) any {
		return searchRequest{Context: context, Query: query}
	})
	if err != nil {
		return music.SearchResults{}, err
	}

	var response searchResponse
	if err := client.post(ctx, call, &response); err != nil {
		return music.SearchResults{}, err
	}
	return parseSearch(response.sections()), nil
}

type suggestionsRequest struct {
	Context requestContext `json:"context"`
	Input   string         `json:"input"`
}

// SearchSuggestions returns the text completions YouTube Music offers while query is typed.
func (client *Client) SearchSuggestions(ctx context.Context, query string) ([]string, error) {
	call, err := client.visitorCall(ctx, WebRemix, "music/get_search_suggestions", func(context requestContext) any {
		return suggestionsRequest{Context: context, Input: query}
	})
	if err != nil {
		return nil, err
	}

	var response suggestionsResponse
	if err := client.post(ctx, call, &response); err != nil {
		return nil, err
	}
	return parseSuggestions(response), nil
}
