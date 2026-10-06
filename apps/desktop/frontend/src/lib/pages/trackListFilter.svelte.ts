import type { Track } from '#lib/api/music.ts';
import { formatArtists } from '#lib/format.ts';

export type TrackSort = 'custom' | 'title' | 'artist' | 'album' | 'duration';

export const trackSorts: { value: TrackSort; label: string }[] = [
	{ value: 'custom', label: 'Custom order' },
	{ value: 'title', label: 'Title' },
	{ value: 'artist', label: 'Artist' },
	{ value: 'album', label: 'Album' },
	{ value: 'duration', label: 'Duration' }
];

/** A query and sort order applied to a track list; "custom" keeps the list's own order. */
export class TrackListFilter {
	query = $state('');
	sort = $state<TrackSort>('custom');

	/** Whether the whole list is needed, so every page should be loaded. */
	get active(): boolean {
		return this.query.trim() !== '' || this.sort !== 'custom';
	}

	apply(tracks: Track[]): Track[] {
		const matching = this.#matching(tracks);
		if (this.sort === 'custom') {
			return matching;
		}
		return [...matching].sort((first, second) => this.#compare(first, second));
	}

	#matching(tracks: Track[]): Track[] {
		const needle = this.query.trim().toLocaleLowerCase();
		if (needle === '') {
			return tracks;
		}
		return tracks.filter((track) => searchableText(track).includes(needle));
	}

	#compare(first: Track, second: Track): number {
		if (this.sort === 'duration') {
			return first.durationSeconds - second.durationSeconds;
		}
		return sortKey(first, this.sort).localeCompare(sortKey(second, this.sort));
	}
}

function searchableText(track: Track): string {
	return [track.title, formatArtists(track.artists), albumName(track)]
		.join(' ')
		.toLocaleLowerCase();
}

function sortKey(track: Track, sort: TrackSort): string {
	if (sort === 'artist') {
		return formatArtists(track.artists);
	}
	if (sort === 'album') {
		return albumName(track);
	}
	return track.title;
}

function albumName(track: Track): string {
	if (track.album) {
		return track.album.name;
	}
	return '';
}
