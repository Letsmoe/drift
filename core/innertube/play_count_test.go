package innertube

import "testing"

func TestParsePlayCount(t *testing.T) {
	cases := map[string]string{
		"6.2M plays":  "6.2M plays",
		"922K views":  "922K plays",
		"1 view":      "1 play",
		"1,234 plays": "1,234 plays",
		"Drake":       "",
		"2019":        "",
		"3:20":        "",
	}
	for text, want := range cases {
		runs := []run{{Text: "Artist"}, {Text: segmentSeparator}, {Text: text}}
		if got := parsePlayCount(runs); got != want {
			t.Errorf("parsePlayCount(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestViewsAsPlays(t *testing.T) {
	cases := map[string]string{
		"Artist • 922K views":      "Artist • 922K plays",
		"1 view • Aug 30":          "1 play • Aug 30",
		"Artist • 6.2M plays":      "Artist • 6.2M plays",
		"Review Show • 1.2K views": "Review Show • 1.2K plays",
	}
	for text, want := range cases {
		if got := viewsAsPlays(text); got != want {
			t.Errorf("viewsAsPlays(%q) = %q, want %q", text, got, want)
		}
	}
}
