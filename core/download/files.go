package download

import (
	"mime"
	"os"
	"path/filepath"
	"strings"
)

var audioExtensions = map[string]string{
	"audio/webm": ".webm",
	"audio/mp4":  ".m4a",
}

var coverExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

const partSuffix = ".part"

func mediaType(contentType string) string {
	parsed, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(contentType))
	}
	return parsed
}

func (manager *Manager) audioPath(videoID, mimeType string) string {
	extension, known := audioExtensions[mediaType(mimeType)]
	if !known {
		extension = ".audio"
	}
	return filepath.Join(manager.directory, videoID+extension)
}

func (manager *Manager) coverPath(videoID, contentType string) string {
	extension, known := coverExtensions[mediaType(contentType)]
	if !known {
		extension = ".jpg"
	}
	return filepath.Join(manager.directory, videoID+".cover"+extension)
}

// fileSize returns 0 for a missing file.
func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// removeDownloadFiles deletes every file stored for videoID: audio, partial and cover.
func removeDownloadFiles(directory, videoID string) {
	matches, err := filepath.Glob(filepath.Join(directory, videoID+".*"))
	if err != nil {
		return
	}
	removeFiles(matches...)
}

// removeOtherParts drops partial files of videoID in a different container
// than keep, which can no longer be resumed.
func removeOtherParts(directory, videoID, keep string) {
	matches, err := filepath.Glob(filepath.Join(directory, videoID+".*"+partSuffix))
	if err != nil {
		return
	}
	for _, match := range matches {
		if match != keep {
			_ = os.Remove(match)
		}
	}
}

func coverMimeType(path string) string {
	extension := filepath.Ext(path)
	for contentType, candidate := range coverExtensions {
		if candidate == extension {
			return contentType
		}
	}
	return "image/jpeg"
}
