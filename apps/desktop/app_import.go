package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/neoworks/otter-music/core/importer"
)

// Events the frontend listens to while an import (Spotify or YouTube Music) runs.
const (
	importProgressEvent = "import-progress"
	importDoneEvent     = "import-done"
	importErrorEvent    = "import-error"
	// importCancelledMessage is the importErrorEvent payload after CancelImport.
	importCancelledMessage = "cancelled"
)

var errImportUnavailable = errors.New("importing needs the local library, which could not be opened")

func (app *App) requireImporter() (*importer.Importer, error) {
	if app.importer == nil {
		return nil, errImportUnavailable
	}
	return app.importer, nil
}

// SpotifyPreviewExport counts what a Spotify data export (zip, folder or JSON file) holds.
func (app *App) SpotifyPreviewExport(path string) (importer.Summary, error) {
	return app.previewSpotify(importer.Source{Kind: importer.SourceExport, Location: path})
}

// SpotifyPreviewURL counts what a public Spotify playlist or album link holds.
func (app *App) SpotifyPreviewURL(link string) (importer.Summary, error) {
	return app.previewSpotify(importer.Source{Kind: importer.SourceURL, Location: link})
}

func (app *App) previewSpotify(source importer.Source) (importer.Summary, error) {
	runner, err := app.requireImporter()
	if err != nil {
		return importer.Summary{}, err
	}
	return runner.Preview(app.ctx, source)
}

// SpotifyImport starts an import in the background (sourceKind is "export" or
// "url") and reports through the import-progress, import-done and
// import-error events.
func (app *App) SpotifyImport(sourceKind, pathOrURL string, options importer.Options) error {
	return app.startImport(importer.Source{Kind: sourceKind, Location: pathOrURL}, options)
}

func (app *App) startImport(source importer.Source, options importer.Options) error {
	runner, err := app.requireImporter()
	if err != nil {
		return err
	}
	return runner.Start(app.ctx, source, options,
		func(progress importer.Progress) { runtime.EventsEmit(app.ctx, importProgressEvent, progress) },
		app.reportImportEnd)
}

func (app *App) reportImportEnd(result importer.Result, err error) {
	if errors.Is(err, context.Canceled) {
		runtime.EventsEmit(app.ctx, importErrorEvent, importCancelledMessage)
		return
	}
	if err != nil {
		runtime.EventsEmit(app.ctx, importErrorEvent, err.Error())
		return
	}
	runtime.EventsEmit(app.ctx, importDoneEvent, result)
}

// CancelImport stops the running import, whatever its source.
func (app *App) CancelImport() {
	if app.importer != nil {
		app.importer.Cancel()
	}
}

// PickSpotifyExport opens a file picker for the data export zip or a JSON
// file; it returns an empty path when the user cancels.
func (app *App) PickSpotifyExport() (string, error) {
	path, err := runtime.OpenFileDialog(app.ctx, runtime.OpenDialogOptions{
		Title: "Choose your Spotify data export",
		Filters: []runtime.FileFilter{
			{DisplayName: "Spotify export (*.zip, *.json)", Pattern: "*.zip;*.json"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("open file dialog: %w", err)
	}
	return path, nil
}

// PickSpotifyExportFolder opens a folder picker for an extracted data export;
// it returns an empty path when the user cancels.
func (app *App) PickSpotifyExportFolder() (string, error) {
	path, err := runtime.OpenDirectoryDialog(app.ctx, runtime.OpenDialogOptions{
		Title: "Choose the extracted Spotify data folder",
	})
	if err != nil {
		return "", fmt.Errorf("open folder dialog: %w", err)
	}
	return path, nil
}
