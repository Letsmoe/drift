package mobile

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/letsmoe/drift/core/palette"
)

func TestCoverColors(t *testing.T) {
	cover := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := range 8 {
		for y := range 8 {
			cover.Set(x, y, color.RGBA{R: 10, G: 200, B: 90, A: 255})
		}
	}
	var encoded bytes.Buffer
	err := png.Encode(&encoded, cover)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write(encoded.Bytes())
	}))
	defer server.Close()
	core, _ := openTestCore(t)

	colorsJSON, err := core.CoverColors(server.URL + "/cover.png")
	if err != nil {
		t.Fatalf("CoverColors: %v", err)
	}

	colors := decodeInto[palette.Palette](t, colorsJSON)
	if colors.Vibrant != "#0ac85a" {
		t.Errorf("vibrant = %s, want #0ac85a", colors.Vibrant)
	}
	if len(colors.Colors) == 0 {
		t.Errorf("colors is empty, want the cover's color")
	}
}
