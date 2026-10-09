package innertube

import (
	"context"

	"github.com/letsmoe/drift/core/music"
)

const homeBrowseID = "FEmusic_home"

// podcastsChipParams selects the "Podcasts" chip of the home feed (chipCloudChipRenderer).
const podcastsChipParams = "ggNCSgQIDBADSgQIBxABSgQIBRABSgQIChABSgQIDRABSgQICRABSgQICBABSgQIDhABSgQIBhABSgQIBBABSgQIAxAB"

type browseRequest struct {
	Context      requestContext `json:"context"`
	BrowseID     string         `json:"browseId,omitempty"`
	Params       string         `json:"params,omitempty"`
	Continuation string         `json:"continuation,omitempty"`
}

// Home returns the first page of the YouTube Music home feed.
func (client *Client) Home(ctx context.Context) (music.HomePage, error) {
	return client.homePage(ctx, browseRequest{BrowseID: homeBrowseID})
}

// PodcastsHome returns the first page of the home feed's "Podcasts" chip: episode
// shelves by topic. Its continuations go through HomeContinuation.
func (client *Client) PodcastsHome(ctx context.Context) (music.HomePage, error) {
	return client.homePage(ctx, browseRequest{BrowseID: homeBrowseID, Params: podcastsChipParams})
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
	return client.browseAs(ctx, WebRemix, request, response)
}

// browseAs is browse with another client profile.
func (client *Client) browseAs(ctx context.Context, profile Profile, request browseRequest, response any) error {
	call, err := client.visitorCall(ctx, profile, "browse", func(context requestContext) any {
		request.Context = context
		return request
	})
	if err != nil {
		return err
	}
	return client.post(ctx, call, response)
}
