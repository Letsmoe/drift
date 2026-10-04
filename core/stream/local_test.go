package stream

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type fakeLocalFiles struct {
	audio LocalFile
	cover LocalFile
}

func (files fakeLocalFiles) AudioFile(videoID string) (LocalFile, bool) {
	return files.audio, videoID == "video"
}

func (files fakeLocalFiles) CoverFile(videoID string) (LocalFile, bool) {
	return files.cover, videoID == "video"
}

func writeLocalFile(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProxyServesDownloadedAudioWithoutTheCDN(t *testing.T) {
	cdn := newFakeCDN(t, 0)
	proxy, players := newTestProxy(cdn.URL)
	proxy.SetLocalFiles(fakeLocalFiles{audio: LocalFile{
		Path:     writeLocalFile(t, "video.webm", "local-audio"),
		MimeType: "audio/webm",
	}})

	response := requestRange(t, proxy, "bytes=6-10")

	if response.statusCode != http.StatusPartialContent || response.body != "audio" {
		t.Errorf("got %d %q, want 206 \"audio\"", response.statusCode, response.body)
	}
	if response.header.Get("Content-Type") != "audio/webm" {
		t.Errorf("Content-Type = %q", response.header.Get("Content-Type"))
	}
	if cdn.requestCount.Load() != 0 || len(players.calls) != 0 {
		t.Error("a downloaded track must not touch the network")
	}
}

func TestProxyFallsBackToCDNForOtherVideos(t *testing.T) {
	cdn := newFakeCDN(t, 0)
	proxy, _ := newTestProxy(cdn.URL)
	proxy.SetLocalFiles(fakeLocalFiles{})

	server := requestRange(t, proxy, "")

	if server.body != audioBytes {
		t.Errorf("body = %q, want the CDN's bytes", server.body)
	}
}

func TestProxyServesCover(t *testing.T) {
	proxy, _ := newTestProxy("http://unused.invalid")
	proxy.SetLocalFiles(fakeLocalFiles{cover: LocalFile{
		Path:     writeLocalFile(t, "video.jpg", "jpeg-bytes"),
		MimeType: "image/jpeg",
	}})

	response := requestPath(t, proxy, CoverPath("video"))

	if response.statusCode != http.StatusOK || response.body != "jpeg-bytes" {
		t.Errorf("got %d %q", response.statusCode, response.body)
	}
	if requestPath(t, proxy, CoverPath("other")).statusCode != http.StatusNotFound {
		t.Error("unknown cover should be 404")
	}
}
