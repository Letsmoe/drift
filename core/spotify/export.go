package spotify

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// maxExportFileBytes bounds one data file (JSON or CSV) read from an export.
const maxExportFileBytes = 256 << 20

// ErrNoLibraryData means a path held no Spotify library or playlist data.
var ErrNoLibraryData = errors.New("spotify: no library or playlist data found")

type exportTrack struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Track  string `json:"track"`
	URI    string `json:"uri"`
}

type exportAlbum struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
	URI    string `json:"uri"`
}

type exportArtist struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type exportShow struct {
	Name      string `json:"name"`
	Publisher string `json:"publisher"`
	URI       string `json:"uri"`
}

type exportPlaylistTrack struct {
	TrackName  string `json:"trackName"`
	ArtistName string `json:"artistName"`
	AlbumName  string `json:"albumName"`
	TrackURI   string `json:"trackUri"`
}

type exportPlaylistItem struct {
	Track *exportPlaylistTrack `json:"track"`
}

type exportPlaylist struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Items       []exportPlaylistItem `json:"items"`
}

// exportFile is the union of YourLibrary.json and PlaylistN.json; unknown
// keys are ignored and absent ones stay empty.
type exportFile struct {
	Tracks    []exportTrack    `json:"tracks"`
	Albums    []exportAlbum    `json:"albums"`
	Artists   []exportArtist   `json:"artists"`
	Shows     []exportShow     `json:"shows"`
	Playlists []exportPlaylist `json:"playlists"`
}

// exportFileContents is one data file of an export with its base name; the
// name matters for Exportify, where it is the playlist's name.
type exportFileContents struct {
	name string
	data []byte
}

// ParseExport reads a Spotify data export from a zip, an extracted folder or
// a single file: YourLibrary.json or PlaylistN.json from Spotify's own
// export, or a CSV from Exportify.
func ParseExport(path string) (Library, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Library{}, fmt.Errorf("spotify: open export: %w", err)
	}
	var files []exportFileContents
	switch {
	case info.IsDir():
		files, err = readDirectory(path)
	case strings.EqualFold(filepath.Ext(path), ".zip"):
		files, err = readZip(path)
	default:
		files, err = readSingleFile(path)
	}
	if err != nil {
		return Library{}, err
	}
	return buildLibrary(files)
}

func readSingleFile(path string) ([]exportFileContents, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("spotify: read %s: %w", path, err)
	}
	return []exportFileContents{{name: filepath.Base(path), data: contents}}, nil
}

func readDirectory(root string) ([]exportFileContents, error) {
	var files []exportFileContents
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isExportDataFile(entry.Name()) {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		files = append(files, exportFileContents{name: entry.Name(), data: contents})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("spotify: read export folder: %w", err)
	}
	return files, nil
}

func readZip(path string) ([]exportFileContents, error) {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("spotify: open zip: %w", err)
	}
	defer func() { _ = archive.Close() }()
	var files []exportFileContents
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() || !isExportDataFile(filepath.Base(entry.Name)) {
			continue
		}
		contents, err := readZipEntry(entry)
		if err != nil {
			return nil, err
		}
		files = append(files, exportFileContents{name: filepath.Base(entry.Name), data: contents})
	}
	return files, nil
}

func readZipEntry(entry *zip.File) ([]byte, error) {
	reader, err := entry.Open()
	if err != nil {
		return nil, fmt.Errorf("spotify: open %s: %w", entry.Name, err)
	}
	defer func() { _ = reader.Close() }()
	contents, err := io.ReadAll(io.LimitReader(reader, maxExportFileBytes))
	if err != nil {
		return nil, fmt.Errorf("spotify: read %s: %w", entry.Name, err)
	}
	return contents, nil
}

// isExportDataFile picks the files that hold library data: YourLibrary and
// Playlist JSON files, and any CSV (whether it is an Exportify file is
// decided by its header). The export's streaming history is not read.
func isExportDataFile(name string) bool {
	lowered := strings.ToLower(name)
	if strings.HasSuffix(lowered, ".csv") {
		return true
	}
	if !strings.HasSuffix(lowered, ".json") {
		return false
	}
	return strings.HasPrefix(lowered, "yourlibrary") || strings.HasPrefix(lowered, "playlist")
}

func buildLibrary(files []exportFileContents) (Library, error) {
	library := Library{LikedTracks: []Track{}, Playlists: []Playlist{}, Albums: []Album{}, Artists: []Artist{}, Shows: []Show{}}
	for _, file := range files {
		addFile(&library, file)
	}
	if library.isEmpty() {
		return Library{}, ErrNoLibraryData
	}
	assignPlaylistKeys(library.Playlists)
	return library, nil
}

func addFile(library *Library, file exportFileContents) {
	if strings.HasSuffix(strings.ToLower(file.name), ".csv") {
		addExportifyFile(library, file.name, file.data)
		return
	}
	var parsed exportFile
	err := json.Unmarshal(file.data, &parsed)
	if err != nil {
		return
	}
	addExportFile(library, parsed)
}

func addExportFile(library *Library, parsed exportFile) {
	for _, track := range parsed.Tracks {
		if strings.TrimSpace(track.Track) == "" {
			continue
		}
		library.LikedTracks = append(library.LikedTracks, Track{
			URI: track.URI, Title: track.Track, Artists: splitArtists(track.Artist), Album: track.Album,
		})
	}
	for _, album := range parsed.Albums {
		if strings.TrimSpace(album.Album) != "" {
			library.Albums = append(library.Albums, Album{URI: album.URI, Title: album.Album, Artist: album.Artist})
		}
	}
	for _, artist := range parsed.Artists {
		if strings.TrimSpace(artist.Name) != "" {
			library.Artists = append(library.Artists, Artist{URI: artist.URI, Name: artist.Name})
		}
	}
	for _, show := range parsed.Shows {
		if strings.TrimSpace(show.Name) != "" {
			library.Shows = append(library.Shows, Show{URI: show.URI, Name: show.Name, Publisher: show.Publisher})
		}
	}
	for _, playlist := range parsed.Playlists {
		library.Playlists = append(library.Playlists, convertPlaylist(playlist))
	}
}

func convertPlaylist(playlist exportPlaylist) Playlist {
	converted := Playlist{Name: playlist.Name, Description: playlist.Description, Tracks: []Track{}}
	for _, item := range playlist.Items {
		if item.Track == nil || strings.TrimSpace(item.Track.TrackName) == "" {
			continue
		}
		converted.Tracks = append(converted.Tracks, Track{
			URI:     item.Track.TrackURI,
			Title:   item.Track.TrackName,
			Artists: splitArtists(item.Track.ArtistName),
			Album:   item.Track.AlbumName,
		})
	}
	return converted
}

// splitArtists keeps the export's single artist string whole: names such as
// "Tyler, The Creator" make comma splitting unsafe.
func splitArtists(artist string) []string {
	artist = strings.TrimSpace(artist)
	if artist == "" {
		return []string{}
	}
	return []string{artist}
}

// assignPlaylistKeys makes keys from names, numbering repeated names.
// Playlists that already carry a key (Exportify's) are left as they are.
func assignPlaylistKeys(playlists []Playlist) {
	used := []string{}
	for index := range playlists {
		if playlists[index].Key != "" {
			continue
		}
		key := "export:playlist:" + playlists[index].Name
		for suffix := 2; slices.Contains(used, key); suffix++ {
			key = fmt.Sprintf("export:playlist:%s#%d", playlists[index].Name, suffix)
		}
		used = append(used, key)
		playlists[index].Key = key
	}
}
