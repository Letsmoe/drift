package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/neoworks/otter-music/core/music"
)

const maxCoverBytes = 4 << 20

// fetchCoverQuietly stores the track's largest thumbnail and returns its path,
// or "" when there is none or it failed: a missing cover never fails a download.
func (manager *Manager) fetchCoverQuietly(ctx context.Context, videoID string) string {
	downloads, err := manager.loadDownloads(ctx, `WHERE video_id = ?`, videoID)
	if err != nil || len(downloads) == 0 {
		return ""
	}
	thumbnail, found := largestThumbnail(downloads[0].item.Track.Thumbnails)
	if !found {
		return ""
	}
	path, err := manager.fetchCover(ctx, videoID, thumbnail.URL)
	if err != nil {
		manager.logger.Debug("download cover failed", "videoId", videoID, "error", err)
		return ""
	}
	return path
}

func largestThumbnail(thumbnails []music.Thumbnail) (music.Thumbnail, bool) {
	var largest music.Thumbnail
	found := false
	for _, thumbnail := range thumbnails {
		if thumbnail.URL == "" {
			continue
		}
		if !found || thumbnail.Width*thumbnail.Height >= largest.Width*largest.Height {
			largest = thumbnail
			found = true
		}
	}
	return largest, found
}

func (manager *Manager) fetchCover(ctx context.Context, videoID, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build cover request: %w", err)
	}
	response, err := manager.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("cover request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cover answered status %d", response.StatusCode)
	}
	path := manager.coverPath(videoID, response.Header.Get("Content-Type"))
	return path, writeCover(path, response.Body)
}

func writeCover(path string, body io.Reader) error {
	partPath := path + partSuffix
	file, err := os.OpenFile(partPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open cover: %w", err)
	}
	_, copyErr := io.Copy(file, io.LimitReader(body, maxCoverBytes))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(partPath)
		return fmt.Errorf("write cover: %w", errorsFirst(copyErr, closeErr))
	}
	err = os.Rename(partPath, path)
	if err != nil {
		return fmt.Errorf("finish cover: %w", err)
	}
	return nil
}

func errorsFirst(first, second error) error {
	if first != nil {
		return first
	}
	return second
}
