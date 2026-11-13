import {
	SpotifyImport,
	SpotifyPreviewExport,
	SpotifyPreviewURL
} from '#lib/backend/generated/App.ts';
import type { ImportOptions, ImportSummary } from '#lib/api/musicImport.ts';
import { pickFile, pickFolder } from '#lib/backend/shell.ts';

const exportExtensions = ['zip', 'json', 'csv'];

function hasExportExtension(path: string): boolean {
	const lowerCasePath = path.toLowerCase();
	return exportExtensions.some((extension) => lowerCasePath.endsWith(`.${extension}`));
}

/**
 * Opens a file picker; resolves to '' when the user cancels. The picker has no file filter, so the
 * extension is checked after the pick.
 */
export async function pickExportFile(): Promise<string> {
	const path = await pickFile({
		title: 'Choose your Spotify data export',
		filters: [{ name: 'Spotify export (*.zip, *.json, *.csv)', extensions: exportExtensions }]
	});
	if (path === '' || hasExportExtension(path)) {
		return path;
	}
	throw new Error('Choose a .zip, .json or .csv file.');
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
