package importer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/spotify"
)

// ErrBusy is returned when an import is already running.
var ErrBusy = errors.New("importer: an import is already running")

// ErrUnknownSource means a Source has a kind the importer does not know.
var ErrUnknownSource = errors.New("importer: unknown source kind")

// Importer runs one import at a time.
type Importer struct {
	library      *library.Library
	catalog      Catalog
	logger       *slog.Logger
	fetcher      *spotify.Fetcher
	requestDelay time.Duration
	// account supplies the signed-in YouTube Music reader for SourceYouTubeAccount.
	account AccountProvider

	running sync.Mutex
	// cancelMutex guards cancelRunning, the cancel function of a Start-ed import.
	cancelMutex   sync.Mutex
	cancelRunning context.CancelFunc
}

// Option configures New.
type Option func(*Importer)

// WithFetcher replaces the public-link fetcher, for tests.
func WithFetcher(fetcher *spotify.Fetcher) Option {
	return func(importer *Importer) {
		importer.fetcher = fetcher
	}
}

// WithRequestDelay sets the pause after each catalog search of a worker.
func WithRequestDelay(delay time.Duration) Option {
	return func(importer *Importer) {
		importer.requestDelay = delay
	}
}

// New returns an Importer writing into lib and searching catalog.
func New(lib *library.Library, catalog Catalog, logger *slog.Logger, options ...Option) *Importer {
	importer := &Importer{
		library:      lib,
		catalog:      catalog,
		logger:       logger,
		fetcher:      spotify.NewFetcher(),
		requestDelay: defaultRequestDelay,
	}
	for _, option := range options {
		option(importer)
	}
	return importer
}

// Preview counts what a source holds. It matches nothing; a public link is
// fetched, a data export is only read.
func (importer *Importer) Preview(ctx context.Context, source Source) (Summary, error) {
	if source.Kind == SourceYouTubeAccount {
		return importer.previewAccount(ctx)
	}
	content, err := importer.load(ctx, source)
	if err != nil {
		return Summary{}, err
	}
	summary := Summary{
		Liked:     len(content.LikedTracks),
		Playlists: make([]PlaylistSummary, 0, len(content.Playlists)),
		Albums:    len(content.Albums),
		Artists:   len(content.Artists),
		Shows:     len(content.Shows),
	}
	for _, playlist := range content.Playlists {
		summary.Playlists = append(summary.Playlists, PlaylistSummary{Key: playlist.Key, Name: playlist.Name, TrackCount: len(playlist.Tracks)})
	}
	if source.Kind == SourceURL {
		summary.TrackLimit = spotify.PublicTrackLimit
	}
	return summary, nil
}

func (importer *Importer) load(ctx context.Context, source Source) (spotify.Library, error) {
	switch source.Kind {
	case SourceExport:
		return spotify.ParseExport(source.Location)
	case SourceURL:
		return importer.fetcher.Fetch(ctx, source.Location)
	default:
		return spotify.Library{}, fmt.Errorf("%w: %q", ErrUnknownSource, source.Kind)
	}
}

// Run imports the selected parts of a source. Matches are looked up in the
// library first and searched on YouTube Music otherwise; cancelling ctx
// stops before anything is written, keeping the matches found so far.
func (importer *Importer) Run(ctx context.Context, source Source, options Options, onProgress func(Progress)) (Result, error) {
	if !importer.running.TryLock() {
		return Result{}, ErrBusy
	}
	defer importer.running.Unlock()
	return importer.run(ctx, source, options, onProgress)
}

// Start runs an import in the background and reports its end through onDone.
// It returns ErrBusy at once when an import is running. Cancel stops it.
func (importer *Importer) Start(ctx context.Context, source Source, options Options, onProgress func(Progress), onDone func(Result, error)) error {
	if !importer.running.TryLock() {
		return ErrBusy
	}
	runCtx, cancel := context.WithCancel(ctx)
	importer.cancelMutex.Lock()
	importer.cancelRunning = cancel
	importer.cancelMutex.Unlock()
	go func() {
		result, err := importer.run(runCtx, source, options, onProgress)
		cancel()
		importer.running.Unlock()
		onDone(result, err)
	}()
	return nil
}

