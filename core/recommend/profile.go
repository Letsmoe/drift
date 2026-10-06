package recommend

import (
	"math"
	"sort"
	"time"

	"github.com/neoworks/otter-music/core/music"
)

// Profile is the user's time-decayed taste: an affinity in [-1, 1] per artist
// and album, normalized per dimension so the strongest entity of each is ±1.
type Profile struct {
	// Artists is keyed by artist browse ID; artists YouTube Music does not link
	// carry no affinity.
	Artists map[string]float64 `json:"artists"`
	// ArtistNames maps the artist IDs seen in the profile to display names.
	ArtistNames map[string]string  `json:"artistNames"`
	Albums      map[string]float64 `json:"albums"`
	// Seeds are the tracks with the strongest positive signal, strongest first.
	Seeds []Seed `json:"seeds"`
	// Disliked holds the tracks rated thumbs-down.
	Disliked map[string]bool `json:"disliked"`
	// BlockedTracks and BlockedArtists are the "Don't recommend" choices: kept
	// out of every recommendation without being a taste signal.
	BlockedTracks  map[string]bool `json:"blockedTracks"`
	BlockedArtists map[string]bool `json:"blockedArtists"`
	// SignalCount is how many signals the profile was built from.
	SignalCount int `json:"signalCount"`
}

// excludes reports whether track must never be recommended: disliked, blocked,
// or by a blocked artist (featured ones included).
func (profile *Profile) excludes(track music.Track) bool {
	if profile.Disliked[track.VideoID] || profile.BlockedTracks[track.VideoID] {
		return true
	}
	for _, artist := range track.Artists {
		if artist.ID != "" && profile.BlockedArtists[artist.ID] {
			return true
		}
	}
	return false
}

// Seed is a track whose radio is worth recommending.
type Seed struct {
	VideoID string  `json:"videoId"`
	Title   string  `json:"title"`
	Weight  float64 `json:"weight"`
}

// profileInput is everything buildProfile reads.
type profileInput struct {
	signals  []signal
	tracks   map[string]music.Track
	saved    savedTaste
	disliked map[string]bool
	blocked  blocks
	now      time.Time
}

// Empty reports whether there is no signal yet (cold start).
func (profile *Profile) Empty() bool {
	return profile.SignalCount == 0 && len(profile.Artists) == 0
}

// decayFactor is the share of a signal's weight left after age.
func decayFactor(age, halfLife time.Duration) float64 {
	if age <= 0 {
		return 1
	}
	return math.Pow(0.5, float64(age)/float64(halfLife))
}

// buildProfile turns signals into per-entity affinities.
func buildProfile(input profileInput) *Profile {
	profile := &Profile{
		Artists:        map[string]float64{},
		ArtistNames:    map[string]string{},
		Albums:         map[string]float64{},
		Disliked:       input.disliked,
		BlockedTracks:  input.blocked.tracks,
		BlockedArtists: input.blocked.artists,
		SignalCount:    len(input.signals),
	}
	trackWeights := decayedTrackWeights(input.signals, input.now)
	for videoID, weight := range trackWeights {
		track, found := input.tracks[videoID]
		if found {
			profile.addTrack(track, weight)
		}
	}
	profile.addSaved(input.saved)
	normalizeAffinities(profile.Artists)
	normalizeAffinities(profile.Albums)
	profile.Seeds = seedsFrom(trackWeights, input.tracks)
	return profile
}

// decayedTrackWeights sums each track's signals after time decay.
func decayedTrackWeights(signals []signal, now time.Time) map[string]float64 {
	weights := map[string]float64{}
	for _, entry := range signals {
		weights[entry.videoID] += entry.weight * decayFactor(now.Sub(entry.at), AffinityHalfLife)
	}
	return weights
}

// linkedArtists returns the track's credited artists YouTube Music links, in
// credit order.
func linkedArtists(track music.Track) []music.Artist {
	artists := make([]music.Artist, 0, len(track.Artists))
	for _, artist := range track.Artists {
		if artist.ID != "" {
			artists = append(artists, artist)
		}
	}
	return artists
}

// albumID returns the track's album ID, or empty when it has none.
func albumID(track music.Track) string {
	if track.Album == nil {
		return ""
	}
	return track.Album.ID
}

// addTrack spreads a track's weight onto its artists and album. The first
// credited artist takes it whole; featured artists share it damped by the
// square root of the credit count.
func (profile *Profile) addTrack(track music.Track, weight float64) {
	artists := linkedArtists(track)
	featuredShare := weight / math.Sqrt(float64(len(artists)))
	for position, artist := range artists {
		profile.ArtistNames[artist.ID] = artist.Name
		if position == 0 {
			profile.Artists[artist.ID] += weight
			continue
		}
		profile.Artists[artist.ID] += featuredShare
	}
	if id := albumID(track); id != "" {
		profile.Albums[id] += weight
	}
}

// addSaved boosts followed artists and saved albums.
func (profile *Profile) addSaved(saved savedTaste) {
	for id, name := range saved.artists {
		profile.Artists[id] += followedArtistBoost
		profile.ArtistNames[id] = name
	}
	for id := range saved.albums {
		profile.Albums[id] += savedAlbumBoost
	}
}

// normalizeAffinities divides every value by the largest magnitude.
func normalizeAffinities(affinities map[string]float64) {
	largest := 0.0
	for _, value := range affinities {
		largest = math.Max(largest, math.Abs(value))
	}
	if largest == 0 {
		return
	}
	for key, value := range affinities {
		affinities[key] = value / largest
	}
}

// seedsFrom picks the tracks strong enough to seed radios, strongest first.
func seedsFrom(trackWeights map[string]float64, tracks map[string]music.Track) []Seed {
	seeds := []Seed{}
	for videoID, weight := range trackWeights {
		track, found := tracks[videoID]
		if found && weight >= seedMinimumWeight {
			seeds = append(seeds, Seed{VideoID: videoID, Title: track.Title, Weight: weight})
		}
	}
	sort.Slice(seeds, func(left, right int) bool {
		if seeds[left].Weight != seeds[right].Weight {
			return seeds[left].Weight > seeds[right].Weight
		}
		return seeds[left].VideoID < seeds[right].VideoID
	})
	if len(seeds) > seedLimit {
		seeds = seeds[:seedLimit]
	}
	return seeds
}

// TopEntities returns up to limit ids with a positive value, highest first
// (ties by id, so the order is stable).
func TopEntities(values map[string]float64, limit int) []string {
	ids := make([]string, 0, len(values))
	for id, value := range values {
		if value > 0 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(left, right int) bool {
		if values[ids[left]] != values[ids[right]] {
			return values[ids[left]] > values[ids[right]]
		}
		return ids[left] < ids[right]
	})
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids
}
