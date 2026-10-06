import { home, homeContinuation, podcastsHome } from '#lib/api/catalog.ts';
import type { HomePage, Item, Shelf } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

export type FeedStatus = 'idle' | 'loading' | 'done' | 'failed';
export type HomeFilter = 'all' | 'music' | 'podcasts';

function isPodcastItem(item: Item): boolean {
	return item.kind === 'podcast' || item.kind === 'episode';
}

function isPodcastShelf(shelf: Shelf): boolean {
	return shelf.items.length > 0 && shelf.items.every(isPodcastItem);
}

/** One shelf feed, loaded page by page as the user scrolls. */
class ShelfFeed {
	shelves = $state<Shelf[]>([]);
	status = $state<FeedStatus>('idle');
	errorMessage = $state('');
	loadingMore = $state(false);

	#continuation = $state('');
	#loadFirstPage: () => Promise<HomePage>;

	constructor(loadFirstPage: () => Promise<HomePage>) {
		this.#loadFirstPage = loadFirstPage;
	}

	get hasMore(): boolean {
		return this.#continuation !== '';
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			this.#append(await this.#loadFirstPage());
			this.status = 'done';
		} catch (error) {
			this.errorMessage = describeError(error);
			this.status = 'failed';
		}
	}

	async loadMore(): Promise<void> {
		if (!this.hasMore || this.loadingMore) {
			return;
		}

		this.loadingMore = true;
		try {
			this.#append(await homeContinuation(this.#continuation));
		} catch {
			// The first page is already on screen; a failed extra page just ends the feed.
			this.#continuation = '';
		} finally {
			this.loadingMore = false;
		}
	}

	#append(page: HomePage): void {
		this.shelves = [...this.shelves, ...page.shelves];
		this.#continuation = page.continuation;
	}
}

/**
 * The home feed's shelves for the selected filter. "Podcasts" is YouTube's own podcasts
 * feed: the regular home feed has (almost) no podcast-only shelves to filter down to.
 */
export class HomeFeed {
	filter = $state<HomeFilter>('all');

	#homeFeed = new ShelfFeed(home);
	#podcastsFeed = new ShelfFeed(podcastsHome);

	#active = $derived(this.#feedFor(this.filter));
	visibleShelves = $derived(this.#filterShelves(this.#active.shelves, this.filter));

	get status(): FeedStatus {
		return this.#active.status;
	}

	get errorMessage(): string {
		return this.#active.errorMessage;
	}

	get hasMore(): boolean {
		return this.#active.hasMore;
	}

	get shelfCount(): number {
		return this.#active.shelves.length;
	}

	load(): void {
		this.#homeFeed.load();
	}

	select(filter: HomeFilter): void {
		this.filter = filter;
		if (this.#active.status === 'idle') {
			this.#active.load();
		}
	}

	loadMore(): Promise<void> {
		return this.#active.loadMore();
	}

	#feedFor(filter: HomeFilter): ShelfFeed {
		if (filter === 'podcasts') {
			return this.#podcastsFeed;
		}
		return this.#homeFeed;
	}

	#filterShelves(shelves: Shelf[], filter: HomeFilter): Shelf[] {
		if (filter === 'music') {
			return shelves.filter((shelf) => !isPodcastShelf(shelf));
		}
		return shelves;
	}
}
