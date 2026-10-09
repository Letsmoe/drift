//go:build live

package podcasts

import (
	"log/slog"
	"testing"

	"github.com/letsmoe/drift-music/core/innertube"
	"github.com/letsmoe/drift-music/core/music"
)

func TestLiveShows(t *testing.T) {
	directory := New(innertube.NewClient(), slog.New(slog.DiscardHandler))

	page, err := directory.Shows(t.Context())
	if err != nil {
		t.Fatalf("shows: %v", err)
	}
	if len(page.Shelves) == 0 || len(page.Shelves[0].Items) == 0 {
		t.Fatal("no show shelves; the podcasts feed or show pages probably changed")
	}
	first := page.Shelves[0].Items[0]
	if first.Kind != music.ItemKindPodcast || first.Title == "" || len(first.Thumbnails) == 0 {
		t.Errorf("first show = %+v, want a titled show with a cover", first)
	}
	for _, shelf := range page.Shelves {
		t.Logf("%s: %d shows, first %q", shelf.Title, len(shelf.Items), shelf.Items[0].Title)
	}
}
