package recommend

import (
	"context"

	"github.com/letsmoe/drift-music/core/music"
)

// sourcePlan is what one ranking pass asks the catalog for and what came
// back. Every fetch task writes its own slot, so tasks never share one.
type sourcePlan struct {
	seeds     []Seed
	artistIDs []string
	queries   []string

	radios      [][]music.Track
	artistPages []*music.ArtistPage
	searches    [][]music.Track
	home        *music.HomePage

	exploreArtists []music.Item
	exploreSongs   [][]music.Track
}

// newSourcePlan picks the seeds, top artists and searches to fetch for.
func newSourcePlan(profile *Profile, queries []string) *sourcePlan {
	artistIDs := TopEntities(profile.Artists, topArtistLimit)
	return &sourcePlan{
		seeds:       profile.Seeds,
		artistIDs:   artistIDs,
		queries:     queries,
		radios:      make([][]music.Track, len(profile.Seeds)),
		artistPages: make([]*music.ArtistPage, len(artistIDs)),
		searches:    make([][]music.Track, len(queries)),
	}
}

// primaryTasks are the requests that need nothing but the profile: seed
// radios, top artist pages, recent searches and the home feed.
func (plan *sourcePlan) primaryTasks(catalog *cachedCatalog) []fetchTask {
	tasks := []fetchTask{plan.homeTask(catalog)}
	for index, seed := range plan.seeds {
		tasks = append(tasks, plan.radioTask(catalog, index, seed.VideoID))
	}
	for index, artistID := range plan.artistIDs {
		tasks = append(tasks, plan.artistTask(catalog, index, artistID))
	}
	for index, query := range plan.queries {
		tasks = append(tasks, plan.searchTask(catalog, index, query))
	}
	return tasks
}

// homeTask fetches the first page of YouTube's home feed.
func (plan *sourcePlan) homeTask(catalog *cachedCatalog) fetchTask {
	return fetchTask{name: SourceYouTube, run: func(ctx context.Context) error {
		page, err := catalog.home(ctx)
		plan.home = &page
		return err
	}}
}

// radioTask fetches the radio of seed number index.
func (plan *sourcePlan) radioTask(catalog *cachedCatalog, index int, videoID string) fetchTask {
	return fetchTask{name: SourceRadio + ":" + videoID, run: func(ctx context.Context) error {
		tracks, err := catalog.radio(ctx, videoID)
		plan.radios[index] = tracks
		return err
	}}
}

// artistTask fetches the page of top artist number index.
func (plan *sourcePlan) artistTask(catalog *cachedCatalog, index int, artistID string) fetchTask {
	return fetchTask{name: SourceArtist + ":" + artistID, run: func(ctx context.Context) error {
		page, err := catalog.artist(ctx, artistID)
		if err == nil {
			plan.artistPages[index] = &page
		}
		return err
	}}
}

// searchTask fetches the songs of recent search number index.
func (plan *sourcePlan) searchTask(catalog *cachedCatalog, index int, query string) fetchTask {
	return fetchTask{name: SourceSearch + ":" + query, run: func(ctx context.Context) error {
		tracks, err := catalog.searchSongs(ctx, query)
		plan.searches[index] = tracks
		return err
	}}
}

// chooseExploreArtists picks the related artists, from the "fans might also
// like" shelves of the top artists' pages, that the profile does not know yet.
func (plan *sourcePlan) chooseExploreArtists(profile *Profile) {
	chosen := map[string]bool{}
	for _, page := range plan.artistPages {
		if page == nil {
			continue
		}
		for _, related := range relatedArtists(*page) {
			_, known := profile.Artists[related.ID]
			if known || chosen[related.ID] || len(plan.exploreArtists) >= exploreArtistLimit {
				continue
			}
			chosen[related.ID] = true
			plan.exploreArtists = append(plan.exploreArtists, related)
		}
	}
	plan.exploreSongs = make([][]music.Track, len(plan.exploreArtists))
}

// relatedArtists lists the artists of the page's shelves made only of artists.
func relatedArtists(page music.ArtistPage) []music.Item {
	related := []music.Item{}
	for _, shelf := range page.Shelves {
		if isArtistShelf(shelf) {
			related = append(related, shelf.Items...)
		}
	}
	return related
}

// isArtistShelf reports whether every item of a non-empty shelf is an artist.
func isArtistShelf(shelf music.Shelf) bool {
	if len(shelf.Items) == 0 {
		return false
	}
	for _, item := range shelf.Items {
		if item.Kind != music.ItemKindArtist || item.ID == "" {
			return false
		}
	}
	return true
}

