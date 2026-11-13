import type { DownloadStatus } from '#lib/api/downloads.ts';

/** Share of a download that is done, or 0 while the size is not known yet. */
export function downloadFraction(status: DownloadStatus | undefined): number {
	if (!status || status.bytesTotal <= 0) {
		return 0;
	}
	return status.bytesDone / status.bytesTotal;
}

/** Percent (0 to 100) for a progress indicator, or undefined while the size is not known, so it spins instead. */
export function progressPercent(fraction: number): number | undefined {
	if (fraction <= 0 || fraction > 1) {
		return undefined;
	}
	return fraction * 100;
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
