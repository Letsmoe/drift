package innertube

import (
	"context"

	"github.com/neoworks/otter-music/core/music"
)

const homeBrowseID = "FEmusic_home"

type browseRequest struct {
	Context      requestContext `json:"context"`
	BrowseID     string         `json:"browseId,omitempty"`
	Continuation string         `json:"continuation,omitempty"`
}

// Home returns the first page of the YouTube Music home feed.
func (client *Client) Home(ctx context.Context) (music.HomePage, error) {
	return client.homePage(ctx, browseRequest{BrowseID: homeBrowseID})
}

// HomeContinuation returns the home feed page after the one that carried continuation.
func (client *Client) HomeContinuation(ctx context.Context, continuation string) (music.HomePage, error) {
	return client.homePage(ctx, browseRequest{Continuation: continuation})
}

func (client *Client) homePage(ctx context.Context, request browseRequest) (music.HomePage, error) {
	var response browseResponse
	if err := client.browse(ctx, request, &response); err != nil {
		return music.HomePage{}, err
	}
	list := response.sectionList()
	return music.HomePage{
		Shelves:      parseShelves(list.Contents),
		Continuation: list.continuation(),
	}, nil
}

// browse posts request to the browse endpoint and decodes the answer into response.
func (client *Client) browse(ctx context.Context, request browseRequest, response any) error {
	call, err := client.visitorCall(ctx, WebRemix, "browse", func(context requestContext) any {
		request.Context = context
		return request
	})
	if err != nil {
		return err
	}
	return client.post(ctx, call, response)
}
