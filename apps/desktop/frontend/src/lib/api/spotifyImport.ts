import {
	SpotifyImport,
	SpotifyPreviewExport,
	SpotifyPreviewURL
} from '#lib/backend/generated/App.ts';
import type { ImportOptions, ImportSummary } from '#lib/api/musicImport.ts';
import { pickFile, pickFolder } from '#lib/backend/shell.ts';

/** Opens a file picker; resolves to '' when the user cancels. */
export function pickExportFile(): Promise<string> {
	return pickFile({
		title: 'Choose your Spotify data export',
		filters: [{ name: 'Spotify export (*.zip, *.json)', extensions: ['zip', 'json'] }]
	});
}

/** Opens a folder picker; resolves to '' when the user cancels. */
export function pickExportFolder(): Promise<string> {
	return pickFolder({ title: 'Choose the extracted Spotify data folder' });
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
