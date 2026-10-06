package recommend

import (
	"context"

	"github.com/neoworks/otter-music/core/music"
)

// Extend implements queue.Extender: autoplay continues with the radio of the
// track that just played, minus disliked, blocked, queued and already played
// tracks, ranked by taste and re-ranked for artist diversity.
func (e *Engine) Extend(ctx context.Context, upcoming, history []music.Track) ([]music.Track, error) {
	if len(history) == 0 {
		return nil, nil
	}
	last := history[len(history)-1]
	radio, err := e.catalog.radio(ctx, last.VideoID)
	if err != nil {
		return nil, err
	}
	profile, _, err := e.profileAt(ctx, e.now())
	if err != nil {
		return nil, err
	}
	seed := Seed{VideoID: last.VideoID, Title: last.Title}
	pool := newCandidatePool()
	pool.add(trackCandidates(radio, SourceRadio, sourcePriors[SourceRadio], relatedReason(seed))...)
	scored := scoreExtension(pool.list(), profile, videoIDSet(upcoming, history))
	return tracksOf(rerank(scored), extendLimit), nil
}

// scoreExtension scores radio candidates by taste and radio rank, skipping
// disliked, blocked and excluded tracks.
func scoreExtension(candidates []*candidate, profile *Profile, excluded map[string]bool) []Item {
	items := make([]Item, 0, len(candidates))
	for _, proposed := range candidates {
		if excluded[proposed.track.VideoID] || profile.excludes(proposed.track) {
			continue
		}
		items = append(items, itemOf(proposed, baseScore(proposed, profile)))
	}
	sortByScore(items)
	return items
}

// videoIDSet collects the video IDs of the track lists.
func videoIDSet(lists ...[]music.Track) map[string]bool {
	ids := map[string]bool{}
	for _, tracks := range lists {
		for _, track := range tracks {
			ids[track.VideoID] = true
		}
	}
	return ids
}

// tracksOf returns the tracks of the first limit items.
func tracksOf(items []Item, limit int) []music.Track {
	tracks := make([]music.Track, 0, min(limit, len(items)))
	for _, item := range items[:min(limit, len(items))] {
		tracks = append(tracks, item.Track)
	}
	return tracks
}
