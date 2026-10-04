package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://open.spotify.com"
	// PublicTrackLimit is how many tracks the public embed page lists at most.
	PublicTrackLimit = 100
	fetchTimeout     = 20 * time.Second
	maxPageBytes     = 8 << 20
	// browserUserAgent: the embed page answers a bare Go client with a different page.
	browserUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"
)

// ErrNoEmbedData means the embed page did not contain the entity data, e.g.
// because the playlist is private or Spotify changed the page.
var ErrNoEmbedData = errors.New("spotify: embed page has no entity data")

var nextDataPattern = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

// Fetcher reads public playlists and albums from Spotify's embed pages.
type Fetcher struct {
	client  *http.Client
	baseURL string
}

// NewFetcher returns a Fetcher for open.spotify.com.
func NewFetcher() *Fetcher {
	return &Fetcher{client: &http.Client{Timeout: fetchTimeout}, baseURL: defaultBaseURL}
}

// NewFetcherFor reads from baseURL with client, for tests.
func NewFetcherFor(client *http.Client, baseURL string) *Fetcher {
	return &Fetcher{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

type embedTrack struct {
	URI      string `json:"uri"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Duration int    `json:"duration"`
}

type embedEntity struct {
	Name      string       `json:"name"`
	Title     string       `json:"title"`
	Subtitle  string       `json:"subtitle"`
	TrackList []embedTrack `json:"trackList"`
}

type embedPage struct {
	Props struct {
		PageProps struct {
			State struct {
				Data struct {
					Entity *embedEntity `json:"entity"`
				} `json:"data"`
			} `json:"state"`
		} `json:"pageProps"`
	} `json:"props"`
}

// Fetch loads the public playlist or album behind link. A playlist comes back
// as one Playlist (at most PublicTrackLimit tracks), an album as one Album.
func (fetcher *Fetcher) Fetch(ctx context.Context, link string) (Library, error) {
	reference, err := ParseURL(link)
	if err != nil {
		return Library{}, err
	}
	entity, err := fetcher.loadEntity(ctx, reference)
	if err != nil {
		return Library{}, err
	}
	return libraryFromEntity(reference, *entity), nil
}

func (fetcher *Fetcher) loadEntity(ctx context.Context, reference Reference) (*embedEntity, error) {
	page, err := fetcher.get(ctx, fmt.Sprintf("%s/embed/%s/%s", fetcher.baseURL, reference.Kind, reference.ID))
	if err != nil {
		return nil, err
	}
	return parseEmbedPage(page)
}

func (fetcher *Fetcher) get(ctx context.Context, address string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("spotify: build request: %w", err)
	}
	request.Header.Set("User-Agent", browserUserAgent)
	request.Header.Set("Accept-Language", "en")
	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("spotify: fetch %s: %w", address, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify: fetch %s: status %d (private or removed?)", address, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxPageBytes))
	if err != nil {
		return nil, fmt.Errorf("spotify: read %s: %w", address, err)
	}
	return body, nil
}

func parseEmbedPage(page []byte) (*embedEntity, error) {
	match := nextDataPattern.FindSubmatch(page)
	if match == nil {
		return nil, ErrNoEmbedData
	}
	var parsed embedPage
	err := json.Unmarshal(match[1], &parsed)
	if err != nil {
		return nil, fmt.Errorf("spotify: decode embed data: %w", err)
	}
	entity := parsed.Props.PageProps.State.Data.Entity
	if entity == nil {
		return nil, ErrNoEmbedData
	}
	return entity, nil
}

func libraryFromEntity(reference Reference, entity embedEntity) Library {
	library := Library{LikedTracks: []Track{}, Playlists: []Playlist{}, Albums: []Album{}, Artists: []Artist{}, Shows: []Show{}}
	name := entity.Name
	if name == "" {
		name = entity.Title
	}
	if reference.Kind == KindAlbum {
		library.Albums = append(library.Albums, Album{URI: reference.URI(), Title: name, Artist: entity.Subtitle})
		return library
	}
	library.Playlists = append(library.Playlists, Playlist{
		Key:    reference.URI(),
		Name:   name,
		Tracks: convertEmbedTracks(entity.TrackList),
	})
	return library
}

func convertEmbedTracks(embedTracks []embedTrack) []Track {
	tracks := make([]Track, 0, len(embedTracks))
	for _, embedded := range embedTracks {
		if !strings.HasPrefix(embedded.URI, "spotify:track:") || embedded.Title == "" {
			continue
		}
		tracks = append(tracks, Track{
			URI:            embedded.URI,
			Title:          embedded.Title,
			Artists:        splitSubtitleArtists(embedded.Subtitle),
			DurationMillis: embedded.Duration,
		})
	}
	return tracks
}

// splitSubtitleArtists splits the embed's "A, B" artist line. Names that
// contain a comma themselves split wrongly; the matcher tolerates that.
func splitSubtitleArtists(subtitle string) []string {
	artists := []string{}
	for _, part := range strings.Split(subtitle, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			artists = append(artists, trimmed)
		}
	}
	return artists
}
