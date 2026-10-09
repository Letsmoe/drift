package innertube

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/letsmoe/drift-music/core/music"
)

// playlistBrowsePrefix turns a playlist ID into the browse ID that opens it.
const playlistBrowsePrefix = "VL"

// errNoPageHeader means a browse response lacks the header the page is read from,
// usually because the ID points at a different kind of page.
var errNoPageHeader = errors.New("response has no page header")

// Album returns the album opened by browseID (MPREb_...) with its tracks.
func (client *Client) Album(ctx context.Context, browseID string) (music.AlbumPage, error) {
	var response pageResponse
	if err := client.browse(ctx, browseRequest{BrowseID: browseID}, &response); err != nil {
		return music.AlbumPage{}, err
	}
	header := response.header()
	if header == nil {
		return music.AlbumPage{}, fmt.Errorf("innertube album %s: %w", browseID, errNoPageHeader)
	}
	return parseAlbumPage(browseID, *header, response.listShelf()), nil
}

// Playlist returns the first page of a playlist; browseID may carry the "VL" prefix or not.
func (client *Client) Playlist(ctx context.Context, browseID string) (music.PlaylistPage, error) {
	browseID = playlistBrowsePrefix + strings.TrimPrefix(browseID, playlistBrowsePrefix)

	var response pageResponse
	if err := client.browse(ctx, browseRequest{BrowseID: browseID}, &response); err != nil {
		return music.PlaylistPage{}, err
	}
	header := response.header()
	if header == nil {
		return music.PlaylistPage{}, fmt.Errorf("innertube playlist %s: %w", browseID, errNoPageHeader)
	}
	return parsePlaylistPage(browseID, *header, response.playlistShelf()), nil
}

// PlaylistContinuation returns the tracks after the page that carried token.
func (client *Client) PlaylistContinuation(ctx context.Context, token string) (music.PlaylistTracks, error) {
	var response pageResponse
	if err := client.browse(ctx, browseRequest{Continuation: token}, &response); err != nil {
		return music.PlaylistTracks{}, err
	}
	entries := response.appendedItems()
	return music.PlaylistTracks{
		Tracks:       parseSongs(rowsOf(entries)),
		Continuation: continuationTokenOf(entries),
	}, nil
}

// Artist returns the artist opened by browseID (UC...) with their songs and shelves.
func (client *Client) Artist(ctx context.Context, browseID string) (music.ArtistPage, error) {
	var response browseResponse
	if err := client.browse(ctx, browseRequest{BrowseID: browseID}, &response); err != nil {
		return music.ArtistPage{}, err
	}
	header := response.artistHeader()
	if header == nil {
		return music.ArtistPage{}, fmt.Errorf("innertube artist %s: %w", browseID, errNoPageHeader)
	}
	return parseArtistPage(browseID, *header, response.sectionList().Contents), nil
}

// Podcast returns the first page of episodes of the show opened by browseID (MPSP...).
func (client *Client) Podcast(ctx context.Context, browseID string) (music.PodcastPage, error) {
	var response pageResponse
	if err := client.browse(ctx, browseRequest{BrowseID: browseID}, &response); err != nil {
		return music.PodcastPage{}, err
	}
	header := response.header()
	if header == nil {
		return music.PodcastPage{}, fmt.Errorf("innertube podcast %s: %w", browseID, errNoPageHeader)
	}
	return parsePodcastPage(browseID, *header, response.listShelf()), nil
}

// PodcastContinuation returns the episodes after the page that carried token.
func (client *Client) PodcastContinuation(ctx context.Context, token string) (music.PodcastEpisodes, error) {
	pageToken, err := decodePodcastToken(token)
	if err != nil {
		return music.PodcastEpisodes{}, err
	}

	var response pageResponse
	if err := client.browse(ctx, browseRequest{Continuation: pageToken.Token}, &response); err != nil {
		return music.PodcastEpisodes{}, err
	}
	shelf := response.ContinuationContents.MusicShelfContinuation
	return music.PodcastEpisodes{
		Episodes:     parseShowEpisodes(shelf.Contents, pageToken.Show),
		Continuation: pageToken.next(shelf.Continuations.token()),
	}, nil
}

type radioRequest struct {
	Context    requestContext `json:"context"`
	VideoID    string         `json:"videoId"`
	PlaylistID string         `json:"playlistId"`
}

// radioPlaylistPrefix is the playlist ID prefix of YouTube Music's "start radio" mix.
const radioPlaylistPrefix = "RDAMVM"

// Radio returns the tracks of the radio YouTube Music builds from videoID. The
// seed track comes first when YouTube includes it.
func (client *Client) Radio(ctx context.Context, videoID string) ([]music.Track, error) {
	call, err := client.visitorCall(ctx, WebRemix, "next", func(context requestContext) any {
		return radioRequest{Context: context, VideoID: videoID, PlaylistID: radioPlaylistPrefix + videoID}
	})
	if err != nil {
		return nil, err
	}

	var response radioResponse
	if err := client.post(ctx, call, &response); err != nil {
		return nil, err
	}
	return parseRadio(response), nil
}
