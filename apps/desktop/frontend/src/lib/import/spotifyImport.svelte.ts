import { SvelteSet } from 'svelte/reactivity';
import {
	cancelImport,
	importCancelledMessage,
	listenToImport,
	pickExportFile,
	pickExportFolder,
	previewExport,
	previewLink,
	startImport,
	type ImportOptions,
	type ImportProgress,
	type ImportResult,
	type ImportSourceKind,
	type ImportSummary
} from '#lib/api/spotifyImport.ts';
import { describeError } from '#lib/format.ts';
import { library } from '#lib/library/library.svelte.ts';

export type ImportStep = 'source' | 'loading' | 'preview' | 'running' | 'done';

/** The Spotify import flow: choose a source, pick what to import, watch it run, read the summary. */
export class SpotifyImportFlow {
	step = $state<ImportStep>('source');
	sourceKind = $state<ImportSourceKind>('export');
	linkText = $state('');
	summary = $state<ImportSummary | null>(null);
	includeLiked = $state(true);
	includeAlbums = $state(true);
	includeArtists = $state(true);
	includeShows = $state(true);
	selectedPlaylistKeys = new SvelteSet<string>();
	progress = $state<ImportProgress | null>(null);
	result = $state<ImportResult | null>(null);
	errorMessage = $state('');
	cancelled = $state(false);

	#location = '';
	#stopListening: (() => void) | null = null;

	/** 0 to 1 share of the items matched so far. */
	get progressFraction(): number {
		if (!this.progress || this.progress.total === 0) {
			return 0;
		}
		return this.progress.done / this.progress.total;
	}

	get canStart(): boolean {
		if (!this.summary) {
			return false;
		}
		return this.selectedItemCount > 0;
	}

	get selectedItemCount(): number {
		if (!this.summary) {
			return 0;
		}
		let count = this.selectedPlaylistKeys.size;
		count += this.#flagCount(this.includeLiked, this.summary.liked);
		count += this.#flagCount(this.includeAlbums, this.summary.albums);
		count += this.#flagCount(this.includeArtists, this.summary.artists);
		count += this.#flagCount(this.includeShows, this.summary.shows);
		return count;
	}

	async chooseExportFile(): Promise<void> {
		await this.#chooseExport(pickExportFile);
	}

	async chooseExportFolder(): Promise<void> {
		await this.#chooseExport(pickExportFolder);
	}

	async loadLink(): Promise<void> {
		const link = this.linkText.trim();
		if (link === '') {
			return;
		}
		this.#location = link;
		this.sourceKind = 'url';
		await this.#loadPreview(() => previewLink(link));
	}

	togglePlaylist(key: string): void {
		if (this.selectedPlaylistKeys.has(key)) {
			this.selectedPlaylistKeys.delete(key);
			return;
		}
		this.selectedPlaylistKeys.add(key);
	}

	toggleAllPlaylists(): void {
		if (!this.summary) {
			return;
		}
		if (this.selectedPlaylistKeys.size === this.summary.playlists.length) {
			this.selectedPlaylistKeys.clear();
			return;
		}
		for (const playlist of this.summary.playlists) {
			this.selectedPlaylistKeys.add(playlist.key);
		}
	}

	async start(): Promise<void> {
		this.errorMessage = '';
		this.cancelled = false;
		this.progress = null;
		this.#stopListening = listenToImport({
			onProgress: (progress) => {
				this.progress = progress;
			},
			onDone: (result) => this.#finish(result),
			onError: (message) => this.#fail(message)
		});
		this.step = 'running';
		try {
			await startImport(this.sourceKind, this.#location, this.#options());
		} catch (error) {
			this.#fail(describeError(error));
		}
	}

	async cancel(): Promise<void> {
		await cancelImport();
	}

	/** Back to the source step, e.g. to import something else. */
	restart(): void {
		this.#stopImportListening();
		this.step = 'source';
		this.summary = null;
		this.result = null;
		this.progress = null;
		this.errorMessage = '';
		this.cancelled = false;
	}

	/** Stops listening to a running import's events; the import itself keeps its own lifetime. */
	dispose(): void {
		this.#stopImportListening();
	}

	async #chooseExport(pick: () => Promise<string>): Promise<void> {
		this.errorMessage = '';
		const path = await pick();
		if (path === '') {
			return;
		}
		this.#location = path;
		this.sourceKind = 'export';
		await this.#loadPreview(() => previewExport(path));
	}

	async #loadPreview(load: () => Promise<ImportSummary>): Promise<void> {
		this.step = 'loading';
		this.errorMessage = '';
		try {
			this.summary = await load();
		} catch (error) {
			this.errorMessage = describeError(error);
			this.step = 'source';
			return;
		}
		this.#selectEverything(this.summary);
		this.step = 'preview';
	}

	#selectEverything(summary: ImportSummary): void {
		this.includeLiked = summary.liked > 0;
		this.includeAlbums = summary.albums > 0;
		this.includeArtists = summary.artists > 0;
		this.includeShows = summary.shows > 0;
		this.selectedPlaylistKeys.clear();
		for (const playlist of summary.playlists) {
			this.selectedPlaylistKeys.add(playlist.key);
		}
	}

	#options(): ImportOptions {
		return {
			liked: this.includeLiked,
			playlists: [...this.selectedPlaylistKeys],
			albums: this.includeAlbums,
			artists: this.includeArtists,
			shows: this.includeShows,
			retryUnmatched: false
		};
	}

	#flagCount(included: boolean, count: number): number {
		if (included && count > 0) {
			return 1;
		}
		return 0;
	}

	async #finish(result: ImportResult): Promise<void> {
		this.#stopImportListening();
		this.result = result;
		this.step = 'done';
		await Promise.all([library.refreshPlaylists(), library.refreshSavedItems()]);
	}

	#fail(message: string): void {
		this.#stopImportListening();
		this.cancelled = message === importCancelledMessage;
		if (!this.cancelled) {
			this.errorMessage = message;
		}
		this.step = 'preview';
	}

	#stopImportListening(): void {
		if (this.#stopListening) {
			this.#stopListening();
			this.#stopListening = null;
		}
	}
}
