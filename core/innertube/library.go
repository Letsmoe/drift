package innertube

import (
	"context"
	"fmt"
	"strings"

	"github.com/neoworks/otter-music/core/music"
)

// Browse IDs of the library pages (ytmusicapi mixins/library.py).
const (
	likedSongsBrowseID    = "VLLM"
	libraryPlaylistsID    = "FEmusic_liked_playlists"
	libraryAlbumsID       = "FEmusic_liked_albums"
	libraryArtistsID      = "FEmusic_library_corpus_artists"
	libraryPodcastsID     = "FEmusic_library_non_music_audio_list"
	maxLibraryPages       = 500
	libraryArtistPrefix   = "MPLA"
	artistChannelIDPrefix = "UC"
)

// LikedSongs returns every song of the user's "Liked Music" playlist, the
// newest like first.
func (account *AccountClient) LikedSongs(ctx context.Context) ([]music.Track, error) {
	return account.playlistTracksOf(ctx, likedSongsBrowseID)
}

// PlaylistTracks returns every track of a playlist of the user; playlistID
// may carry the "VL" prefix or not.
func (account *AccountClient) PlaylistTracks(ctx context.Context, playlistID string) ([]music.Track, error) {
	return account.playlistTracksOf(ctx, playlistBrowsePrefix+strings.TrimPrefix(playlistID, playlistBrowsePrefix))
}

func (account *AccountClient) playlistTracksOf(ctx context.Context, browseID string) ([]music.Track, error) {
	var response pageResponse
	err := account.browse(ctx, browseRequest{BrowseID: browseID}, &response)
	if err != nil {
		return nil, err
	}
	shelf := response.playlistShelf()
	if len(shelf.Contents) == 0 {
		shelf = response.listShelf()
	}
	tracks := parseSongs(rowsOf(shelf.Contents))
	token := firstToken(continuationTokenOf(shelf.Contents), shelf.Continuations.token())
	for page := 0; token != "" && page < maxLibraryPages; page++ {
		var next pageResponse
		err = account.browse(ctx, browseRequest{Continuation: token}, &next)
		if err != nil {
			return nil, fmt.Errorf("playlist %s: %w", browseID, err)
		}
		more, nextToken := next.playlistContinuation()
		tracks = append(tracks, more...)
		token = nextToken
	}
	return tracks, nil
}

func firstToken(tokens ...string) string {
	for _, token := range tokens {
		if token != "" {
			return token
		}
	}
	return ""
}

// playlistContinuation reads the tracks and next token of a playlist continuation,
// in the action form or the older musicShelfContinuation form.
func (response pageResponse) playlistContinuation() ([]music.Track, string) {
	entries := response.appendedItems()
	if len(entries) > 0 {
		return parseSongs(rowsOf(entries)), continuationTokenOf(entries)
	}
	shelf := response.ContinuationContents.MusicShelfContinuation
	return parseSongs(rowsOf(shelf.Contents)), firstToken(continuationTokenOf(shelf.Contents), shelf.Continuations.token())
}

// LibraryPlaylists lists the playlists in the user's library, including the
// auto-generated ones ("Liked Music", "Episodes for later"), which callers skip.
func (account *AccountClient) LibraryPlaylists(ctx context.Context) ([]music.Item, error) {
	return account.libraryItems(ctx, libraryPlaylistsID, music.ItemKindPlaylist)
}

// LibraryAlbums lists the albums the user saved.
func (account *AccountClient) LibraryAlbums(ctx context.Context) ([]music.Item, error) {
	return account.libraryItems(ctx, libraryAlbumsID, music.ItemKindAlbum)
}

// LibraryArtists lists the artists the user subscribed to.
func (account *AccountClient) LibraryArtists(ctx context.Context) ([]music.Item, error) {
	return account.libraryItems(ctx, libraryArtistsID, music.ItemKindArtist)
}

// LibraryPodcasts lists the podcast shows the user saved.
func (account *AccountClient) LibraryPodcasts(ctx context.Context) ([]music.Item, error) {
	return account.libraryItems(ctx, libraryPodcastsID, music.ItemKindPodcast)
}

// libraryItems follows every continuation of a library page and keeps the
// entries of one kind, without duplicates.
func (account *AccountClient) libraryItems(ctx context.Context, browseID string, kind music.ItemKind) ([]music.Item, error) {
	items := []music.Item{}
	seen := map[string]bool{}
	request := browseRequest{BrowseID: browseID}
	for page := 0; page < maxLibraryPages; page++ {
		var response libraryResponse
		err := account.browse(ctx, request, &response)
		if err != nil {
			return nil, fmt.Errorf("library %s: %w", browseID, err)
		}
		libraryPage := response.page()
		items = appendLibraryItems(items, seen, libraryPage.entries, kind)
		if libraryPage.continuation == "" {
			break
		}
		request = browseRequest{Continuation: libraryPage.continuation}
	}
	return items, nil
}

func appendLibraryItems(items []music.Item, seen map[string]bool, entries []libraryEntry, kind music.ItemKind) []music.Item {
	for _, entry := range entries {
		item, ok := parseLibraryEntry(entry)
		if !ok || item.Kind != kind || seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		items = append(items, item)
	}
	return items
}

func parseLibraryEntry(entry libraryEntry) (music.Item, bool) {
	if entry.MusicTwoRowItemRenderer != nil {
		return parseCard(*entry.MusicTwoRowItemRenderer)
	}
	if entry.MusicResponsiveListItemRenderer != nil {
		return parseArtistRow(*entry.MusicResponsiveListItemRenderer)
	}
	return music.Item{}, false
}

// parseArtistRow reads a subscription row: the artist's name, subscriber text
// and channel ID. Library lists may prefix the channel ID with "MPLA".
func parseArtistRow(row listItem) (music.Item, bool) {
	endpoint := row.browseEndpoint()
	if endpoint == nil {
		return music.Item{}, false
	}
	artistID := strings.TrimPrefix(endpoint.BrowseID, libraryArtistPrefix)
	name := joinText(row.column(0))
	if !strings.HasPrefix(artistID, artistChannelIDPrefix) || name == "" {
		return music.Item{}, false
	}
	return music.Item{
		Kind:       music.ItemKindArtist,
		ID:         artistID,
		Title:      name,
		Subtitle:   joinText(row.column(1)),
		Thumbnails: parseThumbnails(row.thumbnails()),
	}, true
}
