package recommend

import "time"

// Every tunable number of the engine lives in this file so the relative
// strengths of signals, sources and penalties can be read side by side.

// Taste profile decay and signal windows.
const (
	// AffinityHalfLife is how fast a signal fades: its weight halves every
	// half-life.
	AffinityHalfLife = 30 * 24 * time.Hour
	// signalWindow bounds how far back signals are read. After six half-lives
	// a signal keeps under 2% of its weight.
	signalWindow = 6 * AffinityHalfLife
	// ignoredImpressionWindow bounds the "shown but not clicked" signal and
	// feed fatigue.
	ignoredImpressionWindow = 30 * 24 * time.Hour
)

// Signal weights per track interaction, strongest positive first. Every play
// row is its own signal, so repeats add up. Negative weights express distaste.
const (
	weightLike              = 3.0
	weightCompletedPlay     = 2.0
	weightPlaylistAdd       = 1.5
	weightPartialPlay       = 1.0
	weightClick             = 0.5
	weightIgnoredImpression = -0.15
	weightEarlySkip         = -1.5
	weightDislike           = -3.0
	// ignoredImpressionCap bounds how many ignored showings of one track count.
	ignoredImpressionCap = 3
)

// Entities the user saved in the library count as a standing signal, added to
// the raw affinity before it is normalized. A follow is worth a like plus a
// completed play; a saved album is slightly less.
const (
	followedArtistBoost = 5.0
	savedAlbumBoost     = 4.0
)

// Play interpretation.
const (
	// A play that stopped before both thresholds is an early skip.
	earlySkipSeconds  = 30
	earlySkipFraction = 0.5
	// A play that reached this share of the duration counts as completed.
	completedFraction = 0.8
	// skipGrace keeps a play that just started from counting as skipped.
	skipGrace = 5 * time.Minute
)

// Candidate generation sizes.
const (
	// seedMinimumWeight is the decayed signal weight a track needs to seed
	// radio candidates (a completed play from up to a month ago).
	seedMinimumWeight = 1.0
	seedLimit         = 8
	// radioTracksPerSeed caps each seed's radio list.
	radioTracksPerSeed = 24
	topArtistLimit     = 5
	// tracksPerArtist caps the top songs taken from one artist page.
	tracksPerArtist = 20
	// exploreArtistLimit is the number of unknown related artists explored.
	exploreArtistLimit     = 4
	tracksPerExploreArtist = 10
	recentSearchLimit      = 3
	recentSearchWindow     = 14 * 24 * time.Hour
	tracksPerSearch        = 20
	// rediscoverWindow is how long a loved track must have been unplayed to
	// be offered again.
	rediscoverWindow = 30 * 24 * time.Hour
	rediscoverLimit  = 30
	homeTrackLimit   = 40
)

// Network access to the catalog.
const (
	// fetchConcurrency is how many catalog requests run at once.
	fetchConcurrency = 4
	// fetchTimeout bounds one catalog request.
	fetchTimeout = 15 * time.Second
	// Cached catalog responses live this long in the library cache.
	radioCacheTTL  = 24 * time.Hour
	artistCacheTTL = 24 * time.Hour
	searchCacheTTL = 24 * time.Hour
	homeCacheTTL   = time.Hour
)

// Source priors: how much a candidate source is trusted before its content is
// scored, in [0, 1]. YouTube Music's radio is collaborative filtering over its
// whole audience, so it ranks highest. SourceArtist is scaled by the artist's
// affinity when candidates are generated.
var sourcePriors = map[string]float64{
	SourceRadio:      1.0,
	SourceArtist:     0.65,
	SourceSearch:     0.4,
	SourceRediscover: 0.35,
	SourceExplore:    0.3,
	SourceYouTube:    0.3,
}

// Prior shaping.
const (
	// rankDecay lowers a candidate's prior with its position within its source.
	rankDecay = 0.05
	// seedWeightInfluence is how much a seed's strength scales the prior of
	// its radio tracks (0 = not at all, 1 = fully proportional).
	seedWeightInfluence = 0.5
	// multiSourceBonus is added to the prior per extra source that found the
	// same track.
	multiSourceBonus = 0.1
)

// Ranking: score = affinity + prior − penalties.
const (
	rankWeightAffinity    = 1.0
	rankWeightSourcePrior = 0.6

	// Recently played tracks sink: heavily within a day, lightly within a week.
	penaltyPlayedToday    = 1.0
	penaltyPlayedThisWeek = 0.3
	playedTodayWindow     = 24 * time.Hour
	playedThisWeekWindow  = 7 * 24 * time.Hour
	// Fatigue: each showing without a click beyond the free ones costs this.
	penaltyFatiguePerShow = 0.15
	fatigueFreeShows      = 1
)

// Affinity match: how a track's artist and album affinities mix into one
// [-1, 1] value.
const (
	matchWeightArtist = 0.7
	matchWeightAlbum  = 0.3
)

// Re-ranking.
const (
	// pageSize is the span the diversity caps apply to.
	pageSize            = 24
	maxPerArtistPerPage = 2
	maxPerAlbumPerPage  = 3
	// maxSameSourceRun is how many items from one source may follow each other.
	maxSameSourceRun = 2
	// explorationEvery reserves every seventh slot (~15%) for exploration.
	explorationEvery = 7
)

// Output shaping.
const (
	// feedCacheTTL is how long a ranked feed serves later pages (offset > 0),
	// so pagination stays stable while impressions shift the scores.
	feedCacheTTL = 30 * time.Minute
	// feedFirstPageTTL is how long the first page (offset 0) keeps serving the
	// same ranking, so reloading Home doesn't reshuffle it. An explicit refresh
	// re-ranks regardless.
	feedFirstPageTTL = 10 * time.Minute
	// minRowItems drops rows too thin to be worth a shelf.
	minRowItems = 3
	// maxRelatedRows caps the "Because you listened to" rows per page, keeping
	// the strongest seeds, so the page is not five rows of the same kind.
	maxRelatedRows = 2
	// madeForYouSize is the size of the "Made for you" feed block the home page
	// shows above the rows (the desktop requests Feed(20, 0)); rows leave out
	// those tracks.
	madeForYouSize = 20
	// extendLimit is how many autoplay tracks one Extend call returns.
	extendLimit = 20
)
