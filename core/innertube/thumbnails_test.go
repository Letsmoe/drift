package innertube

import (
	"testing"

	"github.com/letsmoe/drift/core/music"
)

func TestWithLargeVariantsResizesGoogleImages(t *testing.T) {
	base := "https://yt3.googleusercontent.com/abc-"
	thumbnails := withLargeVariants([]music.Thumbnail{
		{URL: base + "=w60-h60-l90-rj", Width: 60, Height: 60},
		{URL: base + "=w120-h120-l90-rj", Width: 120, Height: 120},
	})
	want := []music.Thumbnail{
		{URL: base + "=w60-h60-l90-rj", Width: 60, Height: 60},
		{URL: base + "=w120-h120-l90-rj", Width: 120, Height: 120},
		{URL: base + "=w544-h544-l90-rj", Width: 544, Height: 544},
		{URL: base + "=w1080-h1080-l90-rj", Width: 1080, Height: 1080},
	}
	if len(thumbnails) != len(want) {
		t.Fatalf("thumbnails = %+v, want %+v", thumbnails, want)
	}
	for index := range want {
		if thumbnails[index] != want[index] {
			t.Errorf("thumbnail %d = %+v, want %+v", index, thumbnails[index], want[index])
		}
	}
}

func TestWithLargeVariantsLeavesOthersAlone(t *testing.T) {
	cases := map[string][]music.Thumbnail{
		"already large": {{URL: "https://lh3.googleusercontent.com/x=w1200-h1200", Width: 1200, Height: 1200}},
		"not resizable": {{URL: "https://i.ytimg.com/vi/id/hqdefault.jpg", Width: 480, Height: 360}},
		"empty":         {},
	}
	for name, thumbnails := range cases {
		got := withLargeVariants(thumbnails)
		if len(got) != len(thumbnails) {
			t.Errorf("%s: got %+v, want it unchanged", name, got)
		}
	}
}

func TestWithLargeVariantsKeepsAspectRatio(t *testing.T) {
	thumbnails := withLargeVariants([]music.Thumbnail{
		{URL: "https://lh3.googleusercontent.com/x=w400-h225", Width: 400, Height: 225},
	})
	if len(thumbnails) != 3 || thumbnails[1].URL != "https://lh3.googleusercontent.com/x=w544-h306" {
		t.Errorf("thumbnails = %+v", thumbnails)
	}
}
