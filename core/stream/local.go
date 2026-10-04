package stream

import (
	"net/http"
	"net/url"
	"os"
	"time"
)

// LocalFile is a file on disk that stands in for a remote stream or cover.
type LocalFile struct {
	Path     string
	MimeType string
}

// LocalFiles finds downloaded media. The proxy asks it before going to the CDN.
type LocalFiles interface {
	AudioFile(videoID string) (LocalFile, bool)
	CoverFile(videoID string) (LocalFile, bool)
}

// SetLocalFiles makes the proxy serve downloaded audio from local and expose
// covers at CoverPath. It is safe to call while the proxy is serving.
func (proxy *Proxy) SetLocalFiles(local LocalFiles) {
	proxy.local.Store(&local)
}

// CoverPath returns the proxy path of videoID's downloaded cover image.
func CoverPath(videoID string) string {
	return "/cover/" + url.PathEscape(videoID)
}

func (proxy *Proxy) localFiles() LocalFiles {
	local := proxy.local.Load()
	if local == nil {
		return nil
	}
	return *local
}

// serveLocalAudio answers from the downloaded file and reports whether it did.
// http.ServeContent handles Range, so seeking works offline too.
func (proxy *Proxy) serveLocalAudio(writer http.ResponseWriter, request *http.Request, videoID string) bool {
	local := proxy.localFiles()
	if local == nil {
		return false
	}
	file, found := local.AudioFile(videoID)
	if !found {
		return false
	}
	served := proxy.serveFile(writer, request, file)
	if served {
		proxy.logger.Debug("stream local", "videoId", videoID, "range", request.Header.Get("Range"))
	}
	return served
}

func (proxy *Proxy) serveCover(writer http.ResponseWriter, request *http.Request) {
	local := proxy.localFiles()
	if local == nil {
		http.NotFound(writer, request)
		return
	}
	file, found := local.CoverFile(request.PathValue("videoId"))
	if !found || !proxy.serveFile(writer, request, file) {
		http.NotFound(writer, request)
	}
}

// serveFile reports false when the file cannot be opened, so callers can fall back.
func (proxy *Proxy) serveFile(writer http.ResponseWriter, request *http.Request, file LocalFile) bool {
	opened, err := os.Open(file.Path)
	if err != nil {
		proxy.logger.Warn("open local file", "path", file.Path, "error", err)
		return false
	}
	defer func() { _ = opened.Close() }()
	info, err := opened.Stat()
	if err != nil {
		return false
	}
	writer.Header().Set("Content-Type", file.MimeType)
	writer.Header().Set("Cache-Control", "no-store")
	http.ServeContent(writer, request, "", info.ModTime().Truncate(time.Second), opened)
	return true
}
