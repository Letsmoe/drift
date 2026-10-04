import { home, homeContinuation } from '#lib/api/catalog.ts';
import type { HomePage, Shelf } from '#lib/api/music.ts';
import { describeError } from '#lib/format.ts';

export type FeedStatus = 'loading' | 'done' | 'failed';

/** The home feed's shelves, loaded page by page as the user scrolls. */
export class HomeFeed {
	shelves = $state<Shelf[]>([]);
	status = $state<FeedStatus>('loading');
	errorMessage = $state('');
	loadingMore = $state(false);

	#continuation = '';

	get hasMore(): boolean {
		return this.#continuation !== '';
	}

	async load(): Promise<void> {
		this.status = 'loading';
		try {
			this.#append(await home());
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
