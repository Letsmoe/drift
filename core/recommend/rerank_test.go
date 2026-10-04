package recommend

import (
	"fmt"
	"testing"
)

// scoredItems builds descending-score items from (source, artist, album) triples.
func scoredItems(specs [][3]string) []Item {
	items := make([]Item, 0, len(specs))
	for index, spec := range specs {
		track := makeTrack(fmt.Sprintf("track%02d", index), spec[1], spec[2])
		items = append(items, Item{Track: track, Source: spec[0], Score: float64(len(specs) - index)})
	}
	return items
}

// TestRerankCapsArtistsPerPage checks at most two tracks per artist on a page.
func TestRerankCapsArtistsPerPage(t *testing.T) {
	specs := [][3]string{}
	for index := range 30 {
		artist := "artist:other" + fmt.Sprint(index)
		if index < 6 {
			artist = "artist:big"
		}
		specs = append(specs, [3]string{SourceArtist + fmt.Sprint(index%3), artist, ""})
	}

	ranked := rerank(scoredItems(specs))
	counts := map[string]int{}
	for _, item := range ranked[:pageSize] {
		counts[item.Track.Artists[0].ID]++
	}
	if counts["artist:big"] > maxPerArtistPerPage {
		t.Errorf("artist:big appears %d times on the first page", counts["artist:big"])
	}
	if len(ranked) != 30 {
		t.Errorf("rerank dropped items: %d of 30", len(ranked))
	}
}

// TestRerankCapsAlbumsPerPage checks at most three tracks per album on a page.
func TestRerankCapsAlbumsPerPage(t *testing.T) {
	specs := [][3]string{}
	for index := range 30 {
		specs = append(specs, [3]string{SourceRadio, fmt.Sprintf("artist:%d", index), "album:same"})
		if index >= 8 {
			specs[index][2] = fmt.Sprintf("album:%d", index)
		}
	}

	ranked := rerank(scoredItems(specs))
	count := 0
	for _, item := range ranked[:pageSize] {
		if item.Track.Album.ID == "album:same" {
			count++
		}
	}
	if count > maxPerAlbumPerPage {
		t.Errorf("album:same appears %d times on the first page", count)
	}
}

// TestRerankReservesExplorationSlots checks that every seventh slot is exploration.
func TestRerankReservesExplorationSlots(t *testing.T) {
	specs := [][3]string{}
	for index := range 20 {
		specs = append(specs, [3]string{SourceRadio, fmt.Sprintf("artist:main%d", index), ""})
	}
	for index := range 5 {
		specs = append(specs, [3]string{SourceExplore, fmt.Sprintf("artist:explore%d", index), ""})
	}
	// Exploration items score lowest, yet must surface in slots 7, 14, 21.
	ranked := rerank(scoredItems(specs))
	for _, slot := range []int{6, 13, 20} {
		if ranked[slot].Source != SourceExplore {
			t.Errorf("slot %d holds %q, want exploration", slot+1, ranked[slot].Source)
		}
	}
}

// TestRerankBreaksSourceRuns checks that no source runs three times in a row when others exist.
func TestRerankBreaksSourceRuns(t *testing.T) {
	specs := [][3]string{}
	for index := range 8 {
		specs = append(specs, [3]string{SourceRadio, fmt.Sprintf("artist:radio%d", index), ""})
	}
	for index := range 8 {
		specs = append(specs, [3]string{SourceSearch, fmt.Sprintf("artist:search%d", index), ""})
	}

	// Once radio runs out only search is left and the run rule yields.
	ranked := rerank(scoredItems(specs))[:10]
	run := 1
	for index := 1; index < len(ranked); index++ {
		if ranked[index].Source == ranked[index-1].Source {
			run++
		} else {
			run = 1
		}
		if run > maxSameSourceRun {
			t.Fatalf("source %q runs %d times at slot %d", ranked[index].Source, run, index+1)
		}
	}
}

// TestRerankDropsWeakestConstraintsFirst checks that a single-source feed still fills up.
func TestRerankDropsWeakestConstraintsFirst(t *testing.T) {
	specs := [][3]string{}
	for index := range 10 {
		specs = append(specs, [3]string{SourceRadio, fmt.Sprintf("artist:%d", index), "album:same"})
	}

	ranked := rerank(scoredItems(specs))
	if len(ranked) != 10 {
		t.Fatalf("got %d items, want all 10", len(ranked))
	}
	for index, item := range ranked {
		if item.Track.VideoID != fmt.Sprintf("track%02d", index) {
			t.Errorf("slot %d holds %s: with nothing else to pick the score order must hold", index+1, item.Track.VideoID)
		}
	}
}
