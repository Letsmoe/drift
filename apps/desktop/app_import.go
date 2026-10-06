package main

import (
	"context"
	"errors"

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
		func(progress importer.Progress) { app.emit(importProgressEvent, progress) },
		app.reportImportEnd)
}

func (app *App) reportImportEnd(result importer.Result, err error) {
	if errors.Is(err, context.Canceled) {
		app.emit(importErrorEvent, importCancelledMessage)
		return
	}
	if err != nil {
		app.emit(importErrorEvent, err.Error())
		return
	}
	app.emit(importDoneEvent, result)
}

// CancelImport stops the running import, whatever its source.
func (app *App) CancelImport() {
	if app.importer != nil {
		app.importer.Cancel()
	}
}
