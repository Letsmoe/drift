//go:build live

package innertube

import "testing"

func TestLiveHome(t *testing.T) {
	client := NewClient()

	page, err := client.Home(t.Context())
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	if len(page.Shelves) == 0 || len(page.Shelves[0].Items) == 0 {
		t.Fatal("home returned no shelves; the response layout probably changed")
	}
	if page.Continuation == "" {
		t.Fatal("home returned no continuation")
	}

	next, err := client.HomeContinuation(t.Context(), page.Continuation)
	if err != nil {
		t.Fatalf("home continuation: %v", err)
	}
	if len(next.Shelves) == 0 {
		t.Error("home continuation returned no shelves")
	}
	for _, shelf := range append(page.Shelves, next.Shelves...) {
		t.Logf("%s: %d items, first %s %q", shelf.Title, len(shelf.Items), shelf.Items[0].Kind, shelf.Items[0].Title)
	}
}
