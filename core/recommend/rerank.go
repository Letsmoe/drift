package recommend

import (
	"slices"

	"github.com/letsmoe/drift-music/core/music"
)

// rerankState tracks the page being filled: diversity counters and the run of
// items from the same source.
type rerankState struct {
	slot          int
	artistCounts  map[string]int
	albumCounts   map[string]int
	lastSource    string
	sameSourceRun int
}

// constraint is one re-ranking rule an item may break.
type constraint func(item Item) bool

// newRerankState starts at the first slot of the first page.
func newRerankState() *rerankState {
	state := &rerankState{}
	state.resetPage()
	return state
}

// resetPage clears the per-page diversity counters.
func (state *rerankState) resetPage() {
	state.artistCounts = map[string]int{}
	state.albumCounts = map[string]int{}
}

// rerank turns the score-ordered list into the feed order: every seventh slot
// goes to exploration, no page holds more than a couple of tracks of one
// artist or album, and no source runs more than twice in a row. Items that
// don't fit a page are deferred, never dropped.
func rerank(scored []Item) []Item {
	main, exploration := splitExploration(scored)
	state := newRerankState()
	ranked := make([]Item, 0, len(scored))
	for len(main)+len(exploration) > 0 {
		var picked Item
		if len(main) == 0 || (state.isExplorationSlot() && len(exploration) > 0) {
			picked, exploration = takeAt(exploration, state.pickIndex(exploration))
		} else {
			picked, main = takeAt(main, state.pickIndex(main))
		}
		state.record(picked)
		ranked = append(ranked, picked)
	}
	return ranked
}

// splitExploration separates exploration items, keeping each list's order.
func splitExploration(scored []Item) (main, exploration []Item) {
	for _, item := range scored {
		if isExploration(item.Source) {
			exploration = append(exploration, item)
			continue
		}
		main = append(main, item)
	}
	return main, exploration
}

// takeAt removes and returns the item at index.
func takeAt(queue []Item, index int) (Item, []Item) {
	item := queue[index]
	return item, slices.Delete(queue, index, index+1)
}

// isExplorationSlot reports whether the next slot is reserved for exploration.
func (state *rerankState) isExplorationSlot() bool {
	return (state.slot+1)%explorationEvery == 0
}

// pickIndex returns the best-scored item that keeps every constraint. When
// none does (e.g. the feed holds a single source), constraints are dropped
// weakest first: the source run, then the album cap; the artist cap goes last.
func (state *rerankState) pickIndex(queue []Item) int {
	tiers := [][]constraint{
		{state.withinArtistCap, state.withinAlbumCap, state.continuesSourceRun},
		{state.withinArtistCap, state.withinAlbumCap},
		{state.withinArtistCap},
	}
	for _, constraints := range tiers {
		index := firstSatisfying(queue, constraints)
		if index >= 0 {
			return index
		}
	}
	return 0
}

// firstSatisfying returns the index of the first item keeping every
// constraint, or -1.
func firstSatisfying(queue []Item, constraints []constraint) int {
	for index, item := range queue {
		if satisfiesAll(item, constraints) {
			return index
		}
	}
	return -1
}

// satisfiesAll reports whether item keeps every constraint.
func satisfiesAll(item Item, constraints []constraint) bool {
	for _, keeps := range constraints {
		if !keeps(item) {
			return false
		}
	}
	return true
}

// artistIDsOf lists the IDs of a track's linked artists.
func artistIDsOf(track music.Track) []string {
	artists := linkedArtists(track)
	ids := make([]string, 0, len(artists))
	for _, artist := range artists {
		ids = append(ids, artist.ID)
	}
	return ids
}

// withinArtistCap reports whether the page still has room for the item's artists.
func (state *rerankState) withinArtistCap(item Item) bool {
	for _, artistID := range artistIDsOf(item.Track) {
		if state.artistCounts[artistID] >= maxPerArtistPerPage {
			return false
		}
	}
	return true
}

// withinAlbumCap reports whether the page still has room for the item's album.
func (state *rerankState) withinAlbumCap(item Item) bool {
	id := albumID(item.Track)
	return id == "" || state.albumCounts[id] < maxPerAlbumPerPage
}

// continuesSourceRun reports whether the item may follow the current run of
// same-source items.
func (state *rerankState) continuesSourceRun(item Item) bool {
	return item.Source != state.lastSource || state.sameSourceRun < maxSameSourceRun
}

// record fills the slot with item and advances, starting a fresh page when
// this one is full.
func (state *rerankState) record(item Item) {
	for _, artistID := range artistIDsOf(item.Track) {
		state.artistCounts[artistID]++
	}
	if id := albumID(item.Track); id != "" {
		state.albumCounts[id]++
	}
	state.recordRun(item)
	state.slot++
	if state.slot%pageSize == 0 {
		state.resetPage()
	}
}

// recordRun extends or restarts the same-source run.
func (state *rerankState) recordRun(item Item) {
	if item.Source == state.lastSource {
		state.sameSourceRun++
		return
	}
	state.lastSource = item.Source
	state.sameSourceRun = 1
}
