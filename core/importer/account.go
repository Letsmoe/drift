package importer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/letsmoe/drift/core/innertube"
	"github.com/letsmoe/drift/core/library"
	"github.com/letsmoe/drift/core/music"
)

// youtubeSource is the import_playlist source name of YouTube Music playlists.
const youtubeSource = "youtube"

// ErrAccountUnavailable means the importer was built without a YouTube Music account reader.
var ErrAccountUnavailable = errors.New("importer: YouTube Music account import is not available")

// Account reads the library of the signed-in YouTube Music user. The
// innertube account client satisfies it.
type Account interface {
	LikedSongs(ctx context.Context) ([]music.Track, error)
	LibraryPlaylists(ctx context.Context) ([]music.Item, error)
	PlaylistTracks(ctx context.Context, playlistID string) ([]music.Track, error)
	LibraryAlbums(ctx context.Context) ([]music.Item, error)
	LibraryArtists(ctx context.Context) ([]music.Item, error)
	LibraryPodcasts(ctx context.Context) ([]music.Item, error)
}

// AccountProvider returns the reader of the current session; it fails when
// nobody is signed in.
type AccountProvider func(ctx context.Context) (Account, error)

// WithAccount enables SourceYouTubeAccount.
func WithAccount(provider AccountProvider) Option {
	return func(importer *Importer) {
		importer.account = provider
	}
}

// playlistKeyPrefix precedes a playlist ID in a browse ID.
const playlistKeyPrefix = "VL"

// autoPlaylistKeys are YouTube's generated playlists: "Liked Music" is
// imported as likes and "Episodes for later" has no local counterpart.
var autoPlaylistKeys = []string{"LM", "SE"}

var trackCountPattern = regexp.MustCompile(`(\d[\d.,]*)\s+(?:song|track)s?\b`)

func (importer *Importer) accountReader(ctx context.Context) (Account, error) {
	if importer.account == nil {
		return nil, ErrAccountUnavailable
	}
	return importer.account(ctx)
}

// previewAccount counts the account's library. It reads every liked song to
// count them, but writes nothing.
func (importer *Importer) previewAccount(ctx context.Context) (Summary, error) {
	reader, err := importer.accountReader(ctx)
	if err != nil {
		return Summary{}, err
	}
	liked, err := reader.LikedSongs(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("importer: liked songs: %w", err)
	}
	playlists, err := importablePlaylists(ctx, reader)
	if err != nil {
		return Summary{}, err
	}
	albums, err := reader.LibraryAlbums(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("importer: albums: %w", err)
	}
	artists, err := reader.LibraryArtists(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("importer: artists: %w", err)
	}
	podcasts, err := reader.LibraryPodcasts(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("importer: podcasts: %w", err)
	}
	summaries := make([]PlaylistSummary, 0, len(playlists))
	for _, playlist := range playlists {
		summaries = append(summaries, PlaylistSummary{Key: playlistKey(playlist), Name: playlist.Title, TrackCount: trackCount(playlist.Subtitle)})
	}
	return Summary{Liked: len(liked), Playlists: summaries, Albums: len(albums), Artists: len(artists), Shows: len(podcasts)}, nil
}

// importablePlaylists lists the library playlists minus YouTube's generated ones.
func importablePlaylists(ctx context.Context, reader Account) ([]music.Item, error) {
	all, err := reader.LibraryPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("importer: playlists: %w", err)
	}
	playlists := make([]music.Item, 0, len(all))
	for _, playlist := range all {
		if !slices.Contains(autoPlaylistKeys, playlistKey(playlist)) {
			playlists = append(playlists, playlist)
		}
	}
	return playlists, nil
}

func playlistKey(playlist music.Item) string {
	return strings.TrimPrefix(playlist.ID, playlistKeyPrefix)
}

// trackCount reads "12 songs" out of a playlist subtitle; 0 means unknown.
func trackCount(subtitle string) int {
	match := trackCountPattern.FindStringSubmatch(subtitle)
	if match == nil {
		return 0
	}
	count, err := strconv.Atoi(strings.NewReplacer(",", "", ".", "").Replace(match[1]))
	if err != nil {
		return 0
	}
	return count
}

// fetchedPlaylist is a library playlist with its tracks.
type fetchedPlaylist struct {
	key    string
	name   string
	tracks []music.Track
}

// accountContent is what a run read from the account.
type accountContent struct {
	liked     []music.Track
	playlists []fetchedPlaylist
	albums    []music.Item
	artists   []music.Item
	podcasts  []music.Item
	// failedPlaylists are the names of playlists that could not be read.
	failedPlaylists []string
}

func (importer *Importer) runAccount(ctx context.Context, options Options, onProgress func(Progress)) (Result, error) {
	if onProgress == nil {
		onProgress = func(Progress) {}
	}
	reader, err := importer.accountReader(ctx)
	if err != nil {
		return Result{}, err
	}
	content, err := importer.fetchAccount(ctx, reader, options, onProgress)
	if err != nil {
		return Result{}, err
	}
	onProgress(Progress{Phase: PhaseSaving, Done: 0, Total: 1})
	result, err := importer.applyAccount(ctx, content)
	if err != nil {
		return Result{}, err
	}
	err = importer.library.MarkYouTubeSynced(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("importer: %w", err)
	}
	onProgress(Progress{Phase: PhaseSaving, Done: 1, Total: 1})
	return result, nil
}

// fetchProgress counts the reads of a run: liked songs, each playlist and
// each saved-item list are one step.
type fetchProgress struct {
	report func(Progress)
	done   int
	total  int
}

func (progress *fetchProgress) step(phase, current string) {
	progress.report(Progress{Phase: phase, Done: progress.done, Total: progress.total, Current: current})
	progress.done++
}

