import {
	PickSpotifyExport,
	PickSpotifyExportFolder,
	SpotifyImport,
	SpotifyPreviewExport,
	SpotifyPreviewURL
} from '#lib/wailsjs/go/main/App.js';
import type { ImportOptions, ImportSummary } from '#lib/api/musicImport.ts';

/** Opens a file picker; resolves to '' when the user cancels. */
export function pickExportFile(): Promise<string> {
	return PickSpotifyExport();
}

/** Opens a folder picker; resolves to '' when the user cancels. */
export function pickExportFolder(): Promise<string> {
	return PickSpotifyExportFolder();
}

export function previewExport(path: string): Promise<ImportSummary> {
	return SpotifyPreviewExport(path);
}

export function previewLink(link: string): Promise<ImportSummary> {
	return SpotifyPreviewURL(link);
}

export function startSpotifyImport(
	sourceKind: 'export' | 'url',
	pathOrLink: string,
	options: ImportOptions
): Promise<void> {
	return SpotifyImport(sourceKind, pathOrLink, options);
}
