package importer

import (
	"context"

	"github.com/letsmoe/drift-music/core/music"
	"github.com/letsmoe/drift-music/core/spotify"
)

func trackLookup(catalog Catalog, track spotify.Track) lookup {
	return lookup{
		uri: track.Key(), kind: music.ItemKindSong, phase: PhaseTracks, label: trackLabel(track),
		find: func(ctx context.Context) (found, error) {
			return findSong(ctx, catalog, track)
		},
	}
}

func albumLookup(catalog Catalog, album spotify.Album) lookup {
	return lookup{
		uri: album.Key(), kind: music.ItemKindAlbum, phase: PhaseAlbums, label: album.Title,
		find: func(ctx context.Context) (found, error) {
			return findNamed(ctx, catalog, music.ItemKindAlbum, searchText(album.Title, album.Artist), func(item music.Item) float64 {
				return scoreNamed(album.Title, album.Artist, item.Title, item.Subtitle, albumWeights)
			}, albumThreshold)
		},
	}
}

func artistLookup(catalog Catalog, artist spotify.Artist) lookup {
	return lookup{
		uri: artist.Key(), kind: music.ItemKindArtist, phase: PhaseArtists, label: artist.Name,
		find: func(ctx context.Context) (found, error) {
			return findNamed(ctx, catalog, music.ItemKindArtist, artist.Name, func(item music.Item) float64 {
				return scoreArtistName(artist.Name, item.Title)
			}, artistThreshold)
		},
	}
}

func showLookup(catalog Catalog, show spotify.Show) lookup {
	return lookup{
		uri: show.Key(), kind: music.ItemKindPodcast, phase: PhaseShows, label: show.Name,
		find: func(ctx context.Context) (found, error) {
			return findNamed(ctx, catalog, music.ItemKindPodcast, searchText(show.Name, show.Publisher), func(item music.Item) float64 {
				return scoreNamed(show.Name, show.Publisher, item.Title, item.Subtitle, showWeights)
			}, showThreshold)
		},
	}
}

func trackLabel(track spotify.Track) string {
	if len(track.Artists) == 0 {
		return track.Title
	}
	return track.Artists[0] + " - " + track.Title
}

func searchText(title, creator string) string {
	cleaned := splitTitle(title).core
	if cleaned == "" {
		cleaned = title
	}
	return cleaned + " " + creator
}

// findSong searches "<title> <first artist>" and keeps the best song above
// the threshold.
func findSong(ctx context.Context, catalog Catalog, track spotify.Track) (found, error) {
	creator := ""
	if len(track.Artists) > 0 {
		creator = track.Artists[0]
	}
	candidates, err := catalog.SearchSongs(ctx, searchText(track.Title, creator))
	if err != nil {
		return found{}, err
	}
	var best found
	for index := range candidates {
		score := scoreSong(track, candidates[index])
		if score >= songThreshold && score > best.score {
			best = found{targetID: candidates[index].VideoID, score: score, track: &candidates[index]}
		}
	}
	return best, nil
}

// findNamed searches everything and scores the results of one kind.
func findNamed(ctx context.Context, catalog Catalog, kind music.ItemKind, query string, score func(music.Item) float64, threshold float64) (found, error) {
	results, err := catalog.Search(ctx, query)
	if err != nil {
		return found{}, err
	}
	var best found
	for _, shelf := range results.Shelves {
		for index := range shelf.Items {
			item := shelf.Items[index]
			if item.Kind != kind || item.ID == "" {
				continue
			}
			itemScore := score(item)
			if itemScore >= threshold && itemScore > best.score {
				best = found{targetID: item.ID, score: itemScore, item: &item}
			}
		}
	}
	return best, nil
}
