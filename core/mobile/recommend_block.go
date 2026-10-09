package mobile

import "github.com/letsmoe/drift/core/library"

// BlockRecommendation keeps a track ("track", video ID) or an artist ("artist",
// browse ID) out of recommendations and autoplay.
func (core *Core) BlockRecommendation(kind, id, name string) error {
	if core.engine == nil {
		return errRecommendUnavailable
	}
	return core.engine.Block(core.ctx, library.BlockKind(kind), id, name)
}

// UnblockRecommendation lifts a "Don't recommend" choice.
func (core *Core) UnblockRecommendation(kind, id string) error {
	if core.engine == nil {
		return errRecommendUnavailable
	}
	return core.engine.Unblock(core.ctx, library.BlockKind(kind), id)
}

// RecommendationBlocks returns the "Don't recommend" choices as library.RecommendationBlock JSON.
func (core *Core) RecommendationBlocks() (string, error) {
	lib, err := core.requireLibrary()
	if err != nil {
		return "", err
	}
	blocks, err := lib.RecommendationBlocks(core.ctx)
	return encodeResult(blocks, err)
}
