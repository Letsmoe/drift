package recommend

import (
	"math"
	"sort"
	"time"

	"github.com/letsmoe/drift/core/music"
)

// scoreCandidates scores every eligible candidate and returns them best first.
func scoreCandidates(candidates []*candidate, state *snapshot) []Item {
	items := make([]Item, 0, len(candidates))
	for _, proposed := range candidates {
		if state.profile.excludes(proposed.track) {
			continue
		}
		total := baseScore(proposed, state.profile) - penalties(proposed.track.VideoID, state)
		items = append(items, itemOf(proposed, total))
	}
	sortByScore(items)
	return items
}

// itemOf builds the ranked item of a candidate.
func itemOf(proposed *candidate, total float64) Item {
	return Item{Track: proposed.track, Source: proposed.source, Reason: proposed.reason, Score: total}
}

// sortByScore orders items best first, ties by video ID so the order is stable.
func sortByScore(items []Item) {
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].Score != items[right].Score {
			return items[left].Score > items[right].Score
		}
		return items[left].Track.VideoID < items[right].Track.VideoID
	})
}

// baseScore is the transparent ranking formula before penalties: affinity
// match plus source prior (with the multi-source bonus).
func baseScore(proposed *candidate, profile *Profile) float64 {
	prior := proposed.prior + multiSourceBonus*float64(proposed.sourceCount-1)
	return rankWeightAffinity*affinityMatch(proposed.track, profile) + rankWeightSourcePrior*prior
}

// affinityMatch mixes the track's artist and album affinities into one value
// in [-1, 1].
func affinityMatch(track music.Track, profile *Profile) float64 {
	artistIDs := make([]string, 0, len(track.Artists))
	for _, artist := range linkedArtists(track) {
		artistIDs = append(artistIDs, artist.ID)
	}
	match := matchWeightArtist * strongestAffinity(artistIDs, profile.Artists)
	return match + matchWeightAlbum*profile.Albums[albumID(track)]
}

// strongestAffinity is the highest affinity among ids, or the lowest when none
// is positive: one loved artist carries a track, an all-disliked cast sinks it.
func strongestAffinity(ids []string, affinities map[string]float64) float64 {
	highest, lowest := 0.0, 0.0
	for _, id := range ids {
		highest = math.Max(highest, affinities[id])
		lowest = math.Min(lowest, affinities[id])
	}
	if highest > 0 {
		return highest
	}
	return lowest
}

// penalties adds up what counts against a track: played lately, or shown
// repeatedly without a click.
func penalties(videoID string, state *snapshot) float64 {
	total := recentPlayPenalty(state.lastPlayed[videoID], state.now)
	return total + fatiguePenalty(state.ignored[videoID].count)
}

// recentPlayPenalty costs a play within a day more than one within a week;
// the zero time means never played.
func recentPlayPenalty(lastPlayed, now time.Time) float64 {
	if lastPlayed.IsZero() {
		return 0
	}
	age := now.Sub(lastPlayed)
	if age < playedTodayWindow {
		return penaltyPlayedToday
	}
	if age < playedThisWeekWindow {
		return penaltyPlayedThisWeek
	}
	return 0
}

// fatiguePenalty grows with every ignored showing past the free ones.
func fatiguePenalty(ignoredShows int) float64 {
	return penaltyFatiguePerShow * float64(max(0, ignoredShows-fatigueFreeShows))
}
