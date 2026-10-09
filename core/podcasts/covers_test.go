package podcasts

import (
	"context"
	"log/slog"
	"testing"

	"github.com/neoworks/otter-music/core/library"
	"github.com/neoworks/otter-music/core/music"
)

type fakeSavedShows struct {
	items []library.SavedItem
	saved []music.Item
}

func (store *fakeSavedShows) SavedItems(context.Context) ([]library.SavedItem, error) {
	return store.items, nil
}

func (store *fakeSavedShows) Save(_ context.Context, item music.Item) error {
	store.saved = append(store.saved, item)
	return nil
}

func savedItem(kind music.ItemKind, id, coverURL string) library.SavedItem {
	return library.SavedItem{Item: music.Item{Kind: kind, ID: id, Title: id, Thumbnails: []music.Thumbnail{{URL: coverURL}}}}
}

func TestRefreshSavedCoversReplacesExpiringShowCovers(t *testing.T) {
	expired := "https://i.ytimg.com/pl_c/PLa/studio_square_thumbnail.jpg?sqp=old"
	fresh := []music.Thumbnail{{URL: "https://i.ytimg.com/pl_c/PLa/studio_square_thumbnail.jpg?sqp=new", Width: 544}}
	catalog := &fakeCatalog{
		shows: map[string]music.PodcastPage{
			"MPSPa": {Title: "A", Thumbnails: fresh},
			"MPSPc": {Title: "C", Thumbnails: fresh},
		},
		opened: map[string]int{},
	}
	store := &fakeSavedShows{items: []library.SavedItem{
		savedItem(music.ItemKindPodcast, "MPSPa", expired),
		savedItem(music.ItemKindPodcast, "MPSPmissing", expired),
		savedItem(music.ItemKindPodcast, "MPSPc", "https://yt3.googleusercontent.com/stable"),
		savedItem(music.ItemKindAlbum, "MPREb", expired),
	}}

	refreshed, err := New(catalog, slog.New(slog.DiscardHandler)).RefreshSavedCovers(context.Background(), store)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed != 1 || len(store.saved) != 1 {
		t.Fatalf("refreshed = %d, saved = %+v, want only MPSPa", refreshed, store.saved)
	}
	if store.saved[0].ID != "MPSPa" || store.saved[0].Thumbnails[0].URL != fresh[0].URL {
		t.Errorf("saved = %+v, want MPSPa with the fresh cover", store.saved[0])
	}
	if catalog.opened["MPSPc"] != 0 {
		t.Error("a show with a stable cover was fetched")
	}
}
