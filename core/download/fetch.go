package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	copyBufferSize = 64 * 1024
	// fullRange is sent when nothing is downloaded yet: without a Range header the
	// CDN throttles to roughly playback speed.
	fullRange = "bytes=0-"
)

var (
	errURLRejected  = errors.New("CDN rejected the stream URL")
	errRangeRefused = errors.New("CDN refused the range of the partial file, restarting")
)

// audioFile is a completely downloaded audio file.
type audioFile struct {
	path     string
	mimeType string
	size     int64
}

// fetchTrack downloads the audio and, best effort, the cover.
func (manager *Manager) fetchTrack(ctx context.Context, videoID string) (finishedFiles, error) {
	audio, err := manager.fetchAudio(ctx, videoID)
	if err != nil {
		return finishedFiles{}, err
	}
	return finishedFiles{
		audioPath: audio.path,
		mimeType:  audio.mimeType,
		sizeBytes: audio.size,
		coverPath: manager.fetchCoverQuietly(ctx, videoID),
	}, nil
}

// fetchAudio retries once with a fresh URL when the CDN rejects the cached one.
func (manager *Manager) fetchAudio(ctx context.Context, videoID string) (audioFile, error) {
	audio, err := manager.transferAudio(ctx, videoID)
	if !errors.Is(err, errURLRejected) {
		return audio, err
	}
	manager.resolver.Invalidate(videoID)
	return manager.transferAudio(ctx, videoID)
}

func (manager *Manager) transferAudio(ctx context.Context, videoID string) (audioFile, error) {
	resolved, err := manager.resolver.Resolve(ctx, videoID)
	if err != nil {
		return audioFile{}, err
	}
	finalPath := manager.audioPath(videoID, resolved.MimeType)
	partPath := finalPath + partSuffix
	removeOtherParts(manager.directory, videoID, partPath)

	response, err := manager.requestFrom(ctx, resolved.URL, fileSize(partPath))
	if err != nil {
		return audioFile{}, err
	}
	defer func() { _ = response.Body.Close() }()

	offset, err := resumeOffset(response, partPath)
	if err != nil {
		return audioFile{}, err
	}
	total := totalBytes(response, offset)
	size, err := manager.writeBody(videoID, response.Body, partPath, offset, total)
	if err != nil {
		return audioFile{}, err
	}
	err = os.Rename(partPath, finalPath)
	if err != nil {
		return audioFile{}, fmt.Errorf("finish %s: %w", videoID, err)
	}
	return audioFile{path: finalPath, mimeType: resolved.MimeType, size: size}, nil
}

// requestFrom asks the CDN for the bytes from offset on.
func (manager *Manager) requestFrom(ctx context.Context, url string, offset int64) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build CDN request: %w", err)
	}
	rangeHeader := fullRange
	if offset > 0 {
		rangeHeader = "bytes=" + strconv.FormatInt(offset, 10) + "-"
	}
	request.Header.Set("Range", rangeHeader)
	response, err := manager.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("CDN request: %w", err)
	}
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusGone {
		_ = response.Body.Close()
		return nil, fmt.Errorf("%w: status %d", errURLRejected, response.StatusCode)
	}
	return response, nil
}

// resumeOffset decides where writing continues. A server that ignores the
// range (200) or refuses it (416) restarts the file from zero.
func resumeOffset(response *http.Response, partPath string) (int64, error) {
	switch response.StatusCode {
	case http.StatusPartialContent:
		return fileSize(partPath), nil
	case http.StatusOK:
		return 0, nil
	case http.StatusRequestedRangeNotSatisfiable:
		_ = os.Remove(partPath)
		return 0, errRangeRefused
	default:
		return 0, fmt.Errorf("CDN answered status %d", response.StatusCode)
	}
}

// totalBytes is the full file size when the response tells it, else 0.
func totalBytes(response *http.Response, offset int64) int64 {
	_, total, found := strings.Cut(response.Header.Get("Content-Range"), "/")
	if found {
		parsed, err := strconv.ParseInt(total, 10, 64)
		if err == nil {
			return parsed
		}
	}
	if response.ContentLength > 0 {
		return offset + response.ContentLength
	}
	return 0
}

// writeBody appends body to the partial file (truncating it when offset is 0)
// and checks the result is complete.
func (manager *Manager) writeBody(videoID string, body io.Reader, partPath string, offset, total int64) (int64, error) {
	flags := os.O_WRONLY | os.O_CREATE
	if offset == 0 {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}
	file, err := os.OpenFile(partPath, flags, 0o600)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", partPath, err)
	}
	written, copyErr := manager.copyWithProgress(videoID, file, body, offset, total)
	closeErr := file.Close()
	if copyErr != nil {
		return 0, copyErr
	}
	if closeErr != nil {
		return 0, fmt.Errorf("close %s: %w", partPath, closeErr)
	}
	if total > 0 && written != total {
		return 0, fmt.Errorf("incomplete transfer: %d of %d bytes", written, total)
	}
	return written, nil
}

// copyWithProgress returns the file size after copying; the partial file keeps
// whatever arrived when the transfer breaks.
func (manager *Manager) copyWithProgress(videoID string, destination io.Writer, source io.Reader, offset, total int64) (int64, error) {
	buffer := make([]byte, copyBufferSize)
	done := offset
	manager.setProgress(videoID, done, total)
	for {
		read, readErr := source.Read(buffer)
		if read > 0 {
			_, err := destination.Write(buffer[:read])
			if err != nil {
				return done, fmt.Errorf("write audio: %w", err)
			}
			done += int64(read)
			manager.setProgress(videoID, done, total)
		}
		if errors.Is(readErr, io.EOF) {
			return done, nil
		}
		if readErr != nil {
			return done, fmt.Errorf("read audio: %w", readErr)
		}
	}
}
