import { SvelteMap, SvelteSet } from 'svelte/reactivity';
import {
	cancelDownload,
	deleteAllDownloads,
	deleteDownload,
	downloadCollection,
	downloadQueue,
	downloadStates,
	downloadSummary,
	downloadTracks,
	downloadedCollections,
	downloadedItems,
	onDownloadUpdate,
	pauseDownloads,
	refreshDownloadedCollection,
	removeDownloadedCollection,
	resumeDownloads,
	retryDownload,
	storageUsage,
	type CollectionKind,
	type DownloadCollectionInfo,
	type DownloadItem,
	type DownloadState,
	type DownloadStatus,
	type DownloadSummaryInfo,
	type DownloadUpdate,
	type StorageUsageInfo
} from '#lib/api/downloads.ts';
import type { ItemKind, Track } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

const listRefreshDelayMilliseconds = 200;

const emptySummary: DownloadSummaryInfo = {
	queued: 0,
	downloading: 0,
	failed: 0,
	done: 0,
	paused: false,
	blocked: false
};

/** Offline downloads: per-track states, collections, and the lists the Downloads page shows. */
export class DownloadsState {
	statuses = new SvelteMap<string, DownloadStatus>();
	summary = $state<DownloadSummaryInfo>(emptySummary);
	collections = $state<DownloadCollectionInfo[]>([]);
	queueItems = $state<DownloadItem[]>([]);
	downloadedItems = $state<DownloadItem[]>([]);
	usage = $state<StorageUsageInfo>({ bytes: 0, count: 0 });
	errorMessage = $state('');

	#requestedIds = new SvelteSet<string>();
	#pendingIds = new SvelteSet<string>();
	#listWatchers = 0;
	#listRefreshTimer: ReturnType<typeof setTimeout> | undefined;
	#ready: Promise<void> = Promise.resolve();

	/** Loads collections and the summary and starts listening for updates; call once at startup. */
	start(): void {
		onDownloadUpdate((update) => this.#applyUpdate(update));
		this.#ready = this.#refreshCollections();
		this.#refreshSummary();
	}

	/** The Downloads page registers while it is open so its lists stay current. */
	watchLists(): () => void {
		this.#listWatchers += 1;
		this.#refreshLists();
		return () => {
			this.#listWatchers -= 1;
		};
	}

	stateOf(videoId: string): DownloadState {
		const status = this.statuses.get(videoId);
		if (!status) {
			return '';
		}
		return status.state as DownloadState;
	}

	statusOf(videoId: string): DownloadStatus | undefined {
		return this.statuses.get(videoId);
	}

	isDownloaded(videoId: string): boolean {
		return this.stateOf(videoId) === 'done';
	}

	/** Looks up the states of videoIds not looked up yet; lookups made in one tick go out as one call. */
	ensureStates(videoIds: string[]): void {
		for (const videoId of videoIds) {
			if (!this.#requestedIds.has(videoId)) {
				this.#requestedIds.add(videoId);
				this.#pendingIds.add(videoId);
			}
		}
		if (this.#pendingIds.size > 0) {
			queueMicrotask(() => this.#flushPending());
		}
	}

	async download(track: Track, kind: ItemKind = 'song'): Promise<void> {
		await this.#attempt(() => downloadTracks([track], kind));
	}

	async cancel(videoId: string): Promise<void> {
		await this.#attempt(() => cancelDownload(videoId));
	}

	async retry(videoId: string): Promise<void> {
		await this.#attempt(() => retryDownload(videoId));
	}

	async remove(videoId: string): Promise<void> {
		await this.#attempt(() => deleteDownload(videoId));
	}

	async deleteAll(): Promise<void> {
		await this.#attempt(() => deleteAllDownloads());
		await this.#refreshCollections();
	}

	async pause(): Promise<void> {
		await this.#attempt(() => pauseDownloads());
	}

	async resume(): Promise<void> {
		await this.#attempt(() => resumeDownloads());
	}

	collection(kind: CollectionKind, id: string): DownloadCollectionInfo | undefined {
		return this.collections.find((candidate) => candidate.kind === kind && candidate.id === id);
	}

	/** Downloads a whole collection and keeps it in sync as it changes. */
	async keepDownloaded(kind: CollectionKind, id: string): Promise<void> {
		await this.#attempt(() => downloadCollection(kind, id, true));
		await this.#refreshCollections();
	}

	/** Deletes the collection's downloads that no other collection or single download needs. */
	async removeCollection(kind: CollectionKind, id: string): Promise<void> {
		await this.#attempt(() => removeDownloadedCollection(kind, id));
		await this.#refreshCollections();
	}

	/** Called when a collection's page opens: kept YouTube collections check for new tracks. */
	async refreshOnOpen(kind: CollectionKind, id: string): Promise<void> {
		await this.#ready;
		const collection = this.collection(kind, id);
		if (!collection || !collection.keep) {
			return;
		}
		try {
			await refreshDownloadedCollection(kind, id);
			await this.#refreshCollections();
		} catch {
			// Offline or YouTube unreachable: the stored copy stays as it is.
		}
	}

	async #attempt(action: () => Promise<void>): Promise<void> {
		try {
			await action();
			this.errorMessage = '';
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	async #flushPending(): Promise<void> {
		if (this.#pendingIds.size === 0) {
			return;
		}
		const videoIds = [...this.#pendingIds];
		this.#pendingIds.clear();
		try {
			const found = await downloadStates(videoIds);
			this.#storeStates(videoIds, found);
		} catch (error) {
			this.errorMessage = describeError(error);
			for (const videoId of videoIds) {
				this.#requestedIds.delete(videoId);
			}
		}
	}

	#storeStates(videoIds: string[], found: Record<string, DownloadStatus>): void {
		for (const videoId of videoIds) {
			// An event that arrived while the lookup was in flight is newer.
			if (this.statuses.has(videoId)) {
				continue;
			}
			const status = found[videoId];
			if (status) {
				this.statuses.set(videoId, status);
			}
		}
	}

	#applyUpdate(update: DownloadUpdate): void {
		this.summary = update.summary;
		const videoId = update.status.videoId;
		if (videoId === '') {
			this.#scheduleListRefresh();
			return;
		}
		const previousState = this.stateOf(videoId);
		this.#requestedIds.add(videoId);
		if (update.status.state === '') {
			this.statuses.delete(videoId);
		} else {
			this.statuses.set(videoId, update.status);
		}
		if (previousState !== update.status.state) {
			this.#scheduleListRefresh();
		}
	}

	// State changes arrive in bursts (a collection queues dozens of tracks), so lists reload once per burst.
	#scheduleListRefresh(): void {
		if (this.#listRefreshTimer !== undefined) {
			return;
		}
		this.#listRefreshTimer = setTimeout(() => {
			this.#listRefreshTimer = undefined;
			this.#refreshLists();
		}, listRefreshDelayMilliseconds);
	}

	async #refreshLists(): Promise<void> {
		await this.#refreshCollections();
		if (this.#listWatchers === 0) {
			return;
		}
		try {
			const [queueItems, items, usage] = await Promise.all([
				downloadQueue(),
				downloadedItems(),
				storageUsage()
			]);
			this.queueItems = queueItems;
			this.downloadedItems = items;
			this.usage = usage;
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	async #refreshCollections(): Promise<void> {
		try {
			this.collections = await downloadedCollections();
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}

	async #refreshSummary(): Promise<void> {
		try {
			this.summary = await downloadSummary();
		} catch (error) {
			this.errorMessage = describeError(error);
		}
	}
}

export const downloads = new DownloadsState();
