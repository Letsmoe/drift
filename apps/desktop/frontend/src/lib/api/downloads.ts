import {
	CancelDownload,
	DeleteAllDownloads,
	DeleteDownload,
	DownloadCollection,
	DownloadQueue,
	DownloadStates,
	DownloadSummary,
	DownloadTracks,
	DownloadedCollections,
	DownloadedItems,
	PauseDownloads,
	RefreshDownloadedCollection,
	RemoveDownloadedCollection,
	ResumeDownloads,
	RetryDownload,
	SetKeepDownloaded,
	StorageUsage
} from '#lib/wailsjs/go/main/App.js';
import { EventsOn } from '#lib/wailsjs/runtime/runtime.js';
import type { download } from '#lib/wailsjs/go/models.ts';
import type { ItemKind, Track } from '#lib/api/music.ts';

export type DownloadStatus = download.Status;
export type DownloadItem = download.Item;
export type DownloadCollectionInfo = download.Collection;
export type DownloadSummaryInfo = download.Summary;
export type StorageUsageInfo = download.Usage;

/** `paused` is reported while the whole queue is paused; an empty state means "not downloaded". */
export type DownloadState = '' | 'queued' | 'downloading' | 'done' | 'failed' | 'paused';

/** Where a collection's tracks come from; `likes` always has the id `liked`. */
export type CollectionKind = 'album' | 'playlist' | 'local' | 'likes';

export const likedCollectionId = 'liked';

/** One download changing, with the queue summary at that moment. */
export interface DownloadUpdate {
	status: DownloadStatus;
	summary: DownloadSummaryInfo;
}

const downloadsEventName = 'downloads';

/** Subscribes to download progress and state changes. Returns the unsubscribe function. */
export function onDownloadUpdate(callback: (update: DownloadUpdate) => void): () => void {
	return EventsOn(downloadsEventName, callback);
}

export function downloadTracks(tracks: Track[], kind: ItemKind): Promise<void> {
	return DownloadTracks(tracks, kind);
}

export function downloadCollection(kind: CollectionKind, id: string, keep: boolean): Promise<void> {
	return DownloadCollection(kind, id, keep);
}

export function setKeepDownloaded(kind: CollectionKind, id: string, keep: boolean): Promise<void> {
	return SetKeepDownloaded(kind, id, keep);
}

export function refreshDownloadedCollection(kind: CollectionKind, id: string): Promise<void> {
	return RefreshDownloadedCollection(kind, id);
}

export function removeDownloadedCollection(kind: CollectionKind, id: string): Promise<void> {
	return RemoveDownloadedCollection(kind, id);
}

export function cancelDownload(videoId: string): Promise<void> {
	return CancelDownload(videoId);
}

export function retryDownload(videoId: string): Promise<void> {
	return RetryDownload(videoId);
}

export function deleteDownload(videoId: string): Promise<void> {
	return DeleteDownload(videoId);
}

export function deleteAllDownloads(): Promise<void> {
	return DeleteAllDownloads();
}

export function pauseDownloads(): Promise<void> {
	return PauseDownloads();
}

export function resumeDownloads(): Promise<void> {
	return ResumeDownloads();
}

/** Download status of the given videos; videos that were never queued are absent. */
export async function downloadStates(videoIds: string[]): Promise<Record<string, DownloadStatus>> {
	return DownloadStates(videoIds);
}

export function downloadQueue(): Promise<DownloadItem[]> {
	return DownloadQueue();
}

export function downloadedItems(): Promise<DownloadItem[]> {
	return DownloadedItems();
}

export function downloadedCollections(): Promise<DownloadCollectionInfo[]> {
	return DownloadedCollections();
}

export function downloadSummary(): Promise<DownloadSummaryInfo> {
	return DownloadSummary();
}

export function storageUsage(): Promise<StorageUsageInfo> {
	return StorageUsage();
}