// exploreTasks fetch the top songs of the chosen explore artists.
func (plan *sourcePlan) exploreTasks(catalog *cachedCatalog) []fetchTask {
	tasks := make([]fetchTask, 0, len(plan.exploreArtists))
	for index, artist := range plan.exploreArtists {
		tasks = append(tasks, fetchTask{name: SourceExplore + ":" + artist.ID, run: func(ctx context.Context) error {
			page, err := catalog.artist(ctx, artist.ID)
			plan.exploreSongs[index] = page.TopSongs
			return err
		}})
	}
	return tasks
}

// candidates turns everything fetched into candidates.
func (plan *sourcePlan) candidates(profile *Profile) []candidate {
	all := plan.radioCandidates()
	all = append(all, plan.artistCandidates(profile)...)
	all = append(all, plan.exploreCandidates()...)
	all = append(all, plan.searchCandidates()...)
	return append(all, plan.homeCandidates()...)
}

// radioCandidates proposes the radio tracks of the seeds. A stronger seed
// lends its tracks a higher prior.
func (plan *sourcePlan) radioCandidates() []candidate {
	if len(plan.seeds) == 0 {
		return nil
	}
	strongestWeight := plan.seeds[0].Weight
	candidates := []candidate{}
	for index, seed := range plan.seeds {
		seedFactor := 1 - seedWeightInfluence + seedWeightInfluence*seed.Weight/strongestWeight
		tracks := withoutTrack(plan.radios[index], seed.VideoID, radioTracksPerSeed)
		prior := sourcePriors[SourceRadio] * seedFactor
		candidates = append(candidates, trackCandidates(tracks, SourceRadio, prior, relatedReason(seed))...)
	}
	return candidates
}

// withoutTrack drops one track from the list and keeps at most limit.
func withoutTrack(tracks []music.Track, videoID string, limit int) []music.Track {
	kept := make([]music.Track, 0, min(limit, len(tracks)))
	for _, track := range tracks {
		if track.VideoID != videoID && len(kept) < limit {
			kept = append(kept, track)
		}
	}
	return kept
}

// artistCandidates proposes the top songs of the top artists, each artist's
// prior scaled by its affinity.
func (plan *sourcePlan) artistCandidates(profile *Profile) []candidate {
	candidates := []candidate{}
	for index, page := range plan.artistPages {
		if page == nil {
			continue
		}
		artistID := plan.artistIDs[index]
		name := page.Name
		if name == "" {
			name = profile.ArtistNames[artistID]
		}
		prior := sourcePriors[SourceArtist] * profile.Artists[artistID]
		tracks := firstTracks(page.TopSongs, tracksPerArtist)
		candidates = append(candidates, trackCandidates(tracks, SourceArtist, prior, artistReason(artistID, name))...)
	}
	return candidates
}

// exploreCandidates proposes the top songs of the unknown related artists.
func (plan *sourcePlan) exploreCandidates() []candidate {
	candidates := []candidate{}
	for index, artist := range plan.exploreArtists {
		tracks := firstTracks(plan.exploreSongs[index], tracksPerExploreArtist)
		candidates = append(candidates, trackCandidates(tracks, SourceExplore, sourcePriors[SourceExplore], exploreReason(artist.ID, artist.Title))...)
	}
	return candidates
}

// searchCandidates proposes the songs recent searches found.
func (plan *sourcePlan) searchCandidates() []candidate {
	candidates := []candidate{}
	for index, query := range plan.queries {
		tracks := firstTracks(plan.searches[index], tracksPerSearch)
		candidates = append(candidates, trackCandidates(tracks, SourceSearch, sourcePriors[SourceSearch], searchReason(query))...)
	}
	return candidates
}

// homeCandidates proposes the song items of YouTube's home feed: the
// cold-start source.
func (plan *sourcePlan) homeCandidates() []candidate {
	if plan.home == nil {
		return nil
	}
	tracks := []music.Track{}
	for _, shelf := range plan.home.Shelves {
		tracks = append(tracks, shelfSongs(shelf)...)
	}
	tracks = firstTracks(tracks, homeTrackLimit)
	return trackCandidates(tracks, SourceYouTube, sourcePriors[SourceYouTube], youTubeReason())
}

// shelfSongs returns the playable song items of a shelf.
func shelfSongs(shelf music.Shelf) []music.Track {
	songs := []music.Track{}
	for _, item := range shelf.Items {
		if item.Kind == music.ItemKindSong && item.Track != nil {
			songs = append(songs, *item.Track)
		}
	}
	return songs
}

// firstTracks keeps at most limit tracks.
func firstTracks(tracks []music.Track, limit int) []music.Track {
	return tracks[:min(limit, len(tracks))]
}
