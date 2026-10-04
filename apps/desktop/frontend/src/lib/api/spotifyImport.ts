import {
	PickSpotifyExport,
	PickSpotifyExportFolder,
	SpotifyCancelImport,
	SpotifyImport,
	SpotifyPreviewExport,
	SpotifyPreviewURL
} from '#lib/wailsjs/go/main/App.js';
import type { importer } from '#lib/wailsjs/go/models.ts';
import { EventsOn } from '#lib/wailsjs/runtime/runtime.js';

export type ImportSummary = importer.Summary;
export type ImportOptions = importer.Options;
export type ImportPlaylistSummary = importer.PlaylistSummary;

// Event payloads are not part of the generated bindings; these mirror the
// json tags of importer.Progress and importer.Result.
export interface ImportProgress {
	phase: string;
	done: number;
	total: number;
	current: string;
}

export interface ImportCount {
	matched: number;
	unmatched: number;
}

export interface ImportPlaylistResult {
	key: string;
	name: string;
	playlistId: string;
	matched: number;
	unmatched: number;
}

export interface UnmatchedItem {
	kind: string;
	title: string;
	artists: string[];
	playlist: string;
}

export interface ImportResult {
	liked: ImportCount;
	albums: ImportCount;
	artists: ImportCount;
	shows: ImportCount;
	playlists: ImportPlaylistResult[];
	unmatched: UnmatchedItem[];
}

export type ImportSourceKind = 'export' | 'url';

/** What the core reports as the import error after the user cancelled it. */
export const importCancelledMessage = 'cancelled';

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

export function startImport(
	sourceKind: ImportSourceKind,
	pathOrLink: string,
	options: ImportOptions
): Promise<void> {
	return SpotifyImport(sourceKind, pathOrLink, options);
}

export function cancelImport(): Promise<void> {
	return SpotifyCancelImport();
}

export interface ImportListener {
	onProgress: (progress: ImportProgress) => void;
	onDone: (result: ImportResult) => void;
	onError: (message: string) => void;
}

/** Listens to the events of a running import; returns the function that stops listening. */
export function listenToImport(listener: ImportListener): () => void {
	const stopProgress = EventsOn('spotify-import-progress', listener.onProgress);
	const stopDone = EventsOn('spotify-import-done', listener.onDone);
	const stopError = EventsOn('spotify-import-error', listener.onError);
	return () => {
		stopProgress();
		stopDone();
		stopError();
	};
}