// Cancel stops the import started by Start; without one it does nothing.
func (importer *Importer) Cancel() {
	importer.cancelMutex.Lock()
	defer importer.cancelMutex.Unlock()
	if importer.cancelRunning != nil {
		importer.cancelRunning()
	}
}

func (importer *Importer) run(ctx context.Context, source Source, options Options, onProgress func(Progress)) (Result, error) {
	if onProgress == nil {
		onProgress = func(Progress) {}
	}
	if source.Kind == SourceYouTubeAccount {
		return importer.runAccount(ctx, options, onProgress)
	}
	content, err := importer.load(ctx, source)
	if err != nil {
		return Result{}, err
	}
	chosen := choose(content, options)
	outcomes, err := importer.matchAll(ctx, chosen, options, onProgress)
	if err != nil {
		return Result{}, err
	}
	onProgress(Progress{Phase: PhaseSaving, Done: 0, Total: 1})
	result, err := importer.apply(ctx, chosen, outcomes)
	if err != nil {
		return Result{}, err
	}
	onProgress(Progress{Phase: PhaseSaving, Done: 1, Total: 1})
	return result, nil
}

// choose narrows a source library to what the options select.
func choose(content spotify.Library, options Options) selection {
	chosen := selection{}
	if options.Liked {
		chosen.liked = content.LikedTracks
	}
	for _, playlist := range content.Playlists {
		if slices.Contains(options.Playlists, playlist.Key) {
			chosen.playlists = append(chosen.playlists, playlist)
		}
	}
	if options.Albums {
		chosen.albums = content.Albums
	}
	if options.Artists {
		chosen.artists = content.Artists
	}
	if options.Shows {
		chosen.shows = content.Shows
	}
	return chosen
}

// matchAll resolves every distinct item of the selection.
func (importer *Importer) matchAll(ctx context.Context, chosen selection, options Options, onProgress func(Progress)) (map[string]resolved, error) {
	lookups := importer.lookupsFor(chosen)
	runner := importer.newMatcher(len(lookups), options.RetryUnmatched, onProgress)
	return runner.resolve(ctx, lookups)
}

func (importer *Importer) lookupsFor(chosen selection) []lookup {
	seen := map[string]bool{}
	lookups := []lookup{}
	add := func(candidate lookup) {
		if seen[candidate.uri] {
			return
		}
		seen[candidate.uri] = true
		lookups = append(lookups, candidate)
	}
	for _, track := range allTracks(chosen) {
		add(trackLookup(importer.catalog, track))
	}
	for _, album := range chosen.albums {
		add(albumLookup(importer.catalog, album))
	}
	for _, artist := range chosen.artists {
		add(artistLookup(importer.catalog, artist))
	}
	for _, show := range chosen.shows {
		add(showLookup(importer.catalog, show))
	}
	return lookups
}

func allTracks(chosen selection) []spotify.Track {
	tracks := slices.Clone(chosen.liked)
	for _, playlist := range chosen.playlists {
		tracks = append(tracks, playlist.Tracks...)
	}
	return tracks
}

// apply writes the matches into the library and counts the outcome.
func (importer *Importer) apply(ctx context.Context, chosen selection, outcomes map[string]resolved) (Result, error) {
	result := Result{Playlists: []PlaylistResult{}, Unmatched: []UnmatchedItem{}}
	err := importer.applyLiked(ctx, chosen.liked, outcomes, &result)
	if err != nil {
		return Result{}, err
	}
	for _, playlist := range chosen.playlists {
		err = importer.applyPlaylist(ctx, playlist, outcomes, &result)
		if err != nil {
			return Result{}, err
		}
	}
	err = importer.applySaved(ctx, chosen, outcomes, &result)
	if err != nil {
		return Result{}, err
	}
	return result, nil
}

