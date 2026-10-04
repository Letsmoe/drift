import { searchSuggestions } from '#lib/api/catalog.ts';
import { recentSearches } from '#lib/api/library.ts';

const debounceMs = 200;
const recentLimit = 8;

/** What the search dropdown offers: recent searches for an empty field, suggestions while typing. */
export class SearchSuggestions {
	items = $state<string[]>([]);
	showsRecent = $state(true);
	activeIndex = $state(-1);
	isOpen = $state(false);

	#timer: ReturnType<typeof setTimeout> | undefined = undefined;
	// Responses can arrive out of order; only the latest request may update state.
	#latestRequestId = 0;

	/** Reacts to the field's text: empty text shows recents at once, other text is debounced. */
	update(query: string): void {
		clearTimeout(this.#timer);
		this.activeIndex = -1;
		this.isOpen = true;
		if (query.trim() === '') {
			this.#fetch(query);
			return;
		}
		this.#timer = setTimeout(() => this.#fetch(query), debounceMs);
	}

	close(): void {
		clearTimeout(this.#timer);
		this.#latestRequestId += 1;
		this.isOpen = false;
		this.activeIndex = -1;
	}

	/** Moves the highlighted row by one, wrapping through "no row" (the typed text). */
	move(step: number): void {
		const positionCount = this.items.length + 1;
		const position = this.activeIndex + 1;
		this.activeIndex = ((position + step + positionCount) % positionCount) - 1;
	}

	activeText(): string {
		if (this.activeIndex < 0) {
			return '';
		}
		return this.items[this.activeIndex];
	}

	async #fetch(query: string): Promise<void> {
		this.#latestRequestId += 1;
		const requestId = this.#latestRequestId;
		const trimmedQuery = query.trim();
		try {
			const items = await this.#lookUp(trimmedQuery);
			if (requestId === this.#latestRequestId) {
				this.showsRecent = trimmedQuery === '';
				this.items = items;
			}
		} catch {
			// Suggestions are a convenience; the field keeps working without them.
			this.items = [];
		}
	}

	#lookUp(trimmedQuery: string): Promise<string[]> {
		if (trimmedQuery === '') {
			return recentSearches(recentLimit);
		}
		return searchSuggestions(trimmedQuery);
	}
}
