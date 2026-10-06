package innertube

import "testing"

func TestParsePlayCount(t *testing.T) {
	cases := map[string]string{
		"6.2M plays":  "6.2M plays",
		"922K views":  "922K views",
		"1 view":      "1 view",
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
