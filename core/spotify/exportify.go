package spotify

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"strings"
)

const (
	exportifyPlaylistKeyPrefix = "exportify:playlist:"
	exportifyLikedSongsName    = "liked_songs"
	exportifySavedAlbumsName   = "saved_albums"
)

var utf8ByteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// exportifyColumns maps a lowercased header name to its column index, so
// exports with more or fewer columns than today's still read.
type exportifyColumns map[string]int

// addExportifyFile adds one Exportify CSV to the library. Files without an
// Exportify header are ignored.
func addExportifyFile(library *Library, fileName string, contents []byte) {
	tracks, isExportify := parseExportifyTracks(contents)
	if !isExportify {
		return
	}
	baseName := exportifyBaseName(fileName)
	switch strings.ToLower(baseName) {
	case exportifyLikedSongsName:
		library.LikedTracks = append(library.LikedTracks, tracks...)
	case exportifySavedAlbumsName:
		library.Albums = append(library.Albums, groupSavedAlbums(tracks)...)
	default:
		library.Playlists = append(library.Playlists, exportifyPlaylist(baseName, tracks))
	}
}

func exportifyBaseName(fileName string) string {
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		return fileName[:len(fileName)-len(".csv")]
	}
	return fileName
}

// exportifyPlaylist keys the playlist by its raw file name: Exportify marks
// duplicate playlist names with a trailing underscore, so the key stays
// unique and stable across exports.
func exportifyPlaylist(baseName string, tracks []Track) Playlist {
	name := strings.TrimSpace(strings.ReplaceAll(baseName, "_", " "))
	return Playlist{Key: exportifyPlaylistKeyPrefix + baseName, Name: name, Tracks: tracks}
}

// parseExportifyTracks reports false when the CSV has no Exportify header.
func parseExportifyTracks(contents []byte) ([]Track, bool) {
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(contents, utf8ByteOrderMark)))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, false
	}
	columns := indexExportifyColumns(header)
	if !columns.hasRequired() {
		return nil, false
	}
	return columns.tracks(readRecords(reader)), true
}

// readRecords returns the rows up to the first malformed one.
func readRecords(reader *csv.Reader) [][]string {
	records := [][]string{}
	for {
		record, err := reader.Read()
		if err != nil {
			return records
		}
		records = append(records, record)
	}
}

func indexExportifyColumns(header []string) exportifyColumns {
	columns := exportifyColumns{}
	for index, name := range header {
		key := strings.ToLower(strings.TrimSpace(name))
		_, seen := columns[key]
		if !seen {
			columns[key] = index
		}
	}
	return columns
}

func (columns exportifyColumns) hasRequired() bool {
	_, hasURI := columns["track uri"]
	_, hasName := columns["track name"]
	return hasURI && hasName
}

func (columns exportifyColumns) tracks(records [][]string) []Track {
	tracks := []Track{}
	for _, record := range records {
		title := columns.cell(record, "track name")
		if title == "" {
			continue
		}
		tracks = append(tracks, Track{
			URI:            columns.cell(record, "track uri"),
			Title:          title,
			Artists:        splitExportifyArtists(columns.cell(record, "artist name(s)")),
			Album:          columns.cell(record, "album name"),
			DurationMillis: parseMillis(columns.cell(record, "duration (ms)")),
		})
	}
	return tracks
}

// cell returns the trimmed value of a named column, empty when the column or
// the record's field is missing.
func (columns exportifyColumns) cell(record []string, name string) string {
	index, found := columns[name]
	if !found || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func parseMillis(value string) int {
	millis, err := strconv.Atoi(value)
	if err != nil || millis < 0 {
		return 0
	}
	return millis
}

// splitExportifyArtists splits on ";" only: a single artist name may hold
// "," or "/".
func splitExportifyArtists(artists string) []string {
	names := []string{}
	for _, name := range strings.Split(artists, ";") {
		name = strings.TrimSpace(name)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

// groupSavedAlbums turns the tracks of Saved_Albums.csv into albums in
// first-seen order. The CSV has no album URI, so albums key by their text.
func groupSavedAlbums(tracks []Track) []Album {
	albums := []Album{}
	seen := map[string]bool{}
	for _, track := range tracks {
		if track.Album == "" {
			continue
		}
		title := strings.ToLower(track.Album)
		if seen[title] {
			continue
		}
		seen[title] = true
		albums = append(albums, savedAlbumOf(track))
	}
	return albums
}

// savedAlbumOf names the album's artist after the track's first artist.
func savedAlbumOf(track Track) Album {
	album := Album{Title: track.Album}
	if len(track.Artists) > 0 {
		album.Artist = track.Artists[0]
	}
	return album
}