func (importer *Importer) applyLiked(ctx context.Context, tracks []spotify.Track, outcomes map[string]resolved, result *Result) error {
	matched := []music.Track{}
	for _, track := range tracks {
		outcome := outcomes[track.Key()]
		if !outcome.matched() {
			result.Liked.Unmatched++
			result.Unmatched = append(result.Unmatched, unmatchedTrack(track, ""))
			continue
		}
		result.Liked.Matched++
		matched = append(matched, *outcome.track)
	}
	return importer.likeAll(ctx, matched)
}

// likeAll likes tracks that are not liked yet, so a re-import keeps the
// order of earlier likes.
func (importer *Importer) likeAll(ctx context.Context, tracks []music.Track) error {
	videoIDs := make([]string, 0, len(tracks))
	for _, track := range tracks {
		videoIDs = append(videoIDs, track.VideoID)
	}
	ratings, err := importer.library.Ratings(ctx, videoIDs)
	if err != nil {
		return fmt.Errorf("importer: %w", err)
	}
	for _, track := range tracks {
		if ratings[track.VideoID] == library.RatingLike {
			continue
		}
		err = importer.library.SetRating(ctx, track, music.ItemKindSong, library.RatingLike)
		if err != nil {
			return fmt.Errorf("importer: like %s: %w", track.VideoID, err)
		}
	}
	return nil
}

func (importer *Importer) applyPlaylist(ctx context.Context, playlist spotify.Playlist, outcomes map[string]resolved, result *Result) error {
	summary := PlaylistResult{Key: playlist.Key, Name: playlist.Name}
	tracks := []music.Track{}
	for _, track := range playlist.Tracks {
		outcome := outcomes[track.Key()]
		if !outcome.matched() {
			summary.Unmatched++
			result.Unmatched = append(result.Unmatched, unmatchedTrack(track, playlist.Name))
			continue
		}
		summary.Matched++
		tracks = append(tracks, *outcome.track)
	}
	playlistID, err := importer.library.ImportPlaylist(ctx, library.ImportedPlaylist{
		Source: matchSource, SourceKey: playlist.Key, Name: playlist.Name, Description: playlist.Description, Tracks: tracks,
	})
	if err != nil {
		return fmt.Errorf("importer: %w", err)
	}
	summary.PlaylistID = playlistID
	result.Playlists = append(result.Playlists, summary)
	return nil
}

func unmatchedTrack(track spotify.Track, playlistName string) UnmatchedItem {
	return UnmatchedItem{Kind: "track", Title: track.Title, Artists: track.Artists, Playlist: playlistName}
}

func (importer *Importer) applySaved(ctx context.Context, chosen selection, outcomes map[string]resolved, result *Result) error {
	for _, album := range chosen.albums {
		err := importer.saveItem(ctx, outcomes[album.Key()], &result.Albums, func() UnmatchedItem {
			return UnmatchedItem{Kind: "album", Title: album.Title, Artists: []string{album.Artist}}
		}, result)
		if err != nil {
			return err
		}
	}
	for _, artist := range chosen.artists {
		err := importer.saveItem(ctx, outcomes[artist.Key()], &result.Artists, func() UnmatchedItem {
			return UnmatchedItem{Kind: "artist", Title: artist.Name, Artists: []string{}}
		}, result)
		if err != nil {
			return err
		}
	}
	for _, show := range chosen.shows {
		err := importer.saveItem(ctx, outcomes[show.Key()], &result.Shows, func() UnmatchedItem {
			return UnmatchedItem{Kind: "show", Title: show.Name, Artists: []string{show.Publisher}}
		}, result)
		if err != nil {
			return err
		}
	}
	return nil
}

// saveItem saves a matched album, artist or show, or records it as unmatched.
func (importer *Importer) saveItem(ctx context.Context, outcome resolved, count *CategoryCount, unmatched func() UnmatchedItem, result *Result) error {
	if !outcome.matched() {
		count.Unmatched++
		result.Unmatched = append(result.Unmatched, unmatched())
		return nil
	}
	count.Matched++
	err := importer.library.Save(ctx, *outcome.item)
	if err != nil {
		return fmt.Errorf("importer: %w", err)
	}
	return nil
}
