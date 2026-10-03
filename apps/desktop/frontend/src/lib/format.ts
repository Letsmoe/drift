import type { Artist, Thumbnail } from '#lib/api/music.ts';

export function formatDuration(totalSeconds: number): string {
	if (totalSeconds <= 0) {
		return '';
	}
	const hours = Math.floor(totalSeconds / 3600);
	const minutes = Math.floor((totalSeconds % 3600) / 60);
	const seconds = String(totalSeconds % 60).padStart(2, '0');
	if (hours > 0) {
		return `${hours}:${String(minutes).padStart(2, '0')}:${seconds}`;
	}
	return `${minutes}:${seconds}`;
}

export function formatArtists(artists: Artist[]): string {
	return artists.map((artist) => artist.name).join(', ');
}

/** Picks the smallest thumbnail at least `minimumSize` wide, or the largest one available. */
export function pickThumbnail(thumbnails: Thumbnail[], minimumSize: number): string | undefined {
	const bySize = [...thumbnails].sort((first, second) => first.width - second.width);
	const fitting = bySize.find((thumbnail) => thumbnail.width >= minimumSize);
	if (fitting) {
		return fitting.url;
	}
	return bySize.at(-1)?.url;
}

export function describeError(error: unknown): string {
	if (error instanceof Error) {
		return error.message;
	}
	if (typeof error === 'string') {
		return error;
	}
	return 'Something went wrong.';
}