func (importer *Importer) fetchAccount(ctx context.Context, reader Account, options Options, onProgress func(Progress)) (accountContent, error) {
	playlists, err := selectedPlaylists(ctx, reader, options)
	if err != nil {
		return accountContent{}, err
	}
	progress := &fetchProgress{report: onProgress, total: len(playlists) + countTrue(options.Liked, options.Albums, options.Artists, options.Shows)}
	content := accountContent{}
	if options.Liked {
		progress.step(PhaseTracks, "Liked songs")
		content.liked, err = reader.LikedSongs(ctx)
		if err != nil {
			return accountContent{}, fmt.Errorf("importer: liked songs: %w", err)
		}
	}
	err = importer.fetchPlaylists(ctx, reader, playlists, progress, &content)
	if err != nil {
		return accountContent{}, err
	}
	return content, fetchSavedItems(ctx, reader, options, progress, &content)
}

func countTrue(flags ...bool) int {
	count := 0
	for _, flag := range flags {
		if flag {
			count++
		}
	}
	return count
}

func selectedPlaylists(ctx context.Context, reader Account, options Options) ([]music.Item, error) {
	if len(options.Playlists) == 0 {
		return nil, nil
	}
	all, err := importablePlaylists(ctx, reader)
	if err != nil {
		return nil, err
	}
	selected := make([]music.Item, 0, len(options.Playlists))
	for _, playlist := range all {
		if slices.Contains(options.Playlists, playlistKey(playlist)) {
			selected = append(selected, playlist)
		}
	}
	return selected, nil
}

// fetchPlaylists reads each playlist; one that cannot be read is reported and
// skipped, except when the session itself is gone.
func (importer *Importer) fetchPlaylists(ctx context.Context, reader Account, playlists []music.Item, progress *fetchProgress, content *accountContent) error {
	for _, playlist := range playlists {
		progress.step(PhaseTracks, playlist.Title)
		tracks, err := reader.PlaylistTracks(ctx, playlistKey(playlist))
		if ctx.Err() != nil {
			return fmt.Errorf("importer: %w", ctx.Err())
		}
		if errors.Is(err, innertube.ErrSessionExpired) {
			return err
		}
		if err != nil {
			importer.logger.Warn("read playlist", "playlist", playlist.Title, "error", err)
			content.failedPlaylists = append(content.failedPlaylists, playlist.Title)
			continue
		}
		content.playlists = append(content.playlists, fetchedPlaylist{key: playlistKey(playlist), name: playlist.Title, tracks: tracks})
	}
	return nil
}

func fetchSavedItems(ctx context.Context, reader Account, options Options, progress *fetchProgress, content *accountContent) error {
	var err error
	if options.Albums {
		progress.step(PhaseAlbums, "Albums")
		content.albums, err = reader.LibraryAlbums(ctx)
		if err != nil {
			return fmt.Errorf("importer: albums: %w", err)
		}
	}
	if options.Artists {
		progress.step(PhaseArtists, "Artists")
		content.artists, err = reader.LibraryArtists(ctx)
		if err != nil {
			return fmt.Errorf("importer: artists: %w", err)
		}
	}
	if options.Shows {
		progress.step(PhaseShows, "Podcasts")
		content.podcasts, err = reader.LibraryPodcasts(ctx)
		if err != nil {
			return fmt.Errorf("importer: podcasts: %w", err)
		}
	}
	return nil
}

func (importer *Importer) applyAccount(ctx context.Context, content accountContent) (Result, error) {
	result := Result{Playlists: []PlaylistResult{}, Unmatched: []UnmatchedItem{}}
	err := importer.applyAccountLikes(ctx, content.liked, &result)
	if err != nil {
		return Result{}, err
	}
	for _, playlist := range content.playlists {
		err = importer.applyAccountPlaylist(ctx, playlist, &result)
		if err != nil {
			return Result{}, err
		}
	}
	for _, name := range content.failedPlaylists {
		result.Unmatched = append(result.Unmatched, UnmatchedItem{Kind: "playlist", Title: name, Artists: []string{}})
	}
	err = importer.saveAll(ctx, content.albums, &result.Albums)
	if err != nil {
		return Result{}, err
	}
	err = importer.saveAll(ctx, content.artists, &result.Artists)
	if err != nil {
		return Result{}, err
	}
	return result, importer.saveAll(ctx, content.podcasts, &result.Shows)
}

// applyAccountLikes likes oldest first: YouTube lists the newest like first,
// and the library shows the most recent like first.
func (importer *Importer) applyAccountLikes(ctx context.Context, liked []music.Track, result *Result) error {
	oldestFirst := slices.Clone(liked)
	slices.Reverse(oldestFirst)
	result.Liked.Matched = len(liked)
	return importer.likeAll(ctx, oldestFirst)
}

func (importer *Importer) applyAccountPlaylist(ctx context.Context, playlist fetchedPlaylist, result *Result) error {
	playlistID, err := importer.library.ImportPlaylist(ctx, library.ImportedPlaylist{
		Source: youtubeSource, SourceKey: playlist.key, Name: playlist.name, Tracks: playlist.tracks,
	})
	if err != nil {
		return fmt.Errorf("importer: %w", err)
	}
	result.Playlists = append(result.Playlists, PlaylistResult{
		Key: playlist.key, Name: playlist.name, PlaylistID: playlistID, Matched: len(playlist.tracks),
	})
	return nil
}

func (importer *Importer) saveAll(ctx context.Context, items []music.Item, count *CategoryCount) error {
	for _, item := range items {
		err := importer.library.Save(ctx, item)
		if err != nil {
			return fmt.Errorf("importer: %w", err)
		}
		count.Matched++
	}
	return nil
}
