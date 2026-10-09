package lyrics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/letsmoe/drift-music/core/music"
)

func TestParseLRC(t *testing.T) {
	text := "[ar:Alan Walker]\n[00:11.03] You were the shadow\n[00:17.7]Another star\n[01:02.345][00:05.00] Twice\n\nuntimed\n[00:30.00]"
	want := []music.LyricLine{
		{Text: "Twice", StartMs: 5000},
		{Text: "You were the shadow", StartMs: 11030},
		{Text: "Another star", StartMs: 17700},
		{Text: "", StartMs: 30000},
		{Text: "Twice", StartMs: 62345},
	}
	if got := ParseLRC(text); !reflect.DeepEqual(got, want) {
		t.Errorf("ParseLRC = %+v", got)
	}
}

var faded = music.Track{
	VideoID:         "pIWaVJPl0-c",
	Title:           "Faded",
	Artists:         []music.Artist{{Name: "Alan Walker"}},
	DurationSeconds: 212,
}

func TestLRCLIBGetsTheExactTrack(t *testing.T) {
	var requested []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requested = append(requested, request.URL.Path+"?"+request.URL.RawQuery)
		if request.Header.Get("User-Agent") != lrclibUserAgent {
			t.Errorf("user agent = %q", request.Header.Get("User-Agent"))
		}
		_, _ = writer.Write([]byte(`{"trackName":"Faded","duration":212,"syncedLyrics":"[00:11.03] You were the shadow","plainLyrics":"You were the shadow"}`))
	}))
	defer server.Close()

	lyrics, found, err := NewLRCLIB(nil, server.URL).Find(t.Context(), faded)
	if err != nil || !found {
		t.Fatalf("found %v, err %v", found, err)
	}
	want := music.Lyrics{Lines: []music.LyricLine{{Text: "You were the shadow", StartMs: 11030}}, Source: "LRCLIB", Synced: true}
	if !reflect.DeepEqual(lyrics, want) {
		t.Errorf("lyrics = %+v", lyrics)
	}
	if len(requested) != 1 || requested[0] != "/api/get?artist_name=Alan+Walker&duration=212&track_name=Faded" {
		t.Errorf("requests = %v", requested)
	}
}

func TestLRCLIBSearchesWhenGetMisses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/get" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = writer.Write([]byte(`[
			{"trackName":"Faded","duration":180,"syncedLyrics":"[00:01.00] wrong version"},
			{"trackName":"Faded","duration":213,"plainLyrics":"plain"},
			{"trackName":"faded","duration":211,"syncedLyrics":"[00:02.00] right version"}
		]`))
	}))
	defer server.Close()

	lyrics, found, err := NewLRCLIB(nil, server.URL).Find(t.Context(), faded)
	if err != nil || !found || !lyrics.Synced || lyrics.Lines[0].Text != "right version" {
		t.Errorf("lyrics %+v, found %v, err %v", lyrics, found, err)
	}
}

func TestLRCLIBSkipsInstrumentals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`{"trackName":"Faded","duration":212,"instrumental":true}`))
	}))
	defer server.Close()

	if _, found, err := NewLRCLIB(nil, server.URL).Find(t.Context(), faded); found || err != nil {
		t.Errorf("found %v, err %v", found, err)
	}
}

type fakeSource struct {
	lyrics music.Lyrics
	found  bool
	err    error
	calls  int
}

func (source *fakeSource) Lyrics(context.Context, string) (music.Lyrics, bool, error) {
	source.calls++
	return source.lyrics, source.found, source.err
}

func (source *fakeSource) Find(context.Context, music.Track) (music.Lyrics, bool, error) {
	source.calls++
	return source.lyrics, source.found, source.err
}

var (
	syncedLyrics = music.Lyrics{Lines: []music.LyricLine{{Text: "a", StartMs: 1}}, Synced: true}
	plainLyrics  = music.Lyrics{Lines: []music.LyricLine{{Text: "a"}}}
)

func TestServicePrefersYouTubeSyncedLyrics(t *testing.T) {
	youtube := &fakeSource{lyrics: syncedLyrics, found: true}
	lrclib := &fakeSource{lyrics: syncedLyrics, found: true}

	if _, found, err := New(youtube, lrclib, nil).Lyrics(t.Context(), faded, Options{AllowLRCLIB: true}); !found || err != nil {
		t.Fatalf("found %v, err %v", found, err)
	}
	if lrclib.calls != 0 {
		t.Error("asked LRCLIB although YouTube had synced lyrics")
	}
}

func TestServiceAsksLRCLIBOnlyWhenAllowed(t *testing.T) {
	youtube := &fakeSource{lyrics: plainLyrics, found: true}
	lrclib := &fakeSource{lyrics: syncedLyrics, found: true}
	service := New(youtube, lrclib, nil)

	lyrics, _, _ := service.Lyrics(t.Context(), faded, Options{})
	if lrclib.calls != 0 || lyrics.Synced {
		t.Errorf("LRCLIB calls = %d without permission", lrclib.calls)
	}
	lyrics, _, _ = service.Lyrics(t.Context(), faded, Options{AllowLRCLIB: true})
	if !lyrics.Synced {
		t.Error("LRCLIB's synced lyrics should win over YouTube's plain ones")
	}
}

func TestServiceKeepsYouTubePlainLyricsOverLRCLIBPlain(t *testing.T) {
	youtubePlain := music.Lyrics{Lines: plainLyrics.Lines, Source: "Musixmatch"}
	youtube := &fakeSource{lyrics: youtubePlain, found: true}
	lrclib := &fakeSource{lyrics: plainLyrics, found: true}

	lyrics, found, _ := New(youtube, lrclib, nil).Lyrics(t.Context(), faded, Options{AllowLRCLIB: true})
	if !found || lyrics.Source != "Musixmatch" {
		t.Errorf("lyrics = %+v", lyrics)
	}
}

func TestServiceFallsBackToLRCLIBWhenYouTubeFails(t *testing.T) {
	youtube := &fakeSource{err: errors.New("offline")}
	lrclib := &fakeSource{lyrics: plainLyrics, found: true}

	if _, found, err := New(youtube, lrclib, nil).Lyrics(t.Context(), faded, Options{AllowLRCLIB: true}); !found || err != nil {
		t.Errorf("found %v, err %v", found, err)
	}
}
