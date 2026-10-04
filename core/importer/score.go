package importer

import (
	"slices"
	"strings"

	"github.com/neoworks/otter-music/core/music"
	"github.com/neoworks/otter-music/core/spotify"
)

// Score thresholds: a candidate below its threshold counts as unmatched.
const (
	songThreshold  = 0.75
	albumThreshold = 0.75
	// artistThreshold wants an almost exact name; a different artist is worse than none.
	artistThreshold = 0.88
	showThreshold   = 0.8
)

const (
	titleWeight    = 0.5
	artistWeight   = 0.3
	durationWeight = 0.2
	albumBonus     = 0.05

	variantMismatchFactor = 0.8
	durationFarFactor     = 0.85
	exactDurationSeconds  = 3
	farDurationSeconds    = 15
	unknownArtistScore    = 0.5
)

// scoreSong rates how well a YouTube Music song is the wanted Spotify track,
// from 0 to 1. It is pure: title similarity, artist overlap and duration
// closeness, lowered when one side is a live/remix/... version and the other
// is not.
func scoreSong(wanted spotify.Track, found music.Track) float64 {
	wantedTitle, foundTitle := splitTitle(wanted.Title), splitTitle(found.Title)
	titleScore := similarity(wantedTitle.core, foundTitle.core)
	artistScore := scoreArtists(wanted.Artists, found.Artists, found.Title)

	score := 0.0
	durationKnown := wanted.DurationMillis > 0 && found.DurationSeconds > 0
	if durationKnown {
		score = titleWeight*titleScore + artistWeight*artistScore + durationWeight*durationScore(wanted.DurationMillis/1000, found.DurationSeconds)
	} else {
		score = (titleWeight*titleScore + artistWeight*artistScore) / (titleWeight + artistWeight)
	}
	if wanted.Album != "" && found.Album != nil && stripNoise(wanted.Album) == stripNoise(found.Album.Name) {
		score += albumBonus
	}
	if !slices.Equal(wantedTitle.variants, foundTitle.variants) {
		score *= variantMismatchFactor
	}
	if durationKnown && absoluteDifference(wanted.DurationMillis/1000, found.DurationSeconds) > farDurationSeconds {
		score *= durationFarFactor
	}
	return min(score, 1)
}

// scoreArtists weighs the first wanted artist (the primary one) over the
// share of all wanted artists the result credits. The title is searched too
// because YouTube sometimes folds featured artists into it.
func scoreArtists(wanted []string, found []music.Artist, foundTitle string) float64 {
	if len(wanted) == 0 || len(found) == 0 {
		return unknownArtistScore
	}
	credited := make([]string, 0, len(found))
	for _, artist := range found {
		credited = append(credited, foldText(artist.Name))
	}
	creditedText := " " + strings.Join(credited, " ") + " " + foldText(foundTitle) + " "
	primary := artistCredited(wanted[0], credited, creditedText)
	matched := 0
	for _, name := range wanted {
		if artistCredited(name, credited, creditedText) > 0.99 {
			matched++
		}
	}
	return 0.7*primary + 0.3*float64(matched)/float64(len(wanted))
}

// artistCredited is 1 when the name is among the credited artists (or reads
// as a whole word run in them), else the best similarity to one of them.
func artistCredited(name string, credited []string, creditedText string) float64 {
	folded := foldText(name)
	if folded == "" {
		return 0
	}
	if strings.Contains(creditedText, " "+folded+" ") {
		return 1
	}
	best := 0.0
	for _, candidate := range credited {
		best = max(best, similarity(folded, candidate))
	}
	return best
}

func durationScore(wantedSeconds, foundSeconds int) float64 {
	difference := absoluteDifference(wantedSeconds, foundSeconds)
	if difference <= exactDurationSeconds {
		return 1
	}
	if difference > farDurationSeconds {
		return 0
	}
	span := float64(farDurationSeconds - exactDurationSeconds)
	return 1 - 0.5*float64(difference-exactDurationSeconds)/span
}

func absoluteDifference(left, right int) int {
	if left > right {
		return left - right
	}
	return right - left
}

// namedWeights splits a name score between title and the creator line.
type namedWeights struct {
	title   float64
	creator float64
}

var (
	albumWeights = namedWeights{title: 0.7, creator: 0.3}
	// Podcast publishers often differ between Spotify and YouTube, so the title dominates.
	showWeights = namedWeights{title: 0.85, creator: 0.15}
)

// scoreNamed rates an album or show result: the title (release noise like
// "Deluxe Edition" ignored) plus whether the creator appears in the result's
// subtitle line.
func scoreNamed(wantedTitle, wantedCreator, foundTitle, foundSubtitle string, weights namedWeights) float64 {
	titleScore := similarity(stripNoise(wantedTitle), stripNoise(foundTitle))
	creatorScore := unknownArtistScore
	folded := foldText(wantedCreator)
	if folded != "" {
		creatorScore = 0
		if strings.Contains(" "+foldText(foundSubtitle)+" ", " "+folded+" ") {
			creatorScore = 1
		}
	}
	return weights.title*titleScore + weights.creator*creatorScore
}

// scoreArtistName rates a found artist against a followed one by name only.
func scoreArtistName(wantedName, foundName string) float64 {
	return similarity(foldText(wantedName), foldText(foundName))
}
