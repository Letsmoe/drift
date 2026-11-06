package mobile

import (
	"net/http"
	"time"

	"github.com/letsmoe/drift/core/palette"
)

const paletteFetchTimeout = 10 * time.Second

func newPaletteService() *palette.Service {
	return palette.NewService(&http.Client{Timeout: paletteFetchTimeout})
}

// CoverColors returns the colors of the cover image at url as palette.Palette
// JSON, for tinting the app with the artwork of what is playing.
func (core *Core) CoverColors(url string) (string, error) {
	colors, err := core.palette.Colors(core.ctx, url)
	return encodeResult(colors, err)
}
