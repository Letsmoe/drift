package lyrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/letsmoe/drift-music/core/music"
)

const (
	lrclibBaseURL = "https://lrclib.net"
	// LRCLIB asks clients to identify themselves.
	lrclibUserAgent = "Otter Music (https://github.com/letsmoe/drift-music)"
	lrclibTimeout   = 10 * time.Second
	lrclibSource    = "LRCLIB"
	// maxDurationDifference is how far (in seconds) a search result may be off the track.
	maxDurationDifference = 3
)

// LRCLIB looks lyrics up on lrclib.net by title, artist and duration.
type LRCLIB struct {
	httpClient *http.Client
	baseURL    string
}

// NewLRCLIB creates a client for lrclib.net; baseURL replaces it when not empty (tests).
func NewLRCLIB(httpClient *http.Client, baseURL string) *LRCLIB {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: lrclibTimeout}
	}
	if baseURL == "" {
		baseURL = lrclibBaseURL
	}
	return &LRCLIB{httpClient: httpClient, baseURL: strings.TrimRight(baseURL, "/")}
}

type lrclibRecord struct {
	TrackName    string
	ArtistName   string
	Duration     float64
	Instrumental bool
	PlainLyrics  string
	SyncedLyrics string
}

// Find returns the lyrics of track. The bool is false when LRCLIB has none or
// lists the track as instrumental.
func (lrclib *LRCLIB) Find(ctx context.Context, track music.Track) (music.Lyrics, bool, error) {
	if track.Title == "" || len(track.Artists) == 0 {
		return music.Lyrics{}, false, nil
	}
	record, found, err := lrclib.lookUp(ctx, track)
	if err != nil || !found || record.Instrumental {
		return music.Lyrics{}, false, err
	}
	return recordLyrics(record)
}

// lookUp asks for the exact track first, which needs a duration, and searches otherwise.
func (lrclib *LRCLIB) lookUp(ctx context.Context, track music.Track) (lrclibRecord, bool, error) {
	if track.DurationSeconds > 0 {
		record, found, err := lrclib.get(ctx, track)
		if err != nil || found {
			return record, found, err
		}
	}
	return lrclib.search(ctx, track)
}

func (lrclib *LRCLIB) get(ctx context.Context, track music.Track) (lrclibRecord, bool, error) {
	query := url.Values{
		"track_name":  {track.Title},
		"artist_name": {track.Artists[0].Name},
		"duration":    {strconv.Itoa(track.DurationSeconds)},
	}
	if track.Album != nil && track.Album.Name != "" {
		query.Set("album_name", track.Album.Name)
	}
	var record lrclibRecord
	found, err := lrclib.request(ctx, "/api/get", query, &record)
	return record, found, err
}

func (lrclib *LRCLIB) search(ctx context.Context, track music.Track) (lrclibRecord, bool, error) {
	query := url.Values{
		"track_name":  {track.Title},
		"artist_name": {track.Artists[0].Name},
	}
	var records []lrclibRecord
	if _, err := lrclib.request(ctx, "/api/search", query, &records); err != nil {
		return lrclibRecord{}, false, err
	}
	return bestRecord(records, track)
}

// bestRecord picks the first result with the same title and a close duration,
// preferring synced lyrics.
func bestRecord(records []lrclibRecord, track music.Track) (lrclibRecord, bool, error) {
	var plainMatch *lrclibRecord
	for index := range records {
		record := &records[index]
		if !recordMatches(*record, track) {
			continue
		}
		if record.SyncedLyrics != "" {
			return *record, true, nil
		}
		if plainMatch == nil {
			plainMatch = record
		}
	}
	if plainMatch == nil {
		return lrclibRecord{}, false, nil
	}
	return *plainMatch, true, nil
}

func recordMatches(record lrclibRecord, track music.Track) bool {
	if !strings.EqualFold(strings.TrimSpace(record.TrackName), strings.TrimSpace(track.Title)) {
		return false
	}
	if track.DurationSeconds == 0 || record.Duration == 0 {
		return true
	}
	difference := record.Duration - float64(track.DurationSeconds)
	return difference >= -maxDurationDifference && difference <= maxDurationDifference
}

// request GETs path and decodes the answer into response; a 404 is "not found".
func (lrclib *LRCLIB) request(ctx context.Context, path string, query url.Values, response any) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, lrclib.baseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		return false, fmt.Errorf("lrclib %s: %w", path, err)
	}
	request.Header.Set("User-Agent", lrclibUserAgent)

	httpResponse, err := lrclib.httpClient.Do(request)
	if err != nil {
		return false, fmt.Errorf("lrclib %s: %w", path, err)
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if httpResponse.StatusCode != http.StatusOK {
		return false, fmt.Errorf("lrclib %s: unexpected status %d", path, httpResponse.StatusCode)
	}
	if err := json.NewDecoder(httpResponse.Body).Decode(response); err != nil {
		return false, fmt.Errorf("lrclib %s: decode response: %w", path, err)
	}
	return true, nil
}

func recordLyrics(record lrclibRecord) (music.Lyrics, bool, error) {
	if record.SyncedLyrics != "" {
		lines := ParseLRC(record.SyncedLyrics)
		if len(lines) > 0 {
			return music.Lyrics{Lines: lines, Source: lrclibSource, Synced: true}, true, nil
		}
	}
	if strings.TrimSpace(record.PlainLyrics) == "" {
		return music.Lyrics{}, false, nil
	}
	return music.Lyrics{Lines: plainLines(record.PlainLyrics), Source: lrclibSource}, true, nil
}

func plainLines(text string) []music.LyricLine {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	rawLines := strings.Split(strings.TrimSpace(text), "\n")
	lines := make([]music.LyricLine, 0, len(rawLines))
	for _, rawLine := range rawLines {
		lines = append(lines, music.LyricLine{Text: strings.TrimSpace(rawLine)})
	}
	return lines
}
