package mobile

import (
	"context"
	"errors"

	"github.com/neoworks/otter-music/core/importer"
)

var errImportUnavailable = errors.New("importing needs the local library, which could not be opened")

// ImportListener receives the events of a running Spotify import. Implemented in Kotlin.
type ImportListener interface {
	// OnImportProgress receives importer.Progress as JSON.
	OnImportProgress(progressJSON string)
	// OnImportDone ends the import: resultJSON is importer.Result JSON, or
	// empty with errorMessage set when the import failed or was cancelled.
	OnImportDone(resultJSON string, errorMessage string)
}

// SpotifyPreviewExport counts what a Spotify data export (a zip, a folder or
// a JSON file path) holds, as importer.Summary JSON.
func (core *Core) SpotifyPreviewExport(path string) (string, error) {
	return core.previewSource(importer.Source{Kind: importer.SourceExport, Location: path})
}

// SpotifyPreviewURL counts what a public Spotify playlist or album link holds, as importer.Summary JSON.
func (core *Core) SpotifyPreviewURL(link string) (string, error) {
	return core.previewSource(importer.Source{Kind: importer.SourceURL, Location: link})
}

func (core *Core) previewSource(source importer.Source) (string, error) {
	runner, err := core.requireImporter()
	if err != nil {
		return "", err
	}
	summary, err := runner.Preview(core.ctx, source)
	return encodeResult(summary, err)
}

// SpotifyImport starts an import in the background. sourceKind is "export" or
// "url", pathOrURL the file path or link, optionsJSON an importer.Options
// object. Progress and the end arrive on listener; it returns an error only
// when the import could not start (e.g. another one is running).
func (core *Core) SpotifyImport(sourceKind, pathOrURL, optionsJSON string, listener ImportListener) error {
	return core.startImport(importer.Source{Kind: sourceKind, Location: pathOrURL}, optionsJSON, listener)
}

func (core *Core) startImport(source importer.Source, optionsJSON string, listener ImportListener) error {
	runner, err := core.requireImporter()
	if err != nil {
		return err
	}
	var options importer.Options
	err = decode(optionsJSON, &options)
	if err != nil {
		return err
	}
	return runner.Start(core.ctx, source, options,
		func(progress importer.Progress) { listener.OnImportProgress(mustEncode(progress)) },
		func(result importer.Result, runErr error) { reportImportDone(listener, result, runErr) })
}

func reportImportDone(listener ImportListener, result importer.Result, err error) {
	if err != nil {
		listener.OnImportDone("", importFailureMessage(err))
		return
	}
	listener.OnImportDone(mustEncode(result), "")
}

func importFailureMessage(err error) string {
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	return err.Error()
}

// SpotifyCancelImport stops the running import; without one it does nothing.
func (core *Core) SpotifyCancelImport() {
	if core.importer != nil {
		core.importer.Cancel()
	}
}

func (core *Core) requireImporter() (*importer.Importer, error) {
	if core.importer == nil {
		return nil, errImportUnavailable
	}
	return core.importer, nil
}
