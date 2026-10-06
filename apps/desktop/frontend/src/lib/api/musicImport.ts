import { CancelImport } from '#lib/backend/generated/App.ts';
import type { importer } from '#lib/backend/generated/models.ts';
import { onEvent } from '#lib/backend/rpc.ts';

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

/** Where an import reads from: the signed-in YouTube Music account or Spotify data. */
export type ImportSourceKind = 'youtube-account' | 'export' | 'url';

/** What the core reports as the import error after the user cancelled it. */
export const importCancelledMessage = 'cancelled';

/** Stops the running import, whatever its source. */
export function cancelImport(): Promise<void> {
	return CancelImport();
}

export interface ImportListener {
	onProgress: (progress: ImportProgress) => void;
	onDone: (result: ImportResult) => void;
	onError: (message: string) => void;
}

/** Listens to the events of a running import; returns the function that stops listening. */
export function listenToImport(listener: ImportListener): () => void {
	const stopProgress = onEvent('import-progress', listener.onProgress);
	const stopDone = onEvent('import-done', listener.onDone);
	const stopError = onEvent('import-error', listener.onError);
	return () => {
		stopProgress();
		stopDone();
		stopError();
	};
}
