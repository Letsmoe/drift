import type { Thumbnail } from '#lib/api/music.ts';
import { itemHref } from '#lib/format.ts';
import { library } from '#lib/library/library.svelte.ts';

export type LibraryFilter = 'all' | 'playlist' | 'album' | 'artist' | 'podcast';

export interface LibraryEntry {
	key: string;
	href: string;
	title: string;
	subtitle: string;
	thumbnails: Thumbnail[];
	round: boolean;
	/** Queue context that plays this entry, to highlight it while it plays. */
	context: string;
}

export const libraryFilters: { filter: LibraryFilter; label: string }[] = [
	{ filter: 'playlist', label: 'Playlists' },
	{ filter: 'album', label: 'Albums' },
	{ filter: 'artist', label: 'Artists' },
	{ filter: 'podcast', label: 'Podcasts' }
];

const kindLabels: Record<string, string> = {
	playlist: 'Playlist',
	album: 'Album',
	artist: 'Artist',
	podcast: 'Podcast'
};

/** The sidebar's rows: the user's own playlists plus saved albums, artists, playlists and podcasts. */
export class LibraryListing {
	filter = $state<LibraryFilter>('all');
	entries = $derived(this.#visibleEntries());

	hasKind(kind: LibraryFilter): boolean {
		return this.#allEntries().some((entry) => entry.key.startsWith(`${kind}:`));
	}

	/** Selecting the active chip clears the filter. */
	toggle(filter: LibraryFilter): void {
		if (this.filter === filter) {
			this.filter = 'all';
			return;
		}
		this.filter = filter;
	}

	#visibleEntries(): LibraryEntry[] {
		const entries = this.#allEntries();
		if (this.filter === 'all') {
			return entries;
		}
		return entries.filter((entry) => entry.key.startsWith(`${this.filter}:`));
	}

	#allEntries(): LibraryEntry[] {
		return [...this.#playlistEntries(), ...this.#savedEntries()];
	}

	#playlistEntries(): LibraryEntry[] {
		return library.playlists.map((summary) => ({
			key: `playlist:library:${summary.id}`,
			href: `/library/playlist/${encodeURIComponent(summary.id)}`,
			title: summary.name,
			subtitle: `Playlist • ${summary.trackCount} songs`,
			thumbnails: summary.thumbnails,
			round: false,
			context: `library:${summary.id}`
		}));
	}

	#savedEntries(): LibraryEntry[] {
		return library.savedItems.map(({ item }) => ({
			key: `${item.kind}:${item.id}`,
			href: itemHref(item.kind, item.id),
			title: item.title,
			subtitle: savedSubtitle(item.kind, item.subtitle),
			thumbnails: item.thumbnails,
			round: item.kind === 'artist',
			context: `${item.kind}:${item.id}`
		}));
	}
}

function savedSubtitle(kind: string, subtitle: string): string {
	const label = kindLabels[kind];
	if (subtitle === '' || kind === 'artist') {
		return label;
	}
	return `${label} • ${subtitle}`;
}
