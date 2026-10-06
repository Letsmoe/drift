import type { Artist, Thumbnail, Track } from '#lib/api/music.ts';

export function formatDuration(totalSeconds: number): string {
	// <audio> reports NaN until a source is loaded.
	if (!Number.isFinite(totalSeconds) || totalSeconds <= 0) {
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

export function formatTotalDuration(tracks: Track[]): string {
	const totalSeconds = tracks.reduce((sum, track) => sum + track.durationSeconds, 0);
	const totalMinutes = Math.round(totalSeconds / 60);
	if (totalMinutes < 60) {
		return `${totalMinutes} min`;
	}
	const hours = Math.floor(totalMinutes / 60);
	return `${hours} hr ${totalMinutes % 60} min`;
}

export function formatSongCount(count: number): string {
	if (count === 1) {
		return '1 song';
	}
	return `${count} songs`;
}

/** Formats a timestamp from the core, or returns an empty string when it is not a date. */
export function formatDate(timestamp: unknown): string {
	const date = new Date(String(timestamp));
	if (Number.isNaN(date.getTime())) {
		return '';
	}
	return date.toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function itemHref(kind: string, id: string): string {
	return `/${kind}/${encodeURIComponent(id)}`;
}

// Podcast show browse IDs start with this; an episode's "album" is its show.
const podcastShowPrefix = 'MPSP';

/** The page a track opens: its album (an episode's show), else its first linked artist. */
export function trackHref(track: Track): string | undefined {
	const albumId = track.album?.id;
	if (albumId && albumId.startsWith(podcastShowPrefix)) {
		return itemHref('podcast', albumId);
	}
	if (albumId) {
		return itemHref('album', albumId);
	}
	const artist = track.artists.find((linked) => linked.id !== '');
	if (artist) {
		return itemHref('artist', artist.id);
	}
	return undefined;
}
