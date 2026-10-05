package palette

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// paint fills the image row by row: each region covers share of the rows.
func paint(regions []struct {
	share float64
	color color.RGBA
},
) *image.RGBA {
	const size = 100
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	row := 0
	for _, region := range regions {
		rows := int(region.share * size)
		for y := row; y < row+rows; y++ {
			for x := range size {
				img.Set(x, y, region.color)
			}
		}
		row += rows
	}
	return img
}

type region = struct {
	share float64
	color color.RGBA
}

func TestExtractFindsDominantAndVibrant(t *testing.T) {
	img := paint([]region{
		{0.7, color.RGBA{60, 70, 90, 255}},   // muted slate background
		{0.25, color.RGBA{0, 0, 0, 255}},     // black bars
		{0.05, color.RGBA{240, 40, 30, 255}}, // red accent
	})

	result := Extract(img)

	if result.Dominant != "#3c465a" {
		t.Errorf("dominant = %s, want #3c465a", result.Dominant)
	}
	if result.Vibrant != "#f0281e" {
		t.Errorf("vibrant = %s, want #f0281e", result.Vibrant)
	}
}

func TestExtractIgnoresSpecksForVibrant(t *testing.T) {
	img := paint([]region{
		{0.995, color.RGBA{120, 120, 120, 255}},
	})
	img.Set(0, 99, color.RGBA{255, 0, 0, 255})

	result := Extract(img)

	if result.Vibrant != result.Dominant {
		t.Errorf("vibrant = %s, want the dominant %s for a grey image", result.Vibrant, result.Dominant)
	}
}

func TestExtractFallsBackOnBlackImage(t *testing.T) {
	img := paint([]region{{1, color.RGBA{5, 5, 5, 255}}})

	result := Extract(img)

	if result.Dominant != "#050505" || result.Vibrant != "#050505" {
		t.Errorf("got %+v, want both #050505", result)
	}
}

func TestExtractTransparentImageIsEmpty(t *testing.T) {
	result := Extract(image.NewRGBA(image.Rect(0, 0, 10, 10)))

	if result != (Palette{}) {
		t.Errorf("got %+v, want empty", result)
	}
}

func TestServiceFetchesOncePerURL(t *testing.T) {
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, paint([]region{{1, color.RGBA{10, 200, 90, 255}}})); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = writer.Write(encoded.Bytes())
	}))
	defer server.Close()
	service := NewService(server.Client())

	for range 2 {
		result, err := service.Colors(context.Background(), server.URL+"/cover.png")
		if err != nil {
			t.Fatal(err)
		}
		if result.Vibrant != "#0ac85a" {
			t.Errorf("vibrant = %s, want #0ac85a", result.Vibrant)
		}
	}
	if requests.Load() != 1 {
		t.Errorf("fetched %d times, want 1", requests.Load())
	}
}

func TestServiceReportsBadStatus(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	_, err := NewService(server.Client()).Colors(context.Background(), server.URL)
	if err == nil {
		t.Fatal("want an error for a 404")
	}
}
