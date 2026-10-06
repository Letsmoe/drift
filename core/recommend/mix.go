package recommend

import (
	"context"
	"fmt"
	"sync"

	"github.com/neoworks/otter-music/core/music"
)

// Mix implements queue.Recommender for magic shuffle: the radios of the seed
// tracks merged, minus disliked, blocked and excluded tracks, ranked by taste and
// re-ranked for artist diversity. It fails only when no seed radio loads.
func (e *Engine) Mix(ctx context.Context, seeds, exclude []music.Track, count int) ([]music.Track, error) {
	if len(seeds) == 0 || count <= 0 {
		return nil, nil
	}
	profile, _, err := e.profileAt(ctx, e.now())
	if err != nil {
		return nil, err
	}
	pool, err := e.seedRadioPool(ctx, seeds)
	if err != nil {
		return nil, err
	}
	scored := scoreExtension(pool.list(), profile, videoIDSet(seeds, exclude))
	return tracksOf(rerank(scored), count), nil
}

// seedRadioPool pools the radios of seeds, fetched concurrently. Failing radios
// are logged and skipped; the error is returned only when all of them failed.
func (e *Engine) seedRadioPool(ctx context.Context, seeds []music.Track) (*candidatePool, error) {
	radios := make([][]music.Track, len(seeds))
	var failures sync.Map
	tasks := make([]fetchTask, 0, len(seeds))
	for index, seed := range seeds {
		tasks = append(tasks, fetchTask{name: SourceRadio, run: func(ctx context.Context) error {
			radio, err := e.catalog.radio(ctx, seed.VideoID)
			if err != nil {
				failures.Store(index, err)
				return err
			}
			radios[index] = radio
			return nil
		}})
	}
	runLimited(ctx, e.logger, tasks)

	pool := newCandidatePool()
	var lastErr error
	for index, seed := range seeds {
		if failed, found := failures.Load(index); found {
			lastErr, _ = failed.(error)
			continue
		}
		reason := relatedReason(Seed{VideoID: seed.VideoID, Title: seed.Title})
		pool.add(trackCandidates(radios[index], SourceRadio, sourcePriors[SourceRadio], reason)...)
	}
	if len(pool.order) == 0 && lastErr != nil {
		return nil, fmt.Errorf("recommend: mix: %w", lastErr)
	}
	return pool, nil
}
