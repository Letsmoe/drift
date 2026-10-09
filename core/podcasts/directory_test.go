package podcasts

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/letsmoe/drift/core/music"
)

type fakeCatalog struct {
	feed   music.HomePage
	shows  map[string]music.PodcastPage
	mutex  sync.Mutex
	opened map[string]int
}

func (catalog *fakeCatalog) PodcastsHome(context.Context) (music.HomePage, error) {
	return catalog.feed, nil
}

func (catalog *fakeCatalog) HomeContinuation(context.Context, string) (music.HomePage, error) {
	return catalog.feed, nil
}

func (catalog *fakeCatalog) Podcast(_ context.Context, browseID string) (music.PodcastPage, error) {
	catalog.mutex.Lock()
	catalog.opened[browseID]++
	catalog.mutex.Unlock()
	page, found := catalog.shows[browseID]
	if !found {
		return music.PodcastPage{}, errors.New("not found")
	}
	return page, nil
}

func episode(videoID, showID string) music.Item {
	track := music.Track{VideoID: videoID, Album: &music.Album{ID: showID}}
	return music.Item{Kind: music.ItemKindEpisode, ID: videoID, Track: &track}
}

func show(title string) music.PodcastPage {
	return music.PodcastPage{Title: title, Author: music.Artist{Name: title + " Author"}, Description: "About " + title}
}

func TestShowsReplaceEpisodesWithTheirShows(t *testing.T) {
	catalog := &fakeCatalog{
		feed: music.HomePage{
			Shelves: []music.Shelf{
				{Title: "Comedy", Items: []music.Item{episode("e1", "MPSPa"), episode("e2", "MPSPb"), episode("e3", "MPSPa")}},
				{Title: "Gone", Items: []music.Item{episode("e4", "MPSPmissing")}},
			},
			Continuation: "next",
		},
		shows:  map[string]music.PodcastPage{"MPSPa": show("A"), "MPSPb": show("B")},
		opened: map[string]int{},
	}
	directory := New(catalog, slog.New(slog.DiscardHandler))

	page, err := directory.Shows(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if page.Continuation != "next" || len(page.Shelves) != 1 {
		t.Fatalf("page = %+v, want only the Comedy shelf", page)
	}
	items := page.Shelves[0].Items
	if len(items) != 2 || items[0].ID != "MPSPa" || items[1].ID != "MPSPb" {
		t.Fatalf("items = %+v, want A then B once each", items)
	}
	first := items[0]
	if first.Kind != music.ItemKindPodcast || first.Title != "A" || first.Subtitle != "A Author" || first.Description != "About A" {
		t.Errorf("show item = %+v", first)
	}

	if _, err := directory.MoreShows(t.Context(), "next"); err != nil {
		t.Fatal(err)
	}
	if catalog.opened["MPSPa"] != 1 {
		t.Errorf("show A opened %d times, want 1 (cached)", catalog.opened["MPSPa"])
	}
}
