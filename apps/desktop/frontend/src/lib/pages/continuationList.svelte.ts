export interface ContinuationPage<Entry> {
	items: Entry[];
	continuation: string;
}

/** A list whose later pages are fetched with a continuation token as the user scrolls. */
export class ContinuationList<Entry> {
	items = $state<Entry[]>([]);
	loadingMore = $state(false);

	#continuation = '';
	#fetchPage: (token: string) => Promise<ContinuationPage<Entry>>;
	#generation = 0;

	constructor(fetchPage: (token: string) => Promise<ContinuationPage<Entry>>) {
		this.#fetchPage = fetchPage;
	}

	get hasMore(): boolean {
		return this.#continuation !== '';
	}

	/** Starts over with the first page; a page still in flight for the old list is dropped. */
	reset(items: Entry[], continuation: string): void {
		this.#generation += 1;
		this.items = items;
		this.#continuation = continuation;
		this.loadingMore = false;
	}

	async loadMore(): Promise<void> {
		if (!this.hasMore || this.loadingMore) {
			return;
		}
		const generation = this.#generation;
		this.loadingMore = true;
		try {
			const page = await this.#fetchPage(this.#continuation);
			if (generation === this.#generation) {
				this.items = [...this.items, ...page.items];
				this.#continuation = page.continuation;
			}
		} catch {
			// The first page is already on screen; a failed extra page just ends the list.
			this.#continuation = '';
		} finally {
			if (generation === this.#generation) {
				this.loadingMore = false;
			}
		}
	}
}
