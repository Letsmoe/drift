// Package lyrics finds a track's lyrics: YouTube Music's own (synced when YouTube
// has timestamps) and, when the user allows it, LRCLIB as a fallback for synced lyrics.
package lyrics

import (
	"context"
	"log/slog"

	"github.com/letsmoe/drift/core/music"
)

// YouTube is the part of the innertube client the service needs.
type YouTube interface {
	Lyrics(ctx context.Context, videoID string) (music.Lyrics, bool, error)
}

// Finder looks lyrics up by track metadata (LRCLIB).
type Finder interface {
	Find(ctx context.Context, track music.Track) (music.Lyrics, bool, error)
}

// Options are per-request choices made by the app.
type Options struct {
	// AllowLRCLIB sends title, artist, album and duration to lrclib.net when
	// YouTube has no synced lyrics. Off by default: it is a third-party service.
	AllowLRCLIB bool
}

// Service combines the lyrics sources.
type Service struct {
	youtube YouTube
	lrclib  Finder
	logger  *slog.Logger
}

// New creates a Service; lrclib may be nil to disable the fallback entirely.
func New(youtube YouTube, lrclib Finder, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{youtube: youtube, lrclib: lrclib, logger: logger}
}

// Lyrics returns the best lyrics for track: YouTube's synced lyrics, then LRCLIB's
// synced lyrics, then plain lyrics from either. The bool is false when no source has any.
func (service *Service) Lyrics(ctx context.Context, track music.Track, options Options) (music.Lyrics, bool, error) {
	youtubeLyrics, youtubeFound, youtubeErr := service.youtube.Lyrics(ctx, track.VideoID)
	if youtubeFound && youtubeLyrics.Synced {
		return youtubeLyrics, true, nil
	}
	if !options.AllowLRCLIB || service.lrclib == nil {
		return youtubeLyrics, youtubeFound, youtubeErr
	}

	lrclibLyrics, lrclibFound, lrclibErr := service.lrclib.Find(ctx, track)
	if lrclibErr != nil {
		service.logger.Debug("lrclib lookup failed", "videoId", track.VideoID, "error", lrclibErr)
	}
	if lrclibFound && lrclibLyrics.Synced {
		return lrclibLyrics, true, nil
	}
	if youtubeFound {
		return youtubeLyrics, true, nil
	}
	if lrclibFound {
		return lrclibLyrics, true, nil
	}
	return music.Lyrics{}, false, firstError(youtubeErr, lrclibErr)
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
