import { search } from '#lib/api/catalog.ts';
import { clearSearches, recentSearches } from '#lib/api/library.ts';
import type { Item, Shelf, Track } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

export type SearchStatus = 'idle' | 'loading' | 'done' | 'failed';

const topResultTitle = 'Top result';
const songsTitle = 'Songs';
const recentSearchLimit = 10;

/** The tracks of a shelf that only holds songs, videos or episodes, or undefined for any other shelf. */
export function shelfTracks(shelf: Shelf): Track[] | undefined {
	const tracks: Track[] = [];
	for (const item of shelf.items) {
		if (!item.track) {
			return undefined;
		}
		tracks.push(item.track);
	}
	return tracks;
}

/** Results of one search plus the filter chip selected on the results page. */
export class SearchPageState {
	shelves = $state<Shelf[]>([]);
	status = $state<SearchStatus>('idle');
	errorMessage = $state('');
	/** Title of the shelf the results are narrowed to; empty shows everything. */
	filter = $state('');
	recentQueries = $state<string[]>([]);

	filterTitles = $derived(this.shelves.map((shelf) => shelf.title).filter(isFilterable));
	topResult = $derived(findTopResult(this.shelves));
	songsShelf = $derived(this.shelves.find((shelf) => shelf.title === songsTitle));
	otherShelves = $derived(this.shelves.filter(isCardRowShelf));
	filteredShelf = $derived(this.shelves.find((shelf) => shelf.title === this.filter));

	// Responses can arrive out of order; only the latest request may update state.
	#latestRequestId = 0;

	async run(query: string): Promise<void> {
		const trimmedQuery = query.trim();
		this.filter = '';
		if (trimmedQuery === '') {
			this.status = 'idle';
			return;
		}

		this.#latestRequestId += 1;
		const requestId = this.#latestRequestId;
		this.status = 'loading';
		try {
			const results = await search(trimmedQuery);
			if (requestId === this.#latestRequestId) {
				this.shelves = results.shelves;
				this.status = 'done';
			}
		} catch (error) {
			if (requestId === this.#latestRequestId) {
				this.errorMessage = describeError(error);
				this.status = 'failed';
			}
		}
	}

	async loadRecent(): Promise<void> {
		try {
			this.recentQueries = await recentSearches(recentSearchLimit);
		} catch {
			this.recentQueries = [];
		}
	}

	async clearRecent(): Promise<void> {
		await clearSearches();
		this.recentQueries = [];
	}

	select(title: string): void {
		this.filter = title;
	}
}

function isFilterable(title: string): boolean {
	return title !== topResultTitle;
}

function isCardRowShelf(shelf: Shelf): boolean {
	return shelf.title !== topResultTitle && shelf.title !== songsTitle;
}

function findTopResult(shelves: Shelf[]): Item | undefined {
	const topShelf = shelves.find((shelf) => shelf.title === topResultTitle);
	if (topShelf) {
		return topShelf.items[0];
	}
	return shelves[0]?.items[0];
}
