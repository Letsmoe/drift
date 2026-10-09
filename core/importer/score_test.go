package importer

import (
	"testing"

	"github.com/letsmoe/drift-music/core/music"
	"github.com/letsmoe/drift-music/core/spotify"
)

func song(title string, artists ...string) music.Track {
	credited := []music.Artist{}
	for _, name := range artists {
		credited = append(credited, music.Artist{Name: name})
	}
	return music.Track{VideoID: "id", Title: title, Artists: credited}
}

func TestSplitTitle(t *testing.T) {
	cases := []struct {
		title    string
		core     string
		variants []string
	}{
		{"Get Lucky (feat. Pharrell Williams & Nile Rodgers)", "get lucky", []string{}},
		{"Blinding Lights - Remastered 2011", "blinding lights", []string{}},
		{"Creep - Live at Glastonbury", "creep", []string{"live"}},
		{"Hurt (Acoustic)", "hurt", []string{"acoustic"}},
		{"Sorry ft. Someone", "sorry", []string{}},
		{"Dancing With Myself", "dancing with myself", []string{}},
		{"Déjà Vu [Remix]", "deja vu", []string{"remix"}},
	}
	for _, testCase := range cases {
		parts := splitTitle(testCase.title)
		if parts.core != testCase.core || len(parts.variants) != len(testCase.variants) {
			t.Errorf("splitTitle(%q) = %+v, want core %q variants %v", testCase.title, parts, testCase.core, testCase.variants)
		}
	}
}

func TestScoreSong(t *testing.T) {
	cases := []struct {
		name    string
		wanted  spotify.Track
		found   music.Track
		matches bool
	}{
		{"exact", spotify.Track{Title: "Hold Up", Artists: []string{"Beyoncé"}}, song("Hold Up", "Beyoncé"), true},
		{"accents folded", spotify.Track{Title: "Hold Up", Artists: []string{"Beyoncé"}}, song("Hold Up", "Beyonce"), true},
		{"feat credit dropped", spotify.Track{Title: "Get Lucky (feat. Pharrell Williams)", Artists: []string{"Daft Punk"}}, song("Get Lucky", "Daft Punk", "Pharrell Williams"), true},
		{"remaster noise", spotify.Track{Title: "Heroes - 2017 Remaster", Artists: []string{"David Bowie"}}, song("Heroes", "David Bowie"), true},
		{"cover by someone else", spotify.Track{Title: "Hold Up", Artists: []string{"Beyoncé"}}, song("Hold Up", "Karaoke Kings"), false},
		{"different song same artist", spotify.Track{Title: "Hold Up", Artists: []string{"Beyoncé"}}, song("Formation", "Beyoncé"), false},
		{"artist folded into title", spotify.Track{Title: "Sorry", Artists: []string{"Justin Bieber"}}, song("Sorry (Justin Bieber)", "Various"), true},
		{"featured artist only", spotify.Track{Title: "Lose Yourself", Artists: []string{"Eminem"}}, song("Lose Yourself", "Eminem"), true},
	}
	for _, testCase := range cases {
		score := scoreSong(testCase.wanted, testCase.found)
		if (score >= songThreshold) != testCase.matches {
			t.Errorf("%s: score %.3f, matches=%v want %v", testCase.name, score, score >= songThreshold, testCase.matches)
		}
	}
}

func TestScoreSongPrefersStudioOverLive(t *testing.T) {
	wanted := spotify.Track{Title: "Creep", Artists: []string{"Radiohead"}, DurationMillis: 238000}
	studio := song("Creep", "Radiohead")
	studio.DurationSeconds = 238
	live := song("Creep (Live)", "Radiohead")
	live.DurationSeconds = 238
	if scoreSong(wanted, studio) <= scoreSong(wanted, live) {
		t.Errorf("studio %.3f not above live %.3f", scoreSong(wanted, studio), scoreSong(wanted, live))
	}
	liveWanted := spotify.Track{Title: "Creep - Live", Artists: []string{"Radiohead"}}
	if scoreSong(liveWanted, live) <= scoreSong(liveWanted, studio) {
		t.Errorf("a wanted live version should favor the live result")
	}
}

func TestScoreSongDuration(t *testing.T) {
	wanted := spotify.Track{Title: "Creep", Artists: []string{"Radiohead"}, DurationMillis: 238000}
	near := song("Creep", "Radiohead")
	near.DurationSeconds = 240
	far := song("Creep", "Radiohead")
	far.DurationSeconds = 300
	nearScore, farScore := scoreSong(wanted, near), scoreSong(wanted, far)
	if nearScore < 0.99 {
		t.Errorf("duration within 3 s scored %.3f", nearScore)
	}
	if farScore >= nearScore-0.1 {
		t.Errorf("duration 60 s off scored %.3f vs %.3f", farScore, nearScore)
	}
}

func TestScoreNamed(t *testing.T) {
	cases := []struct {
		name      string
		score     float64
		threshold float64
		matches   bool
	}{
		{"album exact", scoreNamed("OK Computer", "Radiohead", "OK Computer", "Album • Radiohead • 1997", albumWeights), albumThreshold, true},
		{"album deluxe noise", scoreNamed("OK Computer (Deluxe Edition)", "Radiohead", "OK Computer", "Album • Radiohead • 1997", albumWeights), albumThreshold, true},
		{"album other artist", scoreNamed("Greatest Hits", "Queen", "Greatest Hits", "Album • ABBA • 1976", albumWeights), albumThreshold, false},
		{"album other title", scoreNamed("OK Computer", "Radiohead", "Kid A", "Album • Radiohead • 2000", albumWeights), albumThreshold, false},
		{"show publisher differs", scoreNamed("Darknet Diaries", "Jack Rhysider", "Darknet Diaries", "Podcast • Jack", showWeights), showThreshold, true},
		{"show other title", scoreNamed("Darknet Diaries", "Jack Rhysider", "Dark Net", "Podcast", showWeights), showThreshold, false},
	}
	for _, testCase := range cases {
		if (testCase.score >= testCase.threshold) != testCase.matches {
			t.Errorf("%s: score %.3f, want match=%v", testCase.name, testCase.score, testCase.matches)
		}
	}
}

func TestScoreArtistName(t *testing.T) {
	if scoreArtistName("Bonobo", "Bonobo") < artistThreshold || scoreArtistName("Sigur Rós", "Sigur Ros") < artistThreshold {
		t.Error("equal names should match")
	}
	if scoreArtistName("Bonobo", "Bonobos") >= 1 || scoreArtistName("Queen", "Queens of the Stone Age") >= artistThreshold {
		t.Error("different names should not match")
	}
}
