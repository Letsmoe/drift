import { afterNavigate } from '$app/navigation';

// SvelteKit numbers its history entries under this key of history.state.
const historyIndexKey = 'sveltekit:history';

function currentHistoryIndex(): number {
	const state = window.history.state as Record<string, unknown> | null;
	const index = state?.[historyIndexKey];
	if (typeof index === 'number') {
		return index;
	}
	return 0;
}

/**
 * Back/forward availability of the app's own history. Create it during component
 * initialization: it registers an afterNavigate callback.
 */
export class NavigationHistory {
	#firstIndex: number | undefined = undefined;
	#currentIndex = $state(0);
	#lastIndex = $state(0);

	canGoBack = $derived(this.#firstIndex !== undefined && this.#currentIndex > this.#firstIndex);
	canGoForward = $derived(this.#currentIndex < this.#lastIndex);

	constructor() {
		afterNavigate((navigation) => this.#record(navigation.type));
	}

	back(): void {
		if (this.canGoBack) {
			window.history.back();
		}
	}

	forward(): void {
		if (this.canGoForward) {
			window.history.forward();
		}
	}

	#record(navigationType: string): void {
		const index = currentHistoryIndex();
		if (this.#firstIndex === undefined) {
			this.#firstIndex = index;
		}
		this.#currentIndex = index;
		// A new page drops the entries ahead of it; going back and forth keeps them.
		if (navigationType === 'popstate') {
			this.#lastIndex = Math.max(this.#lastIndex, index);
			return;
		}
		this.#lastIndex = index;
	}
}
