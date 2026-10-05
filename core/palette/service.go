package palette

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // Covers from YouTube are JPEG,
	_ "image/png"  // PNG
	"io"
	"net/http"
	"sync"

	_ "golang.org/x/image/webp" // or WebP.
)

const (
	maxImageBytes = 8 << 20
	// cacheSize bounds the in-memory cache; a palette is two short strings.
	cacheSize = 2048
)

// Service fetches cover images and remembers their palettes by URL.
type Service struct {
	client *http.Client

	mu    sync.Mutex
	cache map[string]Palette
	// order lists the cached URLs oldest first, for eviction.
	order []string
}

// NewService creates a Service that fetches with client.
func NewService(client *http.Client) *Service {
	return &Service{client: client, cache: map[string]Palette{}}
}

// Colors returns the palette of the image at url, fetching it once.
func (service *Service) Colors(ctx context.Context, url string) (Palette, error) {
	if cached, found := service.cached(url); found {
		return cached, nil
	}
	img, err := service.fetch(ctx, url)
	if err != nil {
		return Palette{}, err
	}
	result := Extract(img)
	service.remember(url, result)
	return result, nil
}

func (service *Service) fetch(ctx context.Context, url string) (image.Image, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("palette: request: %w", err)
	}
	response, err := service.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("palette: fetch: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("palette: fetch: status %d", response.StatusCode)
	}
	img, _, err := image.Decode(io.LimitReader(response.Body, maxImageBytes))
	if err != nil {
		return nil, fmt.Errorf("palette: decode: %w", err)
	}
	return img, nil
}

func (service *Service) cached(url string) (Palette, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	result, found := service.cache[url]
	return result, found
}

func (service *Service) remember(url string, result Palette) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if _, found := service.cache[url]; found {
		return
	}
	if len(service.order) >= cacheSize {
		delete(service.cache, service.order[0])
		service.order = service.order[1:]
	}
	service.cache[url] = result
	service.order = append(service.order, url)
}
