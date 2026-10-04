import type { DownloadStatus } from '#lib/api/downloads.ts';

/** Share of a download that is done, or 0 while the size is not known yet. */
export function downloadFraction(status: DownloadStatus | undefined): number {
	if (!status || status.bytesTotal <= 0) {
		return 0;
	}
	return status.bytesDone / status.bytesTotal;
}

export function formatBytes(bytes: number): string {
	if (bytes < 1024) {
		return `${bytes} B`;
	}
	const units = ['KB', 'MB', 'GB', 'TB'];
	let value = bytes / 1024;
	let unitIndex = 0;
	while (value >= 1024 && unitIndex < units.length - 1) {
		value /= 1024;
		unitIndex += 1;
	}
	return `${value.toFixed(1)} ${units[unitIndex]}`;
}
