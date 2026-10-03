// Command desktop is the Otter Music desktop app: the SvelteKit UI from frontend/
// running in a Wails webview, bound to the Go core.
package main

import (
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// frontend/build is the SvelteKit static build (`bun run build` in frontend/).
//
//go:embed all:frontend/build
var assets embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel()}))
	app := NewApp(logger)

	err := wails.Run(&options.App{
		Title:            "Otter Music",
		Width:            1200,
		Height:           800,
		MinWidth:         720,
		MinHeight:        480,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 20, G: 20, B: 22, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []any{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// logLevel enables debug logs (e.g. every stream request) with OTTER_MUSIC_LOG=debug.
func logLevel() slog.Level {
	if os.Getenv("OTTER_MUSIC_LOG") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
